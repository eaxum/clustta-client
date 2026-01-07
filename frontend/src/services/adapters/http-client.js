// =============================================================================
// HTTP CLIENT
// =============================================================================
// Core HTTP utilities for making API calls

import { GLOBAL_API, CLUSTTA_AGENT, STORAGE_KEYS, isDev } from './config.js';

/**
 * Makes an API call to the global Clustta server
 */
export async function globalApiCall(endpoint, method = 'GET', body = null, options = {}) {
  // Note: We use credentials: 'include' to let the browser handle cookies automatically
  // Do NOT manually set Cookie header - browsers forbid it and it won't work
  
  const headers = {
    'Content-Type': 'application/json',
    'Clustta-Agent': CLUSTTA_AGENT,
    ...options.headers,
  };

  const response = await fetch(`${GLOBAL_API}${endpoint}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : null,
    credentials: 'include',  // This tells browser to send and receive cookies
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

  // Handle empty responses
  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return response.json();
  }
  
  // Try to parse as JSON even if content-type is not set correctly
  const text = await response.text();
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

/**
 * Makes an API call to a studio server
 * In development, uses Vite proxy to bypass CORS
 * In production, calls the studio server directly (requires CORS to be configured)
 */
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
  // The proxy reads X-Studio-URL header to know where to forward the request
  let fetchUrl;
  if (isDev) {
    fetchUrl = `/studio-proxy${endpoint}`;
    headers['X-Studio-URL'] = studioUrl;
  } else {
    // In production, call studio directly (assumes CORS is configured on studio servers)
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
  return response.text();
}

/**
 * Get the active studio URL from storage or settings
 */
export function getActiveStudioUrl() {
  return localStorage.getItem(STORAGE_KEYS.STUDIO_URL) || '';
}

/**
 * Set the active studio URL
 */
export function setActiveStudioUrl(url) {
  localStorage.setItem(STORAGE_KEYS.STUDIO_URL, url);
}
