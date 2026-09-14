import { createRouter, createWebHistory } from 'vue-router';
import AccessRequestForm from './components/AccessRequestForm.vue';
import RequestList from './components/RequestList.vue';

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/requests'
    },
    {
      path: '/requests',
      name: 'requests',
      component: RequestList
    },
    {
      path: '/request/new',
      name: 'new-request',
      component: AccessRequestForm
    }
  ],
});
