<template>
  <div class="page-root studio-setup-root">

    <div class="auth-root">

      <!-- back navigation -->
      <div v-if="!platformStore.isWeb && hostingType" class="back-nav" @click="handleBack">
        <img :src="getAppIcon('chevron-left')" class="back-nav-icon small-icons" />
        <span>{{ backLabel }}</span>
      </div>

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" />
        <div class="auth-header">{{ $t('auth.studioSetup.title') }}</div>
      </div>

      <div class="auth-container">

        <!-- hosting type selector -->
        <div v-if="!hostingType" class="hosting-selector">
          <OptionCard :icon="getAppIcon('two-drives')" :title="$t('auth.studioSetup.selfHostedTitle')" :description="$t('auth.studioSetup.selfHostedDescription')" @select="selectHostingType('self-hosted')" />

          <OptionCard :icon="getAppIcon('clustta')" :title="$t('auth.studioSetup.managedTitle')" :description="$t('auth.studioSetup.managedDescription')" @select="selectHostingType('managed')" />

          <!-- back link -->
          <div class="additional-actions">
            <div @click="goBack" class="back-link">
              {{ $t('auth.studioSetup.backToWelcome') }}
            </div>
          </div>
        </div>

        <!-- Self-hosted: Step 1 - Connect to server -->
        <div v-if="hostingType === 'self-hosted' && !isServerConnected" class="auth-form-container">
          <FormInput v-model="studioUrl" :placeholder="$t('auth.studioSetup.studioUrl')" :error="studioUrlError" :info="!studioUrlError ? $t('auth.studioSetup.studioUrlInfo') : ''" @input="validateStudioUrl" />

          <button class="submit-button display-font" :class="{ 'button-inactive': !isUrlValid }" @click="connectToServer">
            <div v-if="!isConnecting">{{ $t('auth.studioSetup.connectButton') }}</div>
            <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" />
          </button>

          <div v-if="connectionError" class="error-message">{{ connectionError }}</div>
        </div>

        <!-- Self-hosted: Step 2 - Registration form (after server connected) -->
        <div v-if="hostingType === 'self-hosted' && isServerConnected" class="auth-form-container">

          <!-- connected server badge -->
          <div class="server-badge">
            <div class="server-badge-info">
              <div class="status-dot"></div>
              <div class="server-badge-details">
                <span class="server-badge-name">{{ connectedServerName }}</span>
                <span class="server-badge-url">{{ studioUrl }}</span>
              </div>
            </div>
            <div @click="disconnectServer" class="server-change-link">{{ $t('auth.studioSetup.changeServer') }}</div>
          </div>

          <!-- registration form -->
          <form @submit.prevent="handleStudioRegister" class="auth-form" autocomplete="off">
            <div class="form-row">
              <FormInput v-model="registerForm.first_name" :placeholder="$t('auth.signUp.firstName')" :error="errors.first_name" />
              <FormInput v-model="registerForm.last_name" :placeholder="$t('auth.signUp.lastName')" />
            </div>

            <FormInput v-model="registerForm.username" :placeholder="$t('auth.signUp.username')" :error="errors.username" />

            <FormInput v-model="registerForm.email" :placeholder="$t('auth.signUp.emailAddress')" :error="errors.email" />

            <FormInput v-model="registerForm.password" :placeholder="$t('auth.signUp.password')" isSecret :error="passwordValidation" />

            <FormInput v-model="registerForm.confirm_password" :placeholder="$t('auth.signUp.confirmPassword')" isSecret :error="!passwordsMatch && registerForm.confirm_password ? errors.confirm_password : ''" />

            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isFormFilled }">
              <div v-if="!isAwaitingResponse">{{ $t('auth.studioSetup.createAccountButton') }}</div>
              <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" />
            </button>
          </form>

          <div v-if="error" class="error-message">{{ error }}</div>

          <!-- back / sign-in links -->
          <div class="additional-actions">
            <div @click="goToLogin" class="back-link">
              {{ $t('auth.studioSetup.haveStudioAccount') }}&nbsp;<span class="bold">{{ $t('auth.studioSetup.signInLink') }}</span>
            </div>
          </div>

          <!-- legal agreement -->
          <div class="legal-agreement">
            <p>{{ $t('auth.signUp.legalPrefix') }} <span class="legal-link" @click="openPrivacyPolicy">{{ $t('auth.signUp.privacyPolicy') }} <ActionButton :icon="getAppIcon('square-arrow-right-up')" :allowDeactivate="true" :isMini="true" /></span> {{ $t('auth.signUp.legalMiddle') }} <span class="legal-link" @click="openTermsOfService">{{ $t('auth.signUp.termsOfService') }} <ActionButton :icon="getAppIcon('square-arrow-right-up')" :allowDeactivate="true" :isMini="true" /></span>.</p>
          </div>
        </div>

        <!-- ClusttaCloud: Studio creation -->
        <div v-if="hostingType === 'managed'" class="auth-form-container">

          <!-- Enterprise: Contact Sales -->
          <div v-if="showEnterpriseSalesForm">
            <div v-if="!salesSubmitted">
              <div class="managed-form-header">
                <div class="managed-title display-font">{{ $t('auth.studioSetup.contactSalesTitle') }}</div>
                <div class="managed-description">{{ $t('auth.studioSetup.contactSalesDescription') }}</div>
              </div>

              <form @submit.prevent="handleContactSales" class="auth-form" autocomplete="off">
                <FormInput v-model="salesForm.name" :placeholder="$t('auth.studioSetup.salesName')" />

                <FormInput v-model="salesForm.email" :placeholder="$t('auth.studioSetup.salesEmail')" :error="salesEmailError" :info="workEmailNudge" />

                <FormInput v-model="salesForm.company" :placeholder="$t('auth.studioSetup.salesCompany')" />

                <div class="dropdown-spacing">
                  <DropDownBox :items="teamSizeItems" :selectedItem="salesForm.team_size" :onSelect="selectTeamSize" :placeHolder="$t('auth.studioSetup.salesTeamSize')" :useFilter="false" />
                </div>

                <div class="dropdown-spacing">
                  <DropDownBox :items="sourceItems" :selectedItem="salesForm.source" :onSelect="selectSource" :placeHolder="$t('auth.studioSetup.salesSource')" :useFilter="false" />
                </div>

                <FormInput v-model="salesForm.website" :placeholder="$t('auth.studioSetup.salesWebsite')" />

                <textarea v-model="salesForm.message" class="desktop-input-long" :placeholder="$t('auth.studioSetup.salesMessage')" rows="3"></textarea>

                <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isSalesFormFilled }">
                  <div v-if="!isSubmittingSales">{{ $t('auth.studioSetup.salesSubmitButton') }}</div>
                  <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" />
                </button>
              </form>

              <div v-if="salesError" class="error-message">{{ salesError }}</div>
            </div>

            <div v-else class="managed-content">
              <div class="managed-icon-container">
                <img :src="getAppIcon('mail')" class="managed-icon gigantic-icons" />
              </div>
              <div class="managed-title display-font">{{ $t('auth.studioSetup.salesSuccessTitle') }}</div>
              <div class="managed-description">{{ $t('auth.studioSetup.salesSuccessDescription') }}</div>
            </div>

            <div class="additional-actions">
              <div @click="showEnterpriseSalesForm = false" class="back-link">{{ $t('auth.studioSetup.backToPlanSelection') }}</div>
            </div>
          </div>

          <!-- Cloud/Pro: Studio name + plan selection -->
          <div v-else>
            <FormInput v-model="cloudStudioName" :placeholder="$t('placeholders.studioName')" :error="cloudStudioNameError" :loading="checkingCloudStudioName" :valid="!!cloudStudioName && !cloudStudioNameError && !checkingCloudStudioName" :showValidation="!!cloudStudioName" :disabled="isEnterprisePlan" @input="checkCloudStudioName" />

            <div v-if="isLoadingPlans" class="plan-loading">Loading plans...</div>

            <div v-else class="plan-select-container">
              <div class="plan-select-label">Select a plan</div>

              <OptionCard v-for="plan in studioPlans" :key="plan.id" :title="formatPlanName(plan.name)" :description="planDescription(plan)" :selectable="true" :selected="selectedPlanId === plan.id" @select="selectPlan(plan)" />
            </div>

            <button class="submit-button display-font" :class="{ 'button-inactive': !canCreateCloud }" @click="createCloudStudioAndCheckout">
              <div v-if="!isCreatingStudio">{{ cloudCreateButtonLabel }}</div>
              <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" />
            </button>

            <div v-if="cloudError" class="error-message">{{ cloudError }}</div>
          </div>

        </div>

      </div>

    </div>

  </div>
</template>

<script setup>
// imports
import { onMounted, ref, reactive, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { Browser } from '@wailsio/runtime';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
import OptionCard from '@/instances/common/components/OptionCard.vue';

// services
import { AuthService, SettingsService, StudioService } from '@/services';

// store imports
import { useAccountStore } from '@/stores/accounts';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useThemeStore } from '@/stores/theme';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

// stores
const accountStore = useAccountStore();
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const themeStore = useThemeStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

// refs
const connectionError = ref('');
const connectedServerName = ref('');
const error = ref('');
const hostingType = ref('');
const isAwaitingResponse = ref(false);
const isConnecting = ref(false);
const isServerConnected = ref(false);
const isSubmittingSales = ref(false);
const salesError = ref('');
const salesSubmitted = ref(false);
const studioUrl = ref('');
const studioUrlError = ref('');

// ClusttaCloud refs
const checkingCloudStudioName = ref(false);
const cloudError = ref('');
const cloudStudioName = ref('');
const cloudStudioNameError = ref('');
const isCloudStudioNameTaken = ref(false);
const isCreatingStudio = ref(false);
const isLoadingPlans = ref(false);
const selectedPlanId = ref(null);
const showEnterpriseSalesForm = ref(false);

const restrictedNames = ['clustta', 'eaxum', 'pixar', 'disney', 'dreamworks'];

const registerForm = reactive({
  first_name: '',
  last_name: '',
  username: '',
  email: '',
  password: '',
  confirm_password: ''
});

const errors = reactive({
  first_name: '',
  last_name: '',
  username: '',
  email: '',
  password: '',
  confirm_password: ''
});

const salesForm = reactive({
  name: '',
  email: '',
  company: '',
  team_size: '',
  source: '',
  website: '',
  message: ''
});

// computed properties
const freeEmailDomains = ['gmail.com', 'yahoo.com', 'hotmail.com', 'outlook.com', 'aol.com', 'icloud.com', 'mail.com', 'protonmail.com', 'zoho.com', 'yandex.com', 'gmx.com', 'live.com'];
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const isUrlValid = computed(() => {
  return studioUrl.value.trim() && !studioUrlError.value;
});

const isFormFilled = computed(() => {
  return isServerConnected.value
    && registerForm.first_name
    && registerForm.username
    && registerForm.email
    && !passwordValidation.value
    && passwordsMatch.value;
});

const isSalesEmailValid = computed(() => {
  return emailRegex.test(salesForm.email);
});

const isSalesFormFilled = computed(() => {
  return salesForm.name
    && salesForm.email
    && salesForm.company
    && salesForm.team_size
    && isSalesEmailValid.value;
});

const salesEmailError = computed(() => {
  if (!salesForm.email) return '';
  return !isSalesEmailValid.value ? t('auth.studioSetup.salesEmailInvalid') : '';
});

const sourceItems = computed(() => [
  t('auth.studioSetup.sourceSearch'),
  t('auth.studioSetup.sourceSocial'),
  t('auth.studioSetup.sourceReferral'),
  t('auth.studioSetup.sourceEvent'),
  t('auth.studioSetup.sourceOther')
]);

const teamSizeItems = ['1–5', '6–15', '16–50', '50+'];

const workEmailNudge = computed(() => {
  if (!salesForm.email || !isSalesEmailValid.value) return '';
  const domain = salesForm.email.split('@')[1]?.toLowerCase();
  return freeEmailDomains.includes(domain) ? t('auth.studioSetup.workEmailNudge') : '';
});

// Returns the label for the back navigation button.
const backLabel = computed(() => {
  if (showEnterpriseSalesForm.value) return t('auth.studioSetup.backToPlanSelection');
  return t('auth.studioSetup.backToWelcome');
});

// Returns whether the cloud studio form is ready to proceed.
const canCreateCloud = computed(() => {
  const plan = studioPlans.value.find(p => p.id === selectedPlanId.value);
  if (plan && plan.name === 'studio_enterprise') return true;
  return cloudStudioName.value && !cloudStudioNameError.value && !isCloudStudioNameTaken.value && !checkingCloudStudioName.value && !!selectedPlanId.value;
});

// Returns the label for the cloud studio create button.
const cloudCreateButtonLabel = computed(() => {
  const plan = studioPlans.value.find(p => p.id === selectedPlanId.value);
  if (!plan) return t('common.create');
  if (plan.name === 'studio_enterprise') return 'Contact Sales';
  return 'Create & Subscribe';
});

// Returns whether the selected plan is enterprise.
const isEnterprisePlan = computed(() => {
  const plan = studioPlans.value.find(p => p.id === selectedPlanId.value);
  return plan && plan.name === 'studio_enterprise';
});

// Returns studio plans for selection (paid plans + enterprise).
const studioPlans = computed(() => {
  return entitlementStore.plans.filter(p => p.type === 'studio' && (p.price_cents !== 0 || p.name === 'studio_enterprise'));
});

const passwordValidation = computed(() => {
  const password = registerForm.password;
  if (!password) return null;

  const lowerPassword = password.toLowerCase();

  const patterns = [
    { value: registerForm.username.toLowerCase(), errorMessage: t('auth.signUp.passwordContainsEmailOrUsername') },
    { value: registerForm.email.toLowerCase().split('@')[0], errorMessage: t('auth.signUp.passwordContainsEmailOrUsername') },
    { value: registerForm.first_name.toLowerCase(), errorMessage: t('auth.signUp.passwordContainsFirstName') },
    { value: registerForm.last_name.toLowerCase(), errorMessage: t('auth.signUp.passwordContainsLastName') }
  ];

  for (const pattern of patterns) {
    if (!pattern.value) continue;
    const escaped = pattern.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    if (new RegExp(escaped, 'i').test(lowerPassword)) return pattern.errorMessage;
  }

  const rules = [
    { regex: /.{8,}/, errorMessage: t('auth.signUp.passwordMinLength') },
    { regex: /[A-Z]/, errorMessage: t('auth.signUp.passwordUppercase') },
    { regex: /[a-z]/, errorMessage: t('auth.signUp.passwordLowercase') },
    { regex: /\d/, errorMessage: t('auth.signUp.passwordNumber') },
    { regex: /[@$!%*?&]/, errorMessage: t('auth.signUp.passwordSpecialChar') }
  ];

  for (const rule of rules) {
    if (!rule.regex.test(password)) return rule.errorMessage;
  }
  return null;
});

const passwordsMatch = computed(() => {
  const match = registerForm.password === registerForm.confirm_password;
  errors.confirm_password = match ? '' : t('auth.signUp.passwordsDoNotMatch');
  return match && registerForm.password.length;
});

// methods/functions

// Checks if the cloud studio name is available.
const checkCloudStudioName = async () => {
  if (!cloudStudioName.value) {
    cloudStudioNameError.value = '';
    isCloudStudioNameTaken.value = false;
    return;
  }

  if (restrictedNames.includes(cloudStudioName.value.toLowerCase())) {
    cloudStudioNameError.value = t('notifications.studioNameReserved');
    isCloudStudioNameTaken.value = true;
    return;
  }

  checkingCloudStudioName.value = true;

  try {
    const nameExists = await StudioService.CheckStudioNameExists(cloudStudioName.value.toLowerCase());
    if (nameExists) {
      cloudStudioNameError.value = t('notifications.studioNameTaken');
      isCloudStudioNameTaken.value = true;
    } else {
      cloudStudioNameError.value = '';
      isCloudStudioNameTaken.value = false;
    }
  } catch (err) {
    cloudStudioNameError.value = '';
    isCloudStudioNameTaken.value = false;
    console.error('Error checking studio name:', err);
  } finally {
    checkingCloudStudioName.value = false;
  }
};

// Creates the ClusttaCloud studio on the free tier, then redirects to Stripe Checkout.
const createCloudStudioAndCheckout = async () => {
  if (!canCreateCloud.value || isCreatingStudio.value) return;

  const plan = studioPlans.value.find(p => p.id === selectedPlanId.value);
  if (!plan) return;

  if (plan.name === 'studio_enterprise') {
    showEnterpriseSalesForm.value = true;
    return;
  }

  isCreatingStudio.value = true;
  cloudError.value = '';

  try {
    // Create the studio (inactive until checkout completes)
    const result = await StudioService.RegisterStudio(cloudStudioName.value, '', 'cloud');

    // Get the studio ID from the creation response
    const studioId = result?.id || '';

    const checkoutUrl = await entitlementStore.createCheckout(plan.id, studioId);
    if (checkoutUrl) {
      Browser.OpenURL(checkoutUrl);
      notificationStore.addNotification('Checkout', 'Complete your payment in the browser. Your studio will be activated once payment is confirmed.', 'info', false);
    } else {
      notificationStore.addNotification('Error', 'Failed to start checkout. Please try again.', 'error', false);
    }

    accountStore.onboardingIntent = null;

    // Navigate to main app — DirOnboardModal will show if needed
    const projectDirectoryExists = await SettingsService.GetProjectDirectory();
    if (!projectDirectoryExists) {
      modals.setModalVisibility('dirOnboardModal', true);
    }
    router.push('/');
  } catch (err) {
    console.error(err);
    cloudError.value = err.message || t('notifications.errorCreatingStudio');
    notificationStore.errorNotification(t('notifications.errorCreatingStudio'), err);
  } finally {
    isCreatingStudio.value = false;
  }
};

// Connects to the self-hosted studio server and retrieves its info.
const connectToServer = async () => {
  if (!isUrlValid.value || isConnecting.value) return;

  isConnecting.value = true;
  connectionError.value = '';

  const normalizedUrl = normalizeStudioUrl(studioUrl.value);

  try {
    const info = await StudioService.GetStudioInfo(normalizedUrl);
    connectedServerName.value = info.name || normalizedUrl;
    studioUrl.value = normalizedUrl;
    isServerConnected.value = true;
  } catch (err) {
    console.log(err);
    connectionError.value = t('auth.studioSetup.connectionFailed');
  } finally {
    isConnecting.value = false;
  }
};

// Disconnects from the currently connected server and resets to URL input.
const disconnectServer = () => {
  isServerConnected.value = false;
  connectedServerName.value = '';
  connectionError.value = '';
  error.value = '';
};

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

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Navigates back based on current context.
const handleBack = () => {
  if (showEnterpriseSalesForm.value) {
    showEnterpriseSalesForm.value = false;
    return;
  }
  goBack();
};

// Navigates back to the welcome page.
const goBack = () => {
  accountStore.onboardingIntent = null;
  router.push('/auth/welcome');
};

// Navigates to the login page with the connected studio URL.
const goToLogin = () => {
  if (isServerConnected.value && studioUrl.value) {
    router.push({
      path: '/auth/login',
      query: { studioUrl: studioUrl.value, name: connectedServerName.value }
    });
  } else {
    router.push('/auth/login');
  }
};

// Navigates to the personal sign-up page.
const goToPersonalSignUp = () => {
  router.push('/auth/sign-up');
};

// Returns a short description with price for the plan card.
const planDescription = (plan) => {
  if (plan.name === 'studio_enterprise') return 'Custom infrastructure — contact us for pricing';
  const price = '$' + (plan.price_cents / 100) + '/mo';
  const storage = formatStorage(plan.storage_bytes) + ' storage';
  const seats = plan.max_collaborators === -1 ? 'Unlimited seats' : plan.max_collaborators + ' seats';
  return price + ' · ' + storage + ' · ' + seats;
};

// Selects a plan and clears studio name for enterprise.
const selectPlan = (plan) => {
  selectedPlanId.value = plan.id;
  if (plan.name === 'studio_enterprise') {
    cloudStudioName.value = '';
    cloudStudioNameError.value = '';
  }
};

// Selects a source option from the dropdown.
const selectSource = (value) => {
  salesForm.source = value;
};

// Selects a team size option from the dropdown.
const selectTeamSize = (value) => {
  salesForm.team_size = value;
};

// Handles contact sales form submission.
const handleContactSales = async () => {
  if (!isSalesFormFilled.value || isSubmittingSales.value) return;

  isSubmittingSales.value = true;
  salesError.value = '';

  try {
    await AuthService.ContactSales(
      salesForm.name,
      salesForm.email,
      salesForm.company,
      salesForm.team_size,
      salesForm.source,
      salesForm.website,
      salesForm.message
    );
    salesSubmitted.value = true;
  } catch (err) {
    console.log(err);
    salesError.value = err.message || t('auth.studioSetup.salesSubmitFailed');
  } finally {
    isSubmittingSales.value = false;
  }
};

// Handles studio registration form submission.
const handleStudioRegister = async () => {
  if (!isFormFilled.value || isAwaitingResponse.value) return;

  isAwaitingResponse.value = true;
  error.value = '';

  const normalizedUrl = normalizeStudioUrl(studioUrl.value);

  try {
    await AuthService.RegisterWithHost(
      registerForm.first_name,
      registerForm.last_name,
      registerForm.username,
      registerForm.email,
      registerForm.password,
      registerForm.confirm_password,
      normalizedUrl
    );

    notificationStore.addNotification(
      t('auth.signUp.registrationSuccessful'),
      t('auth.signUp.studioAccountCreated', { url: normalizedUrl }),
      'success'
    );
    router.push('/auth/login');
  } catch (err) {
    console.log(err);
    const errorMessage = err.message || err.response?.data?.message || t('auth.signUp.registrationFailedDefault');
    error.value = errorMessage;
    notificationStore.errorNotification(t('auth.signUp.registrationFailed'), errorMessage);
  } finally {
    isAwaitingResponse.value = false;
  }
};

// Normalizes the studio URL by ensuring it has a protocol and no trailing slash.
const normalizeStudioUrl = (url) => {
  if (!url) return '';
  let normalized = url.trim();
  normalized = normalized.replace(/\/+$/, '');
  if (!normalized.startsWith('http://') && !normalized.startsWith('https://')) {
    normalized = 'https://' + normalized;
  }
  return normalized;
};

// Opens the privacy policy page in the browser.
const openPrivacyPolicy = () => {
  Browser.OpenURL('https://clustta.com/privacy-policy');
};

// Opens the terms of service page in the browser.
const openTermsOfService = () => {
  Browser.OpenURL('https://clustta.com/terms-of-service');
};

// Resets the hosting type selection.
const resetHostingType = () => {
  hostingType.value = '';
};

// Sets the hosting type (self-hosted or managed).
const selectHostingType = (type) => {
  hostingType.value = type;
};

// lifecycle hooks
onMounted(async () => {
  const queryType = route.query.type;
  const queryUrl = route.query.url;
  const queryName = route.query.name;

  // Auto-select hosting type based on onboarding intent
  if (accountStore.onboardingIntent === 'studio') {
    hostingType.value = 'managed';
  } else if (accountStore.onboardingIntent === 'self-hosted') {
    hostingType.value = 'self-hosted';
  } else if (queryType === 'managed') {
    hostingType.value = 'managed';
  } else if (queryType === 'self-hosted' && queryUrl) {
    hostingType.value = 'self-hosted';
    studioUrl.value = queryUrl;
    connectedServerName.value = queryName || queryUrl;
    await connectToServer();
  }

  // Fetch plans for ClusttaCloud studio creation
  if (!entitlementStore.plans.length) {
    isLoadingPlans.value = true;
    await entitlementStore.fetchPlans();
    isLoadingPlans.value = false;
  }
});

// Validates the studio URL format.
const validateStudioUrl = () => {
  if (!studioUrl.value) {
    studioUrlError.value = '';
    return;
  }

  const urlPattern = /^https?:\/\/[a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)+(:\d+)?(\/.*)?$/;

  if (!studioUrl.value.startsWith('http://') && !studioUrl.value.startsWith('https://')) {
    studioUrlError.value = t('auth.signUp.urlMustStartWith');
  } else if (!urlPattern.test(studioUrl.value)) {
    studioUrlError.value = t('auth.signUp.invalidUrl');
  } else {
    studioUrlError.value = '';
  }
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.studio-setup-root {
  display: flex;
  box-sizing: border-box;
  width: 100%;
  align-items: center;
  height: 100%;
  flex-direction: column;
  background-color: var(--black);
  overflow: hidden;
  overflow-y: auto;
}

.back-nav {
  position: absolute;
  top: 0px;
  left: .5rem;
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.35rem 0.6rem;
  cursor: pointer;
  font-size: 0.8rem;
  font-weight: 300;
  color: var(--white);
  opacity: 0.5;
  transition: opacity 0.2s;
  z-index: 10;
}

.back-nav:hover {
  opacity: 1;
}

.back-nav-icon {
  width: 14px;
  height: 14px;
}

.hosting-selector {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
  min-width: 400px;
}

.additional-actions {
  display: flex;
  box-sizing: border-box;
  padding: 0.5rem;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  width: 100%;
  font-weight: 300;
  font-size: 14px;
  justify-content: center;
}

.back-link {
  color: var(--white);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem;
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 0.2s;
}

.back-link:hover {
  opacity: 1;
}

.legal-agreement {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 1rem 0.5rem;
  font-size: 12px;
  color: var(--white);
  font-weight: 300;
  gap: 0.25rem;
}

.legal-agreement p {
  margin: 0;
}

.legal-link {
  color: var(--white);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  text-decoration: underline;
}

.legal-link:hover {
  color: var(--blue);
}

.managed-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  width: 100%;
  min-width: 400px;
}

.managed-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 2rem;
  text-align: center;
}

.managed-icon-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: var(--large-radius);
  background-color: var(--light-steel);
}

.managed-icon {
  width: 32px;
  height: 32px;
  filter: var(--icon-filter, none);
}

.managed-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--white);
}

.managed-description {
  font-size: 0.9rem;
  color: var(--white);
  opacity: 0.6;
  font-weight: 300;
  line-height: 150%;
  max-width: 380px;
}

.managed-form-header {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  margin-bottom: 1rem;
}



.dropdown-spacing {
  margin-bottom: 0.8rem;
}

.desktop-input-long {
  margin-top: 0px;
  font-weight: 200;
  color: var(--white);
  margin-bottom: .8rem;
}

.button-inactive {
  opacity: 0.5;
  pointer-events: none;
}

.error-message {
  color: var(--red);
  font-size: 0.8rem;
  font-weight: 400;
  text-align: center;
  padding: 0.25rem 0;
}

.server-badge {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.6rem 1rem;
  margin-bottom: 0.5rem;
  border-radius: var(--large-radius);
  background-color: var(--midnight-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  gap: 1rem;
  transition: border-radius .2s;
}

.server-badge:hover {
  border-radius: var(--normal-radius);
}

.server-badge-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.status-dot {
  position: relative;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #22c55e;
  animation: dot-entrance 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
}

.status-dot::before,
.status-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background-color: #22c55e;
  animation: ripple 2.5s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}

.status-dot::after {
  animation-delay: 1.25s;
}

@keyframes dot-entrance {
  0% { transform: scale(0); }
  60% { transform: scale(1.3); }
  80% { transform: scale(0.9); }
  100% { transform: scale(1); }
}

@keyframes ripple {
  0% { opacity: 0.5; transform: scale(1); }
  100% { opacity: 0; transform: scale(3); }
}

.server-badge-details {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.server-badge-name {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--white);
  opacity: 0.9;
}

.server-badge-url {
  font-size: 0.7rem;
  color: var(--white);
  opacity: 0.4;
  font-weight: 300;
}

.server-change-link {
  font-size: 0.75rem;
  color: var(--white);
  opacity: 0.5;
  cursor: pointer;
  transition: opacity 0.2s;
}

.server-change-link:hover {
  opacity: 1;
}

@media (max-width: 768px) {
  .hosting-selector,
  .managed-container {
    min-width: 300px;
  }
}

.plan-loading {
  text-align: center;
  padding: 1rem;
  color: var(--white);
  opacity: 0.6;
}

.plan-select-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.5rem 0;
  width: 100%;
}

.plan-select-label {
  font-size: 0.85rem;
  color: var(--white);
  opacity: 0.6;
  font-weight: 300;
}
</style>
