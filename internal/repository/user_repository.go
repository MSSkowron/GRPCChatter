package repository

import "github.com/MSSkowron/GRPCChatter/internal/model"

type UserRepository interface {
	AddUser(user *model.User) error
	DeleteUser(userID int) error
	GetUserByID(userID int) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetAllUsers() ([]*model.User, error)
	UpdateUser(user *model.User) error
}
