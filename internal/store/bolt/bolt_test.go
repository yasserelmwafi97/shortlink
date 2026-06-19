package bolt

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shortlink/internal/shortcode"
	"shortlink/internal/store"
)

func gen() (string, error) {
	return shortcode.Generate(6)
}

func TestSaveAndLookup(t *testing.T) {
	st := open(t)

	code, err := st.SaveURL("https://example.com", gen)
	require.NoError(t, err)

	got, err := st.Lookup(code)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", got)
}

func TestLookupNotFound(t *testing.T) {
	st := open(t)

	_, err := st.Lookup("missing")
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestSaveIdempotent(t *testing.T) {
	st := open(t)

	first, err := st.SaveURL("https://example.com", gen)
	require.NoError(t, err)
	second, err := st.SaveURL("https://example.com", gen)
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestPersistenceAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "restart.db")

	st1, err := Open(path)
	require.NoError(t, err)
	code, err := st1.SaveURL("https://persisted.example.com", gen)
	require.NoError(t, err)
	require.NoError(t, st1.Close())

	st2, err := Open(path)
	require.NoError(t, err)
	defer st2.Close()

	got, err := st2.Lookup(code)
	require.NoError(t, err)
	assert.Equal(t, "https://persisted.example.com", got)
}

func TestConcurrentSameURL(t *testing.T) {
	st := open(t)

	const workers = 50
	codes := make([]string, workers)
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			code, err := st.SaveURL("https://race.example.com", gen)
			assert.NoError(t, err)
			codes[idx] = code
		}(i)
	}
	wg.Wait()

	for i := 1; i < workers; i++ {
		assert.Equal(t, codes[0], codes[i], "all concurrent writers must get the same code")
	}
}

func open(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	st, err := Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	return st
}
