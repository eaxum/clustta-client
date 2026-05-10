import { defineStore } from "pinia";

// Pending tool-call approval requests from the agent.
// Multiple destructive calls in one assistant turn are queued and shown one at a time.
export const useAgentApprovalStore = defineStore("agentApproval", {
  state: () => ({
    queue: [],
  }),

  getters: {
    current: (state) => (state.queue.length > 0 ? state.queue[0] : null),
    pendingCount: (state) => state.queue.length,
  },

  actions: {
    // enqueueRequest adds a new approval request payload from the backend.
    // payload shape: { id, tool, risk, args, preview }
    enqueueRequest(payload) {
      if (!payload || !payload.id) return;
      this.queue.push(payload);
    },

    // dequeueCurrent removes the head of the queue (called after a decision is made).
    dequeueCurrent() {
      if (this.queue.length > 0) this.queue.shift();
    },

    // clear empties the queue (used on agent cancel/done).
    clear() {
      this.queue = [];
    },
  },
});
