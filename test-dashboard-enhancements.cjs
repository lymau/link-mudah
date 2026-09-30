// Verification script for Dashboard Enhancements
const BASE_URL = 'http://localhost:8080';

async function request(path, options = {}) {
  const url = `${BASE_URL}${path}`;
  const res = await fetch(url, options);
  let payload;
  const contentType = res.headers.get('content-type') || '';
  if (contentType.includes('application/json')) {
    payload = await res.json();
  } else {
    payload = await res.text();
  }
  return { status: res.status, ok: res.ok, data: payload };
}

async function run() {
  console.log('=== TEST 1: Route & Endpoint Auth Protection ===');
  const unauth = await request('/api/me');
  console.log('GET /api/me without token -> Status:', unauth.status);
  if (unauth.status !== 401) throw new Error(`Expected 401, got ${unauth.status}`);

  console.log('\n=== TEST 2: Register User & Login ===');
  const timestamp = Date.now();
  const username = `user-${timestamp}`;
  const email = `user-${timestamp}@test.com`;
  const password = 'Password123!';

  const regRes = await request('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password })
  });
  if (!regRes.ok) throw new Error('Register failed: ' + JSON.stringify(regRes.data));

  const loginRes = await request('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  if (!loginRes.ok) throw new Error('Login failed');
  const token = loginRes.data.token;
  const authHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`
  };

  console.log('\n=== TEST 3: Settings with Judul (Title) & Deskripsi (di bawah avatar) ===');
  const updateSettingsPayload = {
    bg_color: '#0f172a', // Dark Slate
    font_family: 'Inter',
    avatar_url: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150',
    title: 'Budi Handoko',
    description: 'Senior Software Architect & Open Source Enthusiast'
  };

  const putRes = await request('/api/me/settings', {
    method: 'PUT',
    headers: authHeaders,
    body: JSON.stringify(updateSettingsPayload)
  });
  console.log('PUT /api/me/settings -> Status:', putRes.status, 'Data:', putRes.data);
  if (!putRes.ok) throw new Error('Settings update failed');
  if (putRes.data.title !== 'Budi Handoko') throw new Error(`Expected title 'Budi Handoko', got '${putRes.data.title}'`);
  if (putRes.data.description !== 'Senior Software Architect & Open Source Enthusiast') {
    throw new Error(`Expected description match, got '${putRes.data.description}'`);
  }
  if (putRes.data.bg_color !== '#0f172a') throw new Error('bg_color mismatch');

  console.log('\n=== TEST 4: Verify GET /api/me Returns Title & Description ===');
  const meRes = await request('/api/me', { headers: authHeaders });
  console.log('GET /api/me settings:', meRes.data.page_settings);
  if (meRes.data.page_settings.title !== 'Budi Handoko') throw new Error('GET /api/me title mismatch');
  if (meRes.data.page_settings.description !== 'Senior Software Architect & Open Source Enthusiast') {
    throw new Error('GET /api/me description mismatch');
  }

  console.log('\n=== TEST 5: Add a Link for Public Page ===');
  const linkRes = await request('/api/me/links', {
    method: 'POST',
    headers: authHeaders,
    body: JSON.stringify({ title: 'My GitHub Profile', url: 'https://github.com/lymau' })
  });
  if (!linkRes.ok) throw new Error('Create link failed');

  console.log('\n=== TEST 6: Verify Public Endpoint GET /api/public/{username} ===');
  const publicRes = await request(`/api/public/${username}`);
  console.log('GET /api/public/' + username + ' ->', publicRes.data);
  if (!publicRes.ok) throw new Error('Public page request failed');
  if (publicRes.data.title !== 'Budi Handoko') throw new Error('Public title mismatch');
  if (publicRes.data.description !== 'Senior Software Architect & Open Source Enthusiast') {
    throw new Error('Public description mismatch');
  }
  if (publicRes.data.bg_color !== '#0f172a') throw new Error('Public bg_color mismatch');
  if (publicRes.data.links?.length !== 1) throw new Error('Public links count mismatch');

  console.log('\n=== TEST 7: Contrast Ratio WCAG Evaluation ===');
  const { getAccessibleTheme, getContrastRatio } = await import('./frontend/src/lib/contrast.js');

  const blackTheme = getAccessibleTheme('#000000');
  console.log('Theme for #000000:', {
    isDark: blackTheme.isDark,
    contrastRatio: blackTheme.contrastRatio,
    wcagLevel: blackTheme.wcagLevel,
    textColor: blackTheme.textColor,
    mutedTextColor: blackTheme.mutedTextColor
  });
  if (!blackTheme.isDark || blackTheme.contrastRatio < 15 || blackTheme.textColor !== '#ffffff') {
    throw new Error('Black theme contrast ratio failure');
  }

  const darkSlateTheme = getAccessibleTheme('#0f172a');
  console.log('Theme for #0f172a:', {
    isDark: darkSlateTheme.isDark,
    contrastRatio: darkSlateTheme.contrastRatio,
    wcagLevel: darkSlateTheme.wcagLevel,
    textColor: darkSlateTheme.textColor
  });
  if (!darkSlateTheme.isDark || darkSlateTheme.contrastRatio < 10) {
    throw new Error('Dark slate theme contrast ratio failure');
  }

  const whiteTheme = getAccessibleTheme('#ffffff');
  console.log('Theme for #ffffff:', {
    isDark: whiteTheme.isDark,
    contrastRatio: whiteTheme.contrastRatio,
    wcagLevel: whiteTheme.wcagLevel,
    textColor: whiteTheme.textColor
  });
  if (whiteTheme.isDark || whiteTheme.textColor !== '#0f172a') {
    throw new Error('White theme contrast ratio failure');
  }

  console.log('\n======================================================');
  console.log('>>> ALL 4 DASHBOARD ENHANCEMENT TESTS PASSED! <<<');
  console.log('======================================================');
}

run().catch(err => {
  console.error('\n❌ Verification Failed:', err);
  process.exit(1);
});
