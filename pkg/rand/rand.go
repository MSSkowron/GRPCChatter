package rand

import (
	"crypto/rand"
	"math/big"
)

var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// Str generates a random string of the specified length using the characters from 'chars'.
func Str(length int) string {
	b := make([]rune, length)
	for i := range b {
		letterIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[letterIdx.Int64()]
	}
	return string(b)
}
