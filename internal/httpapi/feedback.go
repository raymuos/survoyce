package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/raymuos/survoyce/internal/domain"
)

const maxUploadBytes = 10 << 20 // 10 MB

type FeedbackHandler struct {
	store   domain.FeedbackStore
	storage domain.AudioStorage
	log     *slog.Logger
}

func NewFeedbackHandler(s domain.FeedbackStore, a domain.AudioStorage, log *slog.Logger) *FeedbackHandler {
	return &FeedbackHandler{store: s, storage: a, log: log}
}

func (h *FeedbackHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /businesses/{businessID}/feedback", h.Upload)
	mux.HandleFunc("GET /feedback/{id}", h.Get)
}

func (h *FeedbackHandler) Upload(w http.ResponseWriter, r *http.Request) {
	businessID := r.PathValue("businessID")
	if _, err := uuid.Parse(businessID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid business id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	file, _, err := r.FormFile("audio")
	if err != nil {
		writeError(w, http.StatusBadRequest, "audio file required (max 5 MB)")
		return
	}
	defer file.Close()

	id := uuid.NewString()
	key := "audio/" + id + ".webm"

	if err := h.storage.Put(r.Context(), key, file); err != nil {
		h.log.Error("store audio", "err", err)
		writeError(w, http.StatusInternalServerError, "could not store audio")
		return
	}

	fb := &domain.Feedback{ID: id, BusinessID: businessID, AudioKey: key}
	if err := h.store.Create(r.Context(), fb); err != nil {
		h.log.Error("create feedback", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save feedback")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"id": id, "status": string(fb.Status)})
}

type feedbackResponse struct {
	ID         string   `json:"id"`
	Status     string   `json:"status"`
	Transcript string   `json:"transcript"`
	Sentiment  float64  `json:"sentiment"`
	Themes     []string `json:"themes"`
}

func (h *FeedbackHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	fb, err := h.store.Get(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		h.log.Error("get feedback", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	themes := fb.Themes
	if themes == nil {
		themes = []string{}
	}
	writeJSON(w, http.StatusOK, feedbackResponse{
		ID: fb.ID, Status: string(fb.Status), Transcript: fb.Transcript,
		Sentiment: fb.Sentiment, Themes: themes,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
