package service

import "github.com/MSSkowron/GRPCChatter/internal/dtos"

// UserService is an interface that defines the methods required for users management.
type UserService interface {
	RegisterUser(*dtos.AccountCreateDTO) (*dtos.UserDTO, error)
	LoginUser(*dtos.UserLoginDTO) (*dtos.TokenDTO, error)
}
