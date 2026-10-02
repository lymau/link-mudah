# Pedoman AI Assistant - Link Mudah

Dokumen ini berisi panduan untuk AI Assistant (Antigravity IDE, Cursor, Windsurf, Copilot, Claude, dll.) saat mengembangkan atau memodifikasi kode pada repositori `link-mudah`.

---

## 1. Arsitektur Proyek
- **Backend (`/backend`)**: Golang dengan Gin framework, SQLite, JWT auth, dan middleware rate limiting/security.
- **Frontend (`/frontend`)**: Svelte 5 (Vite), Tailwind CSS v4, `shadcn-svelte`, `lucide-svelte` / `@lucide/svelte`, dan `svelte-spa-router`.
- **Proxy (`/proxy`)**: Caddy reverse proxy untuk routing API dan frontend.

---

## 2. Standar Pengembangan Frontend (shadcn-svelte & Svelte 5)

Seluruh komponen UI frontend wajib mengikuti standar **shadcn-svelte** dan **Svelte 5 runes**:

### A. Konfigurasi Komponen (`frontend/components.json`)
- Path aliases:
  - `ui`: `$lib/components/ui`
  - `components`: `$lib/components`
  - `utils`: `$lib/utils`
  - `lib`: `$lib`
- Proyek menggunakan **JavaScript (non-TypeScript)** (`"typescript": false`).
- Tema Tailwind berbasis `slate` (`src/app.css`).

### B. Aturan Import shadcn-svelte
- **Single-Component Barrels**: Gunakan named import langsung:
  ```javascript
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Badge } from "$lib/components/ui/badge";
  ```
  *(Dilarang menggunakan `import * as Button from ...`)*
- **Multi-Part Components**: Gunakan namespace atau named import sesuai kebutuhan:
  ```javascript
  import * as Card from "$lib/components/ui/card";
  // Atau:
  import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "$lib/components/ui/card";
  ```

### C. Konvensi Svelte 5 Runes
- Props wajib menggunakan rune `$props()` (bukan sintaks Svelte 4 `export let`):
  ```javascript
  let { title, description, children } = $props();
  ```
- State reaktif menggunakan `$state()` dan `$derived()` (bukan sintaks `let x = ...` atau `$: y = ...`).
- Event listeners menggunakan atribut Svelte 5: `onclick`, `onsubmit`, `onchange` (bukan `on:click`, `on:submit`).
- Content projection / slots menggunakan snippet `children` dan `{@render children?.()}` (bukan `<slot />`).

### D. Styling & Tailwind CSS
- Gunakan warna semantik design token (`bg-primary`, `text-primary-foreground`, `text-muted-foreground`, `border-border`, dll.). Hindari warna absolut seperti `bg-blue-500`.
- Gunakan `flex gap-*` untuk jarak antar elemen. Hindari utility class legacy `space-x-*` atau `space-y-*`.
- Gunakan `size-*` jika lebar dan tinggi bernilai sama.
- Gabungkan class dinamis dengan fungsi helper `cn(...)` dari `$lib/utils.js`.

---

## 3. Skills dan Aturan Tambahan
Panduan detail dan referensi shadcn-svelte terpasang di:
- Skill: `.agents/skills/shadcn-svelte/SKILL.md`
- Aturan: `.agents/rules/shadcn-svelte.md`
