package models

import (
	"time"
)

// User holds user attributes.
type User struct {
	ID        string    `json:"id" sql:"user_id,pk"`
	Name      string    `json:"name"`
	Password  string    `json:"password,omitempty"`
	Roles     []string  `json:"roles,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Users defined an array of user.
type Users []User
