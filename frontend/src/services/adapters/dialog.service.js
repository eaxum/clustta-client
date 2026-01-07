// =============================================================================
// DIALOG SERVICE (Browser alternatives)
// =============================================================================

export const DialogService = {
  OpenFile: async (options) => {
    return new Promise((resolve) => {
      const input = document.createElement('input');
      input.type = 'file';
      if (options?.filters) {
        const extensions = options.filters.flatMap(f => f.extensions || []);
        input.accept = extensions.map(ext => `.${ext}`).join(',');
      }
      input.onchange = (e) => {
        const file = e.target.files?.[0];
        resolve(file ? file.name : null);
      };
      input.click();
    });
  },
  
  OpenDirectory: async (options) => {
    console.warn('OpenDirectory has limited support in web browsers');
    return null;
  },
  
  SaveFile: async (options) => {
    console.warn('SaveFile has limited support in web browsers');
    return null;
  },
  
  ShowMessage: async (options) => {
    alert(options?.message || options?.title || '');
  },
  
  ShowError: async (title, message) => {
    alert(`Error: ${title}\n${message}`);
  },
  
  ShowWarning: async (title, message) => {
    alert(`Warning: ${title}\n${message}`);
  },
  
  ShowInfo: async (title, message) => {
    alert(`${title}\n${message}`);
  },
  
  AskConfirmation: async (title, message) => {
    return confirm(`${title}\n${message}`);
  },
};
