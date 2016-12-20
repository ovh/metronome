package routines

import (
	"container/ring"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/runabove/metronome/src/metronome/models"
	"github.com/runabove/metronome/src/scheduler/core"
)

// batch represent a batch of job to send
type batch struct {
	at   time.Time
	jobs map[string][]models.Job
}

// TaskScheduler handle the internal states of the scheduler
type TaskScheduler struct {
	entries     map[string]*core.Entry
	nextExec    *ring.Ring
	plan        *ring.Ring
	now         time.Time
	jobs        chan []models.Job
	stop        chan struct{}
	planning    chan struct{}
	dispatch    chan struct{}
	nextTimer   *time.Timer
	jobProducer *JobProducer
}

// NewTaskScheduler return a new task scheduler
func NewTaskScheduler(tasks <-chan models.Task) *TaskScheduler {
	const buffSize = 20

	ts := &TaskScheduler{
		plan:     ring.New(buffSize),
		entries:  make(map[string]*core.Entry),
		now:      time.Now().UTC(),
		jobs:     make(chan []models.Job, buffSize),
		stop:     make(chan struct{}),
		planning: make(chan struct{}, 1),
		dispatch: make(chan struct{}, 1),
	}
	ts.plan.Value = batch{
		ts.now,
		make(map[string][]models.Job),
	}
	ts.nextExec = ts.plan

	// jobs producer
	ts.jobProducer = NewJobProducer(ts.jobs)

	go func() {
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
					ts.Stop()
				}
				ts.handleTask(t)
			case <-ts.stop:
				return
			}
		}
	}()

	return ts
}

// Start task scheduling
func (ts *TaskScheduler) Start() {
	ts.planning <- struct{}{}
	ts.dispatch <- struct{}{}
}

// Stop the scheduler
func (ts *TaskScheduler) Stop() {
	close(ts.dispatch)
	close(ts.planning)
	ts.nextTimer.Stop()
	ts.stop <- struct{}{}

	ts.jobProducer.Close()
}

// Jobs return the out jobs channel
func (ts *TaskScheduler) Jobs() <-chan []models.Job {
	return ts.jobs
}

// Handle incomming task
func (ts *TaskScheduler) handleTask(t models.Task) {
	if t.Schedule == "" {
		log.Infof("DELETE task: %s", t.GUID)
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
	for ts.nextExec.Value != nil && ts.nextExec.Value.(batch).at.Unix() <= now {
		var jobs []models.Job
		for _, js := range ts.nextExec.Value.(batch).jobs {
			jobs = append(jobs, js...)
		}
		ts.jobs <- jobs
		ts.nextExec.Value = nil
		ts.nextExec = ts.nextExec.Next()
	}

	if ts.nextExec.Value == nil {
		// NOP sleep a bit before next exec
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
			fmt.Printf("*")
			ts.plan.Value.(batch).jobs[k] = jobs
		}
	}

	fmt.Printf(".")

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
