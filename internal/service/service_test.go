package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shortlink/internal/store/memory"
)

func newService() *Service {
	return New(memory.New(), "http://localhost:8080", 6)
}

func TestEncodeValid(t *testing.T) {
	svc := newService()

	short, err := svc.Encode("https://codesubmit.io/library/react")
	require.NoError(t, err)
	assert.Contains(t, short, "http://localhost:8080/")
	assert.Len(t, short, len("http://localhost:8080/")+6)
}

func TestEncodeInvalid(t *testing.T) {
	svc := newService()

	cases := []string{
		"",
		"   ",
		"not-a-url",
		"ftp://example.com",
		"javascript:alert(1)",
		"https://",
		"//example.com",
	}

	for _, in := range cases {
		_, err := svc.Encode(in)
		assert.ErrorIs(t, err, ErrInvalidURL, "input=%q", in)
	}
}

func TestEncodeIdempotent(t *testing.T) {
	svc := newService()

	first, err := svc.Encode("https://example.com/path")
	require.NoError(t, err)
	second, err := svc.Encode("https://example.com/path")
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	svc := newService()

	const original = "https://example.com/a/b/c?q=1"
	short, err := svc.Encode(original)
	require.NoError(t, err)

	got, err := svc.Decode(short)
	require.NoError(t, err)
	assert.Equal(t, original, got)
}

func TestDecodeBareCode(t *testing.T) {
	svc := newService()

	short, err := svc.Encode("https://example.com")
	require.NoError(t, err)

	code := short[len("http://localhost:8080/"):]
	got, err := svc.Decode(code)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", got)
}

func TestDecodeNotFound(t *testing.T) {
	svc := newService()

	_, err := svc.Decode("http://localhost:8080/ZZZZZZ")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDecodeInvalid(t *testing.T) {
	svc := newService()

	cases := []string{"", "  ", "http://localhost:8080/bad code", "abc/def/ghi", "http://localhost:8080/"}
	for _, in := range cases {
		_, err := svc.Decode(in)
		assert.ErrorIs(t, err, ErrInvalidCode, "input=%q", in)
	}
}
