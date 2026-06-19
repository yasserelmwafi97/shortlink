package bolt

import (
	"time"

	bolt "go.etcd.io/bbolt"

	"shortlink/internal/store"
)

const maxAttempts = 10

var (
	bucketURLs  = []byte("urls")
	bucketCodes = []byte("codes")
)

type Store struct {
	db *bolt.DB
}

func Open(path string) (*Store, error) {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketURLs); err != nil {
			return err
		}
		_, err := tx.CreateBucketIfNotExists(bucketCodes)
		return err
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) SaveURL(url string, gen func() (string, error)) (string, error) {
	var code string

	err := s.db.Update(func(tx *bolt.Tx) error {
		codes := tx.Bucket(bucketCodes)
		urls := tx.Bucket(bucketURLs)

		if existing := codes.Get([]byte(url)); existing != nil {
			code = string(existing)
			return nil
		}

		for i := 0; i < maxAttempts; i++ {
			candidate, err := gen()
			if err != nil {
				return err
			}
			if urls.Get([]byte(candidate)) != nil {
				continue
			}
			if err := urls.Put([]byte(candidate), []byte(url)); err != nil {
				return err
			}
			if err := codes.Put([]byte(url), []byte(candidate)); err != nil {
				return err
			}
			code = candidate
			return nil
		}

		return store.ErrCodeExhausted
	})
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Store) Lookup(code string) (string, error) {
	var url string

	err := s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketURLs).Get([]byte(code))
		if v == nil {
			return store.ErrNotFound
		}
		url = string(v)
		return nil
	})
	if err != nil {
		return "", err
	}

	return url, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
