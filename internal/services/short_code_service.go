package services

import (
	"crypto/rand"
	"math/big"
)

// ShortCodeService is an interface that defines the methods that the ShortCodeService must implement.
type ShortCodeService interface {
	GenerateShortCode(string) string
}

// ShortCodeServiceImpl implements the ShortCodeService interface.
type ShortCodeServiceImpl struct {
	shortCodeLength int
}

// NewShortCodeService creates a new ShortCodeServiceImpl.
func NewShortCodeService(shortCodeLength int) *ShortCodeServiceImpl {
	return &ShortCodeServiceImpl{
		shortCodeLength: shortCodeLength,
	}
}

func (s *ShortCodeServiceImpl) GenerateShortCode(roomName string) string {
	return randStr(s.shortCodeLength)
}

var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		letterIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[letterIdx.Int64()]
	}
	return string(b)
}
