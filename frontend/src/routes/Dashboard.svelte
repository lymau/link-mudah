<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle
  } from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Badge } from '$lib/components/ui/badge';
  import { Separator } from '$lib/components/ui/separator';
  import { toast } from '$lib/components/ui/sonner';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
  import { Avatar, AvatarFallback, AvatarImage } from '$lib/components/ui/avatar';
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
    User,
    Globe,
    Smartphone,
    Share2,
    Eye,
    Sparkles,
    Copy,
    ArrowUpRight
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
  let copied = $state(false);

  // Active Tab: 'links' | 'appearance' | 'preview'
  let activeTab = $state('links');

  // Link Form State
  let showLinkForm = $state(false);
  let editingLinkId = $state(null);
  let linkTitle = $state('');
  let linkUrl = $state('');

  // Delete Alert Dialog State
  let isDeleteDialogOpen = $state(false);
  let linkToDelete = $state(null);
  let isDeletingLink = $state(false);

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
      toast.error(err.message || 'Gagal memuat profil.');
    } finally {
      if (!silent) {
        isLoading = false;
      }
    }
  }

  function handleLogout() {
    clearAuthToken();
    toast.info('Berhasil keluar dari akun.');
    push('/login');
  }

  function openAddLinkForm() {
    editingLinkId = null;
    linkTitle = '';
    linkUrl = '';
    showLinkForm = true;
    if (activeTab === 'preview') {
      activeTab = 'links';
    }
  }

  function openEditLinkForm(link) {
    editingLinkId = link.id;
    linkTitle = link.title;
    linkUrl = link.url;
    showLinkForm = true;
    if (activeTab === 'preview') {
      activeTab = 'links';
    }
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
      toast.error('Judul dan URL tautan tidak boleh kosong.');
      return;
    }

    isSavingLink = true;
    try {
      let formattedUrl = linkUrl.trim();
      if (
        !/^(https?|mailto|tel):\/\//i.test(formattedUrl) &&
        !formattedUrl.startsWith('mailto:') &&
        !formattedUrl.startsWith('tel:')
      ) {
        formattedUrl = 'https://' + formattedUrl;
      }

      const isEditing = Boolean(editingLinkId);
      if (isEditing) {
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
      toast.success(isEditing ? 'Tautan berhasil diperbarui!' : 'Tautan baru berhasil disimpan!');
    } catch (err) {
      if (err.status === 401) {
        clearAuthToken();
        push('/login');
        return;
      }
      toast.error(err.message || (editingLinkId ? 'Gagal memperbarui tautan.' : 'Gagal menyimpan tautan.'));
    } finally {
      isSavingLink = false;
    }
  }

  function promptDeleteLink(link) {
    linkToDelete = link;
    isDeleteDialogOpen = true;
  }

  async function handleConfirmDelete() {
    if (!linkToDelete) return;

    isDeletingLink = true;
    const targetTitle = linkToDelete.title;
    try {
      await apiFetch(`/api/me/links/${linkToDelete.id}`, {
        method: 'DELETE'
      });
      await loadUserData(true);
      toast.success(`Tautan "${targetTitle}" berhasil dihapus.`);
      isDeleteDialogOpen = false;
      linkToDelete = null;
    } catch (err) {
      if (err.status === 401) {
        clearAuthToken();
        push('/login');
        return;
      }
      toast.error(err.message || 'Gagal menghapus tautan.');
    } finally {
      isDeletingLink = false;
    }
  }

  async function handleSaveSettings(event) {
    if (event && typeof event.preventDefault === 'function') {
      event.preventDefault();
    }

    isSavingSettings = true;
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

      toast.success('Pengaturan tampilan berhasil diperbarui!');
    } catch (err) {
      if (err.status === 401) {
        clearAuthToken();
        push('/login');
        return;
      }
      toast.error(err.message || 'Gagal menyimpan pengaturan.');
    } finally {
      isSavingSettings = false;
    }
  }

  function copyPublicUrl() {
    const publicUrl = `${window.location.origin}/#/${user.username}`;
    navigator.clipboard.writeText(publicUrl);
    copied = true;
    toast.success('Tautan publik berhasil disalin ke papan klip!');
    setTimeout(() => {
      copied = false;
    }, 2000);
  }

  onMount(() => {
    loadUserData();
  });

  $effect(() => {
    if (typeof window !== 'undefined') {
      const handleResize = () => {
        if (window.innerWidth >= 1024 && activeTab === 'preview') {
          activeTab = 'links';
        }
      };
      window.addEventListener('resize', handleResize);
      return () => window.removeEventListener('resize', handleResize);
    }
  });
</script>

{#snippet phonePreview()}
  <Card class="shadow-md border-border/80 bg-card">
    <CardHeader class="pb-3 flex flex-row items-center justify-between">
      <div class="flex items-center gap-2">
        <Smartphone class="h-4 w-4 text-primary" />
        <CardTitle class="text-sm font-bold">Pratinjau Langsung</CardTitle>
      </div>
      <div class="flex items-center gap-1.5">
        <Badge variant="outline" class="text-[11px] font-normal">
          {previewTheme.isDark ? 'Latar Gelap' : 'Latar Terang'}
        </Badge>
        {#if user.username}
          <Button
            variant="ghost"
            size="icon"
            class="h-7 w-7 text-muted-foreground hover:text-foreground"
            href={`/#/${user.username}`}
            target="_blank"
            title="Buka halaman publik di tab baru"
            aria-label="Buka halaman publik"
          >
            <ExternalLink class="h-3.5 w-3.5" />
          </Button>
        {/if}
      </div>
    </CardHeader>

    <CardContent class="pt-0">
      <!-- Smartphone Mockup Frame -->
      <div
        class="mx-auto max-w-[280px] min-h-[420px] max-h-[520px] rounded-[2.5rem] border-[6px] border-slate-900 p-4 shadow-2xl overflow-y-auto transition-colors relative"
        style={`background-color: ${settingsBgColor}; font-family: ${settingsFontFamily}; color: ${previewTheme.textColor}; scrollbar-width: thin;`}
      >
        <!-- Phone notch / speaker -->
        <div
          class="mx-auto h-3.5 w-20 rounded-full mb-4 transition-colors"
          style={`background-color: ${previewTheme.isDark ? 'rgba(255,255,255,0.2)' : 'rgba(0,0,0,0.15)'}`}
        ></div>

        <!-- Profile Header in Phone -->
        <div class="text-center mb-5">
          {#if settingsAvatarUrl && settingsAvatarUrl.trim()}
            <img
              src={settingsAvatarUrl.trim()}
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
            class="font-bold text-sm tracking-tight break-words"
            style={`color: ${previewTheme.textColor};`}
          >
            {settingsTitle.trim() || `@${user.username || 'username'}`}
          </h3>

          {#if settingsTitle.trim()}
            <p class="text-[10px] font-mono opacity-80 mt-0.5 break-words" style={`color: ${previewTheme.subTextColor};`}>
              @{user.username || 'username'}
            </p>
          {/if}

          <!-- Deskripsi Profil di Pratinjau -->
          <p
            class="text-[11px] mt-1 leading-snug line-clamp-3 break-words"
            style={`color: ${previewTheme.mutedTextColor};`}
          >
            {settingsDescription.trim() || 'Link-in-bio resmi'}
          </p>
        </div>

        <!-- Links in Phone -->
        <div class="space-y-2">
          {#if user.links.length === 0}
            <div
              class="p-2.5 rounded-xl border border-dashed text-center text-[11px]"
              style={`border-color: ${previewTheme.cardBorder}; color: ${previewTheme.mutedTextColor};`}
            >
              Tautan Anda akan muncul di sini
            </div>
          {:else}
            {#each user.links as link}
              <a
                href={link.url}
                target="_blank"
                rel="noreferrer"
                class="block p-2.5 rounded-xl border text-center text-xs font-semibold truncate transition-all shadow-xs hover:opacity-90 cursor-pointer"
                style={`background-color: ${previewTheme.cardBg}; color: ${previewTheme.cardTextColor}; border-color: ${previewTheme.cardBorder}; box-shadow: ${previewTheme.cardShadow};`}
                title={link.title}
              >
                {link.title}
              </a>
            {/each}
          {/if}
        </div>

        <!-- Phone Footer -->
        <div class="mt-6 text-center">
          <span class="text-[10px] font-medium opacity-75" style={`color: ${previewTheme.subTextColor};`}>
            Link Mudah
          </span>
        </div>
      </div>
    </CardContent>
  </Card>
{/snippet}

<div class="min-h-screen bg-muted/20 pb-12">
  <!-- Top Navigation Bar (shadcn style header) -->
  <header class="sticky top-0 z-40 w-full border-b border-border/70 bg-background/85 backdrop-blur-md transition-all">
    <div class="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
      <!-- Left: Logo & Title -->
      <div class="flex items-center gap-3">
        <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm transition-transform hover:scale-105">
          <Link2 class="h-5 w-5" />
        </div>
        <div class="flex items-center gap-2">
          <span class="text-base font-bold tracking-tight text-foreground sm:text-lg">Link Mudah</span>
          <Badge variant="secondary" class="font-medium text-[11px] px-2 py-0.5">Dashboard</Badge>
        </div>
      </div>

      <!-- Right: Action buttons & User profile -->
      <div class="flex items-center gap-2 sm:gap-3">
        {#if user.username}
          <Button
            variant="outline"
            size="sm"
            class="h-8 gap-1.5 text-xs font-medium transition-all shadow-2xs"
            onclick={copyPublicUrl}
            title="Salin tautan halaman publik Anda"
          >
            {#if copied}
              <Check class="h-3.5 w-3.5 text-emerald-600 dark:text-emerald-400" />
              <span class="hidden sm:inline">Tersalin!</span>
            {:else}
              <Share2 class="h-3.5 w-3.5" />
              <span class="hidden sm:inline">Bagikan Link</span>
            {/if}
          </Button>

          <Button
            variant="ghost"
            size="sm"
            class="h-8 gap-1.5 text-xs font-medium"
            href={`/#/${user.username}`}
            target="_blank"
            title="Buka halaman publik di tab baru"
          >
            <ExternalLink class="h-3.5 w-3.5" />
            <span class="hidden md:inline">Halaman Publik</span>
          </Button>
        {/if}

        <Separator orientation="vertical" class="h-5 hidden sm:block" />

        <!-- User profile display -->
        <div class="flex items-center gap-2 pl-1">
          <Avatar class="h-8 w-8 ring-1 ring-border">
            {#if user.page_settings.avatar_url}
              <AvatarImage src={user.page_settings.avatar_url} alt={user.username} />
            {/if}
            <AvatarFallback class="bg-primary/10 text-primary font-bold text-xs">
              {(user.username || 'U').slice(0, 1).toUpperCase()}
            </AvatarFallback>
          </Avatar>

          <span class="text-xs font-medium text-foreground max-w-[120px] truncate hidden lg:inline">
            {user.username || user.email}
          </span>

          <Button
            variant="ghost"
            size="icon"
            class="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
            onclick={handleLogout}
            title="Keluar dari akun"
            aria-label="Logout"
          >
            <LogOut class="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  </header>

  <!-- Main Content Layout -->
  <main class="container mx-auto max-w-7xl px-4 py-6 sm:px-6 sm:py-8">
    {#if isLoading}
      <div class="flex flex-col items-center justify-center py-28 gap-3 text-muted-foreground">
        <Loader2 class="h-8 w-8 animate-spin text-primary" />
        <p class="text-sm font-medium">Memuat data dashboard Anda...</p>
      </div>
    {:else}
      <!-- Header & Welcome Banner -->
      <div class="mb-6 flex flex-col md:flex-row md:items-center md:justify-between gap-4 rounded-3xl border border-border/60 bg-card p-5 sm:p-6 shadow-sm">
        <div>
          <div class="flex items-center gap-2">
            <h1 class="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
              Halo, {user.username || 'Pengguna'}!
            </h1>
            <Badge variant="outline" class="font-medium text-xs">
              {user.links.length} Tautan
            </Badge>
          </div>
          <div class="mt-2 flex flex-wrap items-center gap-2 text-xs sm:text-sm text-muted-foreground">
            <span>Tautan publik aktif:</span>
            {#if user.username}
              <div class="inline-flex items-center gap-1 rounded-lg bg-muted px-2.5 py-1 font-mono text-xs font-semibold text-foreground">
                <span>linkmudah.com/#{user.username}</span>
                <button
                  type="button"
                  onclick={copyPublicUrl}
                  class="ml-1 text-primary hover:text-primary/80 cursor-pointer"
                  title="Salin tautan"
                  aria-label="Salin tautan publik"
                >
                  {#if copied}
                    <Check class="h-3.5 w-3.5 text-emerald-600" />
                  {:else}
                    <Copy class="h-3.5 w-3.5" />
                  {/if}
                </button>
              </div>
            {/if}
          </div>
        </div>

        <div class="flex items-center gap-2 self-start md:self-auto">
          <Button onclick={openAddLinkForm} class="gap-1.5 shadow-sm">
            <Plus class="h-4 w-4" />
            Tambah Tautan
          </Button>
        </div>
      </div>

      <!-- Main Responsive Tabs & 2-Column Grid -->
      <Tabs bind:value={activeTab} class="w-full">
        <!-- Navigation switcher: 2 tabs on desktop, 3 tabs on mobile/tablet (including live phone preview) -->
        <div class="mb-6 flex items-center justify-between">
          <TabsList class="grid grid-cols-3 lg:grid-cols-2 w-full max-w-md">
            <TabsTrigger value="links" class="gap-1.5 text-xs sm:text-sm">
              <Link2 class="h-4 w-4" />
              <span>Tautan ({user.links.length})</span>
            </TabsTrigger>
            <TabsTrigger value="appearance" class="gap-1.5 text-xs sm:text-sm">
              <Palette class="h-4 w-4" />
              <span>Tampilan & Tema</span>
            </TabsTrigger>
            <TabsTrigger value="preview" class="gap-1.5 text-xs sm:text-sm lg:hidden">
              <Smartphone class="h-4 w-4" />
              <span>Pratinjau HP</span>
            </TabsTrigger>
          </TabsList>
        </div>

        <!-- 2-Column Responsive Layout on Desktop -->
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          <!-- Left Column: Links or Appearance Form (7 cols) -->
          <div class="space-y-6 lg:col-span-7">
            <!-- TAB 1: Links Content -->
            <TabsContent value="links" class="space-y-6 m-0">
              <!-- Add / Edit Form Card -->
              {#if showLinkForm}
                <Card class="border-primary/50 shadow-md">
                  <CardHeader class="pb-3">
                    <div class="flex items-center justify-between">
                      <div>
                        <CardTitle class="text-lg font-bold">
                          {editingLinkId ? 'Edit Tautan' : 'Tambah Tautan Baru'}
                        </CardTitle>
                        <CardDescription class="text-xs">
                          Tautan ini akan langsung muncul dan dapat diklik di halaman publik Anda.
                        </CardDescription>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <form onsubmit={handleSaveLink} class="space-y-4">
                      <div class="space-y-1.5">
                        <Label for="link-title" class="text-xs font-semibold">Judul Tautan</Label>
                        <Input
                          id="link-title"
                          bind:value={linkTitle}
                          placeholder="contoh: Instagram Pribadi, Portofolio, WhatsApp Bisnis"
                          required
                          class="h-10"
                        />
                      </div>

                      <div class="space-y-1.5">
                        <Label for="link-url" class="text-xs font-semibold">URL / Tautan Tujuan</Label>
                        <Input
                          id="link-url"
                          bind:value={linkUrl}
                          placeholder="https://instagram.com/username"
                          required
                          class="h-10"
                        />
                        <p class="text-[11px] text-muted-foreground">
                          Awalan https:// otomatis ditambahkan jika belum ditulis.
                        </p>
                      </div>

                      <div class="flex items-center justify-end gap-2 pt-2">
                        <Button type="button" variant="outline" size="sm" onclick={cancelLinkForm}>
                          Batal
                        </Button>
                        <Button type="submit" size="sm" disabled={isSavingLink}>
                          {#if isSavingLink}
                            <Loader2 class="mr-1.5 h-3.5 w-3.5 animate-spin" />
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

              <!-- Links List Card -->
              <Card class="shadow-sm">
                <CardHeader class="flex flex-row items-center justify-between pb-3">
                  <div>
                    <CardTitle class="text-base sm:text-lg font-bold">Daftar Tautan</CardTitle>
                    <CardDescription class="text-xs">
                      Kelola tautan yang ditampilkan kepada pengunjung Anda
                    </CardDescription>
                  </div>
                  {#if !showLinkForm}
                    <Button variant="outline" size="sm" onclick={openAddLinkForm} class="gap-1.5 text-xs">
                      <Plus class="h-3.5 w-3.5" />
                      Tambah
                    </Button>
                  {/if}
                </CardHeader>

                <CardContent>
                  {#if user.links.length === 0}
                    <div class="flex flex-col items-center justify-center py-12 text-center">
                      <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground mb-3">
                        <Link2 class="h-6 w-6" />
                      </div>
                      <h3 class="font-semibold text-foreground text-sm sm:text-base">Belum ada tautan yang ditambahkan</h3>
                      <p class="text-xs sm:text-sm text-muted-foreground max-w-sm mt-1 mb-4">
                        Tambahkan tautan pertama Anda agar pengunjung dapat menemukan profil media sosial atau portofolio Anda.
                      </p>
                      <Button size="sm" onclick={openAddLinkForm} class="gap-1.5">
                        <Plus class="h-4 w-4" />
                        Tambah Tautan Pertama
                      </Button>
                    </div>
                  {:else}
                    <div class="space-y-2.5">
                      {#each user.links as link, i}
                        <div class="flex items-center justify-between p-3.5 rounded-2xl border border-border/80 bg-card hover:bg-muted/40 transition-colors shadow-2xs group min-w-0 max-w-full">
                          <div class="min-w-0 flex-1 pr-2 sm:pr-3 overflow-hidden">
                            <div class="flex items-center gap-2 min-w-0">
                              <span class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary/10 text-[11px] font-bold text-primary">
                                {i + 1}
                              </span>
                              <h4 class="font-semibold text-sm text-foreground truncate min-w-0 flex-1" title={link.title}>{link.title}</h4>
                            </div>
                            <a
                              href={link.url}
                              target="_blank"
                              rel="noreferrer"
                              class="mt-1 flex items-center gap-1 text-xs text-muted-foreground hover:text-primary underline-offset-2 hover:underline min-w-0 max-w-full group/link"
                              title={link.url}
                            >
                              <span class="truncate min-w-0 flex-1">{link.url}</span>
                              <ArrowUpRight class="h-3 w-3 shrink-0 opacity-60 group-hover/link:opacity-100 transition-opacity" />
                            </a>
                          </div>

                          <div class="flex items-center gap-1 shrink-0">
                            <Button
                              variant="ghost"
                              size="icon"
                              class="h-8 w-8 text-muted-foreground hover:text-foreground"
                              onclick={() => openEditLinkForm(link)}
                              title="Edit Tautan"
                              aria-label="Edit Tautan"
                            >
                              <Edit3 class="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              class="h-8 w-8 text-destructive/80 hover:text-destructive hover:bg-destructive/10"
                              onclick={() => promptDeleteLink(link)}
                              title="Hapus Tautan"
                              aria-label="Hapus Tautan"
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
            </TabsContent>

            <!-- TAB 2: Appearance & Theme Settings -->
            <TabsContent value="appearance" class="space-y-6 m-0">
              <Card class="shadow-sm">
                <CardHeader>
                  <div class="flex items-center gap-2">
                    <Palette class="h-5 w-5 text-primary" />
                    <CardTitle class="text-base sm:text-lg font-bold">Pengaturan Halaman Publik</CardTitle>
                  </div>
                  <CardDescription class="text-xs">
                    Sesuaikan judul, bio, foto avatar, dan palet warna halaman publik Anda
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <form onsubmit={handleSaveSettings} class="space-y-5">
                    <!-- Judul Halaman Public -->
                    <div class="space-y-1.5">
                      <Label for="page-title" class="text-xs font-semibold">Judul Halaman Publik</Label>
                      <Input
                        id="page-title"
                        bind:value={settingsTitle}
                        placeholder={`contoh: ${user.username ? user.username : 'Nama Lengkap atau Brand'}`}
                        class="h-10"
                      />
                      <p class="text-[11px] text-muted-foreground">
                        Judul utama yang tampil di bawah avatar. Bila kosong, otomatis menggunakan <code>@{user.username || 'username'}</code>.
                      </p>
                    </div>

                    <!-- Deskripsi Profil (Bio) -->
                    <div class="space-y-1.5">
                      <Label for="page-description" class="text-xs font-semibold">Deskripsi Profil (Bio)</Label>
                      <textarea
                        id="page-description"
                        bind:value={settingsDescription}
                        rows="2"
                        class="flex w-full rounded-2xl border border-input bg-background px-3 py-2 text-sm shadow-xs placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                        placeholder="contoh: Software Engineer, Konten Kreator, atau Digital Artist"
                      ></textarea>
                      <p class="text-[11px] text-muted-foreground">
                        Deskripsi singkat yang tampil di bawah judul profil publik Anda.
                      </p>
                    </div>

                    <!-- URL Foto Avatar -->
                    <div class="space-y-1.5">
                      <Label for="avatar-url" class="text-xs font-semibold">URL Avatar / Foto Profil</Label>
                      <div class="flex items-center gap-3">
                        <Avatar class="h-10 w-10 ring-1 ring-border shrink-0">
                          {#if settingsAvatarUrl}
                            <AvatarImage src={settingsAvatarUrl} alt="Avatar Preview" />
                          {/if}
                          <AvatarFallback class="bg-primary/10 text-primary font-bold text-xs">
                            {(user.username || 'U').slice(0, 1).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <Input
                          id="avatar-url"
                          bind:value={settingsAvatarUrl}
                          placeholder="https://example.com/foto-anda.jpg"
                          class="h-10 flex-1"
                        />
                      </div>
                      <p class="text-[11px] text-muted-foreground">
                        Gunakan tautan gambar langsung (JPG/PNG).
                      </p>
                    </div>

                    <!-- Palet Warna Latar Belakang -->
                    <div class="space-y-2">
                      <Label class="text-xs font-semibold">Warna Latar Belakang (Background)</Label>
                      <div class="flex flex-wrap gap-2.5 pt-1">
                        {#each bgPresets as preset}
                          <button
                            type="button"
                            class="h-8 w-8 rounded-full border shadow-2xs transition-transform hover:scale-110 flex items-center justify-center cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary"
                            style={`background-color: ${preset.value};`}
                            onclick={() => settingsBgColor = preset.value}
                            title={preset.label}
                            aria-label={`Pilih warna ${preset.label}`}
                          >
                            {#if settingsBgColor.toLowerCase() === preset.value.toLowerCase()}
                              <Check class={`h-4 w-4 ${preset.value === '#0f172a' ? 'text-white' : 'text-slate-900'}`} />
                            {/if}
                          </button>
                        {/each}
                      </div>

                      <div class="flex items-center gap-2 pt-1">
                        <Input
                          type="text"
                          bind:value={settingsBgColor}
                          class="font-mono text-xs h-9"
                          placeholder="#f8fafc"
                        />
                        <input
                          type="color"
                          bind:value={settingsBgColor}
                          class="h-9 w-9 rounded-xl border border-input cursor-pointer p-0.5 bg-transparent shrink-0"
                          aria-label="Pilih warna custom"
                        />
                      </div>

                      <!-- Indikator Rasio Kontras WCAG 2.1 -->
                      <div
                        class="mt-2.5 p-3.5 rounded-2xl border text-xs space-y-1.5 transition-colors"
                        style={`background-color: ${previewTheme.isDark ? '#0f172a' : '#f8fafc'}; border-color: ${previewTheme.isDark ? '#334155' : '#e2e8f0'}; color: ${previewTheme.isDark ? '#f8fafc' : '#0f172a'};`}
                      >
                        <div class="flex items-center justify-between">
                          <span class="font-bold flex items-center gap-1.5">
                            <Eye class="h-3.5 w-3.5 text-primary" />
                            Rasio Kontras (WCAG 2.1):
                          </span>
                          <span
                            class="font-bold px-2 py-0.5 rounded-full text-[11px]"
                            style={`background-color: ${previewTheme.isDark ? '#334155' : '#e2e8f0'}; color: ${previewTheme.isDark ? '#38bdf8' : '#0284c7'};`}
                          >
                            {previewTheme.contrastRatio}:1 ({previewTheme.wcagLevel})
                          </span>
                        </div>
                        <p class="text-[11px] leading-relaxed" style={`color: ${previewTheme.isDark ? '#cbd5e1' : '#64748b'};`}>
                          {#if previewTheme.isDark}
                            Latar gelap terdeteksi: Teks otomatis disesuaikan ke warna terang kontras tinggi untuk memastikan keterbacaan (readability & legibility) maksimal.
                          {:else}
                            Latar terang terdeteksi: Teks otomatis memakai warna gelap kontras tajam sesuai standar aksesibilitas WCAG.
                          {/if}
                        </p>
                      </div>
                    </div>

                    <!-- Jenis Huruf (Font Family) -->
                    <div class="space-y-1.5">
                      <Label for="font-family" class="text-xs font-semibold">Jenis Huruf (Font)</Label>
                      <select
                        id="font-family"
                        bind:value={settingsFontFamily}
                        class="flex h-10 w-full rounded-2xl border border-input bg-background px-3 py-1 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      >
                        <option value="Inter">Inter (Modern Clean)</option>
                        <option value="sans-serif">System Sans</option>
                        <option value="serif">Classic Serif</option>
                        <option value="monospace">Developer Mono</option>
                      </select>
                    </div>

                    <Button type="submit" class="w-full h-10 mt-2" disabled={isSavingSettings}>
                      {#if isSavingSettings}
                        <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                        Menyimpan Perubahan...
                      {:else}
                        Simpan Perubahan Tampilan
                      {/if}
                    </Button>
                  </form>
                </CardContent>
              </Card>
            </TabsContent>

            <!-- TAB 3: Mobile Preview Content -->
            <TabsContent value="preview" class="m-0 lg:hidden space-y-4">
              {@render phonePreview()}

              <!-- Quick helper card on Mobile -->
              <Card class="shadow-sm border-border/80">
                <CardContent class="p-4 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs">
                  <div class="flex items-center gap-2 text-muted-foreground">
                    <Sparkles class="h-4 w-4 text-primary shrink-0" />
                    <span>Perubahan pada tautan & tema otomatis tersinkronisasi di pratinjau.</span>
                  </div>
                  <div class="flex items-center gap-2 w-full sm:w-auto shrink-0">
                    <Button
                      variant="outline"
                      size="sm"
                      class="flex-1 sm:flex-initial text-xs h-8"
                      onclick={() => activeTab = 'appearance'}
                    >
                      <Palette class="h-3.5 w-3.5 mr-1.5" />
                      Edit Tampilan
                    </Button>
                    {#if user.username}
                      <Button
                        size="sm"
                        class="flex-1 sm:flex-initial text-xs h-8"
                        href={`/#/${user.username}`}
                        target="_blank"
                      >
                        <ExternalLink class="h-3.5 w-3.5 mr-1.5" />
                        Buka Halaman
                      </Button>
                    {/if}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </div>

          <!-- Right Column: Live Mobile Phone Mockup (Sticky Desktop - 5 cols) -->
          <div class="hidden lg:block lg:col-span-5">
            <div class="sticky top-24 space-y-4">
              {@render phonePreview()}
            </div>
          </div>
        </div>
      </Tabs>
    {/if}
  </main>

  <!-- Alert Dialog Konfirmasi Hapus Tautan (shadcn-svelte) -->
  <AlertDialog.Root bind:open={isDeleteDialogOpen}>
    <AlertDialog.Content>
      <AlertDialog.Header>
        <AlertDialog.Title>Hapus Tautan?</AlertDialog.Title>
        <AlertDialog.Description>
          Apakah Anda yakin ingin menghapus tautan <strong class="font-medium text-foreground">"{linkToDelete?.title}"</strong>? Tautan ini akan dihapus secara permanen dari halaman publik Anda dan tindakan ini tidak dapat dibatalkan.
        </AlertDialog.Description>
      </AlertDialog.Header>
      <AlertDialog.Footer>
        <AlertDialog.Cancel
          disabled={isDeletingLink}
          onclick={() => {
            isDeleteDialogOpen = false;
            linkToDelete = null;
          }}
        >
          Batal
        </AlertDialog.Cancel>
        <AlertDialog.Action
          variant="destructive"
          disabled={isDeletingLink}
          onclick={handleConfirmDelete}
        >
          {#if isDeletingLink}
            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
            Menghapus...
          {:else}
            Hapus Tautan
          {/if}
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
</div>
