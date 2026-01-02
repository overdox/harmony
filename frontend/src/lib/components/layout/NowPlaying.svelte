<script lang="ts">
	import { clsx } from 'clsx';
	import {
		Play,
		Pause,
		SkipBack,
		SkipForward,
		Volume2,
		VolumeX,
		Repeat,
		Shuffle,
		ListMusic,
		Maximize2
	} from 'lucide-svelte';
	import Button from '../ui/Button.svelte';
	import Slider from '../ui/Slider.svelte';
	import { currentTrack, isPlaying, volume, currentTime, duration, shuffle, repeat } from '$lib/stores/player';

	let showQueue = $state(false);
	let previousVolume = $state(1);

	function togglePlay() {
		isPlaying.update((v) => !v);
	}

	function toggleMute() {
		volume.update((v) => {
			if (v > 0) {
				previousVolume = v;
				return 0;
			}
			return previousVolume;
		});
	}

	function toggleShuffle() {
		shuffle.update((v) => !v);
	}

	function toggleRepeat() {
		repeat.update((v) => {
			if (v === 'off') return 'all';
			if (v === 'all') return 'one';
			return 'off';
		});
	}

	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	function handleSeek(value: number) {
		currentTime.set(value);
	}

	function handleVolumeChange(value: number) {
		volume.set(value / 100);
	}
</script>

<footer
	class="fixed bottom-0 left-0 right-0 h-[var(--spacing-nowplaying)] bg-surface-elevated border-t border-surface-border px-4 z-40"
>
	<div class="flex items-center justify-between h-full max-w-screen-2xl mx-auto">
		<!-- Track Info -->
		<div class="flex items-center gap-4 min-w-[180px] w-[30%]">
			{#if $currentTrack}
				<div class="w-14 h-14 bg-surface-hover rounded overflow-hidden flex-shrink-0">
					{#if $currentTrack.coverArtUrl}
						<img
							src={$currentTrack.coverArtUrl}
							alt={$currentTrack.title}
							class="w-full h-full object-cover"
						/>
					{/if}
				</div>
				<div class="min-w-0">
					<p class="text-sm font-medium truncate">{$currentTrack.title}</p>
					<p class="text-xs text-text-secondary truncate">{$currentTrack.artistName}</p>
				</div>
			{:else}
				<div class="w-14 h-14 bg-surface-hover rounded flex-shrink-0"></div>
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
				>
					<Shuffle size={18} />
				</Button>

				<Button variant="icon" size="sm">
					<SkipBack size={20} />
				</Button>

				<Button
					variant="primary"
					size="md"
					class="w-9 h-9 p-0"
					onclick={togglePlay}
				>
					{#if $isPlaying}
						<Pause size={20} />
					{:else}
						<Play size={20} class="ml-0.5" />
					{/if}
				</Button>

				<Button variant="icon" size="sm">
					<SkipForward size={20} />
				</Button>

				<Button
					variant="icon"
					size="sm"
					class={clsx($repeat !== 'off' && 'text-accent')}
					onclick={toggleRepeat}
				>
					<Repeat size={18} />
					{#if $repeat === 'one'}
						<span class="absolute text-[10px] font-bold">1</span>
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
				class={clsx(showQueue && 'text-accent')}
				onclick={() => (showQueue = !showQueue)}
			>
				<ListMusic size={18} />
			</Button>

			<div class="flex items-center gap-2 w-32">
				<Button variant="icon" size="sm" onclick={toggleMute}>
					{#if $volume === 0}
						<VolumeX size={18} />
					{:else}
						<Volume2 size={18} />
					{/if}
				</Button>
				<Slider
					value={$volume * 100}
					max={100}
					class="flex-1"
					oninput={handleVolumeChange}
				/>
			</div>

			<Button variant="icon" size="sm">
				<Maximize2 size={18} />
			</Button>
		</div>
	</div>
</footer>
