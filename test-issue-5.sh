#!/usr/bin/env bash
set -euo pipefail

BASE="http://localhost:8080"

json_pretty() {
  python3 - "$1" <<'PY'
import json, sys
text = sys.argv[1]
try:
    print(json.dumps(json.loads(text), indent=2, ensure_ascii=False))
except Exception:
    print(text)
PY
}

req() {
  local label="$1"
  shift

  echo
  echo "===== $label ====="
  local headers_file
  headers_file=$(mktemp)
  local body_file
  body_file=$(mktemp)

  curl -sS -D "$headers_file" -o "$body_file" "$@"

  echo "--- HEADERS ---"
  sed -n '1,20p' "$headers_file"
  echo "--- BODY ---"
  json_pretty "$(cat "$body_file")"

  rm -f "$headers_file" "$body_file"
}

echo "==> Register user 1"
req "Register user 1" \
  -X POST "$BASE/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"budi123","email":"budi@example.com","password":"password123"}'

echo

echo "==> Register user 2"
req "Register user 2" \
  -X POST "$BASE/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"siti456","email":"siti@example.com","password":"password123"}'

echo

echo "==> Login user 1"
LOGIN_1=$(curl -sS -X POST "$BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"password123"}')

echo "$LOGIN_1"
TOKEN_USER_1=$(echo "$LOGIN_1" | python3 -c 'import sys, json; print(json.load(sys.stdin)["token"])')

echo

echo "==> Login user 2"
LOGIN_2=$(curl -sS -X POST "$BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"siti@example.com","password":"password123"}')

echo "$LOGIN_2"
TOKEN_USER_2=$(echo "$LOGIN_2" | python3 -c 'import sys, json; print(json.load(sys.stdin)["token"])')

echo

req "GET /api/me user 1" \
  "$BASE/api/me" \
  -H "Authorization: Bearer $TOKEN_USER_1"

echo

req "PUT /api/me/settings" \
  -X PUT "$BASE/api/me/settings" \
  -H "Authorization: Bearer $TOKEN_USER_1" \
  -H "Content-Type: application/json" \
  -d '{"bg_color":"#111827","font_family":"Inter","avatar_url":"https://example.com/avatar.png"}'

echo

LINK_CREATE=$(curl -sS -X POST "$BASE/api/me/links" \
  -H "Authorization: Bearer $TOKEN_USER_1" \
  -H "Content-Type: application/json" \
  -d '{"title":"GitHub","url":"https://github.com"}')

echo "==> Create first link"
echo "$LINK_CREATE"
LINK_ID=$(echo "$LINK_CREATE" | python3 -c 'import sys, json; print(json.load(sys.stdin)["id"])')

echo

req "POST second link" \
  -X POST "$BASE/api/me/links" \
  -H "Authorization: Bearer $TOKEN_USER_1" \
  -H "Content-Type: application/json" \
  -d '{"title":"Portfolio","url":"https://example.com"}'

echo

req "PUT /api/me/links/{id}" \
  -X PUT "$BASE/api/me/links/$LINK_ID" \
  -H "Authorization: Bearer $TOKEN_USER_1" \
  -H "Content-Type: application/json" \
  -d '{"title":"GitHub Updated","url":"https://github.com/updated","position":1}'

echo

req "DELETE /api/me/links/{id}" \
  -X DELETE "$BASE/api/me/links/$LINK_ID" \
  -H "Authorization: Bearer $TOKEN_USER_1"

echo

req "GET /api/me after delete" \
  "$BASE/api/me" \
  -H "Authorization: Bearer $TOKEN_USER_1"

echo

req "401 without token" \
  "$BASE/api/me"

echo

req "Ownership check: user 2 deletes user 1 link" \
  -X DELETE "$BASE/api/me/links/$LINK_ID" \
  -H "Authorization: Bearer $TOKEN_USER_2"
