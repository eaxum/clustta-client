// =============================================================================
// CLIPBOARD SERVICE
// =============================================================================

export const ClipboardService = {
  WriteText: async (text) => {
    try {
      await navigator.clipboard.writeText(text);
    } catch (err) {
      console.error('Clipboard write failed:', err);
    }
  },
  
  ReadText: async () => {
    try {
      return await navigator.clipboard.readText();
    } catch (err) {
      console.error('Clipboard read failed:', err);
      return '';
    }
  },
};
