package rand

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStr(t *testing.T) {
	data := []struct {
		inputStrLen          int
		outputStrExpectedLen int
	}{
		{0, 0},
		{1, 1},
		{10, 10},
		{100, 100},
	}

	for _, d := range data {
		t.Run(fmt.Sprintf("Length %d", d.inputStrLen), func(t *testing.T) {
			randomStr := Str(d.inputStrLen)
			assert.Equal(t, d.outputStrExpectedLen, len(randomStr), fmt.Sprintf("Expected string of length %d, but got length %d", d.outputStrExpectedLen, len(randomStr)))
		})
	}

	// Additional test to check for randomness
	t.Run("RandomnessTest", func(t *testing.T) {
		randomStr1 := Str(10)
		randomStr2 := Str(10)

		assert.NotEqual(t, randomStr1, randomStr2, fmt.Sprintf("Random strings are not unique: %s and %s", randomStr1, randomStr2))
	})
}
