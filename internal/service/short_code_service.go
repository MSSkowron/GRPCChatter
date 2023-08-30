package service

import "github.com/MSSkowron/GRPCChatter/pkg/rand"

// ShortCodeService is an interface that defines the methods required for short code management.
type ShortCodeService interface {
	// GenerateShortCode generates a short code for a given room name.
	// It returns the generated short code.
	GenerateShortCode(roomName string) string
}

// ShortCodeServiceImpl implements the ShortCodeService interface.
type ShortCodeServiceImpl struct {
	shortCodeLength int
}

// NewShortCodeService creates a new ShortCodeServiceImpl instance with the specified short code length.
func NewShortCodeService(shortCodeLength int) ShortCodeService {
	return &ShortCodeServiceImpl{
		shortCodeLength: shortCodeLength,
	}
}

// GenerateShortCode generates a short code for a given room name.
// It returns the generated short code.
func (s *ShortCodeServiceImpl) GenerateShortCode(roomName string) string {
	return rand.Str(s.shortCodeLength)
}
