package service

import "github.com/MSSkowron/GRPCChatter/internal/dto"

// UserService is an interface that defines the methods required for users management.
type UserService interface {
	RegisterUser(*dto.AccountCreateDTO) (*dto.UserDTO, error)
	LoginUser(*dto.UserLoginDTO) (*dto.TokenDTO, error)
}
