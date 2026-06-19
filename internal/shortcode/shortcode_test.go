package shortcode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateLength(t *testing.T) {
	for _, length := range []int{1, 6, 7, 12} {
		code, err := Generate(length)
		require.NoError(t, err)
		assert.Len(t, code, length)
	}
}

func TestGenerateCharset(t *testing.T) {
	code, err := Generate(64)
	require.NoError(t, err)
	for _, r := range code {
		assert.True(t, strings.ContainsRune(alphabet, r), "unexpected rune %q", r)
	}
}

func TestGenerateInvalidLength(t *testing.T) {
	_, err := Generate(0)
	assert.ErrorIs(t, err, ErrInvalidLength)
}

func TestGenerateUniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		code, err := Generate(8)
		require.NoError(t, err)
		_, dup := seen[code]
		require.False(t, dup, "unexpected collision at length 8")
		seen[code] = struct{}{}
	}
}
