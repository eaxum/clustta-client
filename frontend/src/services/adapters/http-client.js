import { GLOBAL_API, CLUSTTA_AGENT, STORAGE_KEYS, isDev } from './config.js';

// Makes an API call to the global Clustta server
export async function globalApiCall(endpoint, method = 'GET', body = null, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    'Clustta-Agent': CLUSTTA_AGENT,
    ...options.headers,
  };

  // Uses credentials: 'include' to let the browser handle cookies automatically
  const response = await fetch(`${GLOBAL_API}${endpoint}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : null,
    credentials: 'include',
  });

  if (!response.ok) {
    let errorMessage;
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || response.statusText;
    } catch {
      errorMessage = await response.text() || response.statusText;
    }
    throw new Error(errorMessage);
  }

  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return response.json();
  }
  
  const text = await response.text();
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

// Makes an API call to a studio server
// In production, routes through the global API proxy to handle HTTP studios from HTTPS app
export async function studioApiCall(studioUrl, endpoint, method = 'GET', body = null, options = {}) {
  const user = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || '{}');
  
  const headers = {
    'Content-Type': 'application/json',
    'Clustta-Agent': CLUSTTA_AGENT,
    'UserId': user.id || '',
    'UserData': JSON.stringify(user),
    ...options.headers,
  };

  let fetchUrl;
  
  // Check if studio URL is already HTTPS (can call directly)
  const isStudioHttps = studioUrl.startsWith('https://');
  
  if (isDev) {
    // In development, use the Vite proxy
    fetchUrl = `/studio-proxy${endpoint}`;
    headers['X-Studio-URL'] = studioUrl;
  } else if (isStudioHttps) {
    // Studio has HTTPS, call directly
    fetchUrl = `${studioUrl}${endpoint}`;
  } else {
    // Studio is HTTP, proxy through global API to avoid mixed content
    fetchUrl = `${GLOBAL_API}/v1/studio-proxy`;
    headers['X-Studio-URL'] = `${studioUrl}${endpoint}`;
    headers['X-Studio-Method'] = method;
  }

  const response = await fetch(fetchUrl, {
    method: (!isDev && !isStudioHttps) ? 'POST' : method, // Proxy always uses POST
    headers,
    body: body ? JSON.stringify(body) : null,
    credentials: isDev ? 'omit' : 'include',
  });

  if (!response.ok) {
    let errorMessage;
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || response.statusText;
    } catch {
      errorMessage = await response.text() || response.statusText;
    }
    throw new Error(errorMessage);
  }

  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return response.json();
  }

  const text = await response.text();
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

// Returns the active studio URL from storage
export function getActiveStudioUrl() {
  return localStorage.getItem(STORAGE_KEYS.STUDIO_URL) || '';
}

// Sets the active studio URL in storage
export function setActiveStudioUrl(url) {
  localStorage.setItem(STORAGE_KEYS.STUDIO_URL, url);
}

// Returns the session ID from storage
export function getSessionId() {
  return localStorage.getItem(STORAGE_KEYS.SESSION_ID) || '';
}

/**
 * Fetch binary data from studio server (for protobuf endpoints like /data)
 * Uses special /studio-data proxy that converts POST → GET-with-body
 * Returns ArrayBuffer instead of JSON
 */
export async function studioDataFetch(studioUrl, endpoint, method = 'POST', body = null, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    'Clustta-Agent': CLUSTTA_AGENT,
    ...options.headers,
  };

  // In development, use the special /studio-data proxy
  // This proxy converts browser POST → server GET with body (Node.js can do this)
  let fetchUrl;
  if (isDev) {
    // Use /studio-data proxy for data endpoints (handles GET-with-body limitation)
    fetchUrl = `/studio-data${endpoint}`;
    headers['X-Studio-URL'] = studioUrl;
  } else if (studioUrl.startsWith('http://')) {
    // For HTTP URLs in production, route through our API proxy to avoid mixed content
    fetchUrl = `${GLOBAL_API}/v1/studio-proxy`;
    headers['X-Studio-URL'] = `${studioUrl}${endpoint}`;
    headers['X-Studio-Method'] = 'GET'; // Studio server expects GET with body
  } else {
    // HTTPS URLs can be called directly
    fetchUrl = `${studioUrl}${endpoint}`;
  }

  console.log('[studioDataFetch] Request:', { fetchUrl, method, body });

  const response = await fetch(fetchUrl, {
    method: 'POST', // Always POST to proxy - proxy converts to GET with body
    headers,
    body: body ? JSON.stringify(body) : null,
  });

  if (!response.ok) {
    let errorMessage;
    try {
      errorMessage = await response.text();
      console.error('[studioDataFetch] Server error:', response.status, errorMessage);
    } catch {
      errorMessage = response.statusText;
    }
    throw new Error(errorMessage);
  }

  // Return as ArrayBuffer for binary data
  return response.arrayBuffer();
}
