package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"linkmudah/backend/internal/cache"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

type authHandler struct {
	db        *sql.DB
	jwtSecret []byte
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type authClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type profileHandler struct {
	db  *sql.DB
	rdb *redis.Client
}

type pageSettings struct {
	BGColor    string `json:"bg_color"`
	FontFamily string `json:"font_family"`
	AvatarURL  string `json:"avatar_url"`
}

type linkPayload struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Position int    `json:"position,omitempty"`
}

type meResponse struct {
	Username     string        `json:"username"`
	Email        string        `json:"email"`
	PageSettings pageSettings  `json:"page_settings"`
	Links        []linkPayload `json:"links"`
}

type publicPageResponse struct {
	Username   string       `json:"username"`
	BGColor    string       `json:"bg_color"`
	FontFamily string       `json:"font_family"`
	AvatarURL  string       `json:"avatar_url"`
	Links      []publicLink `json:"links"`
}

type publicLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type settingsRequest struct {
	BGColor    string `json:"bg_color"`
	FontFamily string `json:"font_family"`
	AvatarURL  string `json:"avatar_url"`
}

type linkRequest struct {
	Title    *string `json:"title"`
	URL      *string `json:"url"`
	Position *int    `json:"position"`
}

type createLinkRequest struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

func newAuthHandler(db *sql.DB, secret string) *authHandler {
	return &authHandler{db: db, jwtSecret: []byte(secret)}
}

func newProfileHandler(db *sql.DB, rdb *redis.Client) *profileHandler {
	return &profileHandler{db: db, rdb: rdb}
}

func (h *profileHandler) publicPage(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	key := userPageCacheKey(username)
	if payload, err := cache.Get(r.Context(), h.rdb, key); err == nil && payload != "" {
		log.Printf("cache hit key=%s", key)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload))
		return
	}
	log.Printf("cache miss key=%s", key)

	var page publicPageResponse
	if err := h.db.QueryRowContext(r.Context(), `SELECT username FROM users WHERE username = $1`, username).Scan(&page.Username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	_ = h.db.QueryRowContext(r.Context(), `
		SELECT COALESCE(bg_color, ''), COALESCE(font_family, ''), COALESCE(avatar_url, '')
		FROM page_settings WHERE user_id = (SELECT id FROM users WHERE username = $1)`, username).
		Scan(&page.BGColor, &page.FontFamily, &page.AvatarURL)

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT title, url FROM links
		WHERE user_id = (SELECT id FROM users WHERE username = $1)
		ORDER BY position ASC, created_at ASC`, username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load links")
		return
	}
	defer rows.Close()

	page.Links = make([]publicLink, 0)
	for rows.Next() {
		var link publicLink
		if err := rows.Scan(&link.Title, &link.URL); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read links")
			return
		}
		page.Links = append(page.Links, link)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read links")
		return
	}

	if err := cache.Set(r.Context(), h.rdb, key, page, cache.DefaultTTL); err != nil {
		log.Printf("cache set failed key=%s: %v", key, err)
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *profileHandler) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var username, email string
	if err := h.db.QueryRowContext(r.Context(), `SELECT username, email FROM users WHERE id = $1`, userID).Scan(&username, &email); err != nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	settings := pageSettings{}
	_ = h.db.QueryRowContext(r.Context(), `SELECT bg_color, font_family, avatar_url FROM page_settings WHERE user_id = $1`, userID).Scan(&settings.BGColor, &settings.FontFamily, &settings.AvatarURL)

	rows, err := h.db.QueryContext(r.Context(), `SELECT id, title, url, position FROM links WHERE user_id = $1 ORDER BY position ASC, created_at ASC`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load links")
		return
	}
	defer rows.Close()

	links := make([]linkPayload, 0)
	for rows.Next() {
		var link linkPayload
		if err := rows.Scan(&link.ID, &link.Title, &link.URL, &link.Position); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read links")
			return
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read links")
		return
	}

	writeJSON(w, http.StatusOK, meResponse{Username: username, Email: email, PageSettings: settings, Links: links})
}

func (h *profileHandler) updateSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var input settingsRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var stored pageSettings
	if err := h.db.QueryRowContext(r.Context(), `
		INSERT INTO page_settings (user_id, bg_color, font_family, avatar_url, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET
			bg_color = EXCLUDED.bg_color,
			font_family = EXCLUDED.font_family,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = CURRENT_TIMESTAMP
		RETURNING bg_color, font_family, avatar_url`, userID, input.BGColor, input.FontFamily, input.AvatarURL).Scan(&stored.BGColor, &stored.FontFamily, &stored.AvatarURL); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update settings")
		return
	}

	if err := h.invalidateUserPageCache(r.Context(), userID); err != nil {
		log.Printf("cache invalidation failed for user %s: %v", userID, err)
	}
	writeJSON(w, http.StatusOK, stored)
}

func (h *profileHandler) createLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var input createLinkRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.URL = strings.TrimSpace(input.URL)
	if input.Title == "" || input.URL == "" {
		writeError(w, http.StatusBadRequest, "title and url are required")
		return
	}

	var position int
	if err := h.db.QueryRowContext(r.Context(), `SELECT COALESCE(MAX(position), 0) + 1 FROM links WHERE user_id = $1`, userID).Scan(&position); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to determine link position")
		return
	}

	linkID, err := newUserID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create link")
		return
	}
	if _, err := h.db.ExecContext(r.Context(), `INSERT INTO links (id, user_id, title, url, position) VALUES ($1, $2, $3, $4, $5)`, linkID, userID, input.Title, input.URL, position); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create link")
		return
	}

	if err := h.invalidateUserPageCache(r.Context(), userID); err != nil {
		log.Printf("cache invalidation failed for user %s: %v", userID, err)
	}
	writeJSON(w, http.StatusCreated, linkPayload{ID: linkID, Title: input.Title, URL: input.URL, Position: position})
}

func (h *profileHandler) updateLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	linkID := chi.URLParam(r, "id")
	if linkID == "" {
		writeError(w, http.StatusBadRequest, "link id is required")
		return
	}

	var input linkRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var existingUserID, title, url string
	var position int
	if err := h.db.QueryRowContext(r.Context(), `SELECT user_id, title, url, position FROM links WHERE id = $1`, linkID).Scan(&existingUserID, &title, &url, &position); err != nil {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	if existingUserID != userID {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}

	if input.Title != nil {
		title = strings.TrimSpace(*input.Title)
		if title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
	}
	if input.URL != nil {
		url = strings.TrimSpace(*input.URL)
		if url == "" {
			writeError(w, http.StatusBadRequest, "url is required")
			return
		}
	}
	if input.Position != nil {
		position = *input.Position
	}

	if _, err := h.db.ExecContext(r.Context(), `UPDATE links SET title = $1, url = $2, position = $3 WHERE id = $4 AND user_id = $5`, title, url, position, linkID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update link")
		return
	}

	if err := h.invalidateUserPageCache(r.Context(), userID); err != nil {
		log.Printf("cache invalidation failed for user %s: %v", userID, err)
	}
	writeJSON(w, http.StatusOK, linkPayload{ID: linkID, Title: title, URL: url, Position: position})
}

func (h *profileHandler) deleteLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	linkID := chi.URLParam(r, "id")
	if linkID == "" {
		writeError(w, http.StatusBadRequest, "link id is required")
		return
	}

	result, err := h.db.ExecContext(r.Context(), `DELETE FROM links WHERE id = $1 AND user_id = $2`, linkID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete link")
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}

	if err := h.invalidateUserPageCache(r.Context(), userID); err != nil {
		log.Printf("cache invalidation failed for user %s: %v", userID, err)
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *profileHandler) invalidateUserPageCache(ctx context.Context, userID string) error {
	var username string
	if err := h.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&username); err != nil {
		return err
	}
	return cache.Delete(ctx, h.rdb, userPageCacheKey(username))
}

func userPageCacheKey(username string) string {
	return cache.UserPageKey(username)
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var input registerRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(input.Email)
	if input.Username == "" || !usernamePattern.MatchString(input.Username) {
		writeError(w, http.StatusBadRequest, "username may contain only letters, numbers, and dashes")
		return
	}
	if input.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if utf8.RuneCountInString(input.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if len(h.jwtSecret) == 0 {
		writeError(w, http.StatusInternalServerError, "JWT_SECRET is not configured")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	userID, err := newUserID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	var storedID string
	err = h.db.QueryRowContext(r.Context(), `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, userID, input.Username, input.Email, string(passwordHash)).Scan(&storedID)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "username or email is already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, userResponse{ID: storedID, Username: input.Username, Email: input.Email})
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	input.Email = strings.TrimSpace(input.Email)
	var userID, username, passwordHash string
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, username, password_hash FROM users WHERE email = $1`, input.Email).
		Scan(&userID, &username, &passwordHash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if len(h.jwtSecret) == 0 {
		writeError(w, http.StatusInternalServerError, "JWT_SECRET is not configured")
		return
	}

	now := time.Now()
	claims := authClaims{
		UserID: userID, Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	writeJSON(w, http.StatusOK, tokenResponse{Token: token})
}

func RequireAuth(secret string) func(http.Handler) http.Handler {
	key := []byte(secret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(strings.TrimSpace(r.Header.Get("Authorization")))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "authorization token is required")
				return
			}
			var claims authClaims
			token, err := jwt.ParseWithClaims(parts[1], &claims, func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected signing method")
				}
				return key, nil
			})
			if err != nil || !token.Valid || claims.UserID == "" {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDContextKey, claims.UserID)))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(userIDContextKey).(string)
	return value, ok && value != ""
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(destination)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func newUserID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return hex.EncodeToString(bytes[0:4]) + "-" + hex.EncodeToString(bytes[4:6]) + "-" + hex.EncodeToString(bytes[6:8]) + "-" + hex.EncodeToString(bytes[8:10]) + "-" + hex.EncodeToString(bytes[10:16]), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
