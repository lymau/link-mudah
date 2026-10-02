---
title: "Pemasangan shadcn-svelte Skills untuk AI Assistant"
labels: ["enhancement", "tooling", "ai"]
assignees: []
---

## Deskripsi Singkat
Untuk meningkatkan kualitas dan konsistensi kode yang dihasilkan oleh AI Assistant (seperti Cursor, Windsurf, Copilot, atau Claude), kita perlu memasang **shadcn-svelte skills**. Skill ini akan memberikan konteks dan best practices (seperti penggunaan Svelte 5 runes, Tailwind CSS, dan bits-ui) kepada AI saat menulis komponen UI shadcn-svelte.

## Referensi
- Dokumentasi Resmi: [shadcn-svelte Skills](https://www.shadcn-svelte.com/docs/skills)

## Kriteria Penerimaan (Acceptance Criteria)
- [x] Perintah pemasangan `skills` telah dijalankan di direktori `frontend`.
- [x] File aturan AI untuk Antigravity IDE (seperti `AGENTS.md` atau direktori `.agents/`) berhasil dibuat dan masuk ke dalam version control (Git).
- [x] AI Assistant dapat membaca file konfigurasi `components.json` dan menghasilkan komponen shadcn-svelte sesuai dengan standar (tidak ada error deprecated components).
- [x] Perubahan sudah di-commit dan di-push ke repository.

## Langkah-langkah Implementasi

1. **Masuk ke direktori frontend**:
   Pastikan Anda berada di direktori `frontend` yang memiliki file `components.json`.
   ```bash
   cd frontend
   ```

2. **Jalankan Command Instalasi**:
   Sesuai dokumentasi shadcn-svelte, gunakan `npx` (karena project menggunakan `npm` berdasarkan `package-lock.json`) untuk memasang skill:
   ```bash
   npx skills add huntabyte/shadcn-svelte
   ```

3. **Verifikasi File**:
   Pastikan instruksi untuk AI berhasil ditambahkan. Jika CLI `skills` belum secara otomatis mendukung Antigravity, ubah nama file outputnya (misalnya dari `.cursorrules`) menjadi `.agents/rules/shadcn-svelte.md` atau `AGENTS.md` agar Antigravity dapat memuat aturan tersebut.

4. **Testing AI**:
   Uji coba dengan meminta agen Antigravity untuk membuat komponen form sederhana dengan shadcn-svelte. Pastikan hasil kodenya valid, mengikuti aturan Svelte 5, dan agen mengutip aturan dari instruksi yang terpasang.

## Catatan Penting
Sesuai dengan prinsip project, proses setup instruksi AI ini **hanya perlu dilakukan sekali** oleh implementer issue ini. Setelah file `AGENTS.md` (atau folder `.agents/`) di-commit dan di-push ke Git, developer lain cukup menjalankan `podman compose up` seperti biasa. Mereka otomatis akan mendapatkan manfaat dari skills AI ini tanpa perlu melakukan instalasi command tambahan atau memikirkan setup yang ribet.
