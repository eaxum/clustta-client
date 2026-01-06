// Runtime abstraction for @wailsio/runtime
// Provides mock implementations for web mode where Wails APIs are not available

// CancellablePromise class - mimics Wails' CancellablePromise
export class CancellablePromise extends Promise {
  constructor(executor) {
    let _cancel = () => {};
    super((resolve, reject) => {
      executor(resolve, reject, (cancelFn) => {
        _cancel = cancelFn;
      });
    });
    this._cancel = _cancel;
  }

  cancel() {
    if (this._cancel) {
      this._cancel();
    }
  }

  // Static helper to create a CancellablePromise from a regular promise
  static from(promise) {
    let cancelled = false;
    const cancellable = new CancellablePromise((resolve, reject, onCancel) => {
      onCancel(() => { cancelled = true; });
      promise.then(
        (value) => { if (!cancelled) resolve(value); },
        (error) => { if (!cancelled) reject(error); }
      );
    });
    return cancellable;
  }
}

// WailsEvent class for event handling
export class WailsEvent {
  constructor(name, data = null) {
    this.name = name;
    this.data = data;
  }
}

// Events - mock event system
const eventListeners = new Map();

export const Events = {
  On: (eventName, callback) => {
    if (!eventListeners.has(eventName)) {
      eventListeners.set(eventName, []);
    }
    eventListeners.get(eventName).push(callback);
    return () => Events.Off(eventName, callback);
  },
  Off: (eventName, callback) => {
    const listeners = eventListeners.get(eventName);
    if (listeners) {
      const index = listeners.indexOf(callback);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    }
  },
  Emit: (event) => {
    const eventName = typeof event === 'string' ? event : event.name;
    const eventData = typeof event === 'string' ? null : event.data;
    const listeners = eventListeners.get(eventName);
    if (listeners) {
      listeners.forEach(callback => callback({ name: eventName, data: eventData }));
    }
  },
  Once: (eventName, callback) => {
    const wrapper = (data) => {
      Events.Off(eventName, wrapper);
      callback(data);
    };
    Events.On(eventName, wrapper);
  },
  // Event types enum
  Types: {
    Windows: {
      WindowDragLeave: 'windows:window-drag-leave',
      WindowDragOver: 'windows:window-drag-over',
      WindowDrop: 'windows:window-drop',
    },
    Common: {
      WindowFocus: 'common:window-focus',
      WindowBlur: 'common:window-blur',
      WindowClosing: 'common:window-closing',
      WindowFilesDropped: 'common:window-files-dropped',
    },
  },
  WailsEvent,
};

// Window - mock window controls
export const Window = {
  Name: () => 'main',
  Center: async () => {},
  SetTitle: async (title) => { document.title = title; },
  Fullscreen: async () => {},
  UnFullscreen: async () => {},
  SetSize: async (width, height) => {},
  Size: async () => ({ width: window.innerWidth, height: window.innerHeight }),
  SetMaxSize: async (width, height) => {},
  SetMinSize: async (width, height) => {},
  SetPosition: async (x, y) => {},
  Position: async () => ({ x: 0, y: 0 }),
  Minimise: async () => {},
  Unminimise: async () => {},
  Maximise: async () => {},
  Unmaximise: async () => {},
  IsMaximised: async () => false,
  IsMinimised: async () => false,
  IsFullscreen: async () => false,
  Show: async () => {},
  Hide: async () => {},
  Close: async () => {},
  Focus: async () => {},
  ToggleMaximise: async () => {},
};

// System - mock system info
export const System = {
  IsDebug: false,
  IsDarkMode: () => window.matchMedia('(prefers-color-scheme: dark)').matches,
  Environment: () => ({
    platform: 'web',
    arch: navigator.platform,
    os: 'web',
  }),
};

// Clipboard - use browser clipboard API
export const Clipboard = {
  SetText: async (text) => {
    try {
      await navigator.clipboard.writeText(text);
    } catch (err) {
      console.warn('Clipboard write failed:', err);
    }
  },
  Text: async () => {
    try {
      return await navigator.clipboard.readText();
    } catch (err) {
      console.warn('Clipboard read failed:', err);
      return '';
    }
  },
};

// Browser - open URLs
export const Browser = {
  OpenURL: async (url) => {
    window.open(url, '_blank');
  },
};

// Screens - mock screen info
export const Screens = {
  GetAll: async () => [{
    id: 'primary',
    name: 'Primary',
    width: window.screen.width,
    height: window.screen.height,
    isPrimary: true,
  }],
  GetPrimary: async () => ({
    id: 'primary',
    name: 'Primary',
    width: window.screen.width,
    height: window.screen.height,
    isPrimary: true,
  }),
  GetCurrent: async () => ({
    id: 'primary',
    name: 'Primary',
    width: window.screen.width,
    height: window.screen.height,
    isPrimary: true,
  }),
};

// Application controls
export const Application = {
  Quit: async () => {
    console.warn('Application.Quit() called in web mode - no effect');
  },
  Hide: async () => {},
  Show: async () => {},
};

// Dialog - mock dialogs using browser APIs where possible
export const Dialog = {
  Info: async (options) => {
    alert(options.message || options.title);
  },
  Warning: async (options) => {
    alert(`⚠️ ${options.message || options.title}`);
  },
  Error: async (options) => {
    alert(`❌ ${options.message || options.title}`);
  },
  Question: async (options) => {
    return confirm(options.message || options.title);
  },
  OpenFile: async (options) => {
    // Would need a file input element for web
    console.warn('Dialog.OpenFile() not available in web mode');
    return null;
  },
  SaveFile: async (options) => {
    console.warn('Dialog.SaveFile() not available in web mode');
    return null;
  },
};

// Call - mock for Wails RPC calls (used by bindings)
// In web mode, this should not be called as we use HTTP adapter instead
export const Call = {
  ByID: (id, ...args) => {
    const promise = Promise.reject(
      new Error(`Wails RPC not available in web mode. Call.ByID(${id}) called with args: ${JSON.stringify(args)}`)
    );
    promise.cancel = () => {};
    console.warn(`Call.ByID(${id}) not available in web mode. Use HTTP adapter services.`);
    return promise;
  },
  ByName: (name, ...args) => {
    const promise = Promise.reject(
      new Error(`Wails RPC not available in web mode. Call.ByName(${name}) called.`)
    );
    promise.cancel = () => {};
    console.warn(`Call.ByName(${name}) not available in web mode. Use HTTP adapter services.`);
    return promise;
  },
};

// Create - mock for Wails object creation (used by bindings models)
export const Create = {
  Any: (source) => source,
  Array: (creator) => (source) => {
    if (!Array.isArray(source)) return [];
    return source.map(item => creator ? creator(item) : item);
  },
  Map: (keyCreator, valueCreator) => (source) => {
    if (!source || typeof source !== 'object') return new Map();
    const map = new Map();
    for (const [key, value] of Object.entries(source)) {
      const k = keyCreator ? keyCreator(key) : key;
      const v = valueCreator ? valueCreator(value) : value;
      map.set(k, v);
    }
    return map;
  },
  Nullable: (creator) => (source) => {
    if (source === null || source === undefined) return null;
    return creator ? creator(source) : source;
  },
  Struct: (creator) => (source) => {
    if (!source) return null;
    return creator ? creator(source) : source;
  },
};

// Default export for compatibility
export default {
  Events,
  Window,
  System,
  Clipboard,
  Browser,
  Screens,
  Application,
  Dialog,
  Call,
  Create,
  WailsEvent,
};
