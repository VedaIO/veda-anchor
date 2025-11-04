import 'bootstrap/dist/js/bootstrap.bundle.min.js';
import './global.css';
import App from './App.svelte';

console.log('Svelte app initialized');

const app = new App({
  target: document.getElementById('app'),
  props: {
    url: window.location.pathname,
  },
});

export default app;
