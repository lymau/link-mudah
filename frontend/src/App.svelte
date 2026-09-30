<script>
  import Router, { replace } from 'svelte-spa-router'
  import { wrap } from 'svelte-spa-router/wrap'
  import PublicPage from './routes/PublicPage.svelte'
  import { getAuthToken } from './lib/api'

  // If browser opens direct pathname without hash (e.g. /login or /register or /{username}), convert to hash for svelte-spa-router
  if (typeof window !== 'undefined' && window.location.pathname !== '/' && !window.location.hash) {
    window.location.replace('/#' + window.location.pathname + window.location.search);
  }

  // Precondition guard: Only authorized users can access protected routes like /dashboard
  function requireAuth() {
    const token = getAuthToken();
    if (!token) {
      replace('/login');
      return false;
    }
    return true;
  }

  // Precondition: If user is already logged in, redirect away from guest routes (login/register) to dashboard
  function redirectIfAuthenticated() {
    const token = getAuthToken();
    if (token) {
      replace('/dashboard');
      return false;
    }
    return true;
  }

  const routes = {
    '/': wrap({
      asyncComponent: () => import('./routes/Login.svelte'),
      conditions: [redirectIfAuthenticated]
    }),
    '/login': wrap({
      asyncComponent: () => import('./routes/Login.svelte'),
      conditions: [redirectIfAuthenticated]
    }),
    '/register': wrap({
      asyncComponent: () => import('./routes/Register.svelte'),
      conditions: [redirectIfAuthenticated]
    }),
    '/dashboard': wrap({
      asyncComponent: () => import('./routes/Dashboard.svelte'),
      conditions: [requireAuth]
    }),
    '/*': PublicPage,
    '*': PublicPage,
  }

  function handleConditionsFailed(event) {
    const detail = event.detail || event;
    if (detail && detail.location === '/dashboard') {
      replace('/login');
    }
  }
</script>

<main class="min-h-screen bg-background text-foreground">
  <Router {routes} onConditionsFailed={handleConditionsFailed} />
</main>


