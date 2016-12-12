package core

import (
	"math"
	"strings"
	"time"

	"github.com/runabove/metronome/src/metronome/models"
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

	next int64
}

// NewEntry return a new entry.
func NewEntry(task models.Task) (*Entry, error) {
	segs := strings.Split(string(task.Schedule), "/")

	start, err := time.Parse(time.RFC3339, segs[1])
	if err != nil {
		return nil, err
	}

	matches := durationRegex.FindStringSubmatch(segs[2])

	return &Entry{
		task:     task,
		epsilon:  ParseDuration(strings.Replace(segs[3], "E", "P", 1)).Seconds(),
		start:    start,
		repeat:   99999999, // FIXME
		timeMode: strings.Contains(segs[2], "T"),
		period:   ParseDuration(segs[2]).Seconds(),
		next:     -1,
		years:    ParseInt64(matches[1]),
		months:   ParseInt64(matches[2]),
	}, nil
}

// SameAs check if entry is semanticaly the same as a task.
func (e *Entry) SameAs(t models.Task) bool {
	return e.task.URN == t.URN &&
		e.task.Schedule == t.Schedule
}

// Epsilon return the task epsilon.
func (e *Entry) Epsilon() int64 {
	return int64(e.epsilon)
}

// URN return the task urn.
func (e *Entry) URN() string {
	return e.task.URN
}

// Next return the next execution time.
// Return -1 if invalid.
func (e *Entry) Next() int64 {
	return e.next
}

// Plan the next execution time.
// Return -1 if invalid.
func (e *Entry) Plan(now int64, past bool) (int64, bool) {
	if e.next > now {
		return e.next, false
	}

	if past {
		now = now - int64(e.epsilon)
	}

	if e.timeMode {
		if e.period == 0 {
			return -1, false
		}

		start := e.start.Unix()
		n := int64(0)
		if start < now {
			n = int64(math.Ceil(float64(now-start) / e.period))

			if n > e.repeat {
				e.next = -1
				return -1, false
			}
		}

		next := start + int64(e.period)*int64(n)
		for next <= e.next {
			next += int64(e.period)
		}

		e.next = next
		return e.next, true
	}
	// date mode
	if e.months == 0 && e.years == 0 {
		return -1, false
	}

	n := int(0)
	nowT := time.Unix(now, 0)

	if e.start.Unix() < nowT.Unix() {
		dy := nowT.Year() - e.start.Year()
		dm := int(nowT.Month() - e.start.Month())
		dd := int(nowT.Day() - e.start.Day())

		if dd < 0 {
			dm--
		}

		n = dy*12 + dm + 1
		if int64(n) > e.repeat {
			return -1, false
		}
	}

	next := e.start.AddDate(0, n, 0)

	if e.start.Day() != next.Day() {
		next = next.AddDate(0, 0, -next.Day())
	}

	e.next = next.Unix()
	return e.next, true
}
