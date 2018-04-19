// Package tasksrv handle task Kafka messages.
package tasksrv

import (
	"time"

	log "github.com/sirupsen/logrus"

	acore "github.com/ovh/metronome/src/api/core"
	"github.com/ovh/metronome/src/metronome/core"
	"github.com/ovh/metronome/src/metronome/models"
)

// Create a new task.
// Return true if success.
func Create(task *models.Task) bool {
	task.CreatedAt = time.Now()

	if len(task.ID) == 0 {
		task.ID = core.Sha256(task.UserID + task.Name + string(task.CreatedAt.Unix()))
	}

	k := acore.GetKafka()

	_, _, err := k.Producer.SendMessage(task.ToKafka())
	if err != nil {
		log.Errorf("FAILED to send message: %s\n", err)
		return false
	}
	return true
}

// Delete a task.
// Return true if success.
func Delete(id string, userID string) bool {
	k := acore.GetKafka()

	t := &models.Task{
		ID:     id,
		UserID: userID,
	}

	_, _, err := k.Producer.SendMessage(t.ToKafka())
	if err != nil {
		log.Errorf("FAILED to send message: %s\n", err)
		return false
	}
	return true
}
