import { writable, derived } from 'svelte/store';

export interface Track {
	id: string;
	title: string;
	duration: number;
	trackNumber?: number;
	format: string;
	albumId?: string;
	albumTitle?: string;
	artistId?: string;
	artistName?: string;
	coverArtUrl?: string;
	streamUrl?: string;
}

export type RepeatMode = 'off' | 'all' | 'one';

// Core player state
export const currentTrack = writable<Track | null>(null);
export const queue = writable<Track[]>([]);
export const queueHistory = writable<Track[]>([]);
export const isPlaying = writable(false);
export const currentTime = writable(0);
export const duration = writable(0);
export const volume = writable(1);
export const shuffle = writable(false);
export const repeat = writable<RepeatMode>('off');
export const isLoading = writable(false);

// Derived stores
export const progress = derived(
	[currentTime, duration],
	([$currentTime, $duration]) => ($duration > 0 ? ($currentTime / $duration) * 100 : 0)
);

export const formattedCurrentTime = derived(currentTime, ($time) => formatTime($time));
export const formattedDuration = derived(duration, ($dur) => formatTime($dur));

export const hasNext = derived(
	[queue, repeat],
	([$queue, $repeat]) => $queue.length > 0 || $repeat === 'all'
);

export const hasPrevious = derived(
	[queueHistory, repeat],
	([$history, $repeat]) => $history.length > 0 || $repeat === 'all'
);

// Helper functions
function formatTime(seconds: number): string {
	if (!seconds || isNaN(seconds)) return '0:00';
	const mins = Math.floor(seconds / 60);
	const secs = Math.floor(seconds % 60);
	return `${mins}:${secs.toString().padStart(2, '0')}`;
}

// Queue management
export function addToQueue(track: Track) {
	queue.update((q) => [...q, track]);
}

export function addToQueueNext(track: Track) {
	queue.update((q) => [track, ...q]);
}

export function removeFromQueue(index: number) {
	queue.update((q) => {
		const newQueue = [...q];
		newQueue.splice(index, 1);
		return newQueue;
	});
}

export function clearQueue() {
	queue.set([]);
}

export function reorderQueue(fromIndex: number, toIndex: number) {
	queue.update((q) => {
		const newQueue = [...q];
		const [item] = newQueue.splice(fromIndex, 1);
		newQueue.splice(toIndex, 0, item);
		return newQueue;
	});
}

// Playback control helpers
export function playTrack(track: Track) {
	currentTrack.set(track);
	isPlaying.set(true);
}

export function playTracks(tracks: Track[], startIndex = 0) {
	if (tracks.length === 0) return;

	const [first, ...rest] = tracks.slice(startIndex);
	currentTrack.set(first);
	queue.set(rest);
	queueHistory.set([]);
	isPlaying.set(true);
}

export function togglePlayPause() {
	isPlaying.update((v) => !v);
}

export function setVolume(value: number) {
	volume.set(Math.max(0, Math.min(1, value)));
}

export function seek(time: number) {
	currentTime.set(time);
}

export function seekPercent(percent: number) {
	let dur: number = 0;
	duration.subscribe((d) => (dur = d))();
	currentTime.set((percent / 100) * dur);
}

export function toggleShuffle() {
	shuffle.update((v) => !v);
}

export function toggleRepeat() {
	repeat.update((mode) => {
		if (mode === 'off') return 'all';
		if (mode === 'all') return 'one';
		return 'off';
	});
}

export function cycleRepeat(): RepeatMode {
	let newMode: RepeatMode = 'off';
	repeat.update((mode) => {
		if (mode === 'off') newMode = 'all';
		else if (mode === 'all') newMode = 'one';
		else newMode = 'off';
		return newMode;
	});
	return newMode;
}

// Get next track from queue (handles shuffle)
export function getNextTrack(): Track | null {
	let next: Track | null = null;
	let shuffleEnabled = false;

	shuffle.subscribe((s) => (shuffleEnabled = s))();

	queue.update((q) => {
		if (q.length === 0) return q;

		if (shuffleEnabled) {
			const randomIndex = Math.floor(Math.random() * q.length);
			next = q[randomIndex];
			return [...q.slice(0, randomIndex), ...q.slice(randomIndex + 1)];
		} else {
			next = q[0];
			return q.slice(1);
		}
	});

	return next;
}

// Get previous track from history
export function getPreviousTrack(): Track | null {
	let prev: Track | null = null;

	queueHistory.update((h) => {
		if (h.length === 0) return h;
		prev = h[h.length - 1];
		return h.slice(0, -1);
	});

	return prev;
}

// Move current track to history
export function pushToHistory(track: Track) {
	queueHistory.update((h) => [...h, track]);
}

// Return current track to front of queue
export function returnToQueue(track: Track) {
	queue.update((q) => [track, ...q]);
}
