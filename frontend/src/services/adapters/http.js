// =============================================================================
// HTTP ADAPTER - BACKWARD COMPATIBILITY LAYER
// =============================================================================
// This file re-exports all services from the modular structure for backward compatibility.
// New code should import directly from '@/services/adapters' or '@/services/adapters/index.js'
//
// The adapter has been split into separate modules:
// - config.js         - Configuration constants (API URLs, storage keys)
// - storage.js        - localStorage helpers, multi-account management
// - http-client.js    - Core HTTP utilities (globalApiCall, studioApiCall)
// - *.service.js      - Individual service implementations
// - index.js          - Re-exports all services
//
// Example usage:
//   import { AuthService, ProjectService } from '@/services/adapters';
//   import { AuthService } from '@/services/adapters/auth.service.js';

export * from './index.js';
