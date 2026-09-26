package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestUserPageCacheKey(t *testing.T) {
	if got, want := userPageCacheKey("budi123"), "page:budi123"; got != want {
		t.Fatalf("expected cache key %q, got %q", want, got)
	}
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	r := chi.NewRouter()
	r.With(RequireAuth("test-secret")).Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status %d, got %d with body %s", http.StatusUnauthorized, res.Code, res.Body.String())
	}
}
