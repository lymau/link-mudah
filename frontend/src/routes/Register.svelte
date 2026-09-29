<script>
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
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

<div class="flex min-h-screen items-center justify-center bg-slate-100 p-6">
  <Card class="w-full max-w-md">
    <CardHeader class="space-y-2">
      <CardTitle class="text-2xl">Buat akun</CardTitle>
      <CardDescription>Daftar untuk membuat halaman link-in-bio Anda.</CardDescription>
    </CardHeader>

    <CardContent>
      <form class="space-y-4" onsubmit={handleSubmit}>
        <div class="space-y-2">
          <label for="username" class="text-sm font-medium text-slate-700">Username</label>
          <Input id="username" bind:value={username} type="text" placeholder="contoh: budi123" required />
        </div>

        <div class="space-y-2">
          <label for="email" class="text-sm font-medium text-slate-700">Email</label>
          <Input id="email" bind:value={email} type="email" placeholder="nama@email.com" required />
        </div>

        <div class="space-y-2">
          <label for="password" class="text-sm font-medium text-slate-700">Password</label>
          <Input id="password" bind:value={password} type="password" placeholder="Minimal 8 karakter" required />
        </div>

        {#if error}
          <p class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>
        {/if}

        {#if successMessage}
          <p class="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">{successMessage}</p>
        {/if}

        <Button type="submit" class="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Mendaftar...' : 'Daftar'}
        </Button>
      </form>
    </CardContent>

    <CardFooter class="flex justify-center">
      <Button type="button" variant="link" class="p-0" onclick={() => push('/login')}>
        Sudah punya akun? Masuk di sini
      </Button>
    </CardFooter>
  </Card>
</div>
