package redis

import (
	"sync"

	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285
	"gopkg.in/redis.v5"
)

type db struct {
	DB *redis.Client
}

var d *db
var onceDB sync.Once

// DB get a database instance
func DB() *redis.Client {
	onceDB.Do(func() {
		redis := redis.NewClient(&redis.Options{
			Addr:     viper.GetString("redis.addr"),
			Password: "",
			DB:       0,
		})

		d = &db{
			DB: redis,
		}
	})
	return d.DB
}
