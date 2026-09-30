<script>
  import { onMount } from 'svelte';
  import { push, replace } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import { Link2, ExternalLink, UserX, Loader2, Share2, Check, WifiOff } from 'lucide-svelte';
  import { apiFetch } from '$lib/api';
  import { getAccessibleTheme } from '$lib/contrast';

  let { params = {} } = $props();

  let profile = $state(null);
  let isLoading = $state(true);
  let errorStatus = $state(null);
  let errorMessage = $state('');
  let currentUsername = $state('');
  let copied = $state(false);

  let theme = $derived(getAccessibleTheme(profile?.bg_color || '#f8fafc'));

  function extractUsername() {
    let raw = '';
    if (params && params.wild) {
      raw = params.wild;
    } else if (params && params['*']) {
      raw = params['*'];
    } else if (typeof window !== 'undefined' && window.location.hash) {
      raw = window.location.hash.slice(1).split('?')[0];
    }
    raw = (raw || '').replace(/^\/+|\/+$/g, '');
    if (raw.includes('/')) {
      raw = raw.split('/')[0];
    }
    return raw;
  }

  async function loadProfile(targetUsername) {
    const target = (targetUsername || '').trim();
    if (!target) {
      replace('/');
      return;
    }

    // Do not treat /login, /dashboard, or /register as usernames
    const lower = target.toLowerCase();
    if (lower === 'login' || lower === 'dashboard') {
      replace('/' + lower);
      return;
    }
    if (lower === 'register') {
      replace('/register');
      return;
    }

    currentUsername = target;
    isLoading = true;
    errorStatus = null;
    errorMessage = '';

    try {
      const data = await apiFetch(`/api/public/${encodeURIComponent(target)}`);
      profile = {
        username: data.username,
        title: data.title || '',
        description: data.description || '',
        bg_color: data.bg_color || '#f8fafc',
        font_family: data.font_family || 'Inter',
        avatar_url: data.avatar_url || '',
        links: data.links || []
      };

      if (typeof document !== 'undefined') {
        const displayTitle = profile.title?.trim()
          ? `${profile.title.trim()} (@${profile.username})`
          : `@${profile.username}`;
        document.title = `${displayTitle} | Link Mudah`;
      }
    } catch (err) {
      profile = null;
      if (err.status === 404 || err.message?.toLowerCase().includes('not found') || err.message?.toLowerCase().includes('tidak ditemukan')) {
        errorStatus = 404;
        errorMessage = 'Halaman tidak ditemukan';
      } else {
        errorStatus = err.status || 500;
        errorMessage = err.message || 'Gagal memuat profil.';
      }
    } finally {
      isLoading = false;
    }
  }

  function handleShare() {
    navigator.clipboard.writeText(window.location.href);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  }

  $effect(() => {
    const raw = extractUsername();
    if (raw && raw !== currentUsername) {
      loadProfile(raw);
    }
  });

  onMount(() => {
    const raw = extractUsername();
    if (raw) {
      loadProfile(raw);
    } else {
      replace('/');
    }
  });
</script>

{#if isLoading}
  <div class="flex min-h-screen flex-col items-center justify-center bg-muted/30 p-4">
    <div class="flex flex-col items-center gap-3 text-muted-foreground">
      <Loader2 class="h-8 w-8 animate-spin text-primary" />
      <p class="text-sm font-medium">Memuat profil...</p>
    </div>
  </div>
{:else if errorStatus === 404}
  <div class="flex min-h-screen items-center justify-center bg-muted/40 p-4">
    <Card class="w-full max-w-md shadow-md text-center">
      <CardHeader class="flex flex-col items-center space-y-2">
        <div class="flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10 text-destructive mb-1">
          <UserX class="h-6 w-6" />
        </div>
        <CardTitle class="text-2xl font-bold tracking-tight">Halaman tidak ditemukan</CardTitle>
        <CardDescription class="max-w-xs mx-auto">
          Pengguna '@{currentUsername}' belum terdaftar di Link Mudah.
        </CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col gap-2 pt-2">
        <Button onclick={() => push('/register')} class="w-full">
          Buat Halaman Link Anda Sekarang
        </Button>
        <Button variant="outline" onclick={() => push('/login')} class="w-full">
          Masuk ke Akun
        </Button>
      </CardContent>
    </Card>
  </div>
{:else if errorStatus}
  <div class="flex min-h-screen items-center justify-center bg-muted/40 p-4">
    <Card class="w-full max-w-md shadow-md text-center">
      <CardHeader class="flex flex-col items-center space-y-2">
        <div class="flex h-12 w-12 items-center justify-center rounded-full bg-amber-500/10 text-amber-500 mb-1">
          <WifiOff class="h-6 w-6" />
        </div>
        <CardTitle class="text-2xl font-bold tracking-tight">Gagal Memuat Halaman</CardTitle>
        <CardDescription class="max-w-xs mx-auto">
          {errorMessage || 'Terjadi gangguan saat menghubungi server. Silakan periksa koneksi Anda.'}
        </CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col gap-2 pt-2">
        <Button onclick={() => loadProfile(currentUsername)} class="w-full">
          Coba Lagi
        </Button>
        <Button variant="outline" onclick={() => push('/')} class="w-full">
          Kembali ke Beranda
        </Button>
      </CardContent>
    </Card>
  </div>
{:else}
  <div
    class="min-h-screen flex flex-col items-center justify-between p-4 sm:p-8 transition-colors duration-200"
    style={`background-color: ${profile.bg_color}; font-family: ${profile.font_family}; color: ${theme.textColor};`}
  >
    <!-- Share Button Top Right -->
    <div class="w-full max-w-lg flex justify-end">
      <button
        type="button"
        class="inline-flex items-center gap-1.5 h-8 px-3 rounded-lg text-xs font-medium backdrop-blur shadow-xs border transition-all cursor-pointer hover:opacity-90"
        style="background-color: {theme.actionButtonBg}; color: {theme.actionButtonText}; border-color: {theme.actionButtonBorder};"
        onclick={handleShare}
      >
        {#if copied}
          <Check class="h-3.5 w-3.5 text-emerald-400" />
          <span>Tersalin</span>
        {:else}
          <Share2 class="h-3.5 w-3.5" />
          <span>Bagikan</span>
        {/if}
      </button>
    </div>

    <!-- Main Profile Card Container -->
    <div class="w-full max-w-lg my-auto py-8">
      <!-- Profile Header -->
      <div class="flex flex-col items-center text-center mb-8">
        {#if profile.avatar_url && profile.avatar_url.trim()}
          <img
            src={profile.avatar_url}
            alt={profile.title || profile.username}
            class="h-24 w-24 rounded-full object-cover shadow-md border-4 mb-4"
            style="border-color: {theme.borderColor};"
            onerror={(e) => { e.currentTarget.style.display = 'none'; }}
          />
        {/if}

        <!-- Judul Halaman Publik -->
        <h1 class="text-2xl sm:text-3xl font-bold tracking-tight mb-1" style="color: {theme.textColor};">
          {profile.title.trim() || `@${profile.username}`}
        </h1>

        <!-- Username badge / handle if custom title is active -->
        {#if profile.title.trim()}
          <p class="text-xs sm:text-sm font-mono opacity-85 mb-1" style="color: {theme.subTextColor};">
            @{profile.username}
          </p>
        {/if}

        <!-- Deskripsi Profil (di bawah avatar dan judul) -->
        <p class="text-sm sm:text-base max-w-md mx-auto leading-relaxed mt-1" style="color: {theme.mutedTextColor};">
          {profile.description.trim() || 'Kumpulan tautan resmi'}
        </p>
      </div>

      <!-- Links List -->
      <div class="space-y-3.5">
        {#if profile.links.length === 0}
          <div
            class="rounded-xl border backdrop-blur text-center py-6 px-4 shadow-xs"
            style="background-color: {theme.cardBg}; border-color: {theme.cardBorder}; color: {theme.mutedTextColor};"
          >
            <p class="text-sm">Belum ada tautan yang dipublikasikan.</p>
          </div>
        {:else}
          {#each profile.links as link}
            <a
              href={link.url}
              target="_blank"
              rel="noopener noreferrer"
              class="group relative flex items-center justify-between p-4 rounded-xl border backdrop-blur shadow-xs hover:shadow-md transition-all transform hover:-translate-y-0.5 active:translate-y-0"
              style="background-color: {theme.cardBg}; border-color: {theme.cardBorder}; color: {theme.cardTextColor}; box-shadow: {theme.cardShadow};"
            >
              <div class="flex items-center gap-3 min-w-0 pr-2">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg transition-colors"
                  style="background-color: {theme.isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.06)'}; color: {theme.cardTextColor};"
                >
                  <Link2 class="h-4 w-4" />
                </div>
                <span class="font-semibold text-sm sm:text-base truncate">
                  {link.title}
                </span>
              </div>

              <ExternalLink
                class="h-4 w-4 shrink-0 transition-opacity opacity-70 group-hover:opacity-100"
                style="color: {theme.cardTextColor};"
              />
            </a>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Bottom Footer Brand Badge -->
    <div class="py-6">
      <button
        type="button"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium backdrop-blur transition-colors cursor-pointer border hover:opacity-90"
        style="background-color: {theme.actionButtonBg}; color: {theme.mutedTextColor}; border-color: {theme.actionButtonBorder};"
        onclick={() => push('/register')}
      >
        <div class="flex h-4 w-4 items-center justify-center rounded bg-primary text-primary-foreground">
          <Link2 class="h-2.5 w-2.5" />
        </div>
        <span>Dibuat dengan <strong style="color: {theme.textColor};">Link Mudah</strong></span>
      </button>
    </div>
  </div>
{/if}
