export function getRelativePath(fromPath, toPath) {
  // Convert both paths to arrays and remove empty segments
  const fromParts = fromPath.split("/").filter((part) => part.length > 0);
  const toParts = toPath.split("/").filter((part) => part.length > 0);

  // Find the common path prefix
  let commonPrefixLength = 0;
  const minLength = Math.min(fromParts.length, toParts.length);

  for (let i = 0; i < minLength; i++) {
    if (fromParts[i] === toParts[i]) {
      commonPrefixLength++;
    } else {
      break;
    }
  }

  // Build the relative path
  // First go up as needed
  const upCount = fromParts.length - commonPrefixLength;

  // Then go to the target directory
  const targetParts = toParts.slice(commonPrefixLength);

  // Combine everything
  const relativePath = [...Array(upCount).fill(".."), ...targetParts].join("/");

  // Handle empty result (same directory)
  return relativePath || ".";
}

/**
 * Gets the parent path from a given path string
 * @param {string} path - The path to extract the parent from
 * @returns {string} The parent path or an empty string if no parent exists
 */
export function getParentPath(path) {
  // Handle empty or invalid paths
  if (!path || typeof path !== "string") {
    return "";
  }

  // Normalize the path by replacing backslashes with forward slashes
  // This makes the function work with both Windows and Unix-style paths
  const normalizedPath = path.replace(/\\/g, "/");

  // Remove trailing slash if it exists
  const cleanPath = normalizedPath.endsWith("/")
    ? normalizedPath.slice(0, -1)
    : normalizedPath;

  // Find the last occurrence of '/'
  const lastSlashIndex = cleanPath.lastIndexOf("/");

  // If no slash found, there is no parent path
  if (lastSlashIndex === -1) {
    return "";
  }

  // Return everything before the last slash
  // If the slash is at the beginning (e.g., "/file"), return "/"
  return lastSlashIndex === 0 ? "/" : cleanPath.slice(0, lastSlashIndex);
}
