import { createRouter, createWebHistory, createMemoryHistory } from "vue-router";
import { AuthService, SettingsService, AppService, ProjectService } from "@/services";
import { useUserStore } from '@/stores/users';
import { useAccountStore } from '@/stores/accounts';
import { useThemeStore } from '@/stores/theme';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useDesktopModalStore } from '@/stores/desktopModals';

const isWebMode = import.meta.env.VITE_PLATFORM === 'web';

// Track if stores have been initialized for the current session
let storesInitialized = false;

const routes = [
  {
    path: '/auth',
    component: () => import('@/instances/desktop/pages/AuthGuard.vue'),
    meta: { isAuthLayout: true },
    children: [
      {
        path: '',
        name: 'auth-default',
        redirect: '/auth/login'
      },
      {
        path: 'login',
        name: 'login',
        component: () => import('@/instances/desktop/pages/Login.vue'),
        meta: { requiresAuth: false, isPublic: true, isAuthPage: true }
      },
      {
        path: 'signup',
        name: 'signup',
        component: () => import('@/instances/desktop/pages/SignUp.vue'),
        meta: { requiresAuth: false, isPublic: true, isAuthPage: true }
      },
      {
        path: 'verify-email',
        name: 'verify-email',
        component: () => import('@/instances/desktop/pages/VerifyEmail.vue'),
        meta: { requiresAuth: false, isPublic: true, isAuthPage: true }
      },
      {
        path: 'forgot-password',
        name: 'forgot-password',
        component: () => import('@/instances/desktop/pages/ResetPassword.vue'),
        meta: { requiresAuth: false, isPublic: true, isAuthPage: true }
      },
      {
        path: 'reset-change-password',
        name: 'reset-change-password',
        component: () => import('@/instances/web/ResetChangePassword.vue'),
        meta: { requiresAuth: false, isPublic: true, isAuthPage: true }
      },
    ]
  },
  // Legacy auth redirects (for compatibility)
  { path: '/login', redirect: '/auth/login' },
  { path: '/signup', redirect: '/auth/signup' },
  { path: '/verify-email', redirect: '/auth/verify-email' },
  { path: '/forgot-password', redirect: '/auth/forgot-password' },
  { path: '/reset-change-password', redirect: '/auth/reset-change-password' },
  // Public profile (web only in practice)
  {
    path: '/user/:username',
    name: 'public-profile',
    component: () => import('@/instances/web/PublicUserProfile.vue'),
    meta: { requiresAuth: false, isPublic: true }
  },
  // Protected routes (main app)
  {
    path: '/',
    name: 'home',
    component: () => import('@/instances/desktop/ClusttaDesktop.vue'),
    meta: { requiresAuth: true }
  },
  // Catch-all redirect
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    redirect: '/'
  }
];

const router = createRouter({
  history: isWebMode ? createWebHistory() : createMemoryHistory(),
  routes,
});

// Updates the loading status text in the initial app loader (index.html).
const setLoaderStatus = (message) => {
  const statusEl = document.getElementById('loader-status');
  if (statusEl) {
    statusEl.textContent = message;
  }
};

// Navigation guard for auth
router.beforeEach(async (to, from, next) => {
  // Public routes that don't need any auth check
  if (to.meta.isPublic && !to.meta.isAuthPage) {
    next();
    return;
  }

  // Check authentication status
  setLoaderStatus('Checking authentication...');
  let isAuthenticated = false;
  let userData = null;
  try {
    const result = await AuthService.IsAuthenticated();
    isAuthenticated = result[0] === true;
    userData = result[1]; // User data from auth check
  } catch (error) {
    isAuthenticated = false;
  }

  // Auth pages: redirect to home if already logged in
  if (to.meta.isAuthPage && isAuthenticated) {
    return next('/');
  }

  // Protected routes: redirect to login if not authenticated
  if (to.meta.requiresAuth && !isAuthenticated) {
    return next('/auth/login');
  }

  // Initialize stores if user is authenticated and stores haven't been initialized yet
  if (isAuthenticated && !storesInitialized && to.meta.requiresAuth) {
    try {
      const userStore = useUserStore();
      const accountStore = useAccountStore();
      const themeStore = useThemeStore();
      const projectStore = useProjectStore();
      const trayStates = useTrayStates();
      const modals = useDesktopModalStore();
      
      // Set user data
      userStore.user = userData;
      userStore.isUserAuthenticated = true;
      
      // Initialize stores
      setLoaderStatus('Loading account...');
      await accountStore.initialize();
      
      setLoaderStatus('Applying theme...');
      await themeStore.initializeTheme();
      
      setLoaderStatus('Loading studios...');
      await projectStore.loadStudios();
      
      // Load projects if directory exists
      const projectDirectoryExists = await SettingsService.GetProjectDirectory();
      if (projectDirectoryExists) {
        setLoaderStatus('Loading projects...');
        await projectStore.loadProjects();
        trayStates.refreshData();
      } else if (!isWebMode) {
        modals.setModalVisibility('dirOnboardModal', true);
      }
      
      storesInitialized = true;

      // Check for a .clst file that was opened before the frontend was ready
      if (!isWebMode) {
        try {
          const pendingFile = await AppService.GetPendingOpenFile();
          if (pendingFile) {
            const fileInfo = await ProjectService.InspectClusttaFile(pendingFile);
            if (fileInfo && fileInfo.valid) {
              await projectStore.openClusttaFile(fileInfo);
            }
          }
        } catch (error) {
          console.error('Failed to open pending project file:', error);
        }
      }
    } catch (error) {
      console.error('Failed to initialize stores:', error);
    }
  }

  next();
});

// Mark stores as initialized (call after manual login/signup)
export const markStoresInitialized = () => {
  storesInitialized = true;
};

// Reset initialization flag on logout
export const resetStoreInitialization = () => {
  storesInitialized = false;
};

export default router;
