import { mount } from 'svelte';
import 'bootstrap/dist/js/bootstrap.bundle.min.js';
import './global.css';
import App from './App.svelte';

console.log('Svelte app initialized with Wails');

const target = document.getElementById('app');

if (!target) {
  throw new Error(
    "Failed to find target element 'app'. Check your index.html.",
  );
}

const app = mount(App, {
  target: target,
});

export default app;
