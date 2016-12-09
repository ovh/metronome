package userSrv

import (
	"github.com/runabove/metronome/src/metronome/models"
	"github.com/runabove/metronome/src/metronome/pg"
)

func All(userId string) *models.Tasks {

	var tasks models.Tasks
	db := pg.DB()

	err := db.Model(&tasks).Where("user_id = ?", userId).Select()
	if err != nil {
		panic(err)
	}

	if len(tasks) == 0 {
		return nil
	}

	return &tasks
}
