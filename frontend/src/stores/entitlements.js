import { defineStore } from "pinia";
import { EntitlementService, StudioService } from "@/services";
import { useProjectStore } from "@/stores/projects";
import utils from "@/services/utils";

let entitlementFetchPromise = null;

// Builds a synthetic bundle for self-hosted (private) studios.
function createPrivateStudioBundle(usage = {}, usageUnavailable = false) {
  return {
    plan: 'private',
    plan_type: 'studio',
    status: 'active',
    limits: {
      storage_bytes: usage.storage_total_bytes || -1,
      max_remote_projects: -1,
      max_collaborators: -1,
      ai_credits_monthly: 0,
    },
    usage: {
      storage_bytes: usage.storage_bytes || 0,
      project_count: usage.project_count || 0,
      ai_credits_used: 0,
      storage_available_bytes: usage.storage_available_bytes || 0,
      storage_total_bytes: usage.storage_total_bytes || 0,
    },
    usage_unavailable: usageUnavailable,
    features: ['sync', 'collaboration', 'custom_roles', 'integrations'],
  };
}

// Synthetic fallback for self-hosted (private) studios.
const PRIVATE_STUDIO_BUNDLE = Object.freeze({
  ...createPrivateStudioBundle(),
  limits: Object.freeze({
    storage_bytes: -1,
    max_remote_projects: -1,
    max_collaborators: -1,
    ai_credits_monthly: 0,
  }),
  usage: Object.freeze({ storage_bytes: 0, project_count: 0, ai_credits_used: 0 }),
  features: Object.freeze(['sync', 'collaboration', 'custom_roles', 'integrations']),
});

function isPrivateStudio(studio) {
  return !!studio && studio.name !== 'Personal' && studio.hosting_mode && studio.hosting_mode !== 'cloud';
}

export const useEntitlementStore = defineStore("entitlements", {
  state: () => ({
    plan: 'free',
    planType: 'individual',
    status: 'active',
    limits: {
      storage_bytes: 0,
      max_remote_projects: 0,
      max_collaborators: 0,
      ai_credits_monthly: 0,
    },
    usage: {
      storage_bytes: 0,
      project_count: 0,
      ai_credits_used: 0,
    },
    features: [],
    effectiveFeatures: [],
    plans: [],
    studioEntitlements: {},
    lastFetched: null,
    isLoading: false,
  }),
  getters: {
    // Returns the resolved bundle for the current studio context.
    // Personal/global → user's own bundle from state.
    // Private studio → synthetic unlimited bundle (global server has no authority).
    // Cloud studio → cached studioEntitlements[id] (or empty placeholder).
    activeBundle() {
      const projectStore = useProjectStore();
      const studio = projectStore.selectedStudio;
      if (isPrivateStudio(studio)) return this.studioEntitlements[studio.id] || PRIVATE_STUDIO_BUNDLE;
      if (studio && studio.name !== 'Personal' && studio.id) {
        return this.studioEntitlements[studio.id] || { features: [], limits: this.limits, usage: this.usage };
      }
      return { features: this.features, limits: this.limits, usage: this.usage };
    },
    activeFeatures() { return this.activeBundle.features || []; },
    activeLimits() { return this.activeBundle.limits || this.limits; },
    activeUsage() { return this.activeBundle.usage || this.usage; },
    canSync() { return this.activeFeatures.includes('sync'); },
    canUseAI: (state) => state.features.includes('ai'),
    canCollaborate() { return this.activeFeatures.includes('collaboration'); },
    canCreateRemoteProject() {
      const limits = this.activeLimits;
      const usage = this.activeUsage;
      if (limits.max_remote_projects === -1) return true;
      return usage.project_count < limits.max_remote_projects;
    },
    canDiscoverTalent() { return this.effectiveFeatures.includes('talent_discovery'); },
    canShareLink() { return this.activeFeatures.includes('share_link'); },
    hasCustomRoles() { return this.activeFeatures.includes('custom_roles'); },
    hasIntegrations() { return this.activeFeatures.includes('integrations'); },
    isOverStorage() {
      const limits = this.activeLimits;
      const usage = this.activeUsage;
      if (limits.storage_bytes <= 0) return false;
      return usage.storage_bytes >= limits.storage_bytes;
    },
    storagePercent() {
      const limits = this.activeLimits;
      const usage = this.activeUsage;
      if (limits.storage_bytes <= 0) return 0;
      return Math.min((usage.storage_bytes / limits.storage_bytes) * 100, 100);
    },
    isNearQuota() {
      const limits = this.activeLimits;
      const usage = this.activeUsage;
      if (limits.storage_bytes <= 0) return false;
      return (usage.storage_bytes / limits.storage_bytes) >= 0.9;
    },
    isPaidPlan: (state) => state.plan !== 'free',
    isStudioActive() {
      const projectStore = useProjectStore();
      const studio = projectStore.selectedStudio;
      if (!studio || studio.name === 'Personal') return true;
      // Only cloud-hosted studios can be deactivated (Stripe / admin toggle).
      // Private/self-hosted studios have no such concept and are always active.
      if (studio.hosting_mode && studio.hosting_mode !== 'cloud') return true;
      return studio.active !== false;
    },
    storageUsedFormatted() { return utils.formatBytes(this.activeUsage.storage_bytes, 2); },
    storageLimitFormatted() { return utils.formatBytes(this.activeLimits.storage_bytes, 0); },
  },
  actions: {
    // Fetches the current user's entitlements from the server.
    async fetchEntitlements() {
      if (entitlementFetchPromise) return entitlementFetchPromise;

      entitlementFetchPromise = (async () => {
        this.isLoading = true;
        try {
          const bundle = await EntitlementService.GetEntitlements();
          this.applyBundle(bundle);
          this.lastFetched = Date.now();
          return bundle;
        } catch (error) {
          console.error('Failed to fetch entitlements:', error);
          return null;
        } finally {
          this.isLoading = false;
        }
      })();

      try {
        return await entitlementFetchPromise;
      } finally {
        entitlementFetchPromise = null;
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

    // Fetches VM-local usage for a private/self-hosted studio.
    async fetchPrivateStudioUsage(studio) {
      if (!studio?.id) return null;
      try {
        const projectStore = useProjectStore();
        const studioUrl = await projectStore.resolveStudioUrl(studio);
        const usage = await StudioService.GetStudioUsage(studioUrl);
        const bundle = createPrivateStudioBundle(usage, false);
        this.studioEntitlements[studio.id] = bundle;
        return bundle;
      } catch (error) {
        console.error('Failed to fetch private studio usage:', error);
        const bundle = createPrivateStudioBundle({}, true);
        this.studioEntitlements[studio.id] = bundle;
        return bundle;
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
      this.effectiveFeatures = bundle.effective_features || bundle.features || [];
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

    // Checks if the current context (user or active studio) has a feature.
    hasFeature(feature, studioId) {
      if (this.features.includes(feature)) return true;
      if (studioId) return this.studioHasFeature(studioId, feature);
      return Object.values(this.studioEntitlements).some(
        (e) => e.features?.includes(feature)
      );
    },

    // Checks account-wide access granted by the user's plan or studio memberships.
    hasEffectiveFeature(feature) {
      return this.effectiveFeatures.includes(feature);
    },

    // Resets entitlements to default free-tier state.
    reset() {
      this.plan = 'free';
      this.planType = 'individual';
      this.status = 'active';
      this.limits = { storage_bytes: 0, max_remote_projects: 1, max_collaborators: 0, ai_credits_monthly: 0 };
      this.usage = { storage_bytes: 0, project_count: 0, ai_credits_used: 0 };
      this.features = [];
      this.effectiveFeatures = [];
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
    async createCheckout(planId, studioId = '') {
      try {
        const url = await EntitlementService.CreateCheckout(planId, studioId);
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
