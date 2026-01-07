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
export async function studioApiCall(studioUrl, endpoint, method = 'GET', body = null, options = {}) {
  const user = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || '{}');
  
  const headers = {
    'Content-Type': 'application/json',
    'Clustta-Agent': CLUSTTA_AGENT,
    'UserId': user.id || '',
    'UserData': JSON.stringify(user),
    ...options.headers,
  };

  // In development, use the Vite proxy to bypass CORS
  let fetchUrl;
  if (isDev) {
    fetchUrl = `/studio-proxy${endpoint}`;
    headers['X-Studio-URL'] = studioUrl;
  } else {
    fetchUrl = `${studioUrl}${endpoint}`;
  }

  const response = await fetch(fetchUrl, {
    method,
    headers,
    body: body ? JSON.stringify(body) : null,
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
