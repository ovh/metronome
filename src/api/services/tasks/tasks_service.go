// Package taskssrv handle tasks database operations.
package taskssrv

import (
	amodels "github.com/ovh/metronome/src/api/models"
	"github.com/ovh/metronome/src/metronome/models"
	"github.com/ovh/metronome/src/metronome/pg"
	"github.com/ovh/metronome/src/metronome/redis"
	log "github.com/sirupsen/logrus"
)

// All retrieve all the tasks of a user.
// Return nil if no task.
func All(userID string) (*amodels.TasksAns, error) {

	var tasks models.Tasks
	db := pg.DB()

	err := db.Model(&tasks).Where("user_id = ?", userID).Select()
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	states := redis.DB().HGetAll(userID)
	if states.Err() != nil {
		return nil, states.Err()
	}

	var ans amodels.TasksAns
	for _, t := range tasks {
		var s models.State
		state, ok := states.Val()[t.GUID]
		if !ok {
			log.Warnf("No such entry in map states for key '%s'", t.GUID)
			ans = append(ans, amodels.TaskAns{
				Task: t,
			})
			continue
		}

		if err = s.FromJSON([]byte(state)); err != nil {
			return nil, err
		}

		ans = append(ans, amodels.TaskAns{
			Task:    t,
			RunAt:   s.At,
			RunCode: s.State,
		})
	}

	return &ans, err
}
