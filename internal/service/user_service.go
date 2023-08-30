package service

import (
	"github.com/MSSkowron/GRPCChatter/internal/dto"
	"github.com/MSSkowron/GRPCChatter/internal/repository"
)

type UserService interface {
	RegisterUser(*dto.UserRegisterDTO) (*dto.UserDTO, error)
	LoginUser(*dto.UserLoginDTO) (*dto.TokenDTO, error)
}

type UserServiceImpl struct {
	tokenService   UserTokenService
	userRepository repository.UserRepository
}

func NewUserServiceImpl(tokenService UserTokenService, userRepository repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{
		tokenService:   tokenService,
		userRepository: userRepository,
	}
}

func (us *UserServiceImpl) RegisterUser(*dto.UserRegisterDTO) (*dto.UserDTO, error) {
	return nil, nil
}

func (us *UserServiceImpl) LoginUser(*dto.UserLoginDTO) (*dto.TokenDTO, error) {
	return nil, nil
}
