package handler

import (
	"errors"
	"net/http"

	"shortlink/internal/service"
)

type encodeRequest struct {
	URL string `json:"url"`
}

type encodeResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *Handler) encode(w http.ResponseWriter, r *http.Request) {
	var req encodeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errTooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		writeError(w, r, status, err.Error())
		return
	}

	shortURL, err := h.svc.Encode(req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			writeError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, r, http.StatusInternalServerError, "could not encode url")
		return
	}

	writeJSON(w, http.StatusOK, encodeResponse{ShortURL: shortURL})
}
