<script>
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Alert, AlertDescription } from '$lib/components/ui/alert';
  import { Link2, UserPlus, AlertCircle, CheckCircle2, Loader2 } from 'lucide-svelte';
  import { apiFetch } from '$lib/api';

  let username = '';
  let email = '';
  let password = '';
  let error = '';
  let successMessage = '';
  let isSubmitting = false;

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

      successMessage = 'Pendaftaran berhasil. Mengalihkan ke halaman login...';
      setTimeout(() => push('/login'), 1200);
    } catch (err) {
      error = err.message || 'Registrasi gagal. Silakan coba lagi.';
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="relative flex min-h-screen flex-col items-center justify-center bg-muted/40 p-4 sm:p-6">
  <div class="mb-6 flex items-center gap-2.5">
    <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
      <Link2 class="h-5 w-5" />
    </div>
    <span class="text-2xl font-bold tracking-tight text-foreground">Link Mudah</span>
  </div>

  <Card class="w-full max-w-sm sm:max-w-md shadow-md border-border/80">
    <CardHeader class="space-y-1.5 text-center">
      <CardTitle class="text-2xl">Buat Akun Baru</CardTitle>
      <CardDescription>Daftar sekarang untuk membuat halaman link-in-bio personal Anda.</CardDescription>
    </CardHeader>

    <CardContent>
      <form class="space-y-4" onsubmit={handleSubmit}>
        <div class="space-y-2">
          <Label for="username">Username</Label>
          <div class="relative">
            <Input
              id="username"
              bind:value={username}
              type="text"
              placeholder="username_anda"
              required
              autocomplete="username"
            />
          </div>
          <p class="text-xs text-muted-foreground">Tautan publik Anda: linkmudah.com/#{username || 'username'}</p>
        </div>

        <div class="space-y-2">
          <Label for="email">Email</Label>
          <Input id="email" bind:value={email} type="email" placeholder="nama@email.com" required autocomplete="email" />
        </div>

        <div class="space-y-2">
          <Label for="password">Kata Sandi</Label>
          <Input id="password" bind:value={password} type="password" placeholder="Minimal 8 karakter" required autocomplete="new-password" />
        </div>

        {#if error}
          <Alert variant="destructive">
            <AlertCircle class="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        {/if}

        {#if successMessage}
          <Alert class="border-emerald-500/30 bg-emerald-50 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300">
            <CheckCircle2 class="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
            <AlertDescription>{successMessage}</AlertDescription>
          </Alert>
        {/if}

        <Button type="submit" class="w-full h-10 font-medium" disabled={isSubmitting}>
          {#if isSubmitting}
            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
            Mendaftarkan...
          {:else}
            <UserPlus class="mr-2 h-4 w-4" />
            Daftar Akun
          {/if}
        </Button>
      </form>
    </CardContent>

    <CardFooter class="flex justify-center text-sm text-muted-foreground pt-0">
      <span>
        Sudah punya akun?
        <Button type="button" variant="link" class="p-0 font-medium h-auto" onclick={() => push('/login')}>
          Masuk di sini
        </Button>
      </span>
    </CardFooter>
  </Card>
</div>

