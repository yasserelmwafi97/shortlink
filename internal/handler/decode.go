package handler

import (
	"errors"
	"net/http"

	"shortlink/internal/service"
)

type decodeRequest struct {
	ShortURL string `json:"short_url"`
}

type decodeResponse struct {
	URL string `json:"url"`
}

func (h *Handler) decode(w http.ResponseWriter, r *http.Request) {
	var req decodeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errTooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		writeError(w, r, status, err.Error())
		return
	}

	original, err := h.svc.Decode(req.ShortURL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCode):
			writeError(w, r, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrNotFound):
			writeError(w, r, http.StatusNotFound, err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "could not decode url")
		}
		return
	}

	writeJSON(w, http.StatusOK, decodeResponse{URL: original})
}
