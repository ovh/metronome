package pg

import (
	"sync"

	"github.com/spf13/viper"
	"gopkg.in/pg.v5"
)

type db struct {
	DB *pg.DB
}

var d *db
var onceDB sync.Once

var (
	// ErrNoRows is throwed when SELECT returns nothing
	ErrNoRows = pg.ErrNoRows
)

// DB get a database instance
func DB() *pg.DB {
	onceDB.Do(func() {
		database := pg.Connect(&pg.Options{
			Addr:     viper.GetString("pg.addr"),
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
