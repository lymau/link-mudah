<script>
  import Router from 'svelte-spa-router'
  import Login from './routes/Login.svelte'
  import Dashboard from './routes/Dashboard.svelte'
  import PublicPage from './routes/PublicPage.svelte'

  const routes = {
    '/login': Login,
    '/dashboard': Dashboard,
    '*': PublicPage,
  }
</script>

<main class="min-h-screen bg-slate-100 text-slate-900">
  <Router {routes} />
</main>

