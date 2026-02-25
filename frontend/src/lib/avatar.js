// DiceBear avatar generation utilities.
import { createAvatar } from '@dicebear/core';
import * as bigSmile from '@dicebear/big-smile';

// Cache for generated avatar data URIs to avoid regenerating.
const avatarCache = new Map();

// Generates a DiceBear avatar SVG data URI from a seed string.
// Uses the big-smile style with transparent background.
export function generateAvatar(seed) {
  if (!seed) return null;
  
  if (avatarCache.has(seed)) {
    return avatarCache.get(seed);
  }

  const avatar = createAvatar(bigSmile, {
    seed: seed,
    backgroundColor: ['transparent'],
  });

  const dataUri = avatar.toDataUri();
  avatarCache.set(seed, dataUri);
  
  return dataUri;
}

// Clears the avatar cache (useful for memory management).
export function clearAvatarCache() {
  avatarCache.clear();
}
