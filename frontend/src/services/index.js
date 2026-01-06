// Service abstraction layer
// This file re-exports the appropriate adapter based on build configuration
// Desktop (Wails): Uses adapters/wails.js -> Wails bindings
// Web: Uses adapters/http.js -> REST API calls

export * from './adapters/wails.js';
