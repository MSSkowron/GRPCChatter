package repository

import (
	"github.com/MSSkowron/GRPCChatter/internal/database"
	"github.com/MSSkowron/GRPCChatter/internal/model"
)

type UserRepository interface {
	AddUser(user *model.User) (*model.User, error)
	DeleteUser(userID int) error
	GetUserByID(userID int) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetAllUsers() ([]*model.User, error)
	UpdateUser(user *model.User) error
}

type UserRepositoryImpl struct {
	db database.Database
}

func NewUserRepository(db database.Database) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db: db,
	}
}

func (ur *UserRepositoryImpl) AddUser(user *model.User) (*model.User, error) {
	return nil, nil
}

func (ur *UserRepositoryImpl) DeleteUser(userID int) error {
	return nil
}

func (ur *UserRepositoryImpl) GetUserByID(userID int) (*model.User, error) {
	return nil, nil
}

func (ur *UserRepositoryImpl) GetUserByUsername(username string) (*model.User, error) {
	return nil, nil
}

func (ur *UserRepositoryImpl) GetAllUsers() ([]*model.User, error) {
	return nil, nil
}

func (ur *UserRepositoryImpl) UpdateUser(user *model.User) error {
	return nil
}
