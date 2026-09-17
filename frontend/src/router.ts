import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from './stores/auth';
import AccessRequestForm from './components/AccessRequestForm.vue';
import RequestList from './components/RequestList.vue';
import LoginForm from './components/LoginForm.vue';

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/requests'
    },
    {
      path: '/login',
      name: 'login',
      component: LoginForm
    },
    {
      path: '/requests',
      name: 'requests',
      component: RequestList,
      meta: { requiresAuth: true }
    },
    {
      path: '/request/new',
      name: 'new-request',
      component: AccessRequestForm,
      meta: { requiresAuth: true }
    }
  ],
});

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore();

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'login', query: { redirect: to.fullPath } });
  } else if (to.name === 'login' && authStore.isAuthenticated) {
    next({ name: 'requests' });
  } else {
    next();
  }
});
