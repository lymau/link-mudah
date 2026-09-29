<script>
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Alert, AlertDescription } from '$lib/components/ui/alert';
  import { Link2, LogIn, AlertCircle, Loader2 } from 'lucide-svelte';
  import { apiFetch, setAuthToken } from '$lib/api';

  let email = '';
  let password = '';
  let error = '';
  let isSubmitting = false;

  async function handleSubmit(event) {
    if (event && typeof event.preventDefault === 'function') {
      event.preventDefault();
    }
    error = '';
    isSubmitting = true;

    try {
      const data = await apiFetch('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });

      setAuthToken(data.token);
      push('/dashboard');
    } catch (err) {
      error = err.message || 'Login gagal. Periksa kembali email dan password Anda.';
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
      <CardTitle class="text-2xl">Masuk ke Akun</CardTitle>
      <CardDescription>Masukkan email dan kata sandi Anda untuk mengakses dashboard.</CardDescription>
    </CardHeader>

    <CardContent>
      <form class="space-y-4" onsubmit={handleSubmit}>
        <div class="space-y-2">
          <Label for="email">Email</Label>
          <Input id="email" bind:value={email} type="email" placeholder="nama@email.com" required autocomplete="email" />
        </div>

        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <Label for="password">Kata Sandi</Label>
          </div>
          <Input id="password" bind:value={password} type="password" placeholder="••••••••" required autocomplete="current-password" />
        </div>

        {#if error}
          <Alert variant="destructive">
            <AlertCircle class="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        {/if}

        <Button type="submit" class="w-full h-10 font-medium" disabled={isSubmitting}>
          {#if isSubmitting}
            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
            Memproses...
          {:else}
            <LogIn class="mr-2 h-4 w-4" />
            Masuk
          {/if}
        </Button>
      </form>
    </CardContent>

    <CardFooter class="flex justify-center text-sm text-muted-foreground pt-0">
      <span>
        Belum punya akun?
        <Button type="button" variant="link" class="p-0 font-medium h-auto" onclick={() => push('/register')}>
          Daftar sekarang
        </Button>
      </span>
    </CardFooter>
  </Card>
</div>

