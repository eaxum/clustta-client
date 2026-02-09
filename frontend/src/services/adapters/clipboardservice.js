export const ClipboardService = {
  // Writes text to the system clipboard
  WriteText: async (text) => {
    try {
      await navigator.clipboard.writeText(text);
    } catch (err) {
      console.error('Clipboard write failed:', err);
    }
  },

  // Reads text from the system clipboard
  ReadText: async () => {
    try {
      return await navigator.clipboard.readText();
    } catch (err) {
      console.error('Clipboard read failed:', err);
      return '';
    }
  },
};
