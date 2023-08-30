package usertoken

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testSecret         = "testsecret123"
	testUserName       = "MSSkowron"
	testUserID         = 1
	testExpirationTime = time.Hour
)

func TestGenerate(t *testing.T) {
	tokenString, err := Generate(testUserID, testUserName, testExpirationTime, testSecret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)
}

func TestValidate(t *testing.T) {
	// Valid token
	tokenString, err := Generate(testUserID, testUserName, testExpirationTime, testSecret)
	require.NoError(t, err)

	err = Validate(tokenString, testSecret)
	require.NoError(t, err)

	// Invalid token
	err = Validate("invalidtoken", testSecret)
	require.ErrorIs(t, err, ErrInvalidToken)

	// Token with incorrect secret
	invalidSecret := "invalidsecret321"
	tokenString, err = Generate(testUserID, testUserName, testExpirationTime, testSecret)
	require.NoError(t, err)

	err = Validate(tokenString, invalidSecret)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestGetClaim(t *testing.T) {
	// Valid claim retrieval
	tokenString, err := Generate(testUserID, testUserName, testExpirationTime, testSecret)
	require.NoError(t, err)

	userID, err := GetClaim[float64](tokenString, testSecret, ClaimUserIDKey)
	require.NoError(t, err)
	require.Equal(t, testUserID, int(userID))

	userName, err := GetClaim[string](tokenString, testSecret, ClaimUserNameKey)
	require.NoError(t, err)
	require.Equal(t, testUserName, userName)

	expiresAt, err := GetClaim[float64](tokenString, testSecret, ClaimExpiresAtKey)
	require.NoError(t, err)
	require.GreaterOrEqual(t, expiresAt, float64(0))

	// Token with incorrect secret
	invalidSecret := "invalidsecret321"
	tokenString, err = Generate(testUserID, testUserName, testExpirationTime, testSecret)
	require.NoError(t, err)

	_, err = GetClaim[string](tokenString, invalidSecret, ClaimUserNameKey)
	require.ErrorIs(t, err, ErrInvalidToken)

	// Token with missing claims
	missingClaimsSecret := "missingclaimssecret"
	tokenString, err = Generate(testUserID, testUserName, testExpirationTime, missingClaimsSecret)
	require.NoError(t, err)

	_, err = GetClaim[string](tokenString, missingClaimsSecret, "nonexistentclaim")
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestInvalidToken(t *testing.T) {
	// Invalid token format
	err := Validate("123XDTOKEN", testSecret)
	require.ErrorIs(t, err, ErrInvalidToken)
}
