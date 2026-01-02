import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
	id: string;
	type: ToastType;
	title?: string;
	message: string;
	duration?: number;
}

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	let idCounter = 0;

	function add(toast: Omit<Toast, 'id'>) {
		const id = `toast-${++idCounter}`;
		const duration = toast.duration ?? 5000;

		update((toasts) => [...toasts, { ...toast, id }]);

		if (duration > 0) {
			setTimeout(() => dismiss(id), duration);
		}

		return id;
	}

	function dismiss(id: string) {
		update((toasts) => toasts.filter((t) => t.id !== id));
	}

	function clear() {
		update(() => []);
	}

	return {
		subscribe,
		add,
		dismiss,
		clear,
		success: (message: string, title?: string) =>
			add({ type: 'success', message, title }),
		error: (message: string, title?: string) =>
			add({ type: 'error', message, title }),
		warning: (message: string, title?: string) =>
			add({ type: 'warning', message, title }),
		info: (message: string, title?: string) =>
			add({ type: 'info', message, title })
	};
}

export const toasts = createToastStore();

// Convenience functions
export const toast = {
	success: toasts.success,
	error: toasts.error,
	warning: toasts.warning,
	info: toasts.info
};

export const dismissToast = toasts.dismiss;
