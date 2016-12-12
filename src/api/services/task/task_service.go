// Package taskSrv handle task Kafka messages.
package taskSrv

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/runabove/metronome/src/api/core"
	"github.com/runabove/metronome/src/metronome/models"
)

// Create a new task.
// Return true if success.
func Create(task *models.Task) bool {
	task.CreatedAt = time.Now()

	if len(task.ID) == 0 {
		task.ID = core.Sha256(task.UserID + task.Name + string(task.CreatedAt.Unix()))
	}

	k := core.GetKafka()

	_, _, err := k.Producer.SendMessage(task.ToKafka())
	if err != nil {
		log.Error("FAILED to send message: %s\n", err)
		return false
	}
	return true
}

// Delete a task.
// Return true if success.
func Delete(id string, userID string) bool {
	k := core.GetKafka()

	t := &models.Task{
		ID:     id,
		UserID: userID,
	}

	_, _, err := k.Producer.SendMessage(t.ToKafka())
	if err != nil {
		log.Error("FAILED to send message: %s\n", err)
		return false
	}
	return true
}
