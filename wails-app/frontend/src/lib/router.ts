import { writable } from 'svelte/store';

// Get initial path from hash, defaulting to '/'
function getHashPath(): string {
  const hash = window.location.hash.slice(1); // Remove the '#'
  return hash || '/';
}

export const currentPath = writable(getHashPath());

export function navigate(path: string) {
  window.location.hash = '#{path}';
  currentPath.set(path);
}

// Listen for hash changes and update the store
window.addEventListener('hashchange', () => {
  currentPath.set(getHashPath());
});
