<script lang="ts">
	import { clsx } from 'clsx';
	import {
		Play,
		Pause,
		SkipBack,
		SkipForward,
		ChevronDown,
		Volume2,
		VolumeX,
		Repeat,
		Repeat1,
		Shuffle,
		ListMusic,
		Heart
	} from 'lucide-svelte';
	import { Button, Slider, Equalizer } from '$lib/components/ui';
	import {
		currentTrack,
		isPlaying,
		currentTime,
		duration,
		volume,
		shuffle,
		repeat,
		toggleShuffle,
		toggleRepeat
	} from '$lib/stores/player';
	import { getAudioController } from '$lib/audio';
	import { getArtworkUrl } from '$lib/api/client';
	import { toggleQueuePanel } from '$lib/stores/ui';
	import { formatDuration } from '$lib/utils';

	interface Props {
		open?: boolean;
		onclose?: () => void;
	}

	let { open = false, onclose }: Props = $props();

	const controller = getAudioController();

	function handleSeek(value: number) {
		controller.seek(value);
	}

	function getCoverUrl(track: typeof $currentTrack): string | null {
		if (track?.coverArtUrl) return track.coverArtUrl;
		if (track?.albumId) return getArtworkUrl('album', track.albumId, 'large');
		return null;
	}

	const coverUrl = $derived(getCoverUrl($currentTrack));
</script>

{#if open && $currentTrack}
	<div
		class="fixed inset-0 bg-surface z-50 flex flex-col"
		role="dialog"
		aria-modal="true"
	>
		<!-- Header -->
		<header class="flex items-center justify-between p-4">
			<Button variant="ghost" size="sm" onclick={onclose}>
				<ChevronDown size={24} />
			</Button>
			<span class="text-sm font-medium text-text-secondary">Now Playing</span>
			<Button variant="ghost" size="sm" onclick={toggleQueuePanel}>
				<ListMusic size={24} />
			</Button>
		</header>

		<!-- Album Art -->
		<div class="flex-1 flex items-center justify-center p-8">
			<div class="w-full max-w-sm aspect-square rounded-lg overflow-hidden shadow-2xl bg-surface-elevated">
				{#if coverUrl}
					<img
						src={coverUrl}
						alt={$currentTrack.title}
						class="w-full h-full object-cover"
					/>
				{:else}
					<div class="w-full h-full flex items-center justify-center">
						{#if $isPlaying}
							<Equalizer playing={true} size="lg" />
						{:else}
							<Play size={64} class="text-text-muted" />
						{/if}
					</div>
				{/if}
			</div>
		</div>

		<!-- Track Info -->
		<div class="px-8 text-center">
			<h2 class="text-xl font-bold truncate">{$currentTrack.title}</h2>
			<p class="text-text-secondary truncate">
				{$currentTrack.artistName || 'Unknown Artist'}
			</p>
		</div>

		<!-- Progress -->
		<div class="px-8 mt-6">
			<Slider
				value={$currentTime}
				max={$duration || 100}
				class="w-full"
				oninput={handleSeek}
			/>
			<div class="flex justify-between text-xs text-text-muted mt-1">
				<span>{formatDuration($currentTime)}</span>
				<span>{formatDuration($duration)}</span>
			</div>
		</div>

		<!-- Controls -->
		<div class="flex items-center justify-center gap-6 py-8">
			<Button
				variant="ghost"
				size="lg"
				class={clsx($shuffle && 'text-accent')}
				onclick={toggleShuffle}
			>
				<Shuffle size={24} />
			</Button>

			<Button variant="ghost" size="lg" onclick={() => controller.previous()}>
				<SkipBack size={32} />
			</Button>

			<Button
				variant="primary"
				size="lg"
				class="w-16 h-16 rounded-full"
				onclick={() => controller.toggle()}
			>
				{#if $isPlaying}
					<Pause size={32} />
				{:else}
					<Play size={32} class="ml-1" />
				{/if}
			</Button>

			<Button variant="ghost" size="lg" onclick={() => controller.next()}>
				<SkipForward size={32} />
			</Button>

			<Button
				variant="ghost"
				size="lg"
				class={clsx($repeat !== 'off' && 'text-accent')}
				onclick={toggleRepeat}
			>
				{#if $repeat === 'one'}
					<Repeat1 size={24} />
				{:else}
					<Repeat size={24} />
				{/if}
			</Button>
		</div>

		<!-- Volume (Optional on mobile) -->
		<div class="flex items-center justify-center gap-4 px-8 pb-8">
			<VolumeX size={18} class="text-text-muted" />
			<Slider
				value={$volume * 100}
				max={100}
				class="w-48"
				oninput={(v) => controller.setVolume(v / 100)}
			/>
			<Volume2 size={18} class="text-text-muted" />
		</div>
	</div>
{/if}
