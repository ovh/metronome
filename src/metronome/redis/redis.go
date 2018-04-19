package redis

import (
	"sync"

	"github.com/spf13/viper"
	"gopkg.in/redis.v5"
)

type db struct {
	DB *Client
}

var d *db
var onceDB sync.Once

// Client is a redis client
type Client struct {
	*redis.Client
}

// DB get a database instance
func DB() *Client {
	onceDB.Do(func() {
		redis := redis.NewClient(&redis.Options{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.pass"),
			DB:       0,
		})

		d = &db{
			DB: &Client{redis},
		}
	})
	return d.DB
}

// PublishTopic send a message to a given topic
func (c *Client) PublishTopic(channel, topic, message string) *redis.IntCmd {
	return c.Publish(channel, topic+":"+message)
}
