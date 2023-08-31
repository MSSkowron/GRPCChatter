package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateUsername(t *testing.T) {
	data := []struct {
		username string
		valid    bool
	}{
		{"", false},
		{"short", false},
		{"noDigits", false},
		{"NoSpecialChars", false},
		{"Valid123", true},
		{"MoreValid99", true},
		{"With_SpecialChar1", false},
	}

	for _, d := range data {
		t.Run(d.username, func(t *testing.T) {
			err := ValidateUsername(d.username)
			if d.valid {
				assert.NoError(t, err, fmt.Sprintf("Expected a valid username, but got an error: %s", err))
			} else {
				assert.Error(t, err, "Expected an invalid username, but got no error")
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	data := []struct {
		password string
		valid    bool
	}{
		{"", false},
		{"short", false},
		{"NoUppercase", false},
		{"nolowercase", false},
		{"NoDigits!", false},
		{"Valid123!", true},
		{"MoreValid99@", true},
		{"Special@Char1", true},
	}

	for _, d := range data {
		t.Run(d.password, func(t *testing.T) {
			err := ValidatePassword(d.password)
			if d.valid {
				assert.NoError(t, err, fmt.Sprintf("Expected a valid password, but got an error: %s", err))
			} else {
				assert.Error(t, err, "Expected an invalid password, but got no error")
			}
		})
	}
}
