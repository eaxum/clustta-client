import { defineStore } from "pinia";
import { EntitlementService } from "@/services";

export const useEntitlementStore = defineStore("entitlements", {
  state: () => ({
    plan: 'free',
    planType: 'individual',
    status: 'active',
    limits: {
      storage_bytes: 0,
      max_remote_projects: 1,
      max_collaborators: 0,
      ai_credits_monthly: 0,
    },
    usage: {
      storage_bytes: 0,
      project_count: 0,
      ai_credits_used: 0,
    },
    features: [],
    plans: [],
    studioEntitlements: {},
    lastFetched: null,
    isLoading: false,
  }),
  getters: {
    canSync: (state) => state.features.includes('sync'),
    canUseAI: (state) => state.features.includes('ai'),
    canCollaborate: (state) => state.features.includes('collaboration'),
    canCreateRemoteProject: (state) => {
      if (state.limits.max_remote_projects === -1) return true;
      return state.usage.project_count < state.limits.max_remote_projects;
    },
    isOverStorage: (state) => {
      if (state.limits.storage_bytes <= 0) return false;
      return state.usage.storage_bytes >= state.limits.storage_bytes;
    },
    storagePercent: (state) => {
      if (state.limits.storage_bytes <= 0) return 0;
      return Math.min((state.usage.storage_bytes / state.limits.storage_bytes) * 100, 100);
    },
    isNearQuota: (state) => {
      if (state.limits.storage_bytes <= 0) return false;
      return (state.usage.storage_bytes / state.limits.storage_bytes) >= 0.9;
    },
    isPaidPlan: (state) => state.plan !== 'free',
    storageUsedFormatted: (state) => formatBytes(state.usage.storage_bytes),
    storageLimitFormatted: (state) => formatBytes(state.limits.storage_bytes),
  },
  actions: {
    // Fetches the current user's entitlements from the server.
    async fetchEntitlements() {
      if (this.isLoading) return;
      this.isLoading = true;
      try {
        const bundle = await EntitlementService.GetEntitlements();
        this.applyBundle(bundle);
        this.lastFetched = Date.now();
      } catch (error) {
        console.error('Failed to fetch entitlements:', error);
      } finally {
        this.isLoading = false;
      }
    },

    // Fetches entitlements for a specific studio.
    async fetchStudioEntitlements(studioId) {
      try {
        const bundle = await EntitlementService.GetStudioEntitlements(studioId);
        this.studioEntitlements[studioId] = bundle;
        return bundle;
      } catch (error) {
        console.error('Failed to fetch studio entitlements:', error);
        return null;
      }
    },

    // Applies an entitlement bundle from a login response or fetch.
    applyBundle(bundle) {
      if (!bundle) return;
      this.plan = bundle.plan || 'free';
      this.planType = bundle.plan_type || 'individual';
      this.status = bundle.status || 'active';
      this.limits = bundle.limits || this.limits;
      this.usage = bundle.usage || this.usage;
      this.features = bundle.features || [];
    },

    // Returns studio entitlements, fetching if not cached.
    async getStudioEntitlements(studioId) {
      if (this.studioEntitlements[studioId]) {
        return this.studioEntitlements[studioId];
      }
      return await this.fetchStudioEntitlements(studioId);
    },

    // Checks if a studio has a specific feature.
    studioHasFeature(studioId, feature) {
      const entitlements = this.studioEntitlements[studioId];
      if (!entitlements) return false;
      return entitlements.features?.includes(feature) || false;
    },

    // Resets entitlements to default free-tier state.
    reset() {
      this.plan = 'free';
      this.planType = 'individual';
      this.status = 'active';
      this.limits = { storage_bytes: 0, max_remote_projects: 1, max_collaborators: 0, ai_credits_monthly: 0 };
      this.usage = { storage_bytes: 0, project_count: 0, ai_credits_used: 0 };
      this.features = [];
      this.plans = [];
      this.studioEntitlements = {};
      this.lastFetched = null;
    },

    // Fetches all available plans from the server.
    async fetchPlans() {
      try {
        const plans = await EntitlementService.GetPlans();
        this.plans = plans || [];
        return this.plans;
      } catch (error) {
        console.error('Failed to fetch plans:', error);
        return [];
      }
    },

    // Changes the user's plan and applies the returned entitlement bundle.
    // Only works for free plan downgrades; paid upgrades use createCheckout.
    async changePlan(planId) {
      try {
        const bundle = await EntitlementService.ChangePlan(planId);
        this.applyBundle(bundle);
        this.lastFetched = Date.now();
        return true;
      } catch (error) {
        console.error('Failed to change plan:', error);
        return false;
      }
    },

    // Creates a Stripe Checkout Session for upgrading to a paid plan.
    async createCheckout(planId) {
      try {
        const url = await EntitlementService.CreateCheckout(planId);
        return url || '';
      } catch (error) {
        console.error('Failed to create checkout:', error);
        return '';
      }
    },

    // Opens the Stripe billing portal for managing subscriptions.
    async openBillingPortal() {
      try {
        const url = await EntitlementService.OpenBillingPortal();
        return url || '';
      } catch (error) {
        console.error('Failed to open billing portal:', error);
        return '';
      }
    },
  },
});

function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i];
}
