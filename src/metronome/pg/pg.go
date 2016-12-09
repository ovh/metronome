package pg

import (
	"sync"

	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285
	"gopkg.in/pg.v5"
)

type db struct {
	DB *pg.DB
}

var d *db
var onceDB sync.Once

func DB() *pg.DB {
	onceDB.Do(func() {
		database := pg.Connect(&pg.Options{
			User:     viper.GetString("pg.user"),
			Password: viper.GetString("pg.password"),
			Database: viper.GetString("pg.database"),
		})

		d = &db{
			DB: database,
		}
	})
	return d.DB
}
