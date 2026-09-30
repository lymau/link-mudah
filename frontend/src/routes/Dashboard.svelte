<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Badge } from '$lib/components/ui/badge';
  import { Separator } from '$lib/components/ui/separator';
  import { Alert, AlertDescription } from '$lib/components/ui/alert';
  import {
    Link2,
    Plus,
    Trash2,
    Edit3,
    ExternalLink,
    LogOut,
    Palette,
    Check,
    Loader2,
    AlertCircle,
    User,
    Globe,
    Smartphone,
    Share2,
    CheckCircle2,
    Eye
  } from 'lucide-svelte';
  import { apiFetch, clearAuthToken, getAuthToken } from '$lib/api';
  import { getAccessibleTheme } from '$lib/contrast';

  let user = $state({
    username: '',
    email: '',
    page_settings: {
      bg_color: '#f8fafc',
      font_family: 'Inter',
      avatar_url: '',
      title: '',
      description: ''
    },
    links: []
  });

  let isLoading = $state(true);
  let isSavingSettings = $state(false);
  let isSavingLink = $state(false);
  let errorMessage = $state('');
  let successMessage = $state('');
  let copied = $state(false);

  // Link Form State
  let showLinkForm = $state(false);
  let editingLinkId = $state(null);
  let linkTitle = $state('');
  let linkUrl = $state('');

  // Settings Form State
  let settingsBgColor = $state('#f8fafc');
  let settingsFontFamily = $state('Inter');
  let settingsAvatarUrl = $state('');
  let settingsTitle = $state('');
  let settingsDescription = $state('');

  // Reactive accessible theme calculation based on WCAG 2.1 contrast ratio
  let previewTheme = $derived(getAccessibleTheme(settingsBgColor));

  const bgPresets = [
    { label: 'Putih Bersih', value: '#ffffff' },
    { label: 'Slate Soft', value: '#f1f5f9' },
    { label: 'Sky Blue', value: '#e0f2fe' },
    { label: 'Emerald Mint', value: '#d1fae5' },
    { label: 'Rose Pink', value: '#ffe4e6' },
    { label: 'Dark Slate', value: '#0f172a' },
  ];

  async function loadUserData(silent = false) {
    const token = getAuthToken();
    if (!token) {
      push('/login');
      return;
    }

    if (!silent) {
      isLoading = true;
    }
    errorMessage = '';
    try {
      const data = await apiFetch('/api/me');
      user = {
        username: data.username || '',
        email: data.email || '',
        page_settings: {
          bg_color: data.page_settings?.bg_color || '#f8fafc',
          font_family: data.page_settings?.font_family || 'Inter',
          avatar_url: data.page_settings?.avatar_url || '',
          title: data.page_settings?.title || '',
          description: data.page_settings?.description || ''
        },
        links: data.links || []
      };

      settingsBgColor = user.page_settings.bg_color;
      settingsFontFamily = user.page_settings.font_family;
      settingsAvatarUrl = user.page_settings.avatar_url;
      settingsTitle = user.page_settings.title;
      settingsDescription = user.page_settings.description;
    } catch (err) {
      if (
        err.status === 401 ||
        err.message?.toLowerCase().includes('unauthorized') ||
        err.message?.toLowerCase().includes('token') ||
        err.message?.toLowerCase().includes('auth')
      ) {
        clearAuthToken();
        push('/login');
        return;
      }
      errorMessage = err.message || 'Gagal memuat profil.';
    } finally {
      if (!silent) {
        isLoading = false;
      }
    }
  }

  function handleLogout() {
    clearAuthToken();
    push('/login');
  }

  function openAddLinkForm() {
    editingLinkId = null;
    linkTitle = '';
    linkUrl = '';
    showLinkForm = true;
  }

  function openEditLinkForm(link) {
    editingLinkId = link.id;
    linkTitle = link.title;
    linkUrl = link.url;
    showLinkForm = true;
  }

  function cancelLinkForm() {
    editingLinkId = null;
    linkTitle = '';
    linkUrl = '';
    showLinkForm = false;
  }

  async function handleSaveLink(event) {
    if (event && typeof event.preventDefault === 'function') {
      event.preventDefault();
    }
    if (!linkTitle.trim() || !linkUrl.trim()) {
      errorMessage = 'Judul dan URL tautan tidak boleh kosong.';
      return;
    }

    isSavingLink = true;
    errorMessage = '';
    try {
      let formattedUrl = linkUrl.trim();
      if (
        !/^(https?|mailto|tel):\/\//i.test(formattedUrl) &&
        !formattedUrl.startsWith('mailto:') &&
        !formattedUrl.startsWith('tel:')
      ) {
        formattedUrl = 'https://' + formattedUrl;
      }

      if (editingLinkId) {
        await apiFetch(`/api/me/links/${editingLinkId}`, {
          method: 'PUT',
          body: JSON.stringify({
            title: linkTitle.trim(),
            url: formattedUrl
          })
        });
      } else {
        await apiFetch('/api/me/links', {
          method: 'POST',
          body: JSON.stringify({
            title: linkTitle.trim(),
            url: formattedUrl
          })
        });
      }

      await loadUserData(true);
      cancelLinkForm();
      showSuccessFeedback(editingLinkId ? 'Tautan berhasil diperbarui!' : 'Tautan berhasil disimpan!');
    } catch (err) {
      if (err.status === 401) {
        clearAuthToken();
        push('/login');
        return;
      }
      errorMessage = err.message || 'Gagal menyimpan tautan.';
    } finally {
      isSavingLink = false;
    }
  }

  async function handleDeleteLink(id) {
    if (!confirm('Yakin ingin menghapus tautan ini?')) return;

    errorMessage = '';
    try {
      await apiFetch(`/api/me/links/${id}`, {
        method: 'DELETE'
      });
      await loadUserData(true);
      showSuccessFeedback('Tautan berhasil dihapus.');
    } catch (err) {
      if (err.status === 401) {
        clearAuthToken();
        push('/login');
        return;
      }
      errorMessage = err.message || 'Gagal menghapus tautan.';
    }
  }

  async function handleSaveSettings(event) {
    if (event && typeof event.preventDefault === 'function') {
      event.preventDefault();
    }

    isSavingSettings = true;
    errorMessage = '';
    try {
      const res = await apiFetch('/api/me/settings', {
        method: 'PUT',
        body: JSON.stringify({
          bg_color: settingsBgColor,
          font_family: settingsFontFamily,
          avatar_url: settingsAvatarUrl,
          title: settingsTitle,
          description: settingsDescription
        })
      });

      // Update state directly from API response to ensure synchronization
      user.page_settings.bg_color = res?.bg_color || settingsBgColor;
      user.page_settings.font_family = res?.font_family || settingsFontFamily;
      user.page_settings.avatar_url = res?.avatar_url || settingsAvatarUrl;
      user.page_settings.title = res?.title || settingsTitle;
      user.page_settings.description = res?.description || settingsDescription;

      showSuccessFeedback('Pengaturan tampilan berhasil diperbarui!');
    } catch (err) {
      if (err.status === 401) {
        clearAuthToken();
        push('/login');
        return;
      }
      errorMessage = err.message || 'Gagal menyimpan pengaturan.';
    } finally {
      isSavingSettings = false;
    }
  }

  function showSuccessFeedback(msg) {
    successMessage = msg;
    setTimeout(() => {
      successMessage = '';
    }, 3000);
  }

  function copyPublicUrl() {
    const publicUrl = `${window.location.origin}/#/${user.username}`;
    navigator.clipboard.writeText(publicUrl);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  }

  onMount(() => {
    loadUserData();
  });
</script>

<div class="min-h-screen bg-muted/30">
  <!-- Top Navigation Bar -->
  <header class="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
    <div class="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
      <div class="flex items-center gap-3">
        <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
          <Link2 class="h-5 w-5" />
        </div>
        <div class="flex items-center gap-2">
          <span class="text-lg font-bold tracking-tight text-foreground">Link Mudah</span>
          <Badge variant="secondary" class="font-normal text-xs">Dashboard</Badge>
        </div>
      </div>

      <div class="flex items-center gap-2 sm:gap-4">
        {#if user.username}
          <Button
            variant="outline"
            size="sm"
            class="hidden sm:inline-flex gap-1.5"
            onclick={copyPublicUrl}
          >
            {#if copied}
              <Check class="h-3.5 w-3.5 text-emerald-600" />
              Tersalin!
            {:else}
              <Share2 class="h-3.5 w-3.5" />
              Bagikan
            {/if}
          </Button>

          <Button
            variant="ghost"
            size="sm"
            class="gap-1.5"
            href={`/#/${user.username}`}
            target="_blank"
          >
            <ExternalLink class="h-3.5 w-3.5" />
            <span class="hidden sm:inline">Halaman Publik</span>
          </Button>
        {/if}

        <Separator orientation="vertical" class="h-6 hidden sm:block" />

        <div class="flex items-center gap-2">
          <span class="text-sm font-medium text-muted-foreground hidden md:inline">
            {user.email || user.username}
          </span>
          <Button variant="ghost" size="icon" onclick={handleLogout} title="Keluar">
            <LogOut class="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  </header>

  <!-- Main Content Layout -->
  <main class="container mx-auto max-w-7xl px-4 py-8 sm:px-6">
    {#if isLoading}
      <div class="flex flex-col items-center justify-center py-24 gap-3 text-muted-foreground">
        <Loader2 class="h-8 w-8 animate-spin text-primary" />
        <p class="text-sm">Memuat data akun Anda...</p>
      </div>
    {:else}
      <!-- Notifications -->
      {#if errorMessage}
        <Alert variant="destructive" class="mb-6">
          <AlertCircle class="h-4 w-4" />
          <AlertDescription>{errorMessage}</AlertDescription>
        </Alert>
      {/if}

      {#if successMessage}
        <Alert class="mb-6 border-emerald-500/30 bg-emerald-50 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300">
          <CheckCircle2 class="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
          <AlertDescription>{successMessage}</AlertDescription>
        </Alert>
      {/if}

      <!-- Header Banner -->
      <div class="mb-8 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
            Kelola Tautan
          </h1>
          <div class="flex flex-wrap items-center gap-2 mt-1.5 text-sm text-muted-foreground">
            <span>Lihat halaman publik saya:</span>
            {#if user.username}
              <a
                href={`/#/${user.username}`}
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-1 font-mono font-semibold text-primary hover:underline bg-primary/10 px-2 py-0.5 rounded-md text-xs sm:text-sm"
                title="Buka halaman publik di tab baru"
              >
                /{user.username}
                <ExternalLink class="h-3 w-3" />
              </a>
              <button
                type="button"
                onclick={copyPublicUrl}
                class="text-xs text-muted-foreground hover:text-foreground underline underline-offset-2 ml-1 cursor-pointer"
              >
                {copied ? '✓ Tersalin' : 'Salin URL'}
              </button>
            {/if}
          </div>
        </div>

        <Button onclick={openAddLinkForm} class="gap-2 self-start sm:self-auto">
          <Plus class="h-4 w-4" />
          Tambah Tautan Baru
        </Button>
      </div>

      <!-- Two-column grid -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        <!-- Left: Links Manager (7 cols) -->
        <div class="lg:col-span-7 space-y-6">
          <!-- Add / Edit Form Modal Card -->
          {#if showLinkForm}
            <Card class="border-primary/40 shadow-md">
              <CardHeader>
                <CardTitle class="text-lg font-semibold">
                  {editingLinkId ? 'Edit Tautan' : 'Tambah Tautan Baru'}
                </CardTitle>
                <CardDescription>
                  Tautan ini akan langsung muncul di halaman publik Anda.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onsubmit={handleSaveLink} class="space-y-4">
                  <div class="space-y-2">
                    <Label for="link-title">Judul Tautan</Label>
                    <Input
                      id="link-title"
                      bind:value={linkTitle}
                      placeholder="contoh: Instagram Pribadi, Portofolio, WhatsApp"
                      required
                    />
                  </div>

                  <div class="space-y-2">
                    <Label for="link-url">URL / Tautan Tujuan</Label>
                    <Input
                      id="link-url"
                      bind:value={linkUrl}
                      placeholder="https://instagram.com/username"
                      required
                    />
                  </div>

                  <div class="flex items-center justify-end gap-2 pt-2">
                    <Button type="button" variant="outline" onclick={cancelLinkForm}>
                      Batal
                    </Button>
                    <Button type="submit" disabled={isSavingLink}>
                      {#if isSavingLink}
                        <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                        Menyimpan...
                      {:else}
                        Simpan Tautan
                      {/if}
                    </Button>
                  </div>
                </form>
              </CardContent>
            </Card>
          {/if}

          <!-- List of Links -->
          <Card>
            <CardHeader class="flex flex-row items-center justify-between pb-3">
              <div>
                <CardTitle class="text-lg font-semibold">Daftar Tautan</CardTitle>
                <CardDescription>Total {user.links.length} tautan aktif</CardDescription>
              </div>
            </CardHeader>
            <CardContent>
              {#if user.links.length === 0}
                <div class="flex flex-col items-center justify-center py-12 text-center">
                  <div class="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground mb-3">
                    <Link2 class="h-6 w-6" />
                  </div>
                  <h3 class="font-medium text-foreground">Belum ada tautan</h3>
                  <p class="text-sm text-muted-foreground max-w-xs mt-1 mb-4">
                    Mulai tambahkan tautan media sosial, portofolio, atau kontak bisnis Anda.
                  </p>
                  <Button variant="outline" size="sm" onclick={openAddLinkForm} class="gap-1.5">
                    <Plus class="h-4 w-4" />
                    Tambah Tautan Pertama
                  </Button>
                </div>
              {:else}
                <div class="space-y-3">
                  {#each user.links as link, i}
                    <div class="flex items-center justify-between p-3.5 rounded-lg border bg-card hover:bg-muted/40 transition-colors shadow-2xs">
                      <div class="min-w-0 flex-1 pr-4">
                        <div class="flex items-center gap-2">
                          <span class="flex h-5 w-5 items-center justify-center rounded-full bg-primary/10 text-[11px] font-semibold text-primary">
                            {i + 1}
                          </span>
                          <h4 class="font-medium text-foreground truncate">{link.title}</h4>
                        </div>
                        <a
                          href={link.url}
                          target="_blank"
                          rel="noreferrer"
                          class="text-xs text-muted-foreground hover:text-primary truncate block mt-1 underline-offset-2 hover:underline"
                        >
                          {link.url}
                        </a>
                      </div>

                      <div class="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          class="h-8 w-8 text-muted-foreground hover:text-foreground"
                          onclick={() => openEditLinkForm(link)}
                          title="Edit Tautan"
                        >
                          <Edit3 class="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          class="h-8 w-8 text-destructive/80 hover:text-destructive hover:bg-destructive/10"
                          onclick={() => handleDeleteLink(link.id)}
                          title="Hapus Tautan"
                        >
                          <Trash2 class="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            </CardContent>
          </Card>
        </div>

        <!-- Right: Settings & Live Preview (5 cols) -->
        <div class="lg:col-span-5 space-y-6">
          <!-- Page Appearance Settings -->
          <Card>
            <CardHeader>
              <div class="flex items-center gap-2">
                <Palette class="h-5 w-5 text-primary" />
                <CardTitle class="text-lg font-semibold">Pengaturan Halaman Publik</CardTitle>
              </div>
              <CardDescription>
                Sesuaikan judul, deskripsi bio, warna latar, dan avatar profil publik Anda.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onsubmit={handleSaveSettings} class="space-y-4">
                <!-- Pengaturan Judul Halaman Public -->
                <div class="space-y-2">
                  <Label for="page-title">Judul Halaman Publik</Label>
                  <Input
                    id="page-title"
                    bind:value={settingsTitle}
                    placeholder={`contoh: ${user.username ? user.username : 'Nama Lengkap atau Brand'}`}
                  />
                  <p class="text-xs text-muted-foreground">
                    Judul utama yang tampil di bawah avatar. Jika kosong, sistem otomatis memakai <code>@{user.username || 'username'}</code>.
                  </p>
                </div>

                <!-- Pengaturan Deskripsi (di bawah avatar) -->
                <div class="space-y-2">
                  <Label for="page-description">Deskripsi Profil (di bawah avatar)</Label>
                  <textarea
                    id="page-description"
                    bind:value={settingsDescription}
                    rows="2"
                    class="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm shadow-xs placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                    placeholder="contoh: Creator, Software Engineer & Tech Enthusiast"
                  ></textarea>
                  <p class="text-xs text-muted-foreground">
                    Deskripsi singkat yang tampil tepat di bawah avatar dan judul pada halaman publik.
                  </p>
                </div>

                <div class="space-y-2">
                  <Label for="avatar-url">URL Avatar / Foto Profil</Label>
                  <Input
                    id="avatar-url"
                    bind:value={settingsAvatarUrl}
                    placeholder="https://example.com/avatar.jpg"
                  />
                  <p class="text-xs text-muted-foreground">
                    Gunakan tautan gambar langsung (JPG/PNG).
                  </p>
                </div>

                <div class="space-y-2">
                  <Label>Warna Latar Belakang</Label>
                  <div class="flex flex-wrap gap-2 pt-1">
                    {#each bgPresets as preset}
                      <button
                        type="button"
                        class="h-7 w-7 rounded-full border shadow-2xs transition-transform hover:scale-110 flex items-center justify-center cursor-pointer"
                        style={`background-color: ${preset.value};`}
                        onclick={() => settingsBgColor = preset.value}
                        title={preset.label}
                      >
                        {#if settingsBgColor.toLowerCase() === preset.value.toLowerCase()}
                          <Check class={`h-3.5 w-3.5 ${preset.value === '#0f172a' ? 'text-white' : 'text-slate-900'}`} />
                        {/if}
                      </button>
                    {/each}
                  </div>
                  <div class="flex items-center gap-2 pt-1">
                    <Input
                      type="text"
                      bind:value={settingsBgColor}
                      class="font-mono text-xs"
                      placeholder="#f8fafc"
                    />
                    <input
                      type="color"
                      bind:value={settingsBgColor}
                      class="h-9 w-9 rounded-md border border-input cursor-pointer p-0.5 bg-transparent"
                    />
                  </div>

                  <!-- Indikator Rasio Kontras (Readability & Legibility WCAG 2.1) -->
                  <div
                    class="mt-2.5 p-3 rounded-lg border text-xs space-y-1.5 transition-colors"
                    style={`background-color: ${previewTheme.isDark ? '#0f172a' : '#f8fafc'}; border-color: ${previewTheme.isDark ? '#334155' : '#e2e8f0'}; color: ${previewTheme.isDark ? '#f8fafc' : '#0f172a'};`}
                  >
                    <div class="flex items-center justify-between">
                      <span class="font-semibold flex items-center gap-1.5">
                        <Eye class="h-3.5 w-3.5 text-primary" />
                        Rasio Kontras (WCAG 2.1):
                      </span>
                      <span
                        class="font-bold px-2 py-0.5 rounded text-[11px]"
                        style={`background-color: ${previewTheme.isDark ? '#334155' : '#e2e8f0'}; color: ${previewTheme.isDark ? '#38bdf8' : '#0284c7'};`}
                      >
                        {previewTheme.contrastRatio}:1 ({previewTheme.wcagLevel})
                      </span>
                    </div>
                    <p class="text-[11px] leading-relaxed" style={`color: ${previewTheme.isDark ? '#cbd5e1' : '#64748b'};`}>
                      {#if previewTheme.isDark}
                        Latar gelap terdeteksi: Teks otomatis beralih ke putih/terang dengan kontras tinggi untuk menjamin keterbacaan (readability & legibility) maksimal.
                      {:else}
                        Latar terang terdeteksi: Teks otomatis beralih ke warna gelap dengan kontras tajam sesuai standar WCAG.
                      {/if}
                    </p>
                  </div>
                </div>

                <div class="space-y-2">
                  <Label for="font-family">Jenis Huruf (Font)</Label>
                  <select
                    id="font-family"
                    bind:value={settingsFontFamily}
                    class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  >
                    <option value="Inter">Inter (Modern Clean)</option>
                    <option value="sans-serif">System Sans</option>
                    <option value="serif">Classic Serif</option>
                    <option value="monospace">Developer Mono</option>
                  </select>
                </div>

                <Button type="submit" class="w-full" disabled={isSavingSettings}>
                  {#if isSavingSettings}
                    <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                    Menyimpan...
                  {:else}
                    Simpan Perubahan Tampilan
                  {/if}
                </Button>
              </form>
            </CardContent>
          </Card>

          <!-- Live Preview Phone Frame -->
          <Card>
            <CardHeader class="pb-3">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <Smartphone class="h-4 w-4 text-primary" />
                  <CardTitle class="text-sm font-semibold">Pratinjau Langsung</CardTitle>
                </div>
                <Badge variant="outline" class="text-[11px] font-normal">
                  {previewTheme.isDark ? 'Mode Gelap' : 'Mode Terang'}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div
                class="mx-auto max-w-[280px] rounded-3xl border-4 border-slate-800 p-4 shadow-xl overflow-hidden transition-colors"
                style={`background-color: ${settingsBgColor}; font-family: ${settingsFontFamily}; color: ${previewTheme.textColor};`}
              >
                <!-- Phone top bar -->
                <div
                  class="mx-auto h-3.5 w-20 rounded-full mb-4"
                  style={`background-color: ${previewTheme.isDark ? 'rgba(255,255,255,0.25)' : 'rgba(0,0,0,0.15)'}`}
                ></div>

                <!-- Profile header in phone -->
                <div class="text-center mb-5">
                  {#if settingsAvatarUrl}
                    <img
                      src={settingsAvatarUrl}
                      alt={settingsTitle || user.username}
                      class="h-16 w-16 rounded-full mx-auto mb-2 object-cover shadow-sm border-2"
                      style={`border-color: ${previewTheme.borderColor};`}
                      onerror={(e) => { e.currentTarget.style.display = 'none'; }}
                    />
                  {:else}
                    <div
                      class="h-16 w-16 rounded-full mx-auto mb-2 flex items-center justify-center font-bold text-xl shadow-sm border-2"
                      style={`background-color: ${previewTheme.avatarFallbackBg}; color: ${previewTheme.avatarFallbackText}; border-color: ${previewTheme.borderColor};`}
                    >
                      {user.username ? user.username.slice(0, 1).toUpperCase() : 'U'}
                    </div>
                  {/if}

                  <!-- Judul Halaman Publik di Pratinjau -->
                  <h3
                    class="font-bold text-sm tracking-tight"
                    style={`color: ${previewTheme.textColor};`}
                  >
                    {settingsTitle.trim() || `@${user.username || 'username'}`}
                  </h3>

                  {#if settingsTitle.trim()}
                    <p class="text-[10px] font-mono opacity-80 mt-0.5" style={`color: ${previewTheme.subTextColor};`}>
                      @{user.username || 'username'}
                    </p>
                  {/if}

                  <!-- Deskripsi Profil di Pratinjau -->
                  <p
                    class="text-[11px] mt-1 leading-snug line-clamp-3"
                    style={`color: ${previewTheme.mutedTextColor};`}
                  >
                    {settingsDescription.trim() || 'Link-in-bio resmi'}
                  </p>
                </div>

                <!-- Links preview in phone -->
                <div class="space-y-2">
                  {#if user.links.length === 0}
                    <div
                      class="p-2.5 rounded-lg border border-dashed text-center text-[11px]"
                      style={`border-color: ${previewTheme.cardBorder}; color: ${previewTheme.mutedTextColor};`}
                    >
                      Tautan akan muncul di sini
                    </div>
                  {:else}
                    {#each user.links as link}
                      <div
                        class="p-2.5 rounded-lg border text-center text-xs font-semibold truncate transition-all shadow-xs"
                        style={`background-color: ${previewTheme.cardBg}; color: ${previewTheme.cardTextColor}; border-color: ${previewTheme.cardBorder}; box-shadow: ${previewTheme.cardShadow};`}
                      >
                        {link.title}
                      </div>
                    {/each}
                  {/if}
                </div>

                <div class="mt-6 text-center">
                  <span class="text-[10px]" style={`color: ${previewTheme.subTextColor};`}>Link Mudah</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    {/if}
  </main>
</div>
