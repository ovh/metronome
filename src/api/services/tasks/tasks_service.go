// Package taskSrv handle tasks database operations.
package taskSrv

import (
	log "github.com/Sirupsen/logrus"

	amodels "github.com/runabove/metronome/src/api/models"
	"github.com/runabove/metronome/src/metronome/models"
	"github.com/runabove/metronome/src/metronome/pg"
	"github.com/runabove/metronome/src/metronome/redis"
)

// All retrive all the tasks of a user.
// Return nil if no task.
func All(userID string) *amodels.TasksAns {

	var tasks models.Tasks
	db := pg.DB()

	err := db.Model(&tasks).Where("user_id = ?", userID).Select()
	if err != nil {
		panic(err)
	}

	if len(tasks) == 0 {
		return nil
	}

	states := redis.DB().HGetAll(userID)
	if states.Err() != nil {
		log.Error(states.Err()) // TODO log
	}

	var ans amodels.TasksAns
	for _, t := range tasks {
		var s models.State
		s.FromJSON(states.Val()[t.GUID])
		ans = append(ans, amodels.TaskAns{
			t,
			s.At,
			s.State,
		})
	}

	return &ans
}
