<script lang="ts">
	import { clsx } from 'clsx';
	import { Play, Pause, MoreHorizontal, Plus, ListPlus, Radio } from 'lucide-svelte';
	import { Button, Dropdown } from '$lib/components/ui';
	import { formatDuration } from '$lib/utils';
	import {
		currentTrack,
		isPlaying,
		playTrack,
		addToQueue,
		addToQueueNext,
		type Track
	} from '$lib/stores/player';
	import { getAudioController } from '$lib/audio';
	import type { Track as ApiTrack } from '$lib/api/types';

	interface Props {
		track: ApiTrack;
		index?: number;
		showAlbum?: boolean;
		showArtist?: boolean;
		albumTitle?: string;
		artistName?: string;
		albumId?: string;
		artistId?: string;
		onPlay?: () => void;
		onRemove?: () => void;
		draggable?: boolean;
	}

	let {
		track,
		index,
		showAlbum = false,
		showArtist = true,
		albumTitle,
		artistName,
		albumId,
		artistId,
		onPlay,
		onRemove,
		draggable = false
	}: Props = $props();

	let isHovered = $state(false);
	let showMenu = $state(false);

	const isCurrentTrack = $derived($currentTrack?.id === track.id);
	const isCurrentlyPlaying = $derived(isCurrentTrack && $isPlaying);

	function handlePlay() {
		if (onPlay) {
			onPlay();
		} else {
			const playerTrack: Track = {
				id: track.id,
				title: track.title,
				duration: track.duration,
				trackNumber: track.trackNumber,
				format: track.format,
				albumId: albumId || track.albumId,
				albumTitle: albumTitle,
				artistId: artistId || track.artistId,
				artistName: artistName
			};
			playTrack(playerTrack);
			getAudioController().play();
		}
	}

	function handleTogglePlay() {
		if (isCurrentTrack) {
			getAudioController().toggle();
		} else {
			handlePlay();
		}
	}

	function handleAddToQueue() {
		const playerTrack: Track = {
			id: track.id,
			title: track.title,
			duration: track.duration,
			trackNumber: track.trackNumber,
			format: track.format,
			albumId: albumId || track.albumId,
			albumTitle: albumTitle,
			artistId: artistId || track.artistId,
			artistName: artistName
		};
		addToQueue(playerTrack);
		showMenu = false;
	}

	function handlePlayNext() {
		const playerTrack: Track = {
			id: track.id,
			title: track.title,
			duration: track.duration,
			trackNumber: track.trackNumber,
			format: track.format,
			albumId: albumId || track.albumId,
			albumTitle: albumTitle,
			artistId: artistId || track.artistId,
			artistName: artistName
		};
		addToQueueNext(playerTrack);
		showMenu = false;
	}

	const menuItems = $derived.by(() => {
		const items = [
			{ label: 'Play Next', icon: ListPlus, action: handlePlayNext },
			{ label: 'Add to Queue', icon: Plus, action: handleAddToQueue },
			{ label: 'Go to Artist', icon: Radio, action: () => {}, href: `/artist/${artistId || track.artistId}` }
		];
		if (onRemove) {
			items.push({ label: 'Remove', icon: MoreHorizontal, action: onRemove });
		}
		return items;
	});
</script>

<div
	class={clsx(
		'group flex items-center gap-4 px-4 py-2 rounded-md transition-colors',
		isHovered && 'bg-surface-hover',
		isCurrentTrack && 'bg-surface-hover/50'
	)}
	onmouseenter={() => (isHovered = true)}
	onmouseleave={() => (isHovered = false)}
	ondblclick={handlePlay}
	role="row"
	tabindex="0"
	onkeydown={(e) => e.key === 'Enter' && handlePlay()}
>
	<!-- Track Number / Play Button -->
	<div class="w-8 flex items-center justify-center">
		{#if isHovered || isCurrentTrack}
			<Button variant="ghost" size="sm" class="w-8 h-8 p-0" onclick={handleTogglePlay}>
				{#if isCurrentlyPlaying}
					<Pause size={16} />
				{:else}
					<Play size={16} class="ml-0.5" />
				{/if}
			</Button>
		{:else if index !== undefined}
			<span class="text-sm text-text-muted">{index + 1}</span>
		{:else}
			<span class="text-sm text-text-muted">{track.trackNumber || '-'}</span>
		{/if}
	</div>

	<!-- Track Info -->
	<div class="flex-1 min-w-0">
		<p class={clsx('text-sm truncate', isCurrentTrack && 'text-accent')}>
			{track.title}
		</p>
		{#if showArtist || showAlbum}
			<p class="text-xs text-text-secondary truncate">
				{#if showArtist}
					<a href="/artist/{artistId || track.artistId}" class="hover:underline">
						{artistName || 'Unknown Artist'}
					</a>
				{/if}
				{#if showArtist && showAlbum}
					<span> • </span>
				{/if}
				{#if showAlbum}
					<a href="/album/{albumId || track.albumId}" class="hover:underline">
						{albumTitle || 'Unknown Album'}
					</a>
				{/if}
			</p>
		{/if}
	</div>

	<!-- Duration -->
	<span class="text-sm text-text-muted w-12 text-right">
		{formatDuration(track.duration)}
	</span>

	<!-- Actions -->
	<div class="w-8 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
		<Dropdown
			items={menuItems}
			bind:open={showMenu}
		>
			<Button variant="ghost" size="sm" class="w-8 h-8 p-0">
				<MoreHorizontal size={16} />
			</Button>
		</Dropdown>
	</div>
</div>
