package dto

import "time"

// UserDTO represents a data transfer object (DTO) for a user.
type UserDTO struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"user_name"`
	Password  string    `json:"password"`
}

// UserRegisterDTO represents a data transfer object (DTO) for creating a user account request.
type UserRegisterDTO struct {
	Username string `json:"user_name"`
	Password string `json:"password"`
}

// UserLoginDTO represents a data transfer object (DTO) for user login request.
type UserLoginDTO struct {
	Username string `json:"user_name"`
	Password string `json:"password"`
}
