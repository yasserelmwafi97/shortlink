package service

import "errors"

var (
	ErrInvalidURL  = errors.New("invalid url")
	ErrInvalidCode = errors.New("invalid short url")
	ErrNotFound    = errors.New("not found")
)
