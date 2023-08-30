package model

import "time"

// User represents a model for a user.
type User struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"user_name"`
	Password  string    `json:"password"`
}
