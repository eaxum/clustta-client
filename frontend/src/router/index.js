import { createRouter, createWebHistory } from "vue-router";

// Only use router in web mode
const isWebMode = import.meta.env.VITE_PLATFORM === 'web';

const routes = isWebMode ? [
  {
    path: '/',
    name: 'home',
    component: () => import('@/instances/desktop/ClusttaDesktop.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/user/:identifier',
    name: 'public-profile',
    component: () => import('@/instances/web/PublicUserProfile.vue'),
    meta: { requiresAuth: false, isPublic: true }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    redirect: '/'
  }
] : [];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// Navigation guard for auth (only in web mode)
if (isWebMode) {
  router.beforeEach((to, from, next) => {
    // Public routes don't need auth check
    if (to.meta.isPublic) {
      next();
      return;
    }

    // For protected routes, check if user is authenticated
    // This will be handled by the component/store, just pass through for now
    next();
  });
}

export default router;
