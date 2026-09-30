// Automated verification test script for Issue #11: Halaman Publik (Public Page Renderer)
const fs = require('fs');
const path = require('path');

const BACKEND_URL = 'http://localhost:8080';
const FRONTEND_URL = 'http://localhost:5173';

async function request(url, options = {}) {
  const res = await fetch(url, options);
  let data;
  const ct = res.headers.get('content-type') || '';
  if (ct.includes('application/json')) {
    data = await res.json();
  } else {
    data = await res.text();
  }
  return { status: res.status, ok: res.ok, data, headers: res.headers };
}

async function run() {
  console.log('====================================================');
  console.log('🧪 VERIFYING ISSUE #11: Public Page Renderer (M3)');
  console.log('====================================================\n');

  // STEP 1: Verify Code Splitting / Lazy Loading Requirement
  console.log('--- TEST 1: Lazy Loading Private Routes (Dashboard/Auth) ---');
  const appSveltePath = path.join(__dirname, 'frontend', 'src', 'App.svelte');
  const appSvelteContent = fs.readFileSync(appSveltePath, 'utf8');

  // Check that Dashboard is NOT statically imported at the top
  const hasStaticDashboardImport = /^import\s+Dashboard\s+from/m.test(appSvelteContent);
  if (hasStaticDashboardImport) {
    throw new Error('FAILED: Dashboard is statically imported in App.svelte! It must be lazy loaded.');
  }
  console.log('✓ App.svelte does NOT statically import Dashboard');

  // Check that wrap uses asyncComponent for Dashboard
  if (!appSvelteContent.includes("asyncComponent: () => import('./routes/Dashboard.svelte')")) {
    throw new Error('FAILED: Dashboard is not wrapped with asyncComponent in App.svelte!');
  }
  console.log('✓ App.svelte lazy-loads Dashboard using asyncComponent');

  // Check that /* wildcard route exists
  if (!appSvelteContent.includes("'/*': PublicPage")) {
    throw new Error('FAILED: Route /* is not mapped to PublicPage in App.svelte!');
  }
  console.log("✓ Route '/*' is properly mapped to PublicPage");

  // STEP 2: Verify PublicPage.svelte Implementation Rules
  console.log('\n--- TEST 2: Inspecting PublicPage.svelte Requirements ---');
  const publicPagePath = path.join(__dirname, 'frontend', 'src', 'routes', 'PublicPage.svelte');
  const publicPageContent = fs.readFileSync(publicPagePath, 'utf8');

  // Check that /login and /dashboard are not treated as usernames
  if (!publicPageContent.includes('login') || !publicPageContent.includes('dashboard')) {
    throw new Error('FAILED: PublicPage does not guard against login/dashboard usernames!');
  }
  console.log('✓ PublicPage guards against treating /login and /dashboard as usernames');

  // Check that path segments are URL encoded
  if (!publicPageContent.includes('encodeURIComponent')) {
    throw new Error('FAILED: PublicPage does not encode username path segments!');
  }
  console.log('✓ PublicPage URL-encodes path segment (encodeURIComponent)');

  // Check that 404 message exists with exact phrase
  if (!publicPageContent.includes('Halaman tidak ditemukan')) {
    throw new Error("FAILED: PublicPage does not include exact text 'Halaman tidak ditemukan'!");
  }
  console.log("✓ PublicPage includes exact 404 text 'Halaman tidak ditemukan'");

  // Check external links security attributes
  if (!publicPageContent.includes('target="_blank"') || !publicPageContent.includes('rel="noopener noreferrer"')) {
    throw new Error('FAILED: External links missing target="_blank" or rel="noopener noreferrer"!');
  }
  console.log('✓ Links use target="_blank" and rel="noopener noreferrer"');

  // STEP 3: Test Public API for Non-Existent User (404)
  console.log('\n--- TEST 3: GET /api/public/{username} for Non-Existent User ---');
  const randomUser = `ghost_user_${Date.now()}`;
  const notFoundRes = await request(`${BACKEND_URL}/api/public/${randomUser}`);
  console.log(`GET /api/public/${randomUser} -> Status:`, notFoundRes.status);
  if (notFoundRes.status !== 404) {
    throw new Error(`Expected 404 for non-existent user, got ${notFoundRes.status}`);
  }
  if (!notFoundRes.data?.error) {
    throw new Error('Expected error message in 404 response payload');
  }
  console.log('✓ 404 response received with proper error payload:', notFoundRes.data);

  // STEP 4: Create a User and Configure Theme & Links
  console.log('\n--- TEST 4: Create Test User with Profile & Links ---');
  const ts = Date.now();
  const testUser = `pubtest-${ts}`;
  const testEmail = `pubtest-${ts}@test.com`;
  const testPass = 'Password123!';

  const regRes = await request(`${BACKEND_URL}/api/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: testUser, email: testEmail, password: testPass })
  });
  if (!regRes.ok) throw new Error('Registration failed: ' + JSON.stringify(regRes.data));

  const loginRes = await request(`${BACKEND_URL}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: testEmail, password: testPass })
  });
  if (!loginRes.ok) throw new Error('Login failed: ' + JSON.stringify(loginRes.data));
  const token = loginRes.data.token;
  const authHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`
  };

  // Set theme & title & description & avatar
  const themePayload = {
    bg_color: '#1e293b',
    font_family: 'monospace',
    avatar_url: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=150',
    title: 'Dr. John Doe',
    description: 'Tech Lead & Open Source Maintainer'
  };

  const settingsRes = await request(`${BACKEND_URL}/api/me/settings`, {
    method: 'PUT',
    headers: authHeaders,
    body: JSON.stringify(themePayload)
  });
  if (!settingsRes.ok) throw new Error('Settings update failed: ' + JSON.stringify(settingsRes.data));
  console.log('✓ Configured theme settings:', {
    bg_color: settingsRes.data.bg_color,
    font_family: settingsRes.data.font_family,
    avatar_url: settingsRes.data.avatar_url
  });

  // Add 2 links
  const link1 = await request(`${BACKEND_URL}/api/me/links`, {
    method: 'POST',
    headers: authHeaders,
    body: JSON.stringify({ title: 'Personal Website', url: 'https://johndoe.dev' })
  });
  const link2 = await request(`${BACKEND_URL}/api/me/links`, {
    method: 'POST',
    headers: authHeaders,
    body: JSON.stringify({ title: 'GitHub Repository', url: 'https://github.com/lymau/link-mudah' })
  });
  if (!link1.ok || !link2.ok) throw new Error('Link creation failed');
  console.log('✓ Created 2 links for user');

  // STEP 5: Verify Public API GET /api/public/{username}
  console.log('\n--- TEST 5: GET /api/public/{username} Data Integrity ---');
  const pubRes = await request(`${BACKEND_URL}/api/public/${testUser}`);
  console.log(`GET /api/public/${testUser} -> Status:`, pubRes.status);
  if (!pubRes.ok) throw new Error('Failed to fetch public profile: ' + JSON.stringify(pubRes.data));

  const data = pubRes.data;
  console.log('Public Profile Data:', {
    username: data.username,
    title: data.title,
    description: data.description,
    bg_color: data.bg_color,
    font_family: data.font_family,
    avatar_url: data.avatar_url,
    linksCount: data.links?.length
  });

  if (data.username !== testUser) throw new Error(`Username mismatch: expected ${testUser}, got ${data.username}`);
  if (data.bg_color !== themePayload.bg_color) throw new Error(`bg_color mismatch: ${data.bg_color}`);
  if (data.font_family !== themePayload.font_family) throw new Error(`font_family mismatch: ${data.font_family}`);
  if (data.avatar_url !== themePayload.avatar_url) throw new Error(`avatar_url mismatch: ${data.avatar_url}`);
  if (data.links?.length !== 2) throw new Error(`Expected 2 links, got ${data.links?.length}`);
  if (data.links[0].title !== 'Personal Website' || data.links[0].url !== 'https://johndoe.dev') {
    throw new Error('Link 1 title/url mismatch');
  }
  if (data.links[1].title !== 'GitHub Repository' || data.links[1].url !== 'https://github.com/lymau/link-mudah') {
    throw new Error('Link 2 title/url mismatch');
  }
  console.log('✓ Public API data matches configured settings & links precisely');

  // STEP 6: Verify Frontend Dev Server Serves Direct and Hash URLs
  console.log('\n--- TEST 6: Frontend Route & Direct Pathname Serving ---');
  const feIndex = await request(`${FRONTEND_URL}/`);
  if (!feIndex.ok) throw new Error('Frontend dev server failed to respond on /');
  console.log('✓ Frontend server is healthy and responds on /');

  const feUserDirect = await request(`${FRONTEND_URL}/${testUser}`);
  if (!feUserDirect.ok) throw new Error(`Frontend dev server failed on /${testUser}`);
  console.log(`✓ Frontend server serves SPA HTML on direct pathname /${testUser}`);

  console.log('\n====================================================');
  console.log('🎉 ALL ISSUE #11 VERIFICATION CHECKS PASSED!');
  console.log('====================================================');
}

run().catch(err => {
  console.error('\n❌ VERIFICATION TEST FAILED:', err);
  process.exit(1);
});
