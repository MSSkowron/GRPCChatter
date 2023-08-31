package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/database"
	"github.com/MSSkowron/GRPCChatter/internal/model"
)

type UserRepository interface {
	AddUser(ctx context.Context, user *model.User) (userID int, err error)
	DeleteUser(ctx context.Context, userID int) (err error)
	GetUserByID(ctx context.Context, userID int) (user *model.User, err error)
	GetUserByUsername(ctx context.Context, username string) (user *model.User, err error)
	GetAllUsers(ctx context.Context) (users []*model.User, err error)
	UpdateUser(ctx context.Context, user *model.User) (err error)
}

type UserRepositoryImpl struct {
	db database.Database
}

func NewUserRepository(db database.Database) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db: db,
	}
}

func (ur *UserRepositoryImpl) AddUser(ctx context.Context, user *model.User) (int, error) {
	query := "INSERT INTO users (created_at, user_name, password) VALUES ($1, $2, $3) RETURNING id"

	row, err := ur.db.QueryRowContext(ctx, query, user.Username, user.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to add user: %w", err)
	}

	var userID int
	if err := row.Scan(&userID); err != nil {
		return 0, fmt.Errorf("failed to add user: %w", err)
	}

	return userID, nil
}

func (ur *UserRepositoryImpl) DeleteUser(ctx context.Context, userID int) error {
	query := "DELETE FROM users WHERE id = $1"

	if _, err := ur.db.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (ur *UserRepositoryImpl) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	query := "SELECT id, created_at, user_name, password FROM users WHERE id = $1"

	row, err := ur.db.QueryRowContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	var user model.User
	if err = row.Scan(&user.ID, &user.CreatedAt, &user.Username, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (ur *UserRepositoryImpl) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := "SELECT id, created_at, user_name, password FROM users WHERE user_name  = $1"

	row, err := ur.db.QueryRowContext(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	var user model.User
	if err = row.Scan(&user.ID, &user.CreatedAt, &user.Username, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (ur *UserRepositoryImpl) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	query := "SELECT id, created_at, user_name, password FROM users"

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

func (ur *UserRepositoryImpl) UpdateUser(ctx context.Context, user *model.User) error {
	query := "UPDATE users SET user_name = $1, password = $2 WHERE id = $3"

	_, err := ur.db.ExecContext(ctx, query, user.Username, user.Password, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
