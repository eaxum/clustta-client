<template>
  <div class="modal-container cloud-modal" v-stop-propagation>
    <HeaderArea title="ClusttaCloud" icon="clustta" :notModal="false">
      <template v-if="canManageBilling" #actions>
        <GeneralButton class="billing-header-button" label="Manage Billing" :icon="getAppIcon('square-arrow-right-up')" :colored="false" :fullWidth="false" :buttonFunction="openBillingPortal" />
      </template>
    </HeaderArea>

    <div class="cloud-modal-body">
      <div v-if="isLoadingPlans" class="cloud-loading">Loading plans...</div>

      <div v-else class="plan-cards" :class="`plan-cards-${comparisonPlans.length}`">
        <div v-for="plan in comparisonPlans" :key="plan.id" class="plan-card" :class="{ 'plan-card-current': plan.name === currentPlanName, 'plan-card-highlighted': isRecommended(plan) }">

          <div class="plan-card-header">
            <span class="plan-card-name">{{ formatPlanName(plan.name) }}</span>
            <span v-if="isRecommended(plan)" class="plan-badge">Recommended</span>
          </div>

          <div class="plan-card-price">
            <template v-if="plan.price_cents > 0">
              <span class="price-amount">${{ (plan.price_cents / 100) }}</span>
              <span class="price-period">/mo</span>
            </template>
            <template v-else-if="plan.type === 'studio'">
              <span class="price-amount">Custom</span>
            </template>
            <template v-else>
              <span class="price-amount">Free</span>
            </template>
          </div>

          <GeneralButton :class="{ 'plan-button-current': plan.name === currentPlanName }" :label="planButtonLabel(plan)" :colored="plan.name !== currentPlanName" :isActive="plan.name !== currentPlanName && !isChanging" :loading="isChanging && changingPlanId === plan.id" :fullWidth="true" :buttonFunction="() => selectPlan(plan)" />

          <div class="plan-card-features">
            <div class="plan-feature" v-for="feature in planCardFeatures(plan)" :key="feature.key" v-tooltip="feature.tooltip">
              <img class="feature-icon small-icons" :src="getAppIcon(feature.icon)" alt="" aria-hidden="true" />
              <span>
                {{ feature.label }}
                <StatusBadge v-if="feature.comingSoon" :text="$t('settings.comingSoon')" class="feature-coming-soon-badge" />
              </span>
            </div>
          </div>
        </div>
      </div>

      <section v-if="commonPlanFeatures.length" class="common-features">
        <h3 class="common-features-title">Every plan includes</h3>
        <div class="common-features-grid">
          <div v-for="feature in commonPlanFeatures" :key="feature.key" class="plan-feature" v-tooltip="feature.tooltip">
            <img class="feature-icon small-icons" :src="getAppIcon(feature.icon)" alt="" aria-hidden="true" />
            <span>
              {{ feature.label }}
              <StatusBadge v-if="feature.comingSoon" :text="$t('settings.comingSoon')" class="feature-coming-soon-badge" />
            </span>
          </div>
        </div>
      </section>

      <section v-if="enterprisePlan" class="enterprise-section">
        <h3 class="enterprise-title">Enterprise</h3>
        <p class="enterprise-price">Contact us for pricing</p>
        <p class="enterprise-description">
          Customize Clustta for your studio with enterprise-grade security, dedicated infrastructure, custom integrations, and priority support from our team.
        </p>
        <GeneralButton class="enterprise-button" label="Contact us" :colored="false" :fullWidth="false" :buttonFunction="() => selectPlan(enterprisePlan)" />
      </section>

      <p v-if="showComingSoonFootnote" id="cloud-coming-soon-features" class="coming-soon-footnote" role="note">
        <StatusBadge :text="$t('settings.comingSoon')" class="coming-soon-footnote-badge" />
        These features are currently in development and will be available in a future release.
      </p>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { Browser } from '@wailsio/runtime';

// components
import StatusBadge from '@/instances/common/components/StatusBadge.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs
const activeTab = ref('individual');
const changingPlanId = ref(null);
const isChanging = ref(false);
const isLoadingPlans = ref(false);

// constants
const featureTooltips = {
  storage: 'Cloud storage for your project files and assets',
  remote_projects: 'Projects hosted on ClusttaCloud for remote access',
  collaborators: 'Number of people who can collaborate on each project',
  checkpoints: 'Save versioned snapshots of your assets to track changes over time',
  workflows: 'Reusable templates of assets and collections for recurring work',
  status: 'Track task progress with customizable statuses like To Do, In Progress, Done',
  share_link: 'Generate shareable links to assets and projects for external viewing',
  dependencies: 'Define relationships between assets so changes propagate correctly',
  sync: 'Synchronize your project data to ClusttaCloud across devices',
  resumable_transfers: 'Resume interrupted uploads and downloads without starting over',
  project_templates: 'Start new projects from predefined structure templates',
  kanban: 'Visual board view for managing tasks across columns',
  tags: 'Label and categorize assets and collections for easy filtering',
  interactive_console: 'In-app command console for advanced operations',
  ignore_list: 'Exclude specific files or folders from version tracking',
  ai: 'AI-powered assistant for creative workflows',
  ai_credits: 'Monthly credits for AI-assisted features',
  custom_roles: 'Define custom permission roles beyond the defaults',
  integrations: 'Connect third-party services like Slack, Jira, and more',
  talent_discovery: 'Find creators and review their work from Clustta profiles',
  web_dashboard: 'Manage your studio and projects from the web',
  discord_support: 'Community support through the Clustta Discord server',
  audit_log: 'Track all user actions and changes for compliance and accountability',
  priority_support: 'Dedicated priority support channel with faster response times',
  dedicated_vm: 'Isolated managed infrastructure dedicated to your studio',
  sso_saml: 'Enterprise single sign-on via SAML for centralized authentication',
  two_factor_auth: 'Enforced two-factor authentication for all studio members',
  custom_branding: 'White-label the experience with your studio brand and logo',
  unlimited_users: 'No limit on the number of collaborators across all projects',
  unlimited_projects: 'No limit on the number of remote projects',
};

const featureIcons = {
  storage: 'drive',
  remote_projects: 'folder-arrow-up-right',
  collaborators: 'two-persons',
  checkpoints: 'checkpoint-stone',
  workflows: 'workflow-arrow',
  status: 'status-ready',
  share_link: 'link',
  dependencies: 'dependency',
  sync: 'cloud-sync',
  resumable_transfers: 'arrow-up-ramp',
  project_templates: 'layers',
  kanban: 'kanban',
  tags: 'tag',
  interactive_console: 'console',
  ignore_list: 'eye-cancel',
  integrations: 'plug',
  custom_roles: 'key',
  talent_discovery: 'person-search',
  web_dashboard: 'website',
  discord_support: 'two-persons',
  audit_log: 'clipboard',
  priority_support: 'bell',
  dedicated_vm: 'stall-cog',
  sso_saml: 'key',
  two_factor_auth: 'lock-closed',
  custom_branding: 'palette',
  ai: 'sparkles',
  ai_credits: 'diamond',
};

const comingSoonFeatureKeys = new Set([
  'audit_log',
  'sso_saml',
  'two_factor_auth',
  'custom_branding',
]);

// computed
// Returns whether the selected studio is a cloud studio.
const isCloudStudio = computed(() => {
  return projectStore.isCloudHosted;
});

// Returns the current plan name based on context (personal or studio).
const currentPlanName = computed(() => {
  if (activeTab.value === 'studio' && isCloudStudio.value) {
    const bundle = entitlementStore.studioEntitlements[projectStore.selectedStudio?.id];
    return bundle?.plan || 'free';
  }
  return entitlementStore.plan;
});

// Returns plans filtered by the active tab category.
const filteredPlans = computed(() => {
  const type = activeTab.value === 'individual' ? 'individual' : 'studio';
  return entitlementStore.plans
    .filter(p => p.type === type)
    .sort((a, b) => a.display_order - b.display_order);
});

// Keeps enterprise as a dedicated section, matching the website pricing page.
const comparisonPlans = computed(() => {
  return filteredPlans.value.filter(plan => plan.name !== 'studio_enterprise');
});

const enterprisePlan = computed(() => {
  return filteredPlans.value.find(plan => plan.name === 'studio_enterprise') || null;
});

// Returns whether the user can open the billing portal for the current context.
const canManageBilling = computed(() => {
  return entitlementStore.isPaidPlan || isCloudStudio.value;
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns an icon from the user's selected app icon scheme.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Returns a human-readable plan name.
const formatPlanName = (name) => {
  return name.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
};

// Formats bytes to human-readable storage string.
const formatStorage = (bytes) => {
  if (bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return Math.round(bytes / Math.pow(1024, i)) + ' ' + units[i];
};

// Returns whether a plan is the recommended one.
const isRecommended = (plan) => {
  if (plan.type === 'individual') return plan.name === 'starter';
  return plan.name === 'studio_cloud';
};

// Returns the button label for a plan card.
const planButtonLabel = (plan) => {
  if (plan.name === currentPlanName.value) return 'Current Plan';
  if (!plan.price_cents && plan.type === 'studio') return 'Contact Sales';
  const currentOrder = entitlementStore.plans.find(p => p.name === currentPlanName.value)?.display_order ?? 0;
  return plan.display_order > currentOrder ? 'Upgrade' : 'Downgrade';
};

// Returns the feature list with labels and tooltips for a plan card.
const planFeatures = (plan) => {
  const features = [];
  const keys = plan.feature_keys || [];

  // Quantitative features first
  if (plan.storage_bytes > 0) {
    features.push({ key: 'storage', label: formatStorage(plan.storage_bytes) + ' storage', tooltip: featureTooltips.storage });
  } else if (plan.storage_bytes === 0) {
    features.push({ key: 'storage', label: 'No cloud storage', tooltip: featureTooltips.storage });
  } else {
    features.push({ key: 'storage', label: 'Unlimited storage', tooltip: featureTooltips.storage });
  }

  if (plan.max_remote_projects === -1) {
    features.push({ key: 'remote_projects', label: 'Unlimited remote projects', tooltip: featureTooltips.unlimited_projects });
  } else if (plan.max_remote_projects > 0) {
    features.push({ key: 'remote_projects', label: plan.max_remote_projects + ' remote project' + (plan.max_remote_projects !== 1 ? 's' : ''), tooltip: featureTooltips.remote_projects });
  } else {
    features.push({ key: 'remote_projects', label: '0 remote projects', tooltip: featureTooltips.remote_projects });
  }

  if (plan.max_collaborators === -1) {
    features.push({ key: 'collaborators', label: 'Unlimited collaborators', tooltip: featureTooltips.unlimited_users });
  } else if (plan.max_collaborators > 0) {
    features.push({ key: 'collaborators', label: plan.max_collaborators + ' collaborators per project', tooltip: featureTooltips.collaborators });
  }

  // Boolean features from feature_keys
  const featureLabels = {
    checkpoints: 'Checkpoints',
    workflows: 'Workflows',
    status: 'Status tracking',
    share_link: 'ShareLink',
    dependencies: 'Dependencies',
    sync: 'Cloud sync',
    resumable_transfers: 'Resumable uploads/downloads',
    project_templates: 'Project templates',
    kanban: 'Kanban',
    tags: 'Tags',
    interactive_console: 'Interactive console',
    ignore_list: 'Ignore list',
    integrations: 'Integrations',
    custom_roles: 'Custom roles',
    talent_discovery: 'Talent discovery',
    web_dashboard: 'Web dashboard',
    discord_support: 'Discord support',
    audit_log: 'Audit log',
    priority_support: 'Priority support',
    dedicated_vm: 'Dedicated/Managed VM',
    sso_saml: 'SSO / SAML',
    two_factor_auth: 'Enforced 2FA',
    custom_branding: 'Custom branding',
  };

  // Match the feature sequence used on the website pricing page.
  const orderedKeys = [
    'checkpoints', 'workflows', 'status', 'share_link', 'dependencies',
    'sync', 'resumable_transfers', 'project_templates', 'kanban', 'tags',
    'interactive_console', 'ignore_list', 'custom_roles', 'integrations',
    'talent_discovery', 'web_dashboard', 'discord_support', 'audit_log', 'priority_support',
    'dedicated_vm', 'sso_saml', 'two_factor_auth', 'custom_branding',
  ];

  // The website places the AI allowance before the capability list.
  if (plan.has_ai) {
    features.push({ key: 'ai', label: 'AI assistant', tooltip: featureTooltips.ai });
  }
  if (plan.ai_credits_monthly > 0) {
    features.push({ key: 'ai_credits', label: plan.ai_credits_monthly.toLocaleString() + ' AI credits/mo', tooltip: featureTooltips.ai_credits });
  }

  for (const key of orderedKeys) {
    if (keys.includes(key)) {
      features.push({ key, label: featureLabels[key], tooltip: featureTooltips[key], comingSoon: comingSoonFeatureKeys.has(key) });
    }
  }

  return features.map(feature => ({
    ...feature,
    icon: featureIcons[feature.key] || 'circle-check',
  }));
};

// Returns features whose label and value are identical across every visible plan.
const commonPlanFeatures = computed(() => {
  if (comparisonPlans.value.length < 2) return [];

  const featureLists = comparisonPlans.value.map(planFeatures);
  return featureLists[0].filter(feature => {
    return featureLists.slice(1).every(features => {
      return features.some(candidate => candidate.key === feature.key && candidate.label === feature.label);
    });
  });
});

const showComingSoonFootnote = computed(() => {
  return comparisonPlans.value
    .flatMap(planFeatures)
    .some(feature => feature.comingSoon);
});

// Keeps plan-specific limits and capabilities in each card.
const planCardFeatures = (plan) => {
  const commonFeatureKeys = new Set(commonPlanFeatures.value.map(feature => feature.key));
  return planFeatures(plan).filter(feature => !commonFeatureKeys.has(feature.key));
};

// Handles clicking a plan button to change plans.
const selectPlan = async (plan) => {
  if (plan.name === currentPlanName.value || isChanging.value) return;
  if (!plan.price_cents && plan.type === 'studio') {
    Browser.OpenURL('https://www.clustta.com/contact');
    closeModal();
    return;
  }
  isChanging.value = true;
  changingPlanId.value = plan.id;

  const studioId = (plan.type === 'studio' && isCloudStudio.value) ? projectStore.selectedStudio?.id : '';

  // Paid plans go through Stripe Checkout; free plan is a direct downgrade
  if (plan.price_cents > 0) {
    const { checkoutUrl, subscriptionUpdated } = await entitlementStore.createCheckout(plan.id, studioId);
    isChanging.value = false;
    changingPlanId.value = null;
    if (subscriptionUpdated) {
      if (studioId) {
        await entitlementStore.fetchStudioEntitlements(studioId);
      } else {
        await entitlementStore.fetchEntitlements();
      }
      notificationStore.addNotification('Plan changed', 'Your subscription has been updated.', 'success', false);
      closeModal();
      return;
    }
    if (checkoutUrl) {
      Browser.OpenURL(checkoutUrl);
      notificationStore.addNotification('Checkout', 'Complete your payment in the browser', 'success', false);
      closeModal();
    } else {
      notificationStore.addNotification('Error', 'Failed to start checkout. Please try again.', 'error', false);
    }
  } else if (currentPlanName.value !== 'free') {
    const success = await entitlementStore.cancelSubscription(studioId);
    isChanging.value = false;
    changingPlanId.value = null;
    if (success) {
      notificationStore.addNotification('Cancellation scheduled', 'Your current plan remains active until the end of the billing period.', 'success', false);
      closeModal();
    } else {
      notificationStore.addNotification('Error', 'Failed to schedule cancellation. Please try again.', 'error', false);
    }
  } else {
    isChanging.value = false;
    changingPlanId.value = null;
  }
};

// Opens the Stripe billing portal in the system browser.
const openBillingPortal = async () => {
  const studioId = activeTab.value === 'studio' && isCloudStudio.value
    ? projectStore.selectedStudio?.id
    : '';
  const portalUrl = await entitlementStore.openBillingPortal(studioId);
  if (portalUrl) {
    Browser.OpenURL(portalUrl);
  } else {
    notificationStore.addNotification('Error', 'Failed to open billing portal.', 'error', false);
  }
};

// lifecycle hooks
onMounted(async () => {
  if (!entitlementStore.plans.length) {
    isLoadingPlans.value = true;
    await entitlementStore.fetchPlans();
    isLoadingPlans.value = false;
  }
  if (projectStore.isCloudHosted) {
    activeTab.value = 'studio';
    entitlementStore.getStudioEntitlements(projectStore.selectedStudio.id);
  } else if (entitlementStore.planType === 'studio') {
    activeTab.value = 'studio';
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.cloud-modal {
  max-width: 96%;
  min-width: 600px;
  width: 1200px;
  max-height: 86vh;
}

.cloud-modal-body {
  flex: 1;
  min-height: 0;
  width: 100%;
  padding: 0 1.5rem 1.5rem;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding-top: 1rem;
  overflow-y: auto;
  scrollbar-color: var(--surface-4) transparent;
  scrollbar-width: thin;
}

.cloud-modal-body::-webkit-scrollbar,
.plan-cards::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.cloud-modal-body::-webkit-scrollbar-thumb,
.plan-cards::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--surface-4);
}

.cloud-modal-body::-webkit-scrollbar-track,
.plan-cards::-webkit-scrollbar-track {
  border-radius: 10px;
  background-color: transparent;
}

.cloud-modal-tabs {
  display: flex;
  background-color: var(--bg);
  border-radius: var(--very-large-radius);
  padding: 0.3rem 0.5rem;
  width: 100%;
  max-width: min-content;
}

.cloud-loading {
  text-align: center;
  padding: 2rem;
  color: var(--text);
  opacity: 0.5;
  font-size: 13px;
}

.plan-cards {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 280px));
  align-items: stretch;
  justify-content: center;
  gap: 1rem;
  width: 100%;
}

.plan-cards-2 {
  grid-template-columns: repeat(2, minmax(0, 280px));
  max-width: 700px;
  margin: 0 auto;
}

.plan-cards-4 {
  grid-template-columns: repeat(4, minmax(0, 260px));
}

.plan-cards-4 .plan-card {
  max-width: 260px;
}

.plan-card {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: var(--very-large-radius);
  background-color: var(--surface-2);
  transition: all 0.2s ease-out;
  width: 100%;
  max-width: 280px;
  min-width: 0;
  outline: var(--transparent-line);
  box-sizing: border-box;
  outline-offset: -1px;
}

.plan-card:hover {
  border-radius: var(--small-radius);
  background-color: var(--surface-3);
}

.plan-card-highlighted {
  outline-color: var(--surface-4);
}

.plan-card-current {
  outline-color: var(--attention);
}

.plan-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.plan-card-name {
  font-size: 19px;
  font-weight: 700;
  color: var(--text);
}

.plan-badge {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: var(--small-radius);
  background-color: rgba(155, 89, 208, 0.14);
  color: var(--grape);
}

.plan-card-price {
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.price-amount {
  font-family: 'Bricolage Grotesque', sans-serif;
  font-size: 36px;
  font-weight: 700;
  color: var(--text);
}

.price-period {
  font-size: 16px;
  color: var(--text);
  opacity: 0.5;
}

.plan-card-features {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 0;
  flex: 1;
}

.plan-feature {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text);
  opacity: 0.75;
  cursor: help;
}

.plan-feature:hover {
  opacity: 1;
}

.feature-icon {
  width: 16px;
  height: 16px;
  object-fit: contain;
  flex-shrink: 0;
}

.common-features {
  padding: 2rem;
  border-radius: var(--very-large-radius);
  background-color: var(--surface-2);
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.common-features-title {
  margin: 0 0 1.5rem;
  color: var(--text);
  font-family: 'Bricolage Grotesque', sans-serif;
  font-size: 24px;
  font-weight: 700;
  text-align: center;
}

.common-features-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem 2rem;
}

.common-features .plan-feature {
  font-size: 14px;
}

.feature-coming-soon-badge {
  margin-left: 5px;
  vertical-align: 1px;
}

.enterprise-section {
  padding: 2rem;
  border-radius: var(--very-large-radius);
  background-color: var(--surface-2);
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.enterprise-title {
  margin: 0 0 8px;
  color: var(--text);
  font-family: 'Bricolage Grotesque', sans-serif;
  font-size: 28px;
  line-height: 1.1;
}

.enterprise-price {
  margin: 0 0 1.25rem;
  color: var(--text);
  opacity: 0.65;
  font-size: 16px;
  font-weight: 700;
}

.enterprise-description {
  max-width: 760px;
  margin: 0 0 1.25rem;
  color: var(--text);
  font-size: 14px;
  line-height: 1.6;
}

.enterprise-button {
  min-width: 150px;
  height: 38px;
  border-radius: var(--normal-radius);
  background-color: var(--surface-3);
  color: var(--text);
}

.enterprise-button:hover {
  background-color: var(--surface-4);
}

.coming-soon-footnote {
  scroll-margin-top: 1rem;
  margin: -0.25rem 0 0;
  color: var(--text);
  opacity: 0.7;
  font-size: 12px;
  line-height: 1.5;
  text-align: center;
}

.coming-soon-footnote-badge {
  margin-right: 5px;
}

@media (max-width: 1100px) {
  .cloud-modal {
    width: 720px;
  }

  .plan-cards-4 {
    grid-template-columns: repeat(2, minmax(0, 280px));
  }

  .plan-cards-4 .plan-card {
    max-width: 280px;
  }
}

@media (max-width: 700px) {
  .cloud-modal {
    min-width: 0;
    width: calc(100vw - 32px);
  }

  .plan-cards-4 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .common-features-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.plan-card :deep(.general-button) {
  background-color: var(--grape);
  border-radius: var(--normal-radius);
  width: 100%;
  height: 40px;
  font-size: 15px;
  font-weight: 600;
  min-width: unset;
  padding: 11px 16px;
}

@media (max-width: 480px) {
  .common-features-grid {
    grid-template-columns: 1fr;
  }

  .common-features,
  .enterprise-section {
    padding: 1.25rem;
  }
}

.plan-card :deep(.general-button-text) {
  font-size: 15px;
}

.plan-card :deep(.general-button.plan-button-current) {
  background-color: var(--surface-3);
  color: var(--text);
  opacity: 1;
  cursor: default;
  outline: 1px solid var(--surface-4);
}

.plan-card :deep(.general-button.plan-button-current:hover) {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
}

.plan-card :deep(.general-button:hover) {
  background-color: hsl(270, 50%, 38%);
}

.plan-card :deep(.general-button.item-inactive:not(.plan-button-current)) {
  background-color: var(--surface-3);
  opacity: 0.5;
}

.billing-header-button {
  height: 30px;
  min-width: max-content;
  padding: 7px 12px;
  border-radius: var(--small-radius);
  background-color: var(--surface-2);
  color: var(--text);
}

.billing-header-button:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
}

.billing-header-button :deep(.general-button-text) {
  font-size: 13px;
}
</style>
