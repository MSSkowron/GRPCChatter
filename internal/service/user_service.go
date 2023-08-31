package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MSSkowron/GRPCChatter/internal/dto"
	"github.com/MSSkowron/GRPCChatter/internal/model"
	"github.com/MSSkowron/GRPCChatter/internal/repository"
	"github.com/MSSkowron/GRPCChatter/pkg/crypto"
)

var (
	// ErrInvalidUsername is returned when an invalid username is provided.
	ErrInvalidUsername = errors.New("user name must must not be empty and have at least 6 characters, including digits")
	// ErrInvalidPassword is returned when an invalid password is provided.
	ErrInvalidPassword = errors.New("password must not be empty and must have at least 6 characters, including 1 uppercase letter, 1 lowercase letter, 1 digit and 1 special character")
	// ErrUserAlreadyExists is returned when a user with the same username already exists.
	ErrUserAlreadyExists = errors.New("user with the provided user name already exists")
	// ErrInvalidCredentials is returned when invalid user credentials are provided.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserService defines the interface for user-related operations.
type UserService interface {
	// RegisterUser registers a new user.
	RegisterUser(context.Context, *dto.UserRegisterDTO) (*dto.UserDTO, error)

	// LoginUser performs user authentication.
	LoginUser(context.Context, *dto.UserLoginDTO) (*dto.TokenDTO, error)
}

// UserServiceImpl is the concrete implementation of the UserService interface.
type UserServiceImpl struct {
	tokenService   UserTokenService
	userRepository repository.UserRepository
}

// NewUserService creates a new instance of UserServiceImpl.
func NewUserService(tokenService UserTokenService, userRepository repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{
		tokenService:   tokenService,
		userRepository: userRepository,
	}
}

// RegisterUser registers a new user.
func (us *UserServiceImpl) RegisterUser(ctx context.Context, userRegister *dto.UserRegisterDTO) (*dto.UserDTO, error) {
	if !us.validateUsername(userRegister.Username) {
		return nil, ErrInvalidUsername
	}
	if !us.validatePassword(userRegister.Password) {
		return nil, ErrInvalidPassword
	}

	user, err := us.userRepository.GetUserByUsername(ctx, userRegister.Username)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := crypto.HashPassword(userRegister.Password)
	if err != nil {
		return nil, err
	}

	currTime := time.Now()
	id, err := us.userRepository.AddUser(ctx, &model.User{
		CreatedAt: currTime,
		Username:  userRegister.Username,
		Password:  hashedPassword,
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserDTO{
		ID:        int64(id),
		CreatedAt: currTime,
		Username:  userRegister.Username,
	}, nil
}

// LoginUser performs user authentication.
func (us *UserServiceImpl) LoginUser(ctx context.Context, userLogin *dto.UserLoginDTO) (*dto.TokenDTO, error) {
	user, _ := us.userRepository.GetUserByUsername(ctx, userLogin.Username)
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := crypto.CheckPassword(userLogin.Password, user.Password); err != nil {
		if errors.Is(err, crypto.ErrInvalidCredentials) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	token, err := us.tokenService.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &dto.TokenDTO{
		Token: token,
	}, nil
}

func (us *UserServiceImpl) validateUsername(username string) bool {
	return len(username) >= 6 &&
		strings.ContainsAny(username, "0123456789") &&
		!strings.ContainsAny(username, "!@#$%^&*()_+[]{};':,.<>?/")
}

func (us *UserServiceImpl) validatePassword(password string) bool {
	return len(password) >= 6 &&
		strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
		strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") &&
		strings.ContainsAny(password, "0123456789") &&
		strings.ContainsAny(password, "!@#$%^&*()_+[]{};':,.<>?/")
}
