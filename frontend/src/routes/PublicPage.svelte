<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import { Link2, ExternalLink, UserX, Loader2, Share2, Check } from 'lucide-svelte';
  import { apiFetch } from '$lib/api';

  let { params = {} } = $props();

  let profile = $state(null);
  let isLoading = $state(true);
  let error = $state('');
  let currentUsername = $state('');
  let copied = $state(false);

  function getUsername() {
    if (params && params['*']) {
      return params['*'].replace(/^\//, '');
    }
    if (typeof window !== 'undefined' && window.location.hash) {
      const path = window.location.hash.slice(1).split('?')[0];
      return path.replace(/^\//, '');
    }
    return '';
  }

  async function loadProfile() {
    const rawUsername = getUsername();
    if (!rawUsername) {
      error = 'Username tidak valid.';
      isLoading = false;
      return;
    }

    currentUsername = rawUsername;
    isLoading = true;
    error = '';

    try {
      const data = await apiFetch(`/api/public/${encodeURIComponent(rawUsername)}`);
      profile = {
        username: data.username,
        bg_color: data.bg_color || '#f8fafc',
        font_family: data.font_family || 'Inter',
        avatar_url: data.avatar_url || '',
        links: data.links || []
      };
    } catch (err) {
      error = err.message || 'Profil tidak ditemukan.';
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

  onMount(() => {
    loadProfile();
  });
</script>

{#if isLoading}
  <div class="flex min-h-screen flex-col items-center justify-center bg-muted/30 p-4">
    <div class="flex flex-col items-center gap-3 text-muted-foreground">
      <Loader2 class="h-8 w-8 animate-spin text-primary" />
      <p class="text-sm font-medium">Memuat profil...</p>
    </div>
  </div>
{:else if error || !profile}
  <div class="flex min-h-screen items-center justify-center bg-muted/40 p-4">
    <Card class="w-full max-w-md shadow-md text-center">
      <CardHeader class="flex flex-col items-center space-y-2">
        <div class="flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10 text-destructive mb-1">
          <UserX class="h-6 w-6" />
        </div>
        <CardTitle class="text-2xl">Halaman Tidak Ditemukan</CardTitle>
        <CardDescription class="max-w-xs mx-auto">
          {error || `Pengguna '@${currentUsername}' belum terdaftar di Link Mudah.`}
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
{:else}
  <div
    class="min-h-screen flex flex-col items-center justify-between p-4 sm:p-8"
    style={`background-color: ${profile.bg_color}; font-family: ${profile.font_family};`}
  >
    <!-- Share Button Top Right -->
    <div class="w-full max-w-lg flex justify-end">
      <Button
        variant="secondary"
        size="sm"
        class="h-8 gap-1.5 shadow-xs bg-background/80 backdrop-blur hover:bg-background"
        onclick={handleShare}
      >
        {#if copied}
          <Check class="h-3.5 w-3.5 text-emerald-600" />
          Tersalin
        {:else}
          <Share2 class="h-3.5 w-3.5" />
          Bagikan
        {/if}
      </Button>
    </div>

    <!-- Main Profile Card Container -->
    <div class="w-full max-w-lg my-auto py-8">
      <!-- Profile Header -->
      <div class="flex flex-col items-center text-center mb-8">
        {#if profile.avatar_url}
          <img
            src={profile.avatar_url}
            alt={profile.username}
            class="h-24 w-24 rounded-full object-cover shadow-md border-4 border-background mb-4"
            onerror={(e) => { e.currentTarget.style.display = 'none'; }}
          />
        {:else}
          <div class="flex h-24 w-24 items-center justify-center rounded-full bg-primary text-primary-foreground text-3xl font-bold shadow-md border-4 border-background mb-4">
            {profile.username ? profile.username.slice(0, 1).toUpperCase() : 'U'}
          </div>
        {/if}

        <h1 class="text-2xl font-bold tracking-tight text-foreground flex items-center gap-1.5">
          @{profile.username}
        </h1>
        <p class="text-sm text-muted-foreground mt-1">
          Kumpulan tautan resmi
        </p>
      </div>

      <!-- Links List -->
      <div class="space-y-3.5">
        {#if profile.links.length === 0}
          <Card class="bg-background/80 backdrop-blur shadow-sm text-center py-6">
            <CardContent class="p-0">
              <p class="text-sm text-muted-foreground">Belum ada tautan yang dipublikasikan.</p>
            </CardContent>
          </Card>
        {:else}
          {#each profile.links as link}
            <a
              href={link.url}
              target="_blank"
              rel="noopener noreferrer"
              class="group relative flex items-center justify-between p-4 rounded-xl border bg-background/90 hover:bg-background text-foreground shadow-xs hover:shadow-md transition-all transform hover:-translate-y-0.5 active:translate-y-0"
            >
              <div class="flex items-center gap-3 min-w-0 pr-2">
                <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary group-hover:bg-primary group-hover:text-primary-foreground transition-colors">
                  <Link2 class="h-4 w-4" />
                </div>
                <span class="font-medium text-sm sm:text-base truncate">
                  {link.title}
                </span>
              </div>

              <ExternalLink class="h-4 w-4 text-muted-foreground group-hover:text-foreground shrink-0 transition-colors" />
            </a>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Bottom Footer Brand Badge -->
    <div class="py-6">
      <Button
        variant="ghost"
        size="sm"
        class="gap-1.5 text-xs text-muted-foreground/80 hover:text-foreground hover:bg-background/50 backdrop-blur"
        onclick={() => push('/register')}
      >
        <div class="flex h-4 w-4 items-center justify-center rounded bg-primary text-primary-foreground">
          <Link2 class="h-2.5 w-2.5" />
        </div>
        <span>Dibuat dengan <strong>Link Mudah</strong></span>
      </Button>
    </div>
  </div>
{/if}
