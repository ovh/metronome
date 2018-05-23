package core

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ovh/metronome/src/metronome/models"
)

// Entry is a task with execution time management.
type Entry struct {
	// The task to run
	task models.Task

	timeMode bool
	start    time.Time
	period   float64
	epsilon  float64
	repeat   int64

	months int64
	years  int64

	next    int64
	planned int64

	initialized bool
}

// NewEntry return a new entry.
func NewEntry(task models.Task) (*Entry, error) {
	segs := strings.Split(task.Schedule, "/")
	if len(segs) != 4 {
		return nil, fmt.Errorf("Bad schedule %s", task.Schedule)
	}

	start, err := time.Parse(time.RFC3339, segs[1])
	if err != nil {
		return nil, err
	}

	matches := durationRegex.FindStringSubmatch(segs[2])

	r, rS := int64(-1), segs[0][1:]
	if len(rS) > 0 {
		parsed, err := strconv.Atoi(rS)
		if err != nil {
			return nil, fmt.Errorf("Bad repeat %s", task.Schedule)
		}
		r = int64(parsed)
	}

	e := &Entry{
		task:     task,
		epsilon:  ParseDuration(strings.Replace(segs[3], "E", "P", 1)).Seconds(),
		start:    start,
		repeat:   r,
		timeMode: strings.Contains(segs[2], "T"),
		period:   ParseDuration(segs[2]).Seconds(),
		next:     -1,
		years:    ParseInt64(matches[1]),
		months:   ParseInt64(matches[2]),
	}

	if e.period == 0 {
		return nil, fmt.Errorf("Null period %v", task.Schedule)
	}

	return e, nil
}

// SameAs check if entry is semanticaly the same as a task.
func (e *Entry) SameAs(t models.Task) bool {
	return e.task.URN == t.URN &&
		e.task.Schedule == t.Schedule
}

// UserID return the task user ID.
func (e *Entry) UserID() string {
	return e.task.UserID
}

// Epsilon return the task epsilon.
func (e *Entry) Epsilon() int64 {
	return int64(e.epsilon)
}

// URN return the task urn.
func (e *Entry) URN() string {
	return e.task.URN
}

// GUID return the task GUID.
func (e *Entry) GUID() string {
	return e.task.GUID
}

// GetPayload return the Task payload
func (e *Entry) GetPayload() map[string]interface{} {
	return e.task.Payload
}

// SetPayload upodate Task payload
func (e *Entry) SetPayload(payload map[string]interface{}) {
	e.task.Payload = payload
}

// Next return the next execution time.
// Return -1 if invalid.
func (e *Entry) Next() int64 {
	return e.next
}

// Init the planning system
// Must be called before Plan
func (e *Entry) Init(now time.Time) {
	e.initialized = true

	if e.timeMode {
		e.next = e.initTimeMode(now)
		return // e.next, true
	}
	e.next = e.initDateMode(now)
}

// initTimeMode compute first iteration for time period
func (e *Entry) initTimeMode(now time.Time) int64 {
	start := e.start.Unix()

	if start >= now.Unix() {
		e.planned = 1
		return start
	}

	n := int64(math.Ceil(float64(now.Unix()-start) / e.period))

	if e.repeat >= 0 && n > e.repeat {
		return -1
	}

	next := start + int64(e.period)*int64(n)

	e.planned = (n + 1)
	return next
}

// initDateMode compute first iteration for date period
func (e *Entry) initDateMode(now time.Time) int64 {
	if e.start.Unix() >= now.Unix() {
		e.planned = 1
		return e.start.Unix()
	}

	period := int(e.months + e.years*12)

	dY := now.Year() - e.start.Year()
	dM := int(now.Month() - e.start.Month())
	dD := int(now.Day() - e.start.Day())
	dh := int(now.Hour() - e.start.Hour())
	dm := int(now.Minute() - e.start.Minute())
	ds := int(now.Second() - e.start.Second())

	n := dY*12 + dM/period
	dt := (((dD*24+dh)*60)+dm)*60 + ds

	if dt > 0 {
		n++
	}
	if e.repeat >= 0 && int64(n) > e.repeat {
		return -1
	}

	next := e.start.AddDate(0, n*period, 0)

	dY = next.Year() - e.start.Year()
	dM = int(next.Month() - e.start.Month())

	// overshoot (due to month rollover)
	if dY*12+dM > n*period {
		next = next.AddDate(0, 0, -next.Day())
	}

	e.planned = int64(n + 1)
	return next.Unix()
}

// Plan the next execution time.
// Return true if planning as been updated.
func (e *Entry) Plan(now time.Time) (bool, error) {
	if !e.initialized {
		return false, errors.New("Unitialized entry. Please call init before")
	}

	if e.next >= now.Unix() {
		return false, nil
	}

	if e.next < 0 {
		return false, nil
	}

	if e.repeat >= 0 && e.planned > e.repeat {
		e.next = -1
		return false, nil
	}

	e.planned++
	if e.timeMode {
		e.next += int64(e.period)
	} else {
		// date mode
		period := int(e.months + e.years*12)
		next := e.start.AddDate(0, period*int(e.planned-1), 0)

		dY := next.Year() - e.start.Year()
		dM := int(next.Month() - e.start.Month())

		// overshoot (due to month rollover)
		if dY*12+dM > int(e.planned-1)*period {
			next = next.AddDate(0, 0, -next.Day())
		}

		e.next = next.Unix()
	}

	return true, nil
}
