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

export function getApiBaseUrl() {
  const envUrl = import.meta.env.VITE_API_BASE_URL;
  return (envUrl || '').replace(/\/+$/, '');
}

export function resolveAvatarUrl(url) {
  if (!url || typeof url !== 'string') return '';
  const trimmed = url.trim();
  if (
    trimmed.startsWith('http://') ||
    trimmed.startsWith('https://') ||
    trimmed.startsWith('data:') ||
    trimmed.startsWith('blob:')
  ) {
    return trimmed;
  }
  const baseUrl = getApiBaseUrl();
  if (baseUrl && trimmed.startsWith('/')) {
    return `${baseUrl}${trimmed}`;
  }
  return trimmed;
}

export async function apiFetch(endpoint, options = {}) {
  const headers = new Headers(options.headers || {});

  // Do not send authorization headers on public routes unless explicitly requested
  if (!options.skipAuth && !endpoint.startsWith('/api/public')) {
    const token = getAuthToken();
    if (token && !headers.has('Authorization')) {
      headers.set('Authorization', `Bearer ${token}`);
    }
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
    const message = payload?.error || (typeof payload === 'string' ? payload : 'Request failed');
    const error = new Error(message);
    error.status = response.status;
    error.payload = payload;
    error.response = response;
    throw error;
  }

  return payload;
}

export default apiFetch;
