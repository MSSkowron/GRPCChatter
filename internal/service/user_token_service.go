package service

// UserTokenService is an interface that defines the methods required for user token management.
type UserTokenService interface {
	// GenerateToken generates a user token.
	GenerateToken(int, string) (string, error)
	// ValidateToken validates a user token.
	ValidateToken(string) error
	// GetUserIDFromToken retrieves the user ID from a user token.
	GetUserIDFromToken(string) (int, error)
	// GetUserNameFromToken retrieves the user name from a user token.
	GetUserNameFromToken(string) (string, error)
}
