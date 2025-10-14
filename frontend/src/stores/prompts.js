import { defineStore } from "pinia";
import { Events } from "@wailsio/runtime";

export const usePromptStore = defineStore("prompts", {
  state: () => ({
    prompts: [],
    currentPrompt: null,
  }),
  getters: {
    getPrompts: (state) => state.prompts,
    getCurrentPrompt: (state) => state.currentPrompt,
  },
  actions: {
    addPrompt(message, type = "info", timeout = 0) {
      let prompt = {
        id: Date.now() + Math.random(), // Simple unique ID
        message: message,
        type: type,
        timestamp: Date.now(),
        active: true,
      };
      
      this.prompts.unshift(prompt);
      this.currentPrompt = prompt;
      
      // Emit event for InfoBar to listen to
      let eventData = new Events.WailsEvent("add_prompt", prompt);
      Events.Emit(eventData);

      // Auto-clear prompt after timeout if specified
      if (timeout > 0) {
        setTimeout(() => {
          this.clearPrompt(prompt.id);
        }, timeout);
      }

      return prompt.id;
    },
    
    updatePrompt(id, message, type) {
      const prompt = this.prompts.find(p => p.id === id);
      if (prompt) {
        prompt.message = message;
        prompt.type = type;
        
        // Update current prompt if it's the one being updated
        if (this.currentPrompt && this.currentPrompt.id === id) {
          this.currentPrompt = prompt;
        }
        
        // Emit update event
        let eventData = new Events.WailsEvent("update_prompt", prompt);
        Events.Emit(eventData);
      }
    },
    
    clearPrompt(id) {
      const index = this.prompts.findIndex(p => p.id === id);
      if (index !== -1) {
        const prompt = this.prompts[index];
        prompt.active = false;
        
        // Clear current prompt if it matches
        if (this.currentPrompt && this.currentPrompt.id === id) {
          this.currentPrompt = null;
        }
        
        // Emit clear event
        let eventData = new Events.WailsEvent("clear_prompt", prompt);
        Events.Emit(eventData);
        
        // Remove from array after a short delay to allow for animations
        setTimeout(() => {
          const currentIndex = this.prompts.findIndex(p => p.id === id);
          if (currentIndex !== -1) {
            this.prompts.splice(currentIndex, 1);
          }
        }, 300);
      }
    },
    
    clearAllPrompts() {
      this.prompts.forEach(prompt => {
        prompt.active = false;
      });
      this.currentPrompt = null;
      
      // Emit clear all event
      let eventData = new Events.WailsEvent("clear_all_prompts", {});
      Events.Emit(eventData);
      
      // Clear array after delay
      setTimeout(() => {
        this.prompts = [];
      }, 300);
    },
    
    setCurrentPrompt(message, type = "info") {
      // Clear any existing current prompt
      if (this.currentPrompt) {
        this.clearPrompt(this.currentPrompt.id);
      }
      
      // Add new prompt
      return this.addPrompt(message, type);
    }
  },
});