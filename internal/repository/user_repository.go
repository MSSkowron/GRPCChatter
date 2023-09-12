package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/database"
	"github.com/MSSkowron/GRPCChatter/internal/model"
)

// UserRepository is an interface that defines the methods required for user data management.
type UserRepository interface {
	// AddUser adds a new user to the database.
	AddUser(ctx context.Context, user *model.User) (addedUser *model.User, err error)

	// DeleteUser deletes a user from the database by their userID.
	DeleteUser(ctx context.Context, userID int) (err error)

	// GetUserByID retrieves a user from the database by their userID.
	GetUserByID(ctx context.Context, userID int) (user *model.User, err error)

	// GetUserByUsername retrieves a user from the database by their username.
	GetUserByUsername(ctx context.Context, username string) (user *model.User, err error)

	// GetAllUsers retrieves all users from the database.
	GetAllUsers(ctx context.Context) (users []*model.User, err error)
}

// UserRepositoryImpl implements the UserRepository interface.
type UserRepositoryImpl struct {
	db database.Database
}

// NewUserRepository creates a new UserRepositoryImpl instance with the provided database.
func NewUserRepository(db database.Database) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db: db,
	}
}

func (ur *UserRepositoryImpl) AddUser(ctx context.Context, user *model.User) (*model.User, error) {
	query := "INSERT INTO users (created_at, username, password) VALUES ($1, $2, $3) RETURNING *"

	row, err := ur.db.QueryRowContext(ctx, query, user.CreatedAt, user.Username, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to add user: %w", err)
	}

	var roleID int
	if err = row.Scan(&user.ID, &user.CreatedAt, &user.Username, &user.Password, &roleID); err != nil {
		return nil, fmt.Errorf("failed to add user: %w", err)
	}

	query = "SELECT name FROM roles WHERE id = $1"
	row, err = ur.db.QueryRowContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role ID: %w", err)
	}

	if err = row.Scan(&user.Role); err != nil {
		return nil, fmt.Errorf("failed to get role ID: %w", err)
	}

	return user, nil
}

func (ur *UserRepositoryImpl) DeleteUser(ctx context.Context, userID int) error {
	query := "DELETE FROM users WHERE id = $1"

	if _, err := ur.db.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (ur *UserRepositoryImpl) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	query := `
		SELECT u.id, u.created_at, u.username, u.password, r.name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`

	row, err := ur.db.QueryRowContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	var user model.User
	if err = row.Scan(&user.ID, &user.CreatedAt, &user.Username, &user.Password, &user.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (ur *UserRepositoryImpl) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT u.id, u.created_at, u.username, u.password, r.name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE username  = $1
	`

	row, err := ur.db.QueryRowContext(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	var user model.User
	if err = row.Scan(&user.ID, &user.CreatedAt, &user.Username, &user.Password, &user.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (ur *UserRepositoryImpl) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	query := `
		SELECT u.id, u.created_at, u.username, u.password, r.name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
	`

	rows, err := ur.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	users := make([]*model.User, 0)
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.CreatedAt, &user.Username, &user.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error in result set: %w", err)
	}

	return users, nil
}
