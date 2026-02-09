export const DialogService = {
  // Opens a file picker dialog
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

  // Opens a directory picker dialog
  OpenDirectory: async (options) => {
    console.warn('OpenDirectory has limited support in web browsers');
    return null;
  },

  // Opens a save file dialog
  SaveFile: async (options) => {
    console.warn('SaveFile has limited support in web browsers');
    return null;
  },

  // Shows a message dialog
  ShowMessage: async (options) => {
    alert(options?.message || options?.title || '');
  },

  // Shows an error dialog
  ShowError: async (title, message) => {
    alert(`Error: ${title}\n${message}`);
  },

  // Shows a warning dialog
  ShowWarning: async (title, message) => {
    alert(`Warning: ${title}\n${message}`);
  },

  // Shows an info dialog
  ShowInfo: async (title, message) => {
    alert(`${title}\n${message}`);
  },

  // Shows a confirmation dialog
  AskConfirmation: async (title, message) => {
    return confirm(`${title}\n${message}`);
  },
};
