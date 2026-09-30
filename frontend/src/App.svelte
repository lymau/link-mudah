<script>
  import Router, { replace } from 'svelte-spa-router'
  import { wrap } from 'svelte-spa-router/wrap'
  import Login from './routes/Login.svelte'
  import Register from './routes/Register.svelte'
  import Dashboard from './routes/Dashboard.svelte'
  import PublicPage from './routes/PublicPage.svelte'
  import { getAuthToken } from './lib/api'

  // If browser opens direct pathname without hash (e.g. /login or /register), convert to hash for svelte-spa-router
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
      component: Login,
      conditions: [redirectIfAuthenticated]
    }),
    '/login': wrap({
      component: Login,
      conditions: [redirectIfAuthenticated]
    }),
    '/register': wrap({
      component: Register,
      conditions: [redirectIfAuthenticated]
    }),
    '/dashboard': wrap({
      component: Dashboard,
      conditions: [requireAuth]
    }),
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


