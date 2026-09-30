# Implementasi Platform Link-in-Bio — Daftar Issue

Dokumen ini memecah arsitektur yang sudah ditentukan menjadi issue-issue kecil yang bisa dikerjakan berurutan. **Ikuti instruksi apa adanya, jangan menambah abstraksi/pola desain baru** (contoh: jangan bikin microservice terpisah, jangan bikin sistem plugin, jangan pakai gRPC dulu — REST API saja sudah cukup untuk MVP).

## Prinsip Kerja
1. Kerjakan issue secara berurutan sesuai nomor — banyak issue bergantung pada issue sebelumnya.
2. Setiap issue punya **Tugas** (checklist) dan **Kriteria Selesai**. Jangan lanjut ke issue berikutnya sebelum kriteria selesai terpenuhi.
3. Jangan menambah fitur yang tidak diminta di issue ini (misal: multi-bahasa, dark mode toggle, analytics dashboard) — itu di luar cakupan MVP ini.
4. Gunakan library yang direkomendasikan di setiap issue, jangan cari alternatif sendiri kecuali library tersebut sudah tidak ada/deprecated.

## Struktur Folder Proyek (target akhir)

```
project-root/
├── compose.yaml
├── .env.example
├── backend/
│   ├── go.mod
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── auth/
│   │   ├── handler/
│   │   ├── model/
│   │   ├── repository/
│   │   └── cache/
│   └── migrations/
├── frontend/
│   ├── package.json
│   ├── vite.config.js
│   ├── tailwind.config.js
│   └── src/
│       ├── lib/
│       ├── routes/ (atau pages/)
│       └── App.svelte
└── proxy/
    └── nginx.conf (atau traefik.yaml)
```

---

## Milestone 1 — Infrastruktur Dasar

### Issue #1: Setup Repository & Skeleton Podman Compose

**Depends on:** —

**Deskripsi**
Buat struktur folder dasar sesuai di atas, dan buat `compose.yaml` yang mendefinisikan 4 service kosong (belum berisi kode, cukup skeleton): `frontend`, `backend`, `postgres`, `redis`. Tujuan issue ini hanya memastikan ke-4 container bisa `podman compose up` bersamaan tanpa error, sebelum kode aplikasi ditulis.

**Tugas**
- [ ] Buat folder `backend/`, `frontend/`, `proxy/` sesuai struktur di atas.
- [ ] Buat `compose.yaml` di root dengan 4 service: `frontend`, `backend`, `postgres`, `redis`.
- [ ] Service `postgres` pakai image resmi `postgres:16-alpine`, dengan environment `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` diambil dari file `.env`.
- [ ] Service `redis` pakai image resmi `redis:7-alpine`.
- [ ] Buat file `.env.example` berisi semua variabel environment yang dipakai (jangan commit `.env` asli).
- [ ] Tambahkan named volume untuk `postgres` agar data tidak hilang saat container di-restart.
- [ ] Pastikan `postgres` dan `redis` punya `healthcheck` di compose.yaml.

**Kriteria Selesai**
- `podman compose up postgres redis` berhasil jalan, `podman exec` ke container postgres dan redis bisa dilakukan tanpa error.

---

### Issue #2: Skema Database PostgreSQL & Migration

**Depends on:** Issue #1

**Deskripsi**
Buat skema tabel minimal yang dibutuhkan. Gunakan tool migration sederhana: **`golang-migrate/migrate`** (jangan pakai ORM auto-migrate supaya skema eksplisit dan mudah diaudit).

**Tugas**
- [ ] Install CLI `golang-migrate` di environment development.
- [ ] Buat folder `backend/migrations/`.
- [ ] Buat migration untuk tabel `users`:
  - `id` (UUID, primary key)
  - `username` (text, unique, not null)
  - `email` (text, unique, not null)
  - `password_hash` (text, not null)
  - `created_at`, `updated_at` (timestamp)
- [ ] Buat migration untuk tabel `page_settings` (1 baris per user, untuk tema halaman publik):
  - `user_id` (UUID, foreign key ke `users.id`, unique)
  - `bg_color` (text)
  - `font_family` (text)
  - `avatar_url` (text, nullable)
  - `updated_at` (timestamp)
- [ ] Buat migration untuk tabel `links`:
  - `id` (UUID, primary key)
  - `user_id` (UUID, foreign key ke `users.id`)
  - `title` (text, not null)
  - `url` (text, not null)
  - `position` (integer, untuk urutan tampil)
  - `created_at` (timestamp)
- [ ] Jalankan migration ke database di container `postgres` dan pastikan berhasil.

**Kriteria Selesai**
- `\dt` di psql menampilkan 3 tabel: `users`, `page_settings`, `links`.
- Migration bisa di-rollback (`migrate down`) tanpa error.

---

## Milestone 2 — Backend (Go)

### Issue #3: Inisialisasi Go Project & Health Check

**Depends on:** Issue #2

**Deskripsi**
Inisialisasi project Go dengan router HTTP ringan. Gunakan **`chi`** (`github.com/go-chi/chi/v5`) sebagai router — jangan pakai framework berat seperti Gin/Echo kecuali sudah biasa, `chi` cukup untuk kebutuhan ini dan tetap kompatibel dengan `net/http` standar.

**Tugas**
- [ ] `go mod init` di folder `backend/`.
- [ ] Install `chi` sebagai router.
- [ ] Buat `cmd/server/main.go` yang menjalankan HTTP server di port `8080` (ambil dari env var `PORT`, default `8080`).
- [ ] Buat endpoint `GET /healthz` yang mengembalikan `{"status": "ok"}` dengan status code 200.
- [ ] Baca koneksi ke Postgres dan Redis dari environment variable saat startup, dan gagal start (fatal log) jika salah satu tidak bisa dihubungi.
- [ ] Tambahkan Dockerfile untuk service `backend` di compose.yaml (multi-stage build: build binary Go, lalu jalankan di image `alpine` kecil).

**Kriteria Selesai**
- `curl http://localhost:8080/healthz` dari luar container mengembalikan `{"status":"ok"}`.

---

### Issue #4: Autentikasi (Register, Login, JWT Middleware)

**Depends on:** Issue #3

**Deskripsi**
Implementasi register & login sederhana dengan JWT. Gunakan **`github.com/golang-jwt/jwt/v5`** untuk JWT dan **`golang.org/x/crypto/bcrypt`** untuk hash password. Gunakan **`database/sql`** + driver **`github.com/jackc/pgx/v5/stdlib`** untuk akses Postgres (raw SQL query, tidak perlu ORM).

**Tugas**
- [ ] Buat endpoint `POST /api/auth/register` — terima `username`, `email`, `password`. Hash password dengan bcrypt sebelum simpan ke tabel `users`.
- [ ] Validasi: `username` hanya huruf/angka/dash, `password` minimal 8 karakter. Kembalikan error 400 dengan pesan jelas jika validasi gagal.
- [ ] Kembalikan error 409 jika `username` atau `email` sudah dipakai.
- [ ] Buat endpoint `POST /api/auth/login` — terima `email`, `password`. Jika cocok, generate JWT (masa berlaku 7 hari) berisi `user_id` dan `username` di payload. Kembalikan token di response body.
- [ ] Buat middleware `RequireAuth` yang membaca header `Authorization: Bearer <token>`, validasi JWT, dan menyimpan `user_id` ke context request. Kembalikan 401 jika token tidak ada/invalid/expired.
- [ ] Simpan `JWT_SECRET` sebagai environment variable, jangan hardcode di kode.

**Kriteria Selesai**
- Register user baru berhasil lalu login dengan kredensial yang sama mengembalikan token JWT valid.
- Request ke endpoint yang dilindungi tanpa token mengembalikan 401.

---

### Issue #5: CRUD Profil, Tema, dan Tautan (Links)

**Depends on:** Issue #4

**Deskripsi**
Endpoint yang dipakai dashboard untuk mengatur konten halaman. Semua endpoint di bawah wajib pakai middleware `RequireAuth` dari Issue #4, dan hanya boleh mengubah/menghapus data milik `user_id` yang login (jangan izinkan user A mengedit data user B).

**Tugas**
- [ ] `GET /api/me` — kembalikan data user yang sedang login: `username`, `email`, `page_settings`, daftar `links` (urut berdasarkan `position`).
- [ ] `PUT /api/me/settings` — update `bg_color`, `font_family`, `avatar_url` di `page_settings`. Buat baris baru jika belum ada (upsert).
- [ ] `POST /api/me/links` — tambah link baru (`title`, `url`), `position` = jumlah link saat ini + 1.
- [ ] `PUT /api/me/links/{id}` — update `title`, `url`, atau `position` dari satu link. Kembalikan 404 jika link tidak ditemukan atau bukan milik user tersebut.
- [ ] `DELETE /api/me/links/{id}` — hapus satu link. Kembalikan 404 jika bukan milik user tersebut.
- [ ] Setelah setiap perubahan di atas berhasil disimpan ke Postgres, **hapus** cache Redis dengan key `page:{username}` milik user tersebut (lihat Issue #6 untuk format key) supaya cache tidak basi.

**Kriteria Selesai**
- Semua endpoint di atas bisa diuji lewat `curl`/Postman dengan token JWT valid, dan mengembalikan data yang sesuai.
- Mencoba mengedit/menghapus link milik user lain mengembalikan 404, bukan 200.

---

### Issue #6: Integrasi Redis Caching

**Depends on:** Issue #5

**Deskripsi**
Tambahkan layer cache untuk data halaman publik. Gunakan **`github.com/redis/go-redis/v9`**.

**Tugas**
- [ ] Buat helper `cache.Get(ctx, key)` dan `cache.Set(ctx, key, value, ttl)` yang membungkus client Redis.
- [ ] Format key cache: `page:{username}` (contoh: `page:budi123`).
- [ ] Isi cache berupa JSON string hasil serialisasi: `page_settings` + daftar `links` milik user tersebut.
- [ ] TTL cache: 1 jam (3600 detik) — sebagai fallback jika proses invalidasi di Issue #5 terlewat.
- [ ] Pastikan helper `cache.Delete(ctx, key)` juga tersedia dan sudah dipanggil dari semua endpoint di Issue #5 yang mengubah data.

**Kriteria Selesai**
- Setelah endpoint `PUT /api/me/settings` dipanggil, key `page:{username}` di Redis langsung hilang (bisa dicek dengan `redis-cli GET page:{username}` via `podman exec`).

---

### Issue #7: Endpoint Publik untuk Halaman Pengunjung

**Depends on:** Issue #6

**Deskripsi**
Endpoint tanpa autentikasi yang dipakai frontend untuk menampilkan halaman publik seseorang.

**Tugas**
- [ ] Buat endpoint `GET /api/public/{username}` (endpoint ini **tidak** memakai middleware `RequireAuth`).
- [ ] Alur: cek Redis dengan key `page:{username}` dulu. Jika ada (cache hit), langsung kembalikan datanya.
- [ ] Jika tidak ada di Redis (cache miss): query ke Postgres (`page_settings` + `links` untuk username tersebut), simpan hasilnya ke Redis dengan TTL 1 jam, lalu kembalikan ke client.
- [ ] Jika `username` tidak ditemukan di Postgres, kembalikan status 404 dengan body `{"error": "user not found"}`.
- [ ] Response format konsisten: `{"username": "...", "bg_color": "...", "font_family": "...", "avatar_url": "...", "links": [{"title": "...", "url": "..."}]}`.

**Kriteria Selesai**
- `curl http://localhost:8080/api/public/budi123` mengembalikan data lengkap jika user ada, dan 404 jika tidak.
- Request kedua ke username yang sama jauh lebih cepat dari request pertama (menandakan cache bekerja) — bisa dicek dengan log sederhana "cache hit"/"cache miss".

---

## Milestone 3 — Frontend (Svelte)

### Issue #8: Setup Svelte + Vite + Tailwind + shadcn-svelte

**Depends on:** —  (bisa dikerjakan paralel dengan Milestone 2)

**Deskripsi**
Inisialisasi project frontend. Ini SPA (Single Page App) biasa dengan Svelte + Vite, **bukan** SvelteKit — jangan install SvelteKit.

**Tugas**
- [ ] Jalankan `bun create vite frontend --template svelte` di root project (atau pindahkan hasilnya ke folder `frontend/` yang sudah ada).
- [ ] Install Tailwind CSS sesuai panduan resmi untuk Vite, dan pastikan `tailwind.config.js` mengarah ke folder `src/**/*.{svelte,js,ts}`.
- [ ] Install dan setup `shadcn-svelte` (jalankan `npx shadcn-svelte@latest init` lalu tambahkan komponen dasar: `button`, `input`, `card`).
- [ ] Install `svelte-spa-router` untuk routing client-side sederhana (jangan pakai router yang lebih kompleks).
- [ ] Definisikan 3 route di `App.svelte`: `/login`, `/dashboard`, dan wildcard `/*` (untuk halaman publik berdasarkan username — akan ditangani di Issue #11).
- [ ] Buat file `.env` frontend berisi `VITE_API_BASE_URL` yang menunjuk ke backend (contoh: `http://localhost:8080`).
- [ ] Tambahkan Dockerfile untuk service `frontend` di compose.yaml (build dengan `bun run build`, lalu serve hasil `dist/` pakai static server ringan seperti `serve` atau `nginx`).

**Kriteria Selesai**
- `bun run dev` menjalankan aplikasi kosong dengan Tailwind aktif (bisa dites dengan menambahkan 1 class `bg-red-500` dan melihat efeknya).

---

### Issue #9: Halaman Login & Register

**Depends on:** Issue #8, Issue #4

**Deskripsi**
Form autentikasi yang memanggil API backend.

**Tugas**
- [ ] Buat komponen `Login.svelte`: form email + password, memanggil `POST /api/auth/login`.
- [ ] Buat komponen `Register.svelte`: form username + email + password, memanggil `POST /api/auth/register`.
- [ ] Setelah login berhasil, simpan token JWT di `localStorage` (key: `authToken`), lalu redirect ke `/dashboard`.
- [ ] Tampilkan pesan error dari API (contoh: "email sudah terdaftar") di bawah form jika request gagal.
- [ ] Buat helper `src/lib/api.js` yang otomatis menambahkan header `Authorization: Bearer <token>` dari `localStorage` di setiap request ke endpoint yang butuh autentikasi.

**Kriteria Selesai**
- User baru bisa register, lalu login, dan token tersimpan di localStorage (bisa dicek lewat DevTools browser).

---

### Issue #10: Dashboard — Manajemen Tautan & Tema

**Depends on:** Issue #9, Issue #5

**Deskripsi**
Halaman utama setelah login, tempat user mengatur halaman publik mereka.

**Tugas**
- [ ] Saat halaman dashboard dibuka, panggil `GET /api/me` dan tampilkan data yang ada (redirect ke `/login` jika token tidak ada/401).
- [ ] Buat form pengaturan tema: input warna latar (`bg_color`), dropdown pilihan font (`font_family`), upload/URL avatar. Simpan lewat `PUT /api/me/settings`.
- [ ] Buat daftar tautan yang bisa ditambah (form title + url → `POST /api/me/links`), diedit (`PUT /api/me/links/{id}`), dan dihapus (`DELETE /api/me/links/{id}`).
- [ ] Setelah setiap aksi berhasil, refresh daftar dari state lokal (tidak perlu reload seluruh halaman).
- [ ] Tampilkan link ke halaman publik user sendiri (contoh: `Lihat halaman publik saya: /{username}`).

**Kriteria Selesai**
- User bisa login, ubah warna tema, tambah 2-3 tautan, edit salah satu tautan, dan hapus salah satu tautan, semua tanpa reload manual dan tanpa error di console.

---

### Issue #11: Halaman Publik (Public Page Renderer)

**Depends on:** Issue #8, Issue #7

**Deskripsi**
Halaman yang dilihat pengunjung umum ketika membuka `domain.com/{username}`.

**Tugas**
- [x] Di route wildcard `/*` (dari Issue #8), ambil segmen path sebagai `username`.
- [x] Panggil `GET /api/public/{username}`. Jika 404, tampilkan halaman sederhana "Halaman tidak ditemukan".
- [x] Render halaman publik: background sesuai `bg_color`, font sesuai `font_family`, avatar (jika ada), dan daftar tombol tautan yang bisa diklik (`title` sebagai teks, `url` sebagai tujuan, buka di tab baru).
- [x] Pastikan halaman ini tetap ringan — tidak memuat kode dashboard/auth yang tidak perlu (boleh lazy-load komponen dashboard secara terpisah jika diperlukan).

**Kriteria Selesai**
- Membuka `http://localhost:5173/budi123` (username yang sudah dibuat sebelumnya) menampilkan tema dan daftar tautan yang benar sesuai yang diatur di dashboard.

---

## Milestone 4 — Reverse Proxy & Deployment

### Issue #12: Setup Reverse Proxy (Nginx)

**Depends on:** Issue #7, Issue #11

**Deskripsi**
Gunakan **Nginx** (lebih sederhana untuk setup awal dibanding Traefik, cukup untuk kebutuhan MVP ini).

**Tugas**
- [ ] Buat `proxy/nginx.conf` yang meneruskan path `/api/*` ke service `backend:8080`, dan path lainnya ke service `frontend` (port sesuai static server di Issue #8).
- [ ] Tambahkan service `proxy` di `compose.yaml` yang memakai image `nginx:alpine`, mount `nginx.conf`, dan expose port `80` ke host.
- [ ] (Opsional untuk tahap ini, boleh dilewati dulu) Catat di README bahwa SSL/HTTPS otomatis (misalnya lewat Certbot atau ganti ke Traefik) adalah langkah lanjutan untuk production, bukan bagian dari MVP ini.

**Kriteria Selesai**
- Membuka `http://localhost/budi123` (lewat proxy, port 80) menampilkan halaman publik yang sama seperti Issue #11, dan `http://localhost/api/healthz` mengembalikan response dari backend.

---

### Issue #13: Integrasi Penuh & Uji End-to-End

**Depends on:** Semua issue di atas

**Deskripsi**
Pastikan seluruh sistem berjalan bersamaan dari nol.

**Tugas**
- [ ] Jalankan `podman compose up --build` dari kondisi bersih (volume dan container dihapus dulu) dan pastikan ke-5 service (`frontend`, `backend`, `postgres`, `redis`, `proxy`) start tanpa error.
- [ ] Jalankan migration database (Issue #2) di environment yang baru ini.
- [ ] Uji alur penuh secara manual: register → login → atur tema & tautan di dashboard → buka halaman publik di tab baru/incognito → pastikan semua tautan & tema tampil benar.
- [ ] Uji ulang halaman publik setelah mengubah satu tautan di dashboard, pastikan perubahan muncul di halaman publik (mengonfirmasi invalidasi cache Redis di Issue #5/#6 bekerja).
- [ ] Tulis singkat di `README.md` root: cara menjalankan project (`cp .env.example .env`, lalu `podman compose up --build`) dan cara menjalankan migration.

**Kriteria Selesai**
- Proses di atas berhasil semua tanpa perlu edit manual kode di tengah jalan.

---

## Catatan: Hal yang Sengaja Tidak Dikerjakan Dulu (Bukan Bug)

Ini di luar cakupan MVP, jangan dikerjakan kecuali diminta secara terpisah:
- gRPC (cukup REST dulu; gRPC baru relevan kalau sudah ada klien mobile Flutter).
- Analitik kunjungan halaman.
- Multi-tema/template yang bisa dipilih (cukup 1 layout dengan warna & font yang bisa diubah).
- Rate limiting, email verifikasi, reset password.
- Setup SSL/HTTPS otomatis (Certbot/Let's Encrypt) — dicatat sebagai langkah production, bukan bagian issue ini.