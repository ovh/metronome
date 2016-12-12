package routines

import (
	"container/ring"
	"fmt"
	"math"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/runabove/metronome/src/metronome/models"
	"github.com/runabove/metronome/src/scheduler/core"
)

// batch represent a batch of job to send
type batch struct {
	at   int64
	jobs map[string][]models.Job
}

// TaskScheduler handle the internal states of the scheduler
type TaskScheduler struct {
	entries   map[string]*core.Entry
	nextExec  *ring.Ring
	plan      *ring.Ring
	now       int64
	jobs      chan []models.Job
	stop      chan int
	planning  chan int
	dispatch  chan int
	nextTimer *time.Timer
}

// NewTaskScheduler return a new task scheduler
func NewTaskScheduler(tasks <-chan models.Task) *TaskScheduler {
	const buffSize = 20

	ts := &TaskScheduler{
		plan:     ring.New(buffSize),
		entries:  make(map[string]*core.Entry),
		now:      int64(time.Now().Unix()),
		jobs:     make(chan []models.Job, buffSize),
		stop:     make(chan int),
		planning: make(chan int, 1),
		dispatch: make(chan int, 1),
	}
	ts.plan.Value = batch{
		ts.now,
		make(map[string][]models.Job),
	}
	ts.nextExec = ts.plan

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
			case t := <-tasks:
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
	ts.planning <- 1
	ts.dispatch <- 1
}

// Stop the scheduler
func (ts *TaskScheduler) Stop() {
	close(ts.dispatch)
	close(ts.planning)
	ts.nextTimer.Stop()
	ts.stop <- 1
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
	for i := 0; i < c.Len(); i++ {
		if c.Value != nil {
			jobs := c.Value.(batch).jobs[t.GUID]
			at := c.Value.(batch).at

			if e.Next() < 0 {
				e.Plan(at, !taskUpdate)
			}
			if e.Next() > 0 && e.Next() <= at {
				p, ok := e.Next(), true
				for ok && p <= at {
					jobs = append(jobs, models.Job{t.GUID, p, e.Epsilon(), e.URN()})
					p, ok = e.Plan(at, !taskUpdate)
				}
			}
			c.Value.(batch).jobs[t.GUID] = jobs
		}

		c = c.Next()
	}
}

// Dispatch jobs executions
func (ts *TaskScheduler) handleDispatch() {
	now := time.Now().Unix()
	for ts.nextExec.Value != nil && ts.nextExec.Value.(batch).at <= now {
		var jobs []models.Job
		for _, js := range ts.nextExec.Value.(batch).jobs {
			jobs = append(jobs, js...)
		}
		ts.jobs <- jobs
		ts.nextExec.Value = nil
		ts.nextExec = ts.nextExec.Next()
	}

	if ts.nextExec.Value == nil {
		// NOP
		time.AfterFunc(300*time.Millisecond, func() {
			ts.dispatch <- 1
		})
		return
	}

	ts.nextTimer = time.AfterFunc(time.Duration(ts.nextExec.Value.(batch).at-now)*time.Second, func() {
		ts.dispatch <- 1
	})

	// Trigger planning
	select {
	case ts.planning <- 1:
	default:
	}
}

// Plan next executions
func (ts *TaskScheduler) handlePlanning() {
	if ts.plan.Next().Value != nil {
		return
	}
	ts.plan = ts.plan.Next()

	ts.now = int64(math.Max(float64(ts.now+1), float64(time.Now().Unix())))
	ts.plan.Value = batch{
		ts.now,
		make(map[string][]models.Job),
	}

	for k := range ts.entries {
		e := ts.entries[k]

		jobs := ts.plan.Value.(batch).jobs[k]

		if e.Next() > 0 && e.Next() <= ts.now {
			p, ok := e.Next(), true
			for ok && p <= ts.now {
				jobs = append(jobs, models.Job{k, p, e.Epsilon(), e.URN()})
				fmt.Printf("*")
				p, ok = e.Plan(ts.now, true)
			}
		}
		ts.plan.Value.(batch).jobs[k] = jobs
	}

	fmt.Printf(".")

	next := ts.plan.Next()
	// Plan next batch if available
	if next.Value == nil {
		select {
		case ts.planning <- 1:
		default:
		}
	}
}
