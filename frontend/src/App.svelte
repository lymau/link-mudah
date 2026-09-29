<script>
  import Router from 'svelte-spa-router'
  import Login from './routes/Login.svelte'
  import Register from './routes/Register.svelte'
  import Dashboard from './routes/Dashboard.svelte'
  import PublicPage from './routes/PublicPage.svelte'

  // If browser opens direct pathname without hash (e.g. /login or /register), convert to hash for svelte-spa-router
  if (typeof window !== 'undefined' && window.location.pathname !== '/' && !window.location.hash) {
    window.location.replace('/#' + window.location.pathname + window.location.search);
  }

  const routes = {
    '/': Login,
    '/login': Login,
    '/register': Register,
    '/dashboard': Dashboard,
    '*': PublicPage,
  }
</script>

<main class="min-h-screen bg-slate-100 text-slate-900">
  <Router {routes} />
</main>

