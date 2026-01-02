<script lang="ts">
	import { clsx } from 'clsx';
	import {
		X,
		Play,
		Pause,
		GripVertical,
		Trash2,
		ListPlus,
		Save
	} from 'lucide-svelte';
	import { Button, Modal, Input } from '$lib/components/ui';
	import { formatDuration } from '$lib/utils';
	import {
		currentTrack,
		queue,
		queueHistory,
		isPlaying,
		removeFromQueue,
		clearQueue,
		reorderQueue,
		playTrack,
		type Track
	} from '$lib/stores/player';
	import { getAudioController } from '$lib/audio';
	import { createPlaylist, addTrackToPlaylist } from '$lib/api/playlists';
	import { getArtworkUrl } from '$lib/api/client';

	interface Props {
		open?: boolean;
		onclose?: () => void;
	}

	let { open = false, onclose }: Props = $props();

	const currentTrackCover = $derived(getCoverUrl($currentTrack));
	let showSaveModal = $state(false);
	let newPlaylistName = $state('');
	let draggedIndex = $state<number | null>(null);
	let dragOverIndex = $state<number | null>(null);

	function handleClose() {
		onclose?.();
	}

	function handlePlayTrack(track: Track, index: number) {
		// Remove from queue and play
		removeFromQueue(index);
		playTrack(track);
		getAudioController().play();
	}

	function handleRemoveTrack(index: number) {
		removeFromQueue(index);
	}

	function handleClearQueue() {
		clearQueue();
	}

	function handleDragStart(index: number) {
		draggedIndex = index;
	}

	function handleDragOver(e: DragEvent, index: number) {
		e.preventDefault();
		dragOverIndex = index;
	}

	function handleDragEnd() {
		if (draggedIndex !== null && dragOverIndex !== null && draggedIndex !== dragOverIndex) {
			reorderQueue(draggedIndex, dragOverIndex);
		}
		draggedIndex = null;
		dragOverIndex = null;
	}

	async function handleSaveAsPlaylist() {
		if (!newPlaylistName.trim()) return;

		try {
			const playlist = await createPlaylist({ name: newPlaylistName.trim() });

			// Add current track and queue to playlist
			const allTracks: Track[] = [];
			if ($currentTrack) allTracks.push($currentTrack);
			allTracks.push(...$queue);

			for (const track of allTracks) {
				await addTrackToPlaylist(playlist.id, track.id);
			}

			newPlaylistName = '';
			showSaveModal = false;
		} catch (e) {
			console.error('Failed to save queue as playlist:', e);
		}
	}

	function getCoverUrl(track: Track | null): string | null {
		if (!track) return null;
		if (track.coverArtUrl) return track.coverArtUrl;
		if (track.albumId) return getArtworkUrl('album', track.albumId, 'thumbnail');
		return null;
	}

	const totalTracks = $derived(($currentTrack ? 1 : 0) + $queue.length);
</script>

<!-- Backdrop -->
{#if open}
	<div
		class="fixed inset-0 bg-black/50 z-40"
		onclick={handleClose}
		onkeydown={(e) => e.key === 'Escape' && handleClose()}
		role="button"
		tabindex="-1"
	></div>
{/if}

<!-- Panel -->
<aside
	class={clsx(
		'fixed top-0 right-0 h-full w-80 bg-surface-elevated border-l border-surface-border z-50 transform transition-transform duration-300',
		open ? 'translate-x-0' : 'translate-x-full'
	)}
>
	<!-- Header -->
	<div class="flex items-center justify-between p-4 border-b border-surface-border">
		<h2 class="font-bold text-lg">Queue</h2>
		<div class="flex items-center gap-1">
			<Button
				variant="ghost"
				size="sm"
				onclick={() => (showSaveModal = true)}
				disabled={totalTracks === 0}
				title="Save as playlist"
			>
				<Save size={18} />
			</Button>
			<Button
				variant="ghost"
				size="sm"
				onclick={handleClearQueue}
				disabled={$queue.length === 0}
				title="Clear queue"
			>
				<Trash2 size={18} />
			</Button>
			<Button variant="ghost" size="sm" onclick={handleClose}>
				<X size={18} />
			</Button>
		</div>
	</div>

	<!-- Content -->
	<div class="h-[calc(100%-4rem)] overflow-y-auto">
		<!-- Now Playing -->
		{#if $currentTrack}
			<div class="p-4 border-b border-surface-border">
				<h3 class="text-xs font-medium text-text-muted uppercase tracking-wider mb-3">
					Now Playing
				</h3>
				<div class="flex items-center gap-3 p-2 bg-accent/10 rounded-lg">
					<div class="w-10 h-10 rounded overflow-hidden bg-surface-hover flex-shrink-0">
						{#if currentTrackCover}
							<img src={currentTrackCover} alt="" class="w-full h-full object-cover" />
						{/if}
					</div>
					<div class="flex-1 min-w-0">
						<p class="text-sm font-medium truncate text-accent">{$currentTrack.title}</p>
						<p class="text-xs text-text-secondary truncate">
							{$currentTrack.artistName || 'Unknown Artist'}
						</p>
					</div>
					<Button
						variant="ghost"
						size="sm"
						class="flex-shrink-0"
						onclick={() => getAudioController().toggle()}
					>
						{#if $isPlaying}
							<Pause size={16} />
						{:else}
							<Play size={16} />
						{/if}
					</Button>
				</div>
			</div>
		{/if}

		<!-- Queue -->
		<div class="p-4">
			<div class="flex items-center justify-between mb-3">
				<h3 class="text-xs font-medium text-text-muted uppercase tracking-wider">
					Next Up
				</h3>
				<span class="text-xs text-text-muted">
					{$queue.length} {$queue.length === 1 ? 'track' : 'tracks'}
				</span>
			</div>

			{#if $queue.length === 0}
				<div class="text-center py-8">
					<ListPlus size={32} class="mx-auto mb-2 text-text-muted" />
					<p class="text-sm text-text-secondary">Queue is empty</p>
					<p class="text-xs text-text-muted mt-1">
						Add songs from the library to build your queue
					</p>
				</div>
			{:else}
				<div class="space-y-1">
					{#each $queue as track, i (track.id + '-' + i)}
						{@const cover = getCoverUrl(track)}
						<div
							class={clsx(
								'group flex items-center gap-2 p-2 rounded-lg hover:bg-surface-hover transition-colors cursor-pointer',
								dragOverIndex === i && 'border-t-2 border-accent'
							)}
							draggable="true"
							ondragstart={() => handleDragStart(i)}
							ondragover={(e) => handleDragOver(e, i)}
							ondragend={handleDragEnd}
							ondblclick={() => handlePlayTrack(track, i)}
							role="listitem"
						>
							<div class="cursor-grab opacity-0 group-hover:opacity-100 transition-opacity">
								<GripVertical size={14} class="text-text-muted" />
							</div>

							<span class="w-5 text-center text-xs text-text-muted">{i + 1}</span>

							<div class="w-8 h-8 rounded overflow-hidden bg-surface-hover flex-shrink-0">
								{#if cover}
									<img src={cover} alt="" class="w-full h-full object-cover" />
								{/if}
							</div>

							<div class="flex-1 min-w-0">
								<p class="text-sm truncate">{track.title}</p>
								<p class="text-xs text-text-secondary truncate">
									{track.artistName || 'Unknown Artist'}
								</p>
							</div>

							<span class="text-xs text-text-muted flex-shrink-0">
								{formatDuration(track.duration)}
							</span>

							<Button
								variant="ghost"
								size="sm"
								class="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
								onclick={() => handleRemoveTrack(i)}
							>
								<X size={14} />
							</Button>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- History -->
		{#if $queueHistory.length > 0}
			<div class="p-4 border-t border-surface-border">
				<h3 class="text-xs font-medium text-text-muted uppercase tracking-wider mb-3">
					Recently Played
				</h3>
				<div class="space-y-1 opacity-60">
					{#each $queueHistory.slice(-5).reverse() as track, i (track.id + '-history-' + i)}
						{@const cover = getCoverUrl(track)}
						<div class="flex items-center gap-2 p-2 rounded-lg">
							<div class="w-8 h-8 rounded overflow-hidden bg-surface-hover flex-shrink-0">
								{#if cover}
									<img src={cover} alt="" class="w-full h-full object-cover" />
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-sm truncate">{track.title}</p>
								<p class="text-xs text-text-secondary truncate">
									{track.artistName || 'Unknown Artist'}
								</p>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</aside>

<!-- Save as Playlist Modal -->
<Modal bind:open={showSaveModal} title="Save Queue as Playlist">
	<form onsubmit={(e) => { e.preventDefault(); handleSaveAsPlaylist(); }} class="space-y-4">
		<Input
			type="text"
			placeholder="Playlist name"
			bind:value={newPlaylistName}
			autofocus
		/>
		<p class="text-sm text-text-secondary">
			This will create a new playlist with {totalTracks} {totalTracks === 1 ? 'track' : 'tracks'}.
		</p>
		<div class="flex justify-end gap-2">
			<Button variant="ghost" onclick={() => (showSaveModal = false)}>
				Cancel
			</Button>
			<Button variant="primary" type="submit" disabled={!newPlaylistName.trim()}>
				Save
			</Button>
		</div>
	</form>
</Modal>
