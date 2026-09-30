package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
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

func TestPublicPageCacheHit(t *testing.T) {
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer rdb.Close()

	payload := `{"username":"budi123","bg_color":"#fff","font_family":"Inter","avatar_url":"","links":[]}`
	if err := rdb.Set(t.Context(), userPageCacheKey("budi123"), payload, 0).Err(); err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Get("/api/public/{username}", newProfileHandler(nil, rdb).publicPage)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/public/budi123", nil))

	if res.Code != http.StatusOK || res.Body.String() != payload {
		t.Fatalf("expected cached response, got status %d and body %q", res.Code, res.Body.String())
	}
}

func TestPublicPageCacheMiss(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer rdb.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT username FROM users WHERE username = $1`)).
		WithArgs("budi123").WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("budi123"))
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COALESCE(bg_color, ''), COALESCE(font_family, ''), COALESCE(avatar_url, ''), COALESCE(title, ''), COALESCE(description, '')
		FROM page_settings WHERE user_id = (SELECT id FROM users WHERE username = $1)`)).
		WithArgs("budi123").WillReturnRows(sqlmock.NewRows([]string{"bg_color", "font_family", "avatar_url", "title", "description"}).AddRow("#fff", "Inter", "avatar.png", "Budi Dev", "Software Engineer"))
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT title, url FROM links
		WHERE user_id = (SELECT id FROM users WHERE username = $1)
		ORDER BY position ASC, created_at ASC`)).
		WithArgs("budi123").WillReturnRows(sqlmock.NewRows([]string{"title", "url"}).AddRow("Blog", "https://example.com"))

	r := chi.NewRouter()
	r.Get("/api/public/{username}", newProfileHandler(db, rdb).publicPage)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/public/budi123", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, res.Code, res.Body.String())
	}
	var page publicPageResponse
	if err := json.Unmarshal(res.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Username != "budi123" || len(page.Links) != 1 || page.Links[0].Title != "Blog" {
		t.Fatalf("unexpected public page: %+v", page)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if !redisServer.Exists(userPageCacheKey("budi123")) {
		t.Fatal("expected public page to be cached")
	}
}

func TestPublicPageUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer rdb.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT username FROM users WHERE username = $1`)).
		WithArgs("missing").WillReturnRows(sqlmock.NewRows([]string{"username"}))

	r := chi.NewRouter()
	r.Get("/api/public/{username}", newProfileHandler(db, rdb).publicPage)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/public/missing", nil))

	if res.Code != http.StatusNotFound || res.Body.String() != "{\"error\":\"user not found\"}\n" {
		t.Fatalf("expected user not found response, got status %d and body %q", res.Code, res.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
