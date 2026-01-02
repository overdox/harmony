<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { clsx } from 'clsx';
	import {
		Play,
		Pause,
		SkipBack,
		SkipForward,
		Volume2,
		VolumeX,
		Volume1,
		Repeat,
		Repeat1,
		Shuffle,
		ListMusic,
		Maximize2
	} from 'lucide-svelte';
	import Button from '../ui/Button.svelte';
	import Slider from '../ui/Slider.svelte';
	import {
		currentTrack,
		isPlaying,
		volume,
		currentTime,
		duration,
		shuffle,
		repeat,
		isLoading,
		toggleShuffle as toggleShuffleStore,
		toggleRepeat as toggleRepeatStore
	} from '$lib/stores/player';
	import { getAudioController, destroyAudioController } from '$lib/audio';
	import { getArtworkUrl } from '$lib/api/client';
	import { showQueuePanel, toggleQueuePanel } from '$lib/stores/ui';
	import type AudioController from '$lib/audio/controller';
	let previousVolume = $state(1);
	let controller: AudioController | null = $state(null);

	onMount(() => {
		controller = getAudioController();
	});

	onDestroy(() => {
		destroyAudioController();
	});

	function togglePlay() {
		controller?.toggle();
	}

	function handlePrevious() {
		controller?.previous();
	}

	function handleNext() {
		controller?.next();
	}

	function toggleMute() {
		if (controller) {
			if ($volume > 0) {
				previousVolume = $volume;
				controller.setVolume(0);
			} else {
				controller.setVolume(previousVolume);
			}
		}
	}

	function toggleShuffle() {
		toggleShuffleStore();
	}

	function toggleRepeat() {
		toggleRepeatStore();
	}

	function formatTime(seconds: number): string {
		if (!seconds || isNaN(seconds)) return '0:00';
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	function handleSeek(value: number) {
		controller?.seek(value);
	}

	function handleVolumeChange(value: number) {
		controller?.setVolume(value / 100);
	}

	function getVolumeIcon(vol: number) {
		if (vol === 0) return VolumeX;
		if (vol < 0.5) return Volume1;
		return Volume2;
	}

	function getCoverArtUrl(track: typeof $currentTrack) {
		if (track?.coverArtUrl) return track.coverArtUrl;
		if (track?.albumId) return getArtworkUrl('album', track.albumId, 'small');
		return null;
	}

	$effect(() => {
		// Keyboard shortcuts for playback
		function handleKeydown(e: KeyboardEvent) {
			// Don't trigger if user is typing in an input
			if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

			switch (e.code) {
				case 'Space':
					e.preventDefault();
					togglePlay();
					break;
				case 'ArrowLeft':
					if (e.shiftKey) {
						controller?.seek(Math.max(0, $currentTime - 10));
					}
					break;
				case 'ArrowRight':
					if (e.shiftKey) {
						controller?.seek(Math.min($duration, $currentTime + 10));
					}
					break;
				case 'ArrowUp':
					if (e.shiftKey) {
						controller?.setVolume(Math.min(1, $volume + 0.1));
					}
					break;
				case 'ArrowDown':
					if (e.shiftKey) {
						controller?.setVolume(Math.max(0, $volume - 0.1));
					}
					break;
				case 'KeyM':
					toggleMute();
					break;
				case 'KeyN':
					handleNext();
					break;
				case 'KeyP':
					if (!e.ctrlKey && !e.metaKey) {
						handlePrevious();
					}
					break;
				case 'KeyS':
					if (!e.ctrlKey && !e.metaKey) {
						toggleShuffle();
					}
					break;
				case 'KeyR':
					toggleRepeat();
					break;
				case 'KeyQ':
					toggleQueuePanel();
					break;
			}
		}

		if (typeof window !== 'undefined') {
			window.addEventListener('keydown', handleKeydown);
			return () => window.removeEventListener('keydown', handleKeydown);
		}
	});
</script>

<footer
	class="fixed bottom-0 left-0 right-0 h-[var(--spacing-nowplaying)] bg-surface-elevated border-t border-surface-border px-4 z-40"
>
	<div class="flex items-center justify-between h-full max-w-screen-2xl mx-auto">
		<!-- Track Info -->
		<div class="flex items-center gap-4 min-w-[180px] w-[30%]">
			{#if $currentTrack}
				{@const coverUrl = getCoverArtUrl($currentTrack)}
				<div class="w-14 h-14 bg-surface-hover rounded overflow-hidden flex-shrink-0">
					{#if coverUrl}
						<img
							src={coverUrl}
							alt={$currentTrack.title}
							class="w-full h-full object-cover"
						/>
					{:else}
						<div class="w-full h-full flex items-center justify-center text-text-muted">
							<Play size={20} />
						</div>
					{/if}
				</div>
				<div class="min-w-0">
					<p class="text-sm font-medium truncate">{$currentTrack.title}</p>
					<p class="text-xs text-text-secondary truncate">{$currentTrack.artistName || 'Unknown Artist'}</p>
				</div>
			{:else}
				<div class="w-14 h-14 bg-surface-hover rounded flex-shrink-0 flex items-center justify-center text-text-muted">
					<Play size={20} />
				</div>
				<div>
					<p class="text-sm text-text-muted">No track playing</p>
				</div>
			{/if}
		</div>

		<!-- Playback Controls -->
		<div class="flex flex-col items-center gap-1 max-w-[40%] w-full">
			<div class="flex items-center gap-2">
				<Button
					variant="icon"
					size="sm"
					class={clsx($shuffle && 'text-accent')}
					onclick={toggleShuffle}
					title="Shuffle (S)"
				>
					<Shuffle size={18} />
				</Button>

				<Button
					variant="icon"
					size="sm"
					onclick={handlePrevious}
					disabled={!$currentTrack}
					title="Previous (P)"
				>
					<SkipBack size={20} />
				</Button>

				<Button
					variant="primary"
					size="md"
					class={clsx('w-9 h-9 p-0', $isLoading && 'animate-pulse')}
					onclick={togglePlay}
					disabled={!$currentTrack}
					title="Play/Pause (Space)"
				>
					{#if $isPlaying}
						<Pause size={20} />
					{:else}
						<Play size={20} class="ml-0.5" />
					{/if}
				</Button>

				<Button
					variant="icon"
					size="sm"
					onclick={handleNext}
					disabled={!$currentTrack}
					title="Next (N)"
				>
					<SkipForward size={20} />
				</Button>

				<Button
					variant="icon"
					size="sm"
					class={clsx('relative', $repeat !== 'off' && 'text-accent')}
					onclick={toggleRepeat}
					title="Repeat (R)"
				>
					{#if $repeat === 'one'}
						<Repeat1 size={18} />
					{:else}
						<Repeat size={18} />
					{/if}
				</Button>
			</div>

			<!-- Progress Bar -->
			<div class="flex items-center gap-2 w-full">
				<span class="text-xs text-text-muted w-10 text-right">
					{formatTime($currentTime)}
				</span>
				<Slider
					value={$currentTime}
					max={$duration || 100}
					class="flex-1"
					showTooltip
					formatTooltip={formatTime}
					oninput={handleSeek}
				/>
				<span class="text-xs text-text-muted w-10">
					{formatTime($duration)}
				</span>
			</div>
		</div>

		<!-- Volume & Additional Controls -->
		<div class="flex items-center justify-end gap-2 min-w-[180px] w-[30%]">
			<Button
				variant="icon"
				size="sm"
				class={clsx($showQueuePanel && 'text-accent')}
				onclick={toggleQueuePanel}
				title="Queue (Q)"
			>
				<ListMusic size={18} />
			</Button>

			<div class="flex items-center gap-2 w-32">
				<Button variant="icon" size="sm" onclick={toggleMute} title="Mute (M)">
					<svelte:component this={getVolumeIcon($volume)} size={18} />
				</Button>
				<Slider
					value={$volume * 100}
					max={100}
					class="flex-1"
					oninput={handleVolumeChange}
				/>
			</div>

			<Button variant="icon" size="sm" title="Full screen">
				<Maximize2 size={18} />
			</Button>
		</div>
	</div>
</footer>
