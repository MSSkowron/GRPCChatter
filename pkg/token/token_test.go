package token

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testSecret    = "testsecret123"
	testUserName  = "MSSkowron"
	testShortCode = "ABC123"
)

func TestGenerate(t *testing.T) {
	tokenString, err := Generate(testUserName, testShortCode, testSecret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)
}

func TestValidate(t *testing.T) {
	// Valid token
	tokenString, err := Generate(testUserName, testShortCode, testSecret)
	require.NoError(t, err)

	err = Validate(tokenString, testSecret)
	require.NoError(t, err)

	// Invalid token
	err = Validate("invalidtoken", testSecret)
	require.ErrorIs(t, err, ErrInvalidToken)

	// Token with incorrect secret
	invalidSecret := "invalidsecret321"
	tokenString, err = Generate(testUserName, testShortCode, testSecret)
	require.NoError(t, err)

	err = Validate(tokenString, invalidSecret)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestGetClaim(t *testing.T) {
	// Valid claim retrieval
	tokenString, err := Generate(testUserName, testShortCode, testSecret)
	require.NoError(t, err)

	userName, err := GetClaim(tokenString, testSecret, ClaimUserNameKey)
	require.NoError(t, err)
	require.Equal(t, testUserName, userName)

	shortCode, err := GetClaim(tokenString, testSecret, ClaimShortCodeKey)
	require.NoError(t, err)
	require.Equal(t, testShortCode, shortCode)

	// Token with incorrect secret
	invalidSecret := "invalidsecret321"
	tokenString, err = Generate(testUserName, testShortCode, testSecret)
	require.NoError(t, err)

	_, err = GetClaim(tokenString, invalidSecret, ClaimUserNameKey)
	require.ErrorIs(t, err, ErrInvalidToken)

	// Token with missing claims
	missingClaimsSecret := "missingclaimssecret"
	tokenString, err = Generate(testUserName, testShortCode, missingClaimsSecret)
	require.NoError(t, err)

	_, err = GetClaim(tokenString, missingClaimsSecret, "nonexistentclaim")
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestInvalidToken(t *testing.T) {
	// Invalid token format
	err := Validate("123XDTOKEN", testSecret)
	require.ErrorIs(t, err, ErrInvalidToken)
}
