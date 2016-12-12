// Package taskSrv handle tasks database operations.
package taskSrv

import (
	"github.com/runabove/metronome/src/metronome/models"
	"github.com/runabove/metronome/src/metronome/pg"
)

// All retrive all the tasks of a user.
// Return nil if no task.
func All(userID string) *models.Tasks {

	var tasks models.Tasks
	db := pg.DB()

	err := db.Model(&tasks).Where("user_id = ?", userID).Select()
	if err != nil {
		panic(err)
	}

	if len(tasks) == 0 {
		return nil
	}

	return &tasks
}
