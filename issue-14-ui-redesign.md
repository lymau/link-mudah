### Issue #14: Redesign Dashboard, Login, dan Register dengan Preset shadcn-svelte

**Depends on:** Issue sebelumnya terkait Frontend (Setup SvelteKit & Tailwind)

**Deskripsi**
Kita perlu memperbarui tampilan antarmuka (UI) untuk halaman Dashboard, Login, dan Register. Untuk memastikan desain yang modern dan konsisten, kita akan menggunakan `shadcn-svelte` dengan preset yang telah ditentukan yaitu `--preset bLZBHZ4gs`. Pastikan mengikuti *best practices* dalam penulisan komponen dan integrasi UI.

**Tugas**
- [x] Terapkan preset pada proyek dengan menjalankan instalasi komponen `shadcn-svelte` yang dibutuhkan (menggunakan parameter `--preset bLZBHZ4gs` saat inisialisasi atau penambahan komponen).
- [x] **Halaman Login:** Desain ulang form login menggunakan komponen `shadcn-svelte` (seperti `Card`, `Input`, `Label`, dan `Button`).
- [x] **Halaman Register:** Desain ulang form pendaftaran dengan struktur UI yang senada dengan form login.
- [x] **Halaman Dashboard:** Perbarui struktur layout, navigasi sidebar/topbar, dan daftar tautan/pengaturan pada dashboard agar sejalan dengan gaya visual dari preset baru.
- [x] Hapus style kustom, file CSS, atau utility class Tailwind lama yang berpotensi konflik dengan preset baru.
- [x] Pertahankan fungsionalitas aplikasi (validasi form, pemanggilan API, dsb) agar tetap berjalan normal dengan UI baru.

**Kriteria Selesai**
- Preset `bLZBHZ4gs` dari `shadcn-svelte` telah terintegrasi tanpa *error*.
- Halaman Login, Register, dan Dashboard sepenuhnya menggunakan komponen UI baru yang direkomendasikan.
- Layout bersifat responsif (mobile-first) dan memenuhi standar aksesibilitas (a11y).
- *Best practices* pada penulisan komponen (modular, reusable) diterapkan dengan baik.
