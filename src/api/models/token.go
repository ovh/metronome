package models

import "time"

// Token describe token serialization.
type Token struct {
	Token     string    `db:"token"`
	UserID    string    `db:"user_id"`
	Roles     []string  `db:"roles"`
	Type      string    `db:"type"`
	CreatedAt time.Time `db:"created_at"`
}

// Tokens is a slice of Token
type Tokens []Token
