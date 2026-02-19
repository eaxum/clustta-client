export const DragService = {
  // Initiates a native OS drag operation with the specified file paths.
  // Not available in web mode - returns error.
  StartNativeDrag: async (filePaths) => {
    console.warn('Native drag is not available in web mode');
    return 0;
  },

  // Checks if the left mouse button is currently pressed.
  // Returns false in web mode.
  IsMouseButtonDown: async () => {
    return false;
  },
};
