package service

import (
	"context"
	"errors"
	"time"

	"github.com/MSSkowron/GRPCChatter/internal/dto"
	"github.com/MSSkowron/GRPCChatter/internal/model"
	"github.com/MSSkowron/GRPCChatter/internal/repository"
	"github.com/MSSkowron/GRPCChatter/pkg/crypto"
	"github.com/MSSkowron/GRPCChatter/pkg/validation"
)

const userDefaultRoleID = 1

var (
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

// UserServiceImpl implements the UserService interface.
type UserServiceImpl struct {
	tokenService   UserTokenService
	userRepository repository.UserRepository
	userRoles      map[int]string
}

// NewUserService creates a new UserServiceImpl instance with the provided tokenService and userRepository.
func NewUserService(tokenService UserTokenService, userRepository repository.UserRepository) *UserServiceImpl {
	// TODO: Fetch roles from the database and create a map of roles with their IDs.
	return &UserServiceImpl{
		tokenService:   tokenService,
		userRepository: userRepository,
		userRoles: map[int]string{
			userDefaultRoleID: "USER",
		},
	}
}

func (us *UserServiceImpl) RegisterUser(ctx context.Context, userRegister *dto.UserRegisterDTO) (*dto.UserDTO, error) {
	if err := validation.ValidateUsername(userRegister.Username); err != nil {
		return nil, err
	}
	if err := validation.ValidatePassword(userRegister.Password); err != nil {
		return nil, err
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

	newUser := &model.User{
		CreatedAt: time.Now(),
		Username:  userRegister.Username,
		RoleID:    userDefaultRoleID,
		Password:  hashedPassword,
	}
	newUserID, err := us.userRepository.AddUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &dto.UserDTO{
		ID:        int64(newUserID),
		CreatedAt: newUser.CreatedAt,
		Username:  newUser.Username,
		Role:      us.userRoles[newUser.RoleID],
	}, nil
}

func (us *UserServiceImpl) LoginUser(ctx context.Context, userLogin *dto.UserLoginDTO) (*dto.TokenDTO, error) {
	user, err := us.userRepository.GetUserByUsername(ctx, userLogin.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := crypto.CheckPassword(userLogin.Password, user.Password); err != nil {
		if errors.Is(err, crypto.ErrInvalidCredentials) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	token, err := us.tokenService.GenerateToken(user.ID, user.Username, us.userRoles[user.RoleID])
	if err != nil {
		return nil, err
	}

	return &dto.TokenDTO{
		Token: token,
	}, nil
}
