// Verification script for Avatar Upload Feature & Security Controls
const BASE_URL = 'http://localhost:8080';

// 1x1 transparent PNG bytes
const samplePNG = Buffer.from([
  0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
  0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
  0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
  0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
  0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
  0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
]);

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
  return { status: res.status, ok: res.ok, headers: res.headers, data: payload };
}

async function run() {
  console.log('=== TEST 1: Unauthorized Avatar Upload Rejected ===');
  const unauthRes = await request('/api/me/avatar', { method: 'POST' });
  console.log('POST /api/me/avatar without token -> Status:', unauthRes.status);
  if (unauthRes.status !== 401) throw new Error(`Expected 401, got ${unauthRes.status}`);

  console.log('\n=== TEST 2: Register User & Obtain JWT Token ===');
  const timestamp = Date.now();
  const username = `avataruser-${timestamp}`;
  const email = `avataruser-${timestamp}@test.com`;
  const password = 'Password123!';

  const regRes = await request('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password }),
  });
  if (!regRes.ok) throw new Error('Registration failed: ' + JSON.stringify(regRes.data));

  const loginRes = await request('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!loginRes.ok) throw new Error('Login failed');
  const token = loginRes.data.token;
  const authHeaders = { Authorization: `Bearer ${token}` };

  console.log('\n=== TEST 3: Security - SVG Upload Rejected (Prevent XSS) ===');
  const svgBlob = new Blob([`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`], { type: 'image/svg+xml' });
  const svgForm = new FormData();
  svgForm.append('avatar', svgBlob, 'malicious.svg');

  const svgRes = await request('/api/me/avatar', {
    method: 'POST',
    headers: authHeaders,
    body: svgForm,
  });
  console.log('POST /api/me/avatar with SVG -> Status:', svgRes.status, 'Data:', svgRes.data);
  if (svgRes.status !== 400) throw new Error(`Expected 400 Bad Request for SVG upload, got ${svgRes.status}`);

  console.log('\n=== TEST 4: Security - Oversized File Rejected (>2MB) ===');
  const oversizeBuffer = Buffer.alloc(2 * 1024 * 1024 + 1024, 0x41);
  const oversizeBlob = new Blob([oversizeBuffer], { type: 'image/png' });
  const oversizeForm = new FormData();
  oversizeForm.append('avatar', oversizeBlob, 'large.png');

  const oversizeRes = await request('/api/me/avatar', {
    method: 'POST',
    headers: authHeaders,
    body: oversizeForm,
  });
  console.log('POST /api/me/avatar with >2MB file -> Status:', oversizeRes.status);
  if (oversizeRes.status !== 413 && oversizeRes.status !== 400) {
    throw new Error(`Expected 413 or 400 for oversize upload, got ${oversizeRes.status}`);
  }

  console.log('\n=== TEST 5: Successful PNG Avatar Upload ===');
  const pngBlob = new Blob([samplePNG], { type: 'image/png' });
  const pngForm = new FormData();
  pngForm.append('avatar', pngBlob, 'my-avatar.png');

  const uploadRes = await request('/api/me/avatar', {
    method: 'POST',
    headers: authHeaders,
    body: pngForm,
  });
  console.log('POST /api/me/avatar with PNG -> Status:', uploadRes.status, 'Data:', uploadRes.data);
  if (!uploadRes.ok) throw new Error('Valid avatar upload failed');
  const uploadedUrl = uploadRes.data.avatar_url;
  if (!uploadedUrl || !uploadedUrl.startsWith('/uploads/avatars/avatar-')) {
    throw new Error(`Unexpected avatar_url format: ${uploadedUrl}`);
  }

  console.log('\n=== TEST 6: Verify Avatar URL in GET /api/me ===');
  const meRes = await request('/api/me', { headers: authHeaders });
  console.log('GET /api/me page_settings.avatar_url:', meRes.data.page_settings.avatar_url);
  if (meRes.data.page_settings.avatar_url !== uploadedUrl) {
    throw new Error(`Mismatch in GET /api/me avatar_url: ${meRes.data.page_settings.avatar_url}`);
  }

  console.log('\n=== TEST 7: Verify Avatar URL in Public Page ===');
  const publicRes = await request(`/api/public/${username}`);
  console.log('GET /api/public/' + username + ' avatar_url:', publicRes.data.avatar_url);
  if (publicRes.data.avatar_url !== uploadedUrl) {
    throw new Error(`Mismatch in GET /api/public/${username} avatar_url: ${publicRes.data.avatar_url}`);
  }

  console.log('\n=== TEST 8: Verify File Download & Security Headers ===');
  const fileRes = await request(uploadedUrl);
  console.log('GET ' + uploadedUrl + ' -> Status:', fileRes.status);
  if (fileRes.status !== 200) throw new Error(`File fetch failed with status ${fileRes.status}`);
  const nosniff = fileRes.headers.get('x-content-type-options');
  const csp = fileRes.headers.get('content-security-policy');
  console.log('Security headers -> nosniff:', nosniff, '| csp:', csp);
  if (nosniff !== 'nosniff') throw new Error('Missing X-Content-Type-Options: nosniff');
  if (!csp) throw new Error('Missing Content-Security-Policy header');

  console.log('\n=== TEST 9: Delete Avatar Feature ===');
  const deleteRes = await request('/api/me/avatar', {
    method: 'DELETE',
    headers: authHeaders,
  });
  console.log('DELETE /api/me/avatar -> Status:', deleteRes.status, 'Data:', deleteRes.data);
  if (!deleteRes.ok) throw new Error('Delete avatar failed');
  if (deleteRes.data.avatar_url !== '') throw new Error('Expected empty avatar_url after delete');

  const meAfterDelete = await request('/api/me', { headers: authHeaders });
  if (meAfterDelete.data.page_settings.avatar_url !== '') {
    throw new Error(`Expected empty avatar_url in /api/me, got: ${meAfterDelete.data.page_settings.avatar_url}`);
  }

  // Verify file was cleaned up from disk and returns 404
  const fileAfterDelete = await request(uploadedUrl);
  console.log('GET file after deletion -> Status:', fileAfterDelete.status);
  if (fileAfterDelete.status !== 404) {
    throw new Error(`Expected 404 for deleted file, got ${fileAfterDelete.status}`);
  }

  console.log('\n======================================================');
  console.log('>>> ALL 9 AVATAR UPLOAD & SECURITY TESTS PASSED! <<<');
  console.log('======================================================');
}

run().catch((err) => {
  console.error('TEST FAILED:', err);
  process.exit(1);
});
