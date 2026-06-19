package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shortlink/internal/service"
	"shortlink/internal/store/memory"
)

func newServer() http.Handler {
	svc := service.New(memory.New(), "http://localhost:8080", 6)
	return New(svc, 4<<10, 0)
}

func do(t *testing.T, srv http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	rec := do(t, newServer(), http.MethodGet, "/healthz", "")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEncodeOK(t *testing.T) {
	rec := do(t, newServer(), http.MethodPost, "/encode", `{"url":"https://codesubmit.io/library/react"}`)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp encodeResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, strings.HasPrefix(resp.ShortURL, "http://localhost:8080/"))
}

func TestEncodeInvalidURL(t *testing.T) {
	rec := do(t, newServer(), http.MethodPost, "/encode", `{"url":"not-a-url"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEncodeBadJSON(t *testing.T) {
	rec := do(t, newServer(), http.MethodPost, "/encode", `{"url":`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEncodeUnknownField(t *testing.T) {
	rec := do(t, newServer(), http.MethodPost, "/encode", `{"url":"https://x.com","evil":1}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEncodeTooLarge(t *testing.T) {
	big := strings.Repeat("a", 5<<10)
	rec := do(t, newServer(), http.MethodPost, "/encode", `{"url":"https://x.com/`+big+`"}`)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func TestDecodeRoundTrip(t *testing.T) {
	srv := newServer()

	enc := do(t, srv, http.MethodPost, "/encode", `{"url":"https://example.com/page"}`)
	require.Equal(t, http.StatusOK, enc.Code)
	var encResp encodeResponse
	require.NoError(t, json.Unmarshal(enc.Body.Bytes(), &encResp))

	payload, _ := json.Marshal(decodeRequest{ShortURL: encResp.ShortURL})
	dec := do(t, srv, http.MethodPost, "/decode", string(payload))
	require.Equal(t, http.StatusOK, dec.Code)

	var decResp decodeResponse
	require.NoError(t, json.Unmarshal(dec.Body.Bytes(), &decResp))
	assert.Equal(t, "https://example.com/page", decResp.URL)
}

func TestDecodeNotFound(t *testing.T) {
	rec := do(t, newServer(), http.MethodPost, "/decode", `{"short_url":"http://localhost:8080/ZZZZZZ"}`)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDecodeInvalid(t *testing.T) {
	rec := do(t, newServer(), http.MethodPost, "/decode", `{"short_url":"http://localhost:8080/"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMethodNotAllowed(t *testing.T) {
	rec := do(t, newServer(), http.MethodGet, "/encode", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestContentTypeJSON(t *testing.T) {
	rec := do(t, newServer(), http.MethodPost, "/encode", `{"url":"https://x.com"}`)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
