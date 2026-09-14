import { createApp } from 'vue';
import { createPinia } from 'pinia';
import './style.css';
import App from './App.vue';
import { router } from './router';
import { apolloClient } from './apollo';

const app = createApp(App);

app.config.globalProperties.$apollo = apolloClient;
app.use(createPinia());
app.use(router);

app.mount('#app');
