# shadcn-svelte Rules for Link Mudah

Pedoman dan aturan pengembangan UI menggunakan shadcn-svelte, Svelte 5, dan Tailwind CSS pada proyek `link-mudah`.

## Konteks Proyek Frontend
- **Konfigurasi**: `frontend/components.json`
- **Aliases**:
  - `ui`: `$lib/components/ui`
  - `components`: `$lib/components`
  - `utils`: `$lib/utils`
  - `hooks`: `$lib/hooks`
  - `lib`: `$lib`
- **TypeScript**: `false` (menggunakan JavaScript murni dengan sintaks Svelte 5)
- **Styling**: Tailwind CSS v4 (`src/app.css`, tema warna dasar `slate`)
- **Icon Library**: `lucide` (diimpor dari `@lucide/svelte` atau `lucide-svelte`)

---

## 1. Aturan Import Komponen
Setiap komponen shadcn-svelte berada di direktori `$lib/components/ui/<nama-komponen>` dengan file `index.js`.

- **Single-Component Barrels**: Gunakan named import langsung.
  ```javascript
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Badge } from "$lib/components/ui/badge";
  import { Separator } from "$lib/components/ui/separator";
  import { Label } from "$lib/components/ui/label";
  ```
  *(JANGAN gunakan `import * as Button from ...` lalu `<Button.Root>`)*

- **Multi-Part Components**: Gunakan namespace import atau named barrel exports.
  ```javascript
  // Namespace import:
  import * as Card from "$lib/components/ui/card";
  import * as Dialog from "$lib/components/ui/dialog";
  import * as Alert from "$lib/components/ui/alert";

  // Atau named imports:
  import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "$lib/components/ui/card";
  import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "$lib/components/ui/dialog";
  ```

---

## 2. Aturan Svelte 5 (Runes)
Proyek ini menggunakan **Svelte 5**. Selalu terapkan pola modern Svelte 5:
- **Props**: Gunakan rune `$props()`.
  ```javascript
  let { class: className, variant = "default", children, ...restProps } = $props();
  ```
  *(JANGAN gunakan sintaks Svelte 4 seperti `export let variant = 'default'`)*
- **State**: Gunakan `$state()` untuk variabel reaktif lokal.
- **Derived State**: Gunakan `$derived()` untuk komputasi reaktif (menggantikan `$: derived = ...`).
- **Two-way Binding Props**: Gunakan `$bindable()`.
- **Event Handlers**: Gunakan atribut standar Svelte 5 seperti `onclick`, `onsubmit`, `oninput`, `onkeydown` (bukan `on:click`, `on:submit`).
- **Children / Slots**: Gunakan snippet `children` dan `{@render children?.()}` (bukan `<slot />`).

---

## 3. Aturan Styling & Tailwind CSS
- **Semantic Tokens**: Selalu gunakan warna semantik Tailwind CSS (misal `bg-primary`, `text-primary-foreground`, `text-muted-foreground`, `border-border`, `bg-background`, `bg-muted`). Hindari hardcoded colors seperti `bg-blue-500` atau `text-gray-700`.
- **Spacing**: Gunakan `flex` dengan `gap-*` (misal `flex flex-col gap-4`), **HINDARI** `space-x-*` atau `space-y-*`.
- **Equal Dimensions**: Gunakan `size-*` jika width dan height sama (misal `size-10` bukan `w-10 h-10`).
- **Truncate**: Gunakan class shorthand `truncate`.
- **Conditional Classes**: Gunakan helper `cn(...)` dari `$lib/utils.js`.
- **No manual z-index on overlays**: Dialog, Sheet, dan Popover menangani z-index secara otomatis.

---

## 4. Aturan Form & Aksesibilitas
- Gunakan `<Label for="id">` yang terhubung secara eksplisit dengan elemen input `<Input id="id">`.
- Tampilkan error / feedback menggunakan `<Alert>` dan `<AlertDescription>`.
- State validasi menggunakan `data-invalid` dan atribut aksesibilitas `aria-invalid`.
- Tombol loading: Gunakan icon spinner (seperti `<Loader2 class="animate-spin" />`) di dalam tombol dan set atribut `disabled`. Tombol tidak memiliki properti `isLoading` bawaan.

---

## 5. Referensi Skill Lengkap
Dokumentasi lengkap dan contoh best practices shadcn-svelte tersedia di:
- `.agents/skills/shadcn-svelte/SKILL.md`
- `.agents/skills/shadcn-svelte/rules/forms.md`
- `.agents/skills/shadcn-svelte/rules/styling.md`
- `.agents/skills/shadcn-svelte/rules/composition.md`
- `.agents/skills/shadcn-svelte/rules/icons.md`
