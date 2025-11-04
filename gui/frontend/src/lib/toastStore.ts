import { writable } from 'svelte/store';

export interface ToastMessage {
  message: string;
  type: 'success' | 'error' | 'info';
}

export const toast = writable<ToastMessage | null>(null);

export function showToast(
  message: string,
  type: ToastMessage['type'] = 'info'
) {
  toast.set({ message, type });
}
