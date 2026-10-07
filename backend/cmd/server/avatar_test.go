package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

// 1x1 transparent PNG bytes
var samplePNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
	0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func createMultipartRequest(t *testing.T, fieldName, filename string, content []byte) (*http.Request, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, writer.FormDataContentType()
}

func TestUploadAvatarUnauthorized(t *testing.T) {
	r := chi.NewRouter()
	r.With(RequireAuth("test-secret")).Post("/api/me/avatar", func(w http.ResponseWriter, r *http.Request) {})

	req, _ := createMultipartRequest(t, "avatar", "test.png", samplePNG)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", res.Code)
	}
}

func TestUploadAvatarRejectsSVG(t *testing.T) {
	// SVGs can carry malicious scripts; magic bytes verification must reject it
	tempDir := t.TempDir()
	handler := newProfileHandler(nil, nil, tempDir)

	svgContent := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)
	req, _ := createMultipartRequest(t, "avatar", "malicious.svg", svgContent)
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, "user-123"))

	res := httptest.NewRecorder()
	handler.uploadAvatar(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for SVG upload, got %d", res.Code)
	}
}

func TestUploadAvatarRejectsOverMaxFileSize(t *testing.T) {
	tempDir := t.TempDir()
	handler := newProfileHandler(nil, nil, tempDir)

	// Create payload larger than 2MB
	largePayload := make([]byte, 2*1024*1024+100)
	copy(largePayload, samplePNG)

	req, _ := createMultipartRequest(t, "avatar", "huge.png", largePayload)
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, "user-123"))

	res := httptest.NewRecorder()
	handler.uploadAvatar(res, req)

	if res.Code != http.StatusRequestEntityTooLarge && res.Code != http.StatusBadRequest {
		t.Fatalf("expected 413 or 400 for oversize upload, got %d", res.Code)
	}
}

func TestUploadAvatarSuccessPNG(t *testing.T) {
	tempDir := t.TempDir()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer rdb.Close()

	handler := newProfileHandler(db, rdb, tempDir)

	// Expect user page cache invalidation query to find username
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT avatar_url FROM page_settings WHERE user_id = $1`)).
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"avatar_url"}).AddRow(""))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO page_settings (user_id, avatar_url, updated_at)`)).
		WithArgs("user-123", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT username FROM users WHERE id = $1`)).
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("budi123"))

	req, _ := createMultipartRequest(t, "avatar", "profile.png", samplePNG)
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, "user-123"))

	res := httptest.NewRecorder()
	handler.uploadAvatar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", res.Code, res.Body.String())
	}

	var resp avatarUploadResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp.AvatarURL == "" || !bytes.Contains([]byte(resp.AvatarURL), []byte("/uploads/avatars/avatar-user-123-")) {
		t.Fatalf("unexpected avatar url: %s", resp.AvatarURL)
	}

	// Verify file was written to disk
	avatarFiles, err := os.ReadDir(filepath.Join(tempDir, "avatars"))
	if err != nil || len(avatarFiles) != 1 {
		t.Fatalf("expected 1 file in avatars dir, got %v (err: %v)", avatarFiles, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAvatar(t *testing.T) {
	tempDir := t.TempDir()
	avatarDir := filepath.Join(tempDir, "avatars")
	_ = os.MkdirAll(avatarDir, 0755)
	existingFilename := "avatar-user-123-test.png"
	_ = os.WriteFile(filepath.Join(avatarDir, existingFilename), samplePNG, 0644)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer rdb.Close()

	handler := newProfileHandler(db, rdb, tempDir)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT avatar_url FROM page_settings WHERE user_id = $1`)).
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"avatar_url"}).AddRow("/uploads/avatars/" + existingFilename))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO page_settings (user_id, avatar_url, updated_at)`)).
		WithArgs("user-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT username FROM users WHERE id = $1`)).
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("budi123"))

	req := httptest.NewRequest(http.MethodDelete, "/api/me/avatar", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, "user-123"))

	res := httptest.NewRecorder()
	handler.deleteAvatar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", res.Code, res.Body.String())
	}

	// Verify the local file was removed
	if _, err := os.Stat(filepath.Join(avatarDir, existingFilename)); !os.IsNotExist(err) {
		t.Fatalf("expected avatar file to be deleted, but it still exists")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestServeUploadsSecurityHeaders(t *testing.T) {
	tempDir := t.TempDir()
	avatarDir := filepath.Join(tempDir, "avatars")
	_ = os.MkdirAll(avatarDir, 0755)
	testFile := filepath.Join(avatarDir, "pic.png")
	_ = os.WriteFile(testFile, samplePNG, 0644)

	handler := newProfileHandler(nil, nil, tempDir)
	serveHandler := handler.serveUploads("/uploads")

	// Test successful file serve
	req := httptest.NewRequest(http.MethodGet, "/uploads/avatars/pic.png", nil)
	res := httptest.NewRecorder()
	serveHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.Code)
	}
	if res.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options: nosniff")
	}
	if res.Header().Get("Content-Security-Policy") == "" {
		t.Fatalf("expected Content-Security-Policy header")
	}

	// Test directory listing blocked (404)
	dirReq := httptest.NewRequest(http.MethodGet, "/uploads/avatars", nil)
	dirRes := httptest.NewRecorder()
	serveHandler(dirRes, dirReq)
	if dirRes.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for directory, got %d", dirRes.Code)
	}

	// Test non-existent file
	noneReq := httptest.NewRequest(http.MethodGet, "/uploads/avatars/nonexistent.png", nil)
	noneRes := httptest.NewRecorder()
	serveHandler(noneRes, noneReq)
	if noneRes.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing file, got %d", noneRes.Code)
	}
}
