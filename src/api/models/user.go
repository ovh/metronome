package models

import (
	"time"
)

type User struct {
	Id        string    `json:"id"sql:"user_id,pk"`
	Name      string    `json:"name"`
	Password  string    `json:"password,omitempty"`
	Roles     []string  `json:"roles,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
type Users []User
