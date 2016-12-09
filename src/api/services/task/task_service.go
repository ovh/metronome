package taskSrv

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/runabove/metronome/src/api/core"
	"github.com/runabove/metronome/src/metronome/models"
)

func Create(task *models.Task) bool {
	task.CreatedAt = time.Now()

	if len(task.Id) == 0 {
		task.Id = core.Sha256(task.UserId + task.Name + string(task.CreatedAt.Unix()))
	}

	k := core.Kafka()

	_, _, err := k.Producer.SendMessage(task.ToKafka())
	if err != nil {
		log.Error("FAILED to send message: %s\n", err)
		return false
	}
	return true
}

func Delete(id string, userId string) bool {
	k := core.Kafka()

	t := &models.Task{
		Id:     id,
		UserId: userId,
	}

	_, _, err := k.Producer.SendMessage(t.ToKafka())
	if err != nil {
		log.Error("FAILED to send message: %s\n", err)
		return false
	}
	return true
}
