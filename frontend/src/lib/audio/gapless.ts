/**
 * Gapless Playback Module
 *
 * Uses Web Audio API for seamless audio transitions between tracks.
 * Falls back to HTML5 Audio preloading if Web Audio is unavailable.
 */

import { get } from 'svelte/store';
import { queue, currentTrack, isPlaying } from '$lib/stores/player';
import { getStreamUrl } from '$lib/api/client';
import type { Track } from '$lib/stores/player';

export interface GaplessOptions {
	crossfadeDuration?: number; // Duration of crossfade in seconds (0 = no crossfade)
	preloadTime?: number; // When to start preloading next track (seconds before end)
	bufferAhead?: number; // Number of tracks to buffer ahead
}

interface BufferedTrack {
	track: Track;
	buffer: AudioBuffer;
}

class GaplessPlayer {
	private audioContext: AudioContext | null = null;
	private currentSource: AudioBufferSourceNode | null = null;
	private nextSource: AudioBufferSourceNode | null = null;
	private gainNode: GainNode | null = null;
	private crossfadeGain: GainNode | null = null;
	private bufferedTracks: Map<string, BufferedTrack> = new Map();
	private options: Required<GaplessOptions>;
	private isSupported: boolean = false;
	private startTime: number = 0;
	private pauseTime: number = 0;

	constructor(options: GaplessOptions = {}) {
		this.options = {
			crossfadeDuration: options.crossfadeDuration ?? 0,
			preloadTime: options.preloadTime ?? 5,
			bufferAhead: options.bufferAhead ?? 2
		};

		this.checkSupport();
	}

	private checkSupport(): void {
		this.isSupported = typeof AudioContext !== 'undefined' || typeof (window as any).webkitAudioContext !== 'undefined';
	}

	async initialize(): Promise<boolean> {
		if (!this.isSupported) return false;

		try {
			const AudioContextClass = AudioContext || (window as any).webkitAudioContext;
			this.audioContext = new AudioContextClass();

			// Create gain nodes for volume and crossfade
			this.gainNode = this.audioContext.createGain();
			this.crossfadeGain = this.audioContext.createGain();

			this.gainNode.connect(this.crossfadeGain);
			this.crossfadeGain.connect(this.audioContext.destination);

			return true;
		} catch (error) {
			console.warn('Failed to initialize Web Audio API:', error);
			this.isSupported = false;
			return false;
		}
	}

	async loadTrack(track: Track, quality: string = 'original'): Promise<AudioBuffer | null> {
		if (!this.audioContext) return null;

		// Check if already buffered
		if (this.bufferedTracks.has(track.id)) {
			return this.bufferedTracks.get(track.id)!.buffer;
		}

		try {
			const url = getStreamUrl(track.id, quality);
			const response = await fetch(url);
			const arrayBuffer = await response.arrayBuffer();
			const audioBuffer = await this.audioContext.decodeAudioData(arrayBuffer);

			this.bufferedTracks.set(track.id, { track, buffer: audioBuffer });

			// Clean up old buffers if we have too many
			this.cleanupBuffers();

			return audioBuffer;
		} catch (error) {
			console.error('Failed to load track:', error);
			return null;
		}
	}

	private cleanupBuffers(): void {
		const maxBuffers = this.options.bufferAhead + 2; // Current + next + buffer ahead

		if (this.bufferedTracks.size > maxBuffers) {
			const current = get(currentTrack);
			const queueTracks = get(queue);
			const keepIds = new Set<string>();

			// Keep current track
			if (current) keepIds.add(current.id);

			// Keep upcoming tracks
			queueTracks.slice(0, this.options.bufferAhead).forEach((t) => keepIds.add(t.id));

			// Remove tracks not in keep list
			for (const [id] of this.bufferedTracks) {
				if (!keepIds.has(id)) {
					this.bufferedTracks.delete(id);
				}
			}
		}
	}

	async preloadNext(): Promise<void> {
		const queueTracks = get(queue);

		// Preload upcoming tracks
		for (let i = 0; i < Math.min(this.options.bufferAhead, queueTracks.length); i++) {
			const track = queueTracks[i];
			if (!this.bufferedTracks.has(track.id)) {
				await this.loadTrack(track);
			}
		}
	}

	async play(track: Track): Promise<void> {
		if (!this.audioContext || !this.gainNode) {
			console.warn('Gapless playback not available');
			return;
		}

		// Resume audio context if suspended (required for autoplay policy)
		if (this.audioContext.state === 'suspended') {
			await this.audioContext.resume();
		}

		// Load and decode the track
		const buffer = await this.loadTrack(track);
		if (!buffer) return;

		// Stop current source if playing
		this.stopCurrent();

		// Create new source
		this.currentSource = this.audioContext.createBufferSource();
		this.currentSource.buffer = buffer;
		this.currentSource.connect(this.gainNode);

		// Handle track end
		this.currentSource.onended = () => {
			this.handleTrackEnd();
		};

		// Start playback
		this.startTime = this.audioContext.currentTime;
		this.currentSource.start(0);

		// Start preloading next tracks
		this.preloadNext();
	}

	private stopCurrent(): void {
		if (this.currentSource) {
			try {
				this.currentSource.stop();
				this.currentSource.disconnect();
			} catch {
				// Ignore errors if source already stopped
			}
			this.currentSource = null;
		}
	}

	pause(): void {
		if (this.audioContext) {
			this.pauseTime = this.audioContext.currentTime - this.startTime;
			this.audioContext.suspend();
		}
	}

	resume(): void {
		if (this.audioContext) {
			this.audioContext.resume();
		}
	}

	private async handleTrackEnd(): Promise<void> {
		const queueTracks = get(queue);

		if (queueTracks.length > 0) {
			const nextTrack = queueTracks[0];

			// If crossfade is enabled and next track is buffered, do crossfade
			if (this.options.crossfadeDuration > 0 && this.bufferedTracks.has(nextTrack.id)) {
				await this.crossfadeTo(nextTrack);
			}
		}
	}

	private async crossfadeTo(track: Track): Promise<void> {
		if (!this.audioContext || !this.crossfadeGain) return;

		const buffer = this.bufferedTracks.get(track.id)?.buffer;
		if (!buffer) return;

		// Create next source
		this.nextSource = this.audioContext.createBufferSource();
		this.nextSource.buffer = buffer;

		// Create separate gain for crossfade
		const nextGain = this.audioContext.createGain();
		nextGain.gain.value = 0;
		this.nextSource.connect(nextGain);
		nextGain.connect(this.audioContext.destination);

		const now = this.audioContext.currentTime;
		const fadeDuration = this.options.crossfadeDuration;

		// Fade out current
		if (this.gainNode) {
			this.gainNode.gain.linearRampToValueAtTime(0, now + fadeDuration);
		}

		// Fade in next
		nextGain.gain.linearRampToValueAtTime(1, now + fadeDuration);

		// Start next track
		this.nextSource.start();

		// After crossfade, swap sources
		setTimeout(() => {
			this.stopCurrent();
			this.currentSource = this.nextSource;
			this.nextSource = null;

			// Reset gain
			if (this.gainNode) {
				this.gainNode.gain.value = 1;
			}
		}, fadeDuration * 1000);
	}

	setVolume(level: number): void {
		if (this.gainNode) {
			this.gainNode.gain.value = Math.max(0, Math.min(1, level));
		}
	}

	getCurrentTime(): number {
		if (!this.audioContext) return 0;
		return this.audioContext.currentTime - this.startTime;
	}

	getDuration(): number {
		return this.currentSource?.buffer?.duration ?? 0;
	}

	seek(time: number): void {
		const track = get(currentTrack);
		if (!track || !this.currentSource || !this.audioContext) return;

		const buffer = this.currentSource.buffer;
		if (!buffer) return;

		// Stop current and create new source at new position
		this.stopCurrent();

		this.currentSource = this.audioContext.createBufferSource();
		this.currentSource.buffer = buffer;
		this.currentSource.connect(this.gainNode!);
		this.currentSource.onended = () => this.handleTrackEnd();

		this.startTime = this.audioContext.currentTime - time;
		this.currentSource.start(0, time);
	}

	isAvailable(): boolean {
		return this.isSupported && this.audioContext !== null;
	}

	destroy(): void {
		this.stopCurrent();
		if (this.nextSource) {
			try {
				this.nextSource.stop();
				this.nextSource.disconnect();
			} catch {
				// Ignore
			}
		}
		this.bufferedTracks.clear();
		if (this.audioContext) {
			this.audioContext.close();
			this.audioContext = null;
		}
	}
}

// Singleton instance
let gaplessPlayer: GaplessPlayer | null = null;

export async function getGaplessPlayer(options?: GaplessOptions): Promise<GaplessPlayer | null> {
	if (!gaplessPlayer) {
		gaplessPlayer = new GaplessPlayer(options);
		const success = await gaplessPlayer.initialize();
		if (!success) {
			gaplessPlayer = null;
		}
	}
	return gaplessPlayer;
}

export function destroyGaplessPlayer(): void {
	if (gaplessPlayer) {
		gaplessPlayer.destroy();
		gaplessPlayer = null;
	}
}

export default GaplessPlayer;
