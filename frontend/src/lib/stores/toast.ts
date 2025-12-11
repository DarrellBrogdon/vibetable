import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info';

export interface Toast {
	id: string;
	type: ToastType;
	message: string;
}

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	function add(type: ToastType, message: string, duration = 3000) {
		const id = Math.random().toString(36).slice(2);
		const toast: Toast = { id, type, message };

		update(toasts => [...toasts, toast]);

		if (duration > 0) {
			setTimeout(() => remove(id), duration);
		}

		return id;
	}

	function remove(id: string) {
		update(toasts => toasts.filter(t => t.id !== id));
	}

	return {
		subscribe,
		success: (message: string, duration?: number) => add('success', message, duration),
		error: (message: string, duration?: number) => add('error', message, duration ?? 5000),
		info: (message: string, duration?: number) => add('info', message, duration),
		remove
	};
}

export const toastStore = createToastStore();
