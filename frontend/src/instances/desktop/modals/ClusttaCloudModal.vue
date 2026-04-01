<template>
  <div class="modal-container cloud-modal" v-stop-propagation>
    <HeaderArea title="ClusttaCloud" icon="clustta" :notModal="false" />

    <div class="cloud-modal-body">
      <div class="cloud-modal-tabs">
        <PaneHeaderTabs :dataTypes="planTabs" :selectedTab="activeTab" :fullWidth="true" @filter="activeTab = $event" />
      </div>

      <div v-if="isLoadingPlans" class="cloud-loading">Loading plans...</div>

      <div v-else class="plan-cards">
        <div v-for="plan in filteredPlans" :key="plan.id" class="plan-card" :class="{ 'plan-card-current': plan.name === entitlementStore.plan, 'plan-card-highlighted': isRecommended(plan) }">

          <div class="plan-card-header">
            <span class="plan-card-name">{{ formatPlanName(plan.name) }}</span>
            <span v-if="isRecommended(plan)" class="plan-badge">Recommended</span>
          </div>

          <div class="plan-card-tagline">{{ planTagline(plan) }}</div>

          <div class="plan-card-price">
            <span v-if="plan.name === 'studio_enterprise'" class="price-amount price-contact">Contact</span>
            <span v-else-if="plan.price_cents === 0" class="price-amount">$0</span>
            <template v-else>
              <span class="price-amount">${{ (plan.price_cents / 100) }}</span>
              <span class="price-period">/mo</span>
            </template>
          </div>

          <GeneralButton :label="planButtonLabel(plan)" :colored="plan.name !== entitlementStore.plan" :isActive="plan.name !== entitlementStore.plan && !isChanging" :loading="isChanging && changingPlanId === plan.id" :fullWidth="true" :buttonFunction="() => selectPlan(plan)" />

          <div class="plan-card-features">
            <div class="plan-feature" v-for="feature in planFeatures(plan)" :key="feature.key" v-tooltip="feature.tooltip">
              <span class="feature-check">✓</span>
              <span>{{ feature.label }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PaneHeaderTabs from '@/instances/common/components/PaneHeaderTabs.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useNotificationStore } from '@/stores/notifications';

const entitlementStore = useEntitlementStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();

// refs
const activeTab = ref('individual');
const changingPlanId = ref(null);
const isChanging = ref(false);
const isLoadingPlans = ref(false);

// constants
const planTabs = [
  { name: 'individual', icon: 'person' },
  { name: 'studio', icon: 'stall' },
];

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

// computed
// Returns plans filtered by the active tab category.
const filteredPlans = computed(() => {
  const type = activeTab.value === 'individual' ? 'individual' : 'studio';
  return entitlementStore.plans.filter(p => p.type === type);
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns a human-readable plan name.
const formatPlanName = (name) => {
  const names = {
    free: 'Free',
    starter: 'Starter',
    pro: 'Pro',
    studio_cloud: 'Studio Cloud',
    studio_pro: 'Studio Pro',
    studio_enterprise: 'Enterprise',
  };
  return names[name] || name.charAt(0).toUpperCase() + name.slice(1).replace(/_/g, ' ');
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
  if (plan.name === entitlementStore.plan) return 'Current Plan';
  if (plan.name === 'studio_enterprise') return 'Contact Sales';
  const currentOrder = entitlementStore.plans.find(p => p.name === entitlementStore.plan)?.display_order ?? 0;
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
    discord_support: 'Discord support',
    audit_log: 'Audit log',
    priority_support: 'Priority support',
    dedicated_vm: 'Dedicated/Managed VM',
    sso_saml: 'SSO / SAML',
    two_factor_auth: '2FA authentication',
    custom_branding: 'Custom branding',
  };

  // Ordered display of boolean features
  const orderedKeys = [
    'checkpoints', 'workflows', 'status', 'share_link', 'dependencies',
    'sync', 'resumable_transfers', 'project_templates', 'kanban', 'tags',
    'interactive_console', 'ignore_list', 'integrations', 'custom_roles',
    'discord_support', 'audit_log', 'priority_support',
    'dedicated_vm', 'sso_saml', 'two_factor_auth', 'custom_branding',
  ];

  for (const key of orderedKeys) {
    if (keys.includes(key)) {
      features.push({ key, label: featureLabels[key], tooltip: featureTooltips[key] });
    }
  }

  // AI features from plan columns
  if (plan.has_ai) {
    features.push({ key: 'ai', label: 'AI assistant', tooltip: featureTooltips.ai });
  }
  if (plan.ai_credits_monthly > 0) {
    features.push({ key: 'ai_credits', label: plan.ai_credits_monthly.toLocaleString() + ' AI credits/mo', tooltip: featureTooltips.ai_credits });
  }

  return features;
};

// Returns the tagline for a plan.
const planTagline = (plan) => {
  const taglines = {
    free: 'Get started with Clustta',
    starter: 'For creators who need remote sync',
    pro: 'Full power for professionals',
    studio_cloud: 'For teams collaborating on creative work',
    studio_pro: 'For studios at scale',
    studio_enterprise: 'Custom infrastructure for large studios',
  };
  return taglines[plan.name] || '';
};

// Handles clicking a plan button to change plans.
const selectPlan = async (plan) => {
  if (plan.name === entitlementStore.plan || isChanging.value) return;
  if (plan.name === 'studio_enterprise') return;
  isChanging.value = true;
  changingPlanId.value = plan.id;
  const success = await entitlementStore.changePlan(plan.id);
  isChanging.value = false;
  changingPlanId.value = null;
  if (success) {
    notificationStore.addNotification('Plan changed', 'You are now on the ' + formatPlanName(plan.name) + ' plan', 'success', false);
    closeModal();
  } else {
    notificationStore.addNotification('Error', 'Failed to change plan. Please try again.', 'error', false);
  }
};

// lifecycle hooks
onMounted(async () => {
  if (!entitlementStore.plans.length) {
    isLoadingPlans.value = true;
    await entitlementStore.fetchPlans();
    isLoadingPlans.value = false;
  }
  if (entitlementStore.planType === 'studio') {
    activeTab.value = 'studio';
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.cloud-modal {
  max-width: 820px;
  min-width: 600px;
  width: 820px;
}

.cloud-modal-body {
  width: 100%;
  padding: 0 1.5rem 1.5rem;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-top: 1rem;
}

.cloud-modal-tabs {
  display: flex;
  background-color: var(--midnight-steel);
  border-radius: var(--very-large-radius);
  padding: 0.3rem 0.5rem;
  width: 100%;
  max-width: min-content;
}

.cloud-loading {
  text-align: center;
  padding: 2rem;
  color: var(--white);
  opacity: 0.5;
  font-size: 13px;
}

.plan-cards {
  display: flex;
  gap: 12px;
  width: 100%;
}

.plan-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 1.2rem;
  border-radius: var(--very-large-radius);
  background-color: var(--dark-steel);
  transition: all 0.2s ease-out;
  min-width: 0;
  outline: var(--transparent-line);
  box-sizing: border-box;
  outline-offset: -1px;
}

.plan-card:hover {
  border-radius: var(--small-radius);
  background-color: var(--steel);
}

.plan-card-current {
  border-color: var(--grape);
}

.plan-card-highlighted {
  border-color: var(--light-steel);
}

.plan-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.plan-card-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--white);
}

.plan-badge {
  font-size: 10px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: var(--small-radius);
  background-color: var(--grape);
  color: white;
}

.plan-card-tagline {
  font-size: 12px;
  color: var(--white);
  opacity: 0.5;
  min-height: 32px;
}

.plan-card-price {
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.price-amount {
  font-size: 28px;
  font-weight: 700;
  color: var(--white);
}

.price-contact {
  font-size: 22px;
  font-weight: 600;
}

.price-period {
  font-size: 13px;
  color: var(--white);
  opacity: 0.5;
}

.plan-card-features {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 4px;
  border-top: 1px solid var(--light-steel);
  padding-top: 10px;
  max-height: 280px;
  overflow-y: auto;
}

.plan-card-features::-webkit-scrollbar {
  width: 4px;
}

.plan-card-features::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.plan-card-features::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.plan-feature {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 13px;
  color: var(--white);
  opacity: 0.75;
  cursor: help;
}

.plan-feature:hover {
  opacity: 1;
}

.feature-check {
  color: var(--grape);
  font-weight: 600;
  flex-shrink: 0;
}

.plan-card :deep(.general-button) {
  background-color: var(--grape);
  border-radius: var(--small-radius);
  height: 32px;
  font-size: 13px;
  min-width: unset;
  padding: 8px 12px;
}

.plan-card :deep(.general-button:hover) {
  background-color: hsl(270, 50%, 38%);
}

.plan-card :deep(.general-button.item-inactive) {
  background-color: var(--steel);
  opacity: 0.5;
}
</style>
