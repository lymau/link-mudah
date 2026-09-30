// Verification script for Issue #10
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

async function runTests() {
  console.log('--- 1. Testing Unauthenticated Access ---');
  const unauthMe = await request('/api/me');
  console.log('GET /api/me without token -> Status:', unauthMe.status);
  if (unauthMe.status !== 401) throw new Error(`Expected 401, got ${unauthMe.status}`);

  const invalidTokenMe = await request('/api/me', {
    headers: { Authorization: 'Bearer invalid.jwt.token' }
  });
  console.log('GET /api/me with invalid token -> Status:', invalidTokenMe.status);
  if (invalidTokenMe.status !== 401) throw new Error(`Expected 401, got ${invalidTokenMe.status}`);

  console.log('\n--- 2. Register & Login ---');
  const timestamp = Date.now();
  const username = `dev-${timestamp}`;
  const email = `dev-${timestamp}@example.com`;
  const password = 'Password123!';

  const regRes = await request('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password })
  });
  console.log('Register new user:', username, '-> Status:', regRes.status);
  if (!regRes.ok) throw new Error(`Register failed: ${JSON.stringify(regRes.data)}`);

  const loginRes = await request('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  console.log('Login -> Status:', loginRes.status);
  if (!loginRes.ok) throw new Error(`Login failed: ${JSON.stringify(loginRes.data)}`);
  const token = loginRes.data.token;
  if (!token) throw new Error('No token returned from login');

  const authHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`
  };

  console.log('\n--- 3. Testing GET /api/me ---');
  const meRes = await request('/api/me', { headers: authHeaders });
  console.log('GET /api/me -> Status:', meRes.status);
  console.log('User data:', { username: meRes.data.username, email: meRes.data.email, linksCount: meRes.data.links?.length });
  if (meRes.data.username !== username) throw new Error(`Expected username ${username}, got ${meRes.data.username}`);

  console.log('\n--- 4. Testing PUT /api/me/settings ---');
  const newSettings = {
    bg_color: '#d1fae5',
    font_family: 'sans-serif',
    avatar_url: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150'
  };
  const settingsRes = await request('/api/me/settings', {
    method: 'PUT',
    headers: authHeaders,
    body: JSON.stringify(newSettings)
  });
  console.log('PUT /api/me/settings -> Status:', settingsRes.status, 'Response:', settingsRes.data);
  if (!settingsRes.ok || settingsRes.data.bg_color !== newSettings.bg_color) {
    throw new Error('Settings update failed or did not match response');
  }

  console.log('\n--- 5. Testing POST /api/me/links (Adding 2 links) ---');
  const link1Res = await request('/api/me/links', {
    method: 'POST',
    headers: authHeaders,
    body: JSON.stringify({ title: 'My GitHub', url: 'https://github.com/lymau' })
  });
  console.log('POST /api/me/links (1) -> Status:', link1Res.status, 'ID:', link1Res.data.id);
  if (!link1Res.ok || !link1Res.data.id) throw new Error('Create link 1 failed');
  const link1Id = link1Res.data.id;

  const link2Res = await request('/api/me/links', {
    method: 'POST',
    headers: authHeaders,
    body: JSON.stringify({ title: 'Portfolio', url: 'https://portfolio.example.com' })
  });
  console.log('POST /api/me/links (2) -> Status:', link2Res.status, 'ID:', link2Res.data.id);
  if (!link2Res.ok || !link2Res.data.id) throw new Error('Create link 2 failed');
  const link2Id = link2Res.data.id;

  console.log('\n--- 6. Testing PUT /api/me/links/{id} (Editing 1 link) ---');
  const editRes = await request(`/api/me/links/${link1Id}`, {
    method: 'PUT',
    headers: authHeaders,
    body: JSON.stringify({ title: 'GitHub Profile (Updated)', url: 'https://github.com/lymau/link-mudah' })
  });
  console.log('PUT /api/me/links/{id} -> Status:', editRes.status, 'Title:', editRes.data.title);
  if (!editRes.ok || editRes.data.title !== 'GitHub Profile (Updated)') {
    throw new Error('Edit link failed');
  }

  console.log('\n--- 7. Testing DELETE /api/me/links/{id} (Deleting 1 link) ---');
  const deleteRes = await request(`/api/me/links/${link2Id}`, {
    method: 'DELETE',
    headers: authHeaders
  });
  console.log('DELETE /api/me/links/{id} -> Status:', deleteRes.status);
  if (!deleteRes.ok) throw new Error('Delete link failed');

  console.log('\n--- 8. Verifying Final State via GET /api/me ---');
  const finalMe = await request('/api/me', { headers: authHeaders });
  console.log('Final links count:', finalMe.data.links?.length);
  if (finalMe.data.links?.length !== 1) throw new Error(`Expected 1 link left, got ${finalMe.data.links?.length}`);
  if (finalMe.data.links[0].title !== 'GitHub Profile (Updated)') {
    throw new Error(`Link title mismatch: ${finalMe.data.links[0].title}`);
  }

  console.log('\n--- 9. Verifying Public Page Endpoint GET /api/public/{username} ---');
  const publicRes = await request(`/api/public/${username}`);
  console.log('GET /api/public/' + username + ' -> Status:', publicRes.status);
  console.log('Public data:', {
    username: publicRes.data.username,
    bg_color: publicRes.data.bg_color,
    font_family: publicRes.data.font_family,
    links: publicRes.data.links
  });
  if (publicRes.data.bg_color !== newSettings.bg_color) throw new Error('Public page bg_color mismatch');
  if (publicRes.data.links.length !== 1) throw new Error('Public page links mismatch');

  console.log('\n========================================');
  console.log('>>> ALL VERIFICATION TESTS PASSED SUCCESSFULLY! <<<');
  console.log('========================================');
}

runTests().catch(err => {
  console.error('\n❌ Test failed:', err);
  process.exit(1);
});
