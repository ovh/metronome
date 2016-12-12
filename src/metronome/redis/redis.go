package redis

import (
	"sync"

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
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})

		d = &db{
			DB: redis,
		}
	})
	return d.DB
}
