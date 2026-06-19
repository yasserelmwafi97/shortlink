package store

import "errors"

var (
	ErrNotFound      = errors.New("store: code not found")
	ErrCodeExhausted = errors.New("store: could not generate a unique code")
)

type Store interface {
	SaveURL(url string, gen func() (string, error)) (string, error)
	Lookup(code string) (string, error)
	Close() error
}
