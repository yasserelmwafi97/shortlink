package shortcode

import (
	"crypto/rand"
	"errors"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

const base = byte(len(alphabet))

var ErrInvalidLength = errors.New("shortcode: length must be positive")

func Generate(length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidLength
	}

	out := make([]byte, length)
	buf := make([]byte, length)
	filled := 0

	for filled < length {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if b >= max {
				continue
			}
			out[filled] = alphabet[b%base]
			filled++
			if filled == length {
				break
			}
		}
	}

	return string(out), nil
}

const max = byte(256 - (256 % int(base)))
