import { writable, derived } from 'svelte/store';
import { getSetupStatus, type SetupStatus } from '$lib/api/setup';

// Setup status store
export const setupStatus = writable<SetupStatus | null>(null);
export const setupLoading = writable(true);
export const setupError = writable<string | null>(null);

// Derived store: is setup required?
export const setupRequired = derived(
	[setupStatus, setupLoading],
	([$setupStatus, $setupLoading]) => {
		if ($setupLoading) return false;
		return $setupStatus !== null && !$setupStatus.completed;
	}
);

// Derived store: is setup completed?
export const setupCompleted = derived(
	[setupStatus, setupLoading],
	([$setupStatus, $setupLoading]) => {
		if ($setupLoading) return false;
		return $setupStatus !== null && $setupStatus.completed;
	}
);

/**
 * Check setup status from API
 */
export async function checkSetupStatus(): Promise<SetupStatus | null> {
	setupLoading.set(true);
	setupError.set(null);

	try {
		const status = await getSetupStatus();
		setupStatus.set(status);
		return status;
	} catch (error) {
		console.error('Failed to check setup status:', error);
		setupError.set('Failed to check setup status');
		return null;
	} finally {
		setupLoading.set(false);
	}
}

/**
 * Mark setup as completed (local state only)
 */
export function markSetupCompleted(): void {
	setupStatus.update((status) => {
		if (status) {
			return { ...status, completed: true };
		}
		return status;
	});
}
