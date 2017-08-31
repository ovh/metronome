package routines

import (
	"container/ring"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	redisV5 "gopkg.in/redis.v5"

	"github.com/ovh/metronome/src/metronome/models"
	"github.com/ovh/metronome/src/metronome/redis"
	"github.com/ovh/metronome/src/scheduler/core"
)

// batch represent a batch of job to send
type batch struct {
	at   time.Time
	jobs map[string][]models.Job
}

type state struct {
	At      int64           `json:"at"`
	Indexes map[int32]int64 `json:"indexes"`
}

// TaskScheduler handle the internal states of the scheduler
type TaskScheduler struct {
	entries      map[string]*core.Entry
	nextExec     *ring.Ring
	plan         *ring.Ring
	now          time.Time
	jobs         chan []models.Job
	halt         chan struct{}
	planning     chan struct{}
	dispatch     chan struct{}
	nextTimer    *time.Timer
	jobProducer  *JobProducer
	partition    int32
	alive        sync.WaitGroup
	entriesMutex sync.Mutex
	// metrics
	taskGauge   prometheus.Gauge
	planCounter prometheus.Counter
}

// NewTaskScheduler return a new task scheduler
func NewTaskScheduler(partition int32, tasks <-chan models.Task) *TaskScheduler {
	const buffSize = 20

	ts := &TaskScheduler{
		plan:      ring.New(buffSize),
		entries:   make(map[string]*core.Entry),
		now:       time.Now().UTC(),
		jobs:      make(chan []models.Job, buffSize),
		halt:      make(chan struct{}),
		planning:  make(chan struct{}, 1),
		dispatch:  make(chan struct{}, 1),
		partition: partition,
	}
	ts.plan.Value = batch{
		ts.now,
		make(map[string][]models.Job),
	}
	ts.nextExec = ts.plan
	ts.alive.Add(1)

	// metrics
	ts.taskGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "metronome",
		Subsystem:   "scheduler",
		Name:        "managed",
		Help:        "Number of tasks managed.",
		ConstLabels: prometheus.Labels{"partition": strconv.Itoa(int(ts.partition))},
	})
	prometheus.MustRegister(ts.taskGauge)
	ts.planCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "metronome",
		Subsystem:   "scheduler",
		Name:        "plan",
		Help:        "Number of tasks plan.",
		ConstLabels: prometheus.Labels{"partition": strconv.Itoa(int(ts.partition))},
	})
	prometheus.MustRegister(ts.planCounter)

	// jobs producer
	ts.jobProducer = NewJobProducer(ts.jobs)

	go func() {
		defer ts.alive.Done()
		for {
			// dispatch first
			select {
			case _, ok := <-ts.dispatch:
				if ok {
					ts.handleDispatch()
					continue
				}
			default:
			}

			// planning then
			select {
			case _, ok := <-ts.planning:
				if ok {
					ts.handlePlanning()
					continue
				}
			default:
			}

			select {
			case _, ok := <-ts.dispatch:
				if ok {
					ts.handleDispatch()
				}
			case _, ok := <-ts.planning:
				if ok {
					ts.handlePlanning()
				}
			case t, ok := <-tasks:
				if !ok {
					// shutdown
					go ts.stop()
					return
				}
				ts.entriesMutex.Lock()
				ts.handleTask(t)
				ts.entriesMutex.Unlock()
			case <-ts.halt:
				return
			}
		}
	}()

	return ts
}

// Start task scheduling
func (ts *TaskScheduler) Start() {
	ts.entriesMutex.Lock()

	defer func() {
		ts.entriesMutex.Unlock()

		ts.planning <- struct{}{}
		ts.dispatch <- struct{}{}
	}()

	val, err := redis.DB().Get(strconv.Itoa(int(ts.partition))).Result()
	if err == redisV5.Nil {
		// no state save
		return
	} else if err != nil {
		log.Error(err)
		return
	}

	var state state
	err = json.Unmarshal([]byte(val), &state)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Scheduler %v restored state %v", ts.partition, state)

	// Re-init from last know scheduler
	for _, e := range ts.entries {
		e.Init(time.Unix(state.At+1, 0))
	}

	// Look if we have already schedule some jobs
	if len(state.Indexes) > 0 {

		jobs := make(chan models.Job)
		NewJobComsumer(state.Indexes, jobs)

	jobsLoader:
		for {
			select {
			case j, ok := <-jobs:
				if !ok {
					break jobsLoader
				}
				if ts.entries[j.GUID] == nil {
					break
				}
				ts.entries[j.GUID].Init(time.Unix(j.At+1, 0))
			}
		}
	}

	// Plan
	for guid, e := range ts.entries {
		jobs := planEntryInBatch(e, ts.nextExec.Value.(batch).at)
		ts.nextExec.Value.(batch).jobs[guid] = jobs
	}
}

// stop the scheduler
func (ts *TaskScheduler) stop() {
	close(ts.dispatch)
	close(ts.planning)
	if ts.nextTimer != nil {
		ts.nextTimer.Stop()
	}
	select {
	case ts.halt <- struct{}{}:
	default:
	}

	ts.jobProducer.Close()
}

// Halted wait for scheduler to be halt
func (ts *TaskScheduler) Halted() {
	ts.alive.Wait()
}

// Jobs return the out jobs channel
func (ts *TaskScheduler) Jobs() <-chan []models.Job {
	return ts.jobs
}

// Handle incomming task
func (ts *TaskScheduler) handleTask(t models.Task) {
	if ts.entries[t.GUID] != nil && t.Schedule == "" {
		log.Infof("DELETE task: %s", t.GUID)
		ts.taskGauge.Dec()
		delete(ts.entries, t.GUID)
		c := ts.nextExec
		for i := 0; i < c.Len(); i++ {
			if c.Value != nil {
				// Clear schedule execution
				c.Value.(batch).jobs[t.GUID] = make([]models.Job, 0)
			}

			c = c.Next()
		}
		return
	}

	taskUpdate := false
	if ts.entries[t.GUID] != nil {
		taskUpdate = true
		if ts.entries[t.GUID].SameAs(t) {
			log.Infof("NOP task: %s", t.GUID)
			return
		}

		// Clear schedule execution on task update
		c := ts.nextExec
		for i := 0; i < c.Len(); i++ {
			if c.Value != nil {
				c.Value.(batch).jobs[t.GUID] = make([]models.Job, 0)
			}
			c = c.Next()
		}
	}

	if !taskUpdate {
		log.Infof("NEW task: %s", t.GUID)
		ts.taskGauge.Inc()
	} else {
		log.Infof("UPDATE task: %s", t.GUID)
	}

	// Update entries
	e, err := core.NewEntry(t)
	if err != nil {
		log.Errorf("unprocessable task(%v)", t)
		return
	}
	ts.entries[t.GUID] = e

	// Plan executions
	c := ts.nextExec
	e.Init(c.Value.(batch).at)
	for i := 0; i < c.Len(); i++ {
		if c.Value != nil {
			jobs := planEntryInBatch(e, c.Value.(batch).at)

			if len(jobs) > 0 {
				c.Value.(batch).jobs[t.GUID] = jobs
			}
		}

		c = c.Next()
	}
}

// Dispatch jobs executions
func (ts *TaskScheduler) handleDispatch() {
	now := time.Now().UTC().Unix()
	at := int64(0)
	send := 0
	for ts.nextExec.Value != nil && ts.nextExec.Value.(batch).at.Unix() <= now {
		at = ts.nextExec.Value.(batch).at.Unix()
		var jobs []models.Job
		for _, js := range ts.nextExec.Value.(batch).jobs {
			send += len(js)
			jobs = append(jobs, js...)
		}
		ts.jobs <- jobs
		ts.nextExec.Value = nil
		ts.nextExec = ts.nextExec.Next()
	}

	if at > 0 {
		ts.planCounter.Add(float64(send))
		log.WithFields(log.Fields{
			"at":       at,
			"partiton": ts.partition,
			"indexes":  ts.jobProducer.Indexes(),
			"do":       send,
		}).Info("Dispatch")
		out, err := json.Marshal(state{at, ts.jobProducer.Indexes()})
		if err != nil {
			log.Error(err)
		} else {
			if err := redis.DB().Set(strconv.Itoa(int(ts.partition)), string(out), 0).Err(); err != nil {
				log.Error(err)
			}
		}
	}

	if ts.nextExec.Value == nil {
		// NOP wait to be trig
		time.AfterFunc(300*time.Millisecond, func() {
			ts.dispatch <- struct{}{}
		})
		return
	}

	nextRun := ts.nextExec.Value.(batch).at.Unix()
	ts.nextTimer = time.AfterFunc(time.Duration(nextRun-now)*time.Second, func() {
		ts.dispatch <- struct{}{}
	})

	// Trigger planning
	select {
	case ts.planning <- struct{}{}:
	default:
	}
}

// Plan next executions
func (ts *TaskScheduler) handlePlanning() {
	if ts.plan.Next().Value != nil {
		return
	}
	ts.plan = ts.plan.Next()

	ts.now = ts.now.Add(1 * time.Second)
	if ts.now.Before(time.Now()) {
		ts.now = time.Now().UTC()
	}

	ts.plan.Value = batch{
		ts.now,
		make(map[string][]models.Job),
	}

	for k := range ts.entries {
		jobs := planEntryInBatch(ts.entries[k], ts.now)

		if len(jobs) > 0 {
			ts.plan.Value.(batch).jobs[k] = jobs
		}
	}

	next := ts.plan.Next()
	// Plan next batch if available
	if next.Value == nil {
		select {
		case ts.planning <- struct{}{}:
		default:
		}
	}
}

func planEntryInBatch(entry *core.Entry, at time.Time) []models.Job {
	jobs := make([]models.Job, 0)
	entry.Plan(at)
	for entry.Next() > 0 && entry.Next() <= at.Unix() {
		jobs = append(jobs, models.Job{GUID: entry.GUID(), UserID: entry.UserID(), At: entry.Next(), Epsilon: entry.Epsilon(), URN: entry.URN()})
		if !entry.Plan(at) {
			break
		}
	}
	return jobs
}
