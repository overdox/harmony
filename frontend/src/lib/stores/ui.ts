import { writable } from 'svelte/store';

// Queue panel visibility
export const showQueuePanel = writable(false);

export function toggleQueuePanel() {
	showQueuePanel.update((v) => !v);
}

export function openQueuePanel() {
	showQueuePanel.set(true);
}

export function closeQueuePanel() {
	showQueuePanel.set(false);
}

// Mobile player visibility
export const showMobilePlayer = writable(false);

export function toggleMobilePlayer() {
	showMobilePlayer.update((v) => !v);
}

export function openMobilePlayer() {
	showMobilePlayer.set(true);
}

export function closeMobilePlayer() {
	showMobilePlayer.set(false);
}
