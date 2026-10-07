package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxAvatarUploadSize = 2 * 1024 * 1024 // 2 Megabytes

type avatarUploadResponse struct {
	AvatarURL string `json:"avatar_url"`
	Message   string `json:"message"`
}

// uploadAvatar handles secure avatar upload for the authenticated user.
// Security features implemented:
// 1. Authentication check via JWT middleware context
// 2. Request body size limit via http.MaxBytesReader (prevents DoS)
// 3. Magic bytes / Content-Type sniffing (validates genuine raster images, disallows SVG/HTML/XSS)
// 4. Safe random filename generation (unpredictable, client filename discarded)
// 5. Strict directory path validation (prevents directory traversal)
// 6. Secure file storage permissions (0644)
// 7. Automatic cleanup of previous uploaded avatar to avoid disk bloating
// 8. Cache invalidation on successful update
func (h *profileHandler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Limit request body to prevent DoS (2MB file + 1MB multipart boundary overhead)
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarUploadSize+(1024*1024))
	if err := r.ParseMultipartForm(maxAvatarUploadSize); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			writeError(w, http.StatusRequestEntityTooLarge, "ukuran file melebihi batas maksimal 2MB")
			return
		}
		writeError(w, http.StatusBadRequest, "gagal memproses form upload")
		return
	}

	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		file, fileHeader, err = r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "file avatar tidak ditemukan dalam form")
			return
		}
	}
	defer file.Close()

	if fileHeader.Size > maxAvatarUploadSize {
		writeError(w, http.StatusRequestEntityTooLarge, "ukuran file melebihi batas maksimal 2MB")
		return
	}

	// Read first 512 bytes to sniff actual file signature (magic bytes)
	sniffBuffer := make([]byte, 512)
	n, err := file.Read(sniffBuffer)
	if err != nil && err != io.EOF {
		writeError(w, http.StatusBadRequest, "gagal membaca data file avatar")
		return
	}

	detectedContentType := http.DetectContentType(sniffBuffer[:n])

	// Validate allowed MIME types (strictly raster images; SVG is rejected to prevent Stored XSS)
	var ext string
	switch detectedContentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	case "image/gif":
		ext = ".gif"
	default:
		writeError(w, http.StatusBadRequest, "format file tidak didukung. Harap gunakan format JPG, PNG, WEBP, atau GIF")
		return
	}

	// Construct combined reader so the first 512 bytes are not lost when saving
	combinedReader := io.MultiReader(bytes.NewReader(sniffBuffer[:n]), file)

	// Generate safe, unpredictable filename using crypto/rand (discard client filename)
	randBytes := make([]byte, 12)
	if _, err := rand.Read(randBytes); err != nil {
		writeError(w, http.StatusInternalServerError, "gagal membuat nama file avatar")
		return
	}
	safeFilename := fmt.Sprintf("avatar-%s-%d-%s%s", userID, time.Now().Unix(), hex.EncodeToString(randBytes), ext)

	// Ensure upload directory exists
	avatarDir := filepath.Join(h.uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		log.Printf("failed to create avatar directory %s: %v", avatarDir, err)
		writeError(w, http.StatusInternalServerError, "gagal menyiapkan folder penyimpanan")
		return
	}

	// Verify safe destination path
	destPath := filepath.Clean(filepath.Join(avatarDir, safeFilename))
	if !strings.HasPrefix(destPath, filepath.Clean(avatarDir)) {
		writeError(w, http.StatusBadRequest, "path tujuan tidak valid")
		return
	}

	// Write file to disk with restricted permissions (0644)
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("failed to open destination file %s: %v", destPath, err)
		writeError(w, http.StatusInternalServerError, "gagal menyimpan file avatar")
		return
	}

	copiedBytes, err := io.Copy(destFile, io.LimitReader(combinedReader, maxAvatarUploadSize))
	closeErr := destFile.Close()
	if err != nil || closeErr != nil || copiedBytes > maxAvatarUploadSize {
		_ = os.Remove(destPath)
		writeError(w, http.StatusRequestEntityTooLarge, "ukuran file melebihi batas maksimal 2MB")
		return
	}

	// Query old avatar url to clean up previously uploaded file if local
	var oldAvatarURL sql.NullString
	_ = h.db.QueryRowContext(r.Context(), `SELECT avatar_url FROM page_settings WHERE user_id = $1`, userID).Scan(&oldAvatarURL)

	// Relative URL served through our secure file server
	newAvatarURL := "/uploads/avatars/" + safeFilename

	// Update page_settings record
	_, err = h.db.ExecContext(r.Context(), `
		INSERT INTO page_settings (user_id, avatar_url, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET
			avatar_url = EXCLUDED.avatar_url,
			updated_at = CURRENT_TIMESTAMP`,
		userID, newAvatarURL)
	if err != nil {
		log.Printf("failed to update avatar_url in database: %v", err)
		_ = os.Remove(destPath)
		writeError(w, http.StatusInternalServerError, "gagal memperbarui data avatar")
		return
	}

	// Clean up previous avatar file if it was stored locally
	if oldAvatarURL.Valid && strings.HasPrefix(oldAvatarURL.String, "/uploads/avatars/") {
		oldFilename := strings.TrimPrefix(oldAvatarURL.String, "/uploads/avatars/")
		if !strings.Contains(oldFilename, "/") && !strings.Contains(oldFilename, "\\") && !strings.Contains(oldFilename, "..") {
			_ = os.Remove(filepath.Join(avatarDir, oldFilename))
		}
	}

	// Invalidate cache
	if err := h.invalidateUserPageCache(r.Context(), userID); err != nil {
		log.Printf("failed to invalidate user page cache for %s: %v", userID, err)
	}

	writeJSON(w, http.StatusOK, avatarUploadResponse{
		AvatarURL: newAvatarURL,
		Message:   "Avatar berhasil diunggah",
	})
}

// deleteAvatar removes the avatar URL and deletes any uploaded local avatar file.
func (h *profileHandler) deleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var oldAvatarURL sql.NullString
	_ = h.db.QueryRowContext(r.Context(), `SELECT avatar_url FROM page_settings WHERE user_id = $1`, userID).Scan(&oldAvatarURL)

	_, err := h.db.ExecContext(r.Context(), `
		INSERT INTO page_settings (user_id, avatar_url, updated_at)
		VALUES ($1, '', CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET
			avatar_url = '',
			updated_at = CURRENT_TIMESTAMP`,
		userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal menghapus avatar")
		return
	}

	// Clean up local file if exists
	if oldAvatarURL.Valid && strings.HasPrefix(oldAvatarURL.String, "/uploads/avatars/") {
		avatarDir := filepath.Join(h.uploadDir, "avatars")
		oldFilename := strings.TrimPrefix(oldAvatarURL.String, "/uploads/avatars/")
		if !strings.Contains(oldFilename, "/") && !strings.Contains(oldFilename, "\\") && !strings.Contains(oldFilename, "..") {
			_ = os.Remove(filepath.Join(avatarDir, oldFilename))
		}
	}

	_ = h.invalidateUserPageCache(r.Context(), userID)

	writeJSON(w, http.StatusOK, avatarUploadResponse{
		AvatarURL: "",
		Message:   "Avatar berhasil dihapus",
	})
}

// serveUploads returns an http.HandlerFunc that safely serves uploaded files.
// Security features:
// 1. Path traversal protection via filepath.Clean and prefix validation
// 2. Directory listing disabled (returns 404 for directories)
// 3. Security headers: X-Content-Type-Options: nosniff
// 4. Content-Security-Policy: default-src 'none'; sandbox (no script execution)
// 5. Cache-Control: public, max-age=86400
func (h *profileHandler) serveUploads(urlPrefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		relPath := strings.TrimPrefix(r.URL.Path, urlPrefix)
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			http.NotFound(w, r)
			return
		}

		cleanUploadDir := filepath.Clean(h.uploadDir)
		fullPath := filepath.Clean(filepath.Join(cleanUploadDir, filepath.FromSlash(relPath)))

		// Prevent directory traversal
		if !strings.HasPrefix(fullPath, cleanUploadDir) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check existence and prevent directory index listing
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
		w.Header().Set("Cache-Control", "public, max-age=86400")

		http.ServeFile(w, r, fullPath)
	}
}
