<script>
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
  import { Alert, AlertDescription } from '$lib/components/ui/alert';
  import {
    Link2,
    UserPlus,
    AlertCircle,
    CheckCircle2,
    Loader2,
    Eye,
    EyeOff,
    Sparkles,
    ArrowLeft,
    Globe
  } from 'lucide-svelte';
  import { apiFetch } from '$lib/api';
  import { toast } from '$lib/components/ui/sonner';

  let username = $state('');
  let email = $state('');
  let password = $state('');
  let showPassword = $state(false);
  let error = $state('');
  let successMessage = $state('');
  let isSubmitting = $state(false);

  async function handleSubmit(event) {
    if (event && typeof event.preventDefault === 'function') {
      event.preventDefault();
    }
    error = '';
    successMessage = '';
    isSubmitting = true;

    try {
      await apiFetch('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify({ username, email, password }),
      });

      const successText = 'Pendaftaran akun berhasil! Mengalihkan ke halaman login...';
      successMessage = successText;
      toast.success(successText);
      setTimeout(() => push('/login'), 1200);
    } catch (err) {
      error = err.message || 'Registrasi gagal. Silakan coba lagi.';
      toast.error(error);
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="relative flex min-h-screen flex-col items-center justify-center p-4 sm:p-6 overflow-hidden bg-muted/20">
  <!-- Subtle decorative background gradient accents -->
  <div class="pointer-events-none absolute -top-40 -right-40 h-96 w-96 rounded-full bg-primary/10 blur-3xl"></div>
  <div class="pointer-events-none absolute -bottom-40 -left-40 h-96 w-96 rounded-full bg-accent/10 blur-3xl"></div>

  <!-- Branding header -->
  <header class="mb-8 flex flex-col items-center text-center">
    <div class="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-md ring-4 ring-primary/10 transition-transform hover:scale-105">
      <Link2 class="h-6 w-6" />
    </div>
    <div class="flex items-center gap-2">
      <h1 class="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">Link Mudah</h1>
      <span class="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2 py-0.5 text-[11px] font-semibold text-primary">
        <Sparkles class="h-3 w-3" /> Gratis
      </span>
    </div>
    <p class="mt-1 text-sm text-muted-foreground">Mulai buat halaman link-in-bio personal Anda sekarang</p>
  </header>

  <!-- Register Card (shadcn-svelte preset) -->
  <Card class="w-full max-w-md border-border/60 bg-card/95 backdrop-blur-sm shadow-xl transition-all">
    <CardHeader class="space-y-1.5 pb-4 text-center">
      <CardTitle class="text-2xl font-bold tracking-tight">Buat Akun Baru</CardTitle>
      <CardDescription class="text-sm text-muted-foreground">
        Lengkapi formulir di bawah ini untuk memulai dalam hitungan detik.
      </CardDescription>
    </CardHeader>

    <CardContent>
      <form class="space-y-4" onsubmit={handleSubmit}>
        {#if error}
          <Alert variant="destructive" class="py-2.5">
            <AlertCircle class="h-4 w-4" />
            <AlertDescription class="text-xs sm:text-sm font-medium">{error}</AlertDescription>
          </Alert>
        {/if}

        {#if successMessage}
          <Alert class="py-2.5 border-emerald-500/30 bg-emerald-50 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300">
            <CheckCircle2 class="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
            <AlertDescription class="text-xs sm:text-sm font-medium">{successMessage}</AlertDescription>
          </Alert>
        {/if}

        <div class="space-y-1.5">
          <Label for="username" class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Username
          </Label>
          <div class="relative">
            <Input
              id="username"
              bind:value={username}
              type="text"
              placeholder="username_anda"
              required
              autocomplete="username"
              class="h-10 transition-colors"
            />
          </div>
          <div class="flex items-center gap-1.5 text-[11px] text-muted-foreground pt-0.5">
            <Globe class="h-3 w-3 text-primary" />
            <span>URL publik Anda:</span>
            <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-foreground">
              linkmudah.com/#{username ? username.trim() : 'username'}
            </code>
          </div>
        </div>

        <div class="space-y-1.5">
          <Label for="email" class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Alamat Email
          </Label>
          <Input
            id="email"
            bind:value={email}
            type="email"
            placeholder="nama@domain.com"
            required
            autocomplete="email"
            class="h-10 transition-colors"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="password" class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Kata Sandi
          </Label>
          <div class="relative">
            <Input
              id="password"
              bind:value={password}
              type={showPassword ? 'text' : 'password'}
              placeholder="Minimal 8 karakter"
              required
              autocomplete="new-password"
              class="h-10 pr-10 transition-colors"
            />
            <button
              type="button"
              class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground hover:text-foreground cursor-pointer focus:outline-none"
              onclick={() => showPassword = !showPassword}
              aria-label={showPassword ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'}
            >
              {#if showPassword}
                <EyeOff class="h-4 w-4" />
              {:else}
                <Eye class="h-4 w-4" />
              {/if}
            </button>
          </div>
        </div>

        <Button
          type="submit"
          class="w-full h-11 text-sm font-medium tracking-wide shadow-sm transition-all mt-2"
          disabled={isSubmitting}
        >
          {#if isSubmitting}
            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
            Mendaftarkan Akun...
          {:else}
            <UserPlus class="mr-2 h-4 w-4" />
            Daftar Sekarang
          {/if}
        </Button>
      </form>
    </CardContent>

    <CardFooter class="flex flex-col items-center gap-2 border-t border-border/40 pt-4 text-center text-sm text-muted-foreground">
      <div class="flex items-center gap-1.5">
        <span>Sudah memiliki akun?</span>
        <Button
          type="button"
          variant="link"
          class="p-0 font-semibold text-primary inline-flex items-center gap-1 h-auto hover:underline"
          onclick={() => push('/login')}
        >
          <ArrowLeft class="h-3.5 w-3.5" />
          Masuk di sini
        </Button>
      </div>
    </CardFooter>
  </Card>

  <footer class="mt-8 text-center text-xs text-muted-foreground/80">
    &copy; {new Date().getFullYear()} Link Mudah. Cepat, aman, dan mudah digunakan.
  </footer>
</div>
