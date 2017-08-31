package models

import (
	"github.com/ovh/metronome/src/metronome/models"
)

// TaskAns hold Task attributes and state fields.
type TaskAns struct {
	models.Task
	RunAt   int64 `json:"runAt"`
	RunCode int64 `json:"runCode"`
}

// TasksAns is an array of TaskAns.
type TasksAns []TaskAns
