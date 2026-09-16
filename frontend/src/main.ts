import { createApp } from 'vue';
import { createPinia } from 'pinia';
import './style.css';
import App from './App.vue';
import { router } from './router';
import { apolloClient } from './apollo';
import { DefaultApolloClient } from '@vue/apollo-composable';

const app = createApp(App);

app.provide(DefaultApolloClient, apolloClient);
app.use(createPinia());
app.use(router);

app.mount('#app');
