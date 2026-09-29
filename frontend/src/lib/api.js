export function getAuthToken() {
  if (typeof window === 'undefined') {
    return '';
  }

  return localStorage.getItem('authToken') || '';
}

export function setAuthToken(token) {
  if (typeof window === 'undefined') {
    return;
  }

  if (token) {
    localStorage.setItem('authToken', token);
    return;
  }

  localStorage.removeItem('authToken');
}

export function clearAuthToken() {
  if (typeof window === 'undefined') {
    return;
  }

  localStorage.removeItem('authToken');
}

async function parseResponse(response) {
  const contentType = response.headers.get('content-type') || '';

  if (response.status === 204) {
    return null;
  }

  if (contentType.includes('application/json')) {
    const text = await response.text();
    return text ? JSON.parse(text) : null;
  }

  return response.text();
}

function getApiBaseUrl() {
  const envUrl = import.meta.env.VITE_API_BASE_URL;

  // In browser, if we are browsing the Vite dev server, prefer relative path
  // so requests go through Vite's dev server proxy to avoid cross-origin / localhost resolution issues
  if (typeof window !== 'undefined') {
    if (!envUrl || envUrl.includes('localhost:8080') || envUrl.includes('127.0.0.1:8080')) {
      return '';
    }
  }

  return envUrl || '';
}

export async function apiFetch(endpoint, options = {}) {
  const headers = new Headers(options.headers || {});
  const token = getAuthToken();

  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  if (
    options.body !== undefined &&
    options.body !== null &&
    !(options.body instanceof FormData) &&
    !headers.has('Content-Type')
  ) {
    headers.set('Content-Type', 'application/json');
  }

  const baseUrl = getApiBaseUrl();
  const url = `${baseUrl}${endpoint}`;

  let response;
  try {
    response = await fetch(url, {
      ...options,
      headers,
    });
  } catch (err) {
    // If request to full URL failed due to network / CORS / resolution, retry with relative path via dev proxy
    if (baseUrl && (err.name === 'TypeError' || err.message?.includes('NetworkError') || err.message?.includes('Failed to fetch'))) {
      try {
        response = await fetch(endpoint, {
          ...options,
          headers,
        });
      } catch (retryErr) {
        throw err;
      }
    } else {
      throw err;
    }
  }

  const payload = await parseResponse(response);

  if (!response.ok) {
    const message = payload?.error || payload || 'Request failed';
    throw new Error(message);
  }

  return payload;
}

export default apiFetch;
