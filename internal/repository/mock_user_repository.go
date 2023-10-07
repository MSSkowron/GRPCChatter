package repository

import (
	"context"

	"github.com/MSSkowron/GRPCChatter/internal/model"
)

// MockUserRepository is a mock implementation of UserRepository for testing purposes.
type MockUserRepository struct {
	Users          map[int]*model.User // Map to store users by ID
	LastInsertedID int                 // To simulate auto-increment behavior
}

// NewMockUserRepository creates a new instance of MockUserRepository.
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[int]*model.User),
	}
}

// AddUser is a mock implementation of AddUser method.
func (m *MockUserRepository) AddUser(ctx context.Context, user *model.User) (*model.User, error) {
	m.LastInsertedID++
	user.ID = m.LastInsertedID
	m.Users[user.ID] = user
	return user, nil
}

// DeleteUser is a mock implementation of DeleteUser method.
func (m *MockUserRepository) DeleteUser(ctx context.Context, userID int) error {
	delete(m.Users, userID)
	return nil
}

// GetUserByID is a mock implementation of GetUserByID method.
func (m *MockUserRepository) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	user, ok := m.Users[userID]
	if !ok {
		return nil, nil
	}
	return user, nil
}

// GetUserByUsername is a mock implementation of GetUserByUsername method.
func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	for _, user := range m.Users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, nil
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	users := make([]*model.User, 0, len(m.Users))
	for _, user := range m.Users {
		users = append(users, user)
	}

	return users, nil
}
