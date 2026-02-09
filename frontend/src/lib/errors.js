// Maps error codes to user-friendly messages.
const errorCodeMap = {
  // HTTP status codes
  400: "Bad request",
  401: "Authentication required",
  403: "Access denied",
  404: "Server not found",
  408: "Request timed out",
  429: "Too many requests, try again later",
  500: "Server error",
  502: "Server unreachable",
  503: "Server unavailable",
  504: "Server timed out",
  // Windows/network error codes
  1033: "Server unreachable",
  10053: "Connection lost",
  10054: "Connection reset",
  10060: "Connection timed out",
  10061: "Server unreachable",
};

// Patterns matched against the full error message string.
const errorPatternMap = [
  { pattern: /connection\s*(was\s*)?refused/i, message: "Server unreachable" },
  { pattern: /forcibly closed by the remote host/i, message: "Connection lost" },
  { pattern: /cannot .+ in offline mode/i, message: "Unavailable in offline mode" },
  { pattern: /no auth host configured/i, message: "Server not configured" },
  { pattern: /secret not found in keyring/i, message: "Please sign in" },
];

// Extracts a user-friendly message from raw error text.
// Returns null if no known pattern matches.
export function friendlyErrorMessage(raw) {
  if (!raw) return null;

  // Check for "error code: XXXX" pattern
  const codeMatch = raw.match(/error code:\s*(\d+)/i);
  if (codeMatch) {
    const code = parseInt(codeMatch[1]);
    if (errorCodeMap[code]) return errorCodeMap[code];
  }

  // Check for "code - XXX" pattern from backend HTTP calls
  const httpCodeMatch = raw.match(/code\s*-\s*(\d+)/i);
  if (httpCodeMatch) {
    const code = parseInt(httpCodeMatch[1]);
    if (errorCodeMap[code]) return errorCodeMap[code];
  }

  // Check for "status code XXX" or "status XXX" pattern
  const statusMatch = raw.match(/status\s*(?:code\s*)?(\d+)/i);
  if (statusMatch) {
    const code = parseInt(statusMatch[1]);
    if (errorCodeMap[code]) return errorCodeMap[code];
  }

  // Check string patterns
  for (const { pattern, message } of errorPatternMap) {
    if (pattern.test(raw)) return message;
  }

  return null;
}
