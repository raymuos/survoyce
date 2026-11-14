package httpapi_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raymuos/survoyce/internal/domain"
	"github.com/raymuos/survoyce/internal/httpapi"
)

const bizID = "00000000-0000-0000-0000-000000000001"

type fakeStore struct{ created []*domain.Feedback }

func (f *fakeStore) Create(_ context.Context, fb *domain.Feedback) error {
	fb.Status = domain.StatusPending
	f.created = append(f.created, fb)
	return nil
}
func (f *fakeStore) Get(context.Context, string) (*domain.Feedback, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeStore) UpdateStatus(context.Context, string, domain.Status) error { return nil }
func (f *fakeStore) SaveResult(context.Context, string, domain.Analysis, string) error {
	return nil
}

type fakeStorage struct{ objects map[string][]byte }

func (f *fakeStorage) Put(_ context.Context, key string, r io.Reader) error {
	b, _ := io.ReadAll(r)
	f.objects[key] = b
	return nil
}
func (f *fakeStorage) Get(context.Context, string) (io.ReadCloser, error) { return nil, nil }

func newServer() (*http.ServeMux, *fakeStore, *fakeStorage) {
	st, sg := &fakeStore{}, &fakeStorage{objects: map[string][]byte{}}
	mux := http.NewServeMux()
	httpapi.NewFeedbackHandler(st, sg, slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)
	return mux, st, sg
}

func multipartBody(withFile bool) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if withFile {
		fw, _ := mw.CreateFormFile("audio", "clip.webm")
		fw.Write([]byte("fake audio bytes"))
	}
	mw.Close()
	return &buf, mw.FormDataContentType()
}

func TestUpload(t *testing.T) {
	tests := []struct {
		name       string
		businessID string
		withFile   bool
		wantStatus int
		wantStored bool
	}{
		{"ok", bizID, true, http.StatusAccepted, true},
		{"missing file", bizID, false, http.StatusBadRequest, false},
		{"bad business id", "not-a-uuid", true, http.StatusBadRequest, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux, st, sg := newServer()
			body, ct := multipartBody(tc.withFile)
			req := httptest.NewRequest("POST", "/businesses/"+tc.businessID+"/feedback", body)
			req.Header.Set("Content-Type", ct)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body)
			}
			if got := len(st.created) == 1 && len(sg.objects) == 1; got != tc.wantStored {
				t.Errorf("stored = %v, want %v", got, tc.wantStored)
			}
		})
	}
}
