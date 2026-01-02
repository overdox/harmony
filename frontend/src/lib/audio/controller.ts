import { get } from 'svelte/store';
import {
	type Track,
	type RepeatMode,
	currentTrack,
	queue,
	isPlaying,
	currentTime,
	duration,
	volume,
	repeat,
	isLoading,
	getNextTrack,
	getPreviousTrack,
	pushToHistory,
	returnToQueue
} from '$lib/stores/player';
import { getStreamUrl, getArtworkUrl } from '$lib/api/client';

export type AudioQuality = 'original' | 'high' | 'medium' | 'low';

interface AudioControllerOptions {
	quality?: AudioQuality;
	preloadNext?: boolean;
	crossfadeDuration?: number;
}

class AudioController {
	private audio: HTMLAudioElement;
	private nextAudio: HTMLAudioElement | null = null;
	private quality: AudioQuality = 'original';
	private preloadNext: boolean = true;
	private crossfadeDuration: number = 0;
	private isInitialized: boolean = false;
	private unsubscribers: (() => void)[] = [];

	constructor(options: AudioControllerOptions = {}) {
		this.audio = new Audio();
		this.quality = options.quality || 'original';
		this.preloadNext = options.preloadNext ?? true;
		this.crossfadeDuration = options.crossfadeDuration || 0;

		this.setupAudioElement();
		this.setupStoreSubscriptions();
		this.setupMediaSession();
		this.isInitialized = true;
	}

	private setupAudioElement(): void {
		// Audio element configuration
		this.audio.preload = 'auto';

		// Event listeners
		this.audio.addEventListener('timeupdate', this.handleTimeUpdate.bind(this));
		this.audio.addEventListener('durationchange', this.handleDurationChange.bind(this));
		this.audio.addEventListener('ended', this.handleEnded.bind(this));
		this.audio.addEventListener('error', this.handleError.bind(this));
		this.audio.addEventListener('waiting', this.handleWaiting.bind(this));
		this.audio.addEventListener('canplay', this.handleCanPlay.bind(this));
		this.audio.addEventListener('play', () => isPlaying.set(true));
		this.audio.addEventListener('pause', () => isPlaying.set(false));
	}

	private setupStoreSubscriptions(): void {
		// Subscribe to current track changes
		const trackUnsub = currentTrack.subscribe((track) => {
			if (track && this.isInitialized) {
				this.loadTrack(track);
			}
		});
		this.unsubscribers.push(trackUnsub);

		// Subscribe to play state changes
		const playUnsub = isPlaying.subscribe((playing) => {
			if (this.isInitialized && this.audio.src) {
				if (playing && this.audio.paused) {
					this.audio.play().catch(this.handlePlayError.bind(this));
				} else if (!playing && !this.audio.paused) {
					this.audio.pause();
				}
			}
		});
		this.unsubscribers.push(playUnsub);

		// Subscribe to volume changes
		const volUnsub = volume.subscribe((vol) => {
			this.audio.volume = vol;
		});
		this.unsubscribers.push(volUnsub);

		// Subscribe to seek requests
		const seekUnsub = currentTime.subscribe((time) => {
			// Only seek if the difference is significant (avoid feedback loop)
			if (Math.abs(this.audio.currentTime - time) > 1) {
				this.audio.currentTime = time;
			}
		});
		this.unsubscribers.push(seekUnsub);
	}

	private setupMediaSession(): void {
		if (!('mediaSession' in navigator)) return;

		navigator.mediaSession.setActionHandler('play', () => this.play());
		navigator.mediaSession.setActionHandler('pause', () => this.pause());
		navigator.mediaSession.setActionHandler('previoustrack', () => this.previous());
		navigator.mediaSession.setActionHandler('nexttrack', () => this.next());
		navigator.mediaSession.setActionHandler('seekto', (details) => {
			if (details.seekTime !== undefined) {
				this.seek(details.seekTime);
			}
		});
		navigator.mediaSession.setActionHandler('seekbackward', (details) => {
			const skipTime = details.seekOffset || 10;
			this.seek(Math.max(0, this.audio.currentTime - skipTime));
		});
		navigator.mediaSession.setActionHandler('seekforward', (details) => {
			const skipTime = details.seekOffset || 10;
			this.seek(Math.min(this.audio.duration, this.audio.currentTime + skipTime));
		});
	}

	private updateMediaSessionMetadata(track: Track): void {
		if (!('mediaSession' in navigator)) return;

		const artwork: MediaImage[] = [];
		if (track.albumId) {
			artwork.push(
				{ src: getArtworkUrl('album', track.albumId, 'thumbnail'), sizes: '64x64', type: 'image/jpeg' },
				{ src: getArtworkUrl('album', track.albumId, 'small'), sizes: '150x150', type: 'image/jpeg' },
				{ src: getArtworkUrl('album', track.albumId, 'medium'), sizes: '300x300', type: 'image/jpeg' },
				{ src: getArtworkUrl('album', track.albumId, 'large'), sizes: '600x600', type: 'image/jpeg' }
			);
		}

		navigator.mediaSession.metadata = new MediaMetadata({
			title: track.title,
			artist: track.artistName || 'Unknown Artist',
			album: track.albumTitle || 'Unknown Album',
			artwork
		});

		// Update position state
		this.updateMediaSessionPosition();
	}

	private updateMediaSessionPosition(): void {
		if (!('mediaSession' in navigator) || !navigator.mediaSession.setPositionState) return;

		try {
			if (this.audio.duration && !isNaN(this.audio.duration)) {
				navigator.mediaSession.setPositionState({
					duration: this.audio.duration,
					playbackRate: this.audio.playbackRate,
					position: this.audio.currentTime
				});
			}
		} catch {
			// Ignore errors from invalid position state
		}
	}

	private async loadTrack(track: Track): Promise<void> {
		isLoading.set(true);

		const streamUrl = getStreamUrl(track.id, this.quality);
		this.audio.src = streamUrl;

		this.updateMediaSessionMetadata(track);

		// Start playback if isPlaying is true
		if (get(isPlaying)) {
			try {
				await this.audio.play();
			} catch (err) {
				this.handlePlayError(err);
			}
		}

		// Preload next track
		if (this.preloadNext) {
			this.preloadNextTrack();
		}
	}

	private preloadNextTrack(): void {
		const queueTracks = get(queue);
		if (queueTracks.length > 0) {
			const nextTrack = queueTracks[0];
			this.nextAudio = new Audio();
			this.nextAudio.preload = 'auto';
			this.nextAudio.src = getStreamUrl(nextTrack.id, this.quality);
		} else {
			this.nextAudio = null;
		}
	}

	// Event handlers
	private handleTimeUpdate(): void {
		currentTime.set(this.audio.currentTime);
		this.updateMediaSessionPosition();
	}

	private handleDurationChange(): void {
		duration.set(this.audio.duration || 0);
	}

	private handleEnded(): void {
		const repeatMode = get(repeat);
		const track = get(currentTrack);

		if (repeatMode === 'one' && track) {
			// Repeat current track
			this.audio.currentTime = 0;
			this.audio.play().catch(this.handlePlayError.bind(this));
		} else {
			// Try to play next track
			this.next();
		}
	}

	private handleError(event: Event): void {
		const error = (event.target as HTMLAudioElement).error;
		console.error('Audio playback error:', error?.message || 'Unknown error');
		isLoading.set(false);
		isPlaying.set(false);
	}

	private handlePlayError(error: unknown): void {
		console.error('Failed to play:', error);
		isLoading.set(false);
		isPlaying.set(false);
	}

	private handleWaiting(): void {
		isLoading.set(true);
	}

	private handleCanPlay(): void {
		isLoading.set(false);
	}

	// Public methods
	play(): void {
		if (this.audio.src) {
			this.audio.play().catch(this.handlePlayError.bind(this));
		}
	}

	pause(): void {
		this.audio.pause();
	}

	toggle(): void {
		if (get(isPlaying)) {
			this.pause();
		} else {
			this.play();
		}
	}

	stop(): void {
		this.audio.pause();
		this.audio.currentTime = 0;
		isPlaying.set(false);
		currentTime.set(0);
	}

	seek(time: number): void {
		if (!isNaN(time) && isFinite(time)) {
			this.audio.currentTime = Math.max(0, Math.min(time, this.audio.duration || 0));
			currentTime.set(this.audio.currentTime);
		}
	}

	seekPercent(percent: number): void {
		const time = (percent / 100) * (this.audio.duration || 0);
		this.seek(time);
	}

	next(): void {
		const track = get(currentTrack);
		const repeatMode = get(repeat);

		// Push current track to history
		if (track) {
			pushToHistory(track);
		}

		// Use preloaded audio if available
		if (this.nextAudio && this.nextAudio.src) {
			const queueTracks = get(queue);
			if (queueTracks.length > 0) {
				const nextTrack = queueTracks[0];
				queue.update((q) => q.slice(1));

				// Swap audio elements
				this.audio.pause();
				this.audio = this.nextAudio;
				this.setupAudioElement();
				this.nextAudio = null;

				currentTrack.set(nextTrack);
				this.updateMediaSessionMetadata(nextTrack);
				this.audio.play().catch(this.handlePlayError.bind(this));

				// Preload the next one
				this.preloadNextTrack();
				return;
			}
		}

		// Fallback: get next track normally
		const nextTrack = getNextTrack();

		if (nextTrack) {
			currentTrack.set(nextTrack);
		} else if (repeatMode === 'all' && track) {
			// No more tracks, but repeat all is on - restart queue
			// In a real scenario, you'd restore the full queue here
			this.audio.currentTime = 0;
			this.audio.play().catch(this.handlePlayError.bind(this));
		} else {
			// End of queue
			this.stop();
			currentTrack.set(null);
		}
	}

	previous(): void {
		// If we're more than 3 seconds into the track, restart it
		if (this.audio.currentTime > 3) {
			this.seek(0);
			return;
		}

		const track = get(currentTrack);
		const prevTrack = getPreviousTrack();

		if (prevTrack) {
			// Return current track to front of queue
			if (track) {
				returnToQueue(track);
			}
			currentTrack.set(prevTrack);
		} else {
			// No previous track, just restart current
			this.seek(0);
		}
	}

	setVolume(level: number): void {
		const clamped = Math.max(0, Math.min(1, level));
		volume.set(clamped);
		this.audio.volume = clamped;
	}

	mute(): void {
		this.audio.muted = true;
	}

	unmute(): void {
		this.audio.muted = false;
	}

	toggleMute(): boolean {
		this.audio.muted = !this.audio.muted;
		return this.audio.muted;
	}

	isMuted(): boolean {
		return this.audio.muted;
	}

	setQuality(quality: AudioQuality): void {
		this.quality = quality;
		// Reload current track with new quality
		const track = get(currentTrack);
		if (track) {
			const currentPos = this.audio.currentTime;
			const wasPlaying = !this.audio.paused;
			this.loadTrack(track).then(() => {
				this.seek(currentPos);
				if (wasPlaying) {
					this.play();
				}
			});
		}
	}

	getQuality(): AudioQuality {
		return this.quality;
	}

	getCurrentTime(): number {
		return this.audio.currentTime;
	}

	getDuration(): number {
		return this.audio.duration || 0;
	}

	getBuffered(): TimeRanges | null {
		return this.audio.buffered;
	}

	getBufferedPercent(): number {
		if (this.audio.buffered.length === 0 || !this.audio.duration) return 0;
		return (this.audio.buffered.end(this.audio.buffered.length - 1) / this.audio.duration) * 100;
	}

	// Cleanup
	destroy(): void {
		this.unsubscribers.forEach((unsub) => unsub());
		this.audio.pause();
		this.audio.src = '';
		this.nextAudio?.pause();
		this.nextAudio = null;
	}
}

// Singleton instance
let audioController: AudioController | null = null;

export function getAudioController(options?: AudioControllerOptions): AudioController {
	if (!audioController) {
		audioController = new AudioController(options);
	}
	return audioController;
}

export function destroyAudioController(): void {
	if (audioController) {
		audioController.destroy();
		audioController = null;
	}
}

export default AudioController;
