<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		Play,
		Shuffle,
		MoreHorizontal,
		Pencil,
		Trash2,
		Clock,
		GripVertical
	} from 'lucide-svelte';
	import { Button, Skeleton, Dropdown, Modal, Input } from '$lib/components/ui';
	import TrackRow from '$lib/components/TrackRow.svelte';
	import {
		getPlaylist,
		updatePlaylist,
		deletePlaylist,
		removeTrackFromPlaylist,
		getPlaylistCoverUrl
	} from '$lib/api/playlists';
	import { formatTotalDuration, pluralize, shuffleArray } from '$lib/utils';
	import { playTracks } from '$lib/stores/player';
	import { getAudioController } from '$lib/audio';
	import type { Playlist } from '$lib/api/types';

	let playlist = $state<Playlist | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let showEditModal = $state(false);
	let showDeleteModal = $state(false);
	let editName = $state('');
	let editDescription = $state('');
	let draggedIndex = $state<number | null>(null);
	let dragOverIndex = $state<number | null>(null);

	const playlistId = $derived($page.params.id);

	onMount(() => {
		loadPlaylist();
	});

	$effect(() => {
		if (playlistId) {
			loadPlaylist();
		}
	});

	async function loadPlaylist() {
		loading = true;
		error = null;

		try {
			playlist = await getPlaylist(playlistId);
			editName = playlist.name;
			editDescription = playlist.description || '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load playlist';
		} finally {
			loading = false;
		}
	}

	function handlePlayAll() {
		if (!playlist?.tracks) return;

		const playerTracks = playlist.tracks.map((t) => ({
			id: t.id,
			title: t.title,
			duration: t.duration,
			trackNumber: t.trackNumber,
			format: t.format,
			albumId: t.albumId,
			artistId: t.artistId
		}));

		playTracks(playerTracks);
		getAudioController().play();
	}

	function handleShuffle() {
		if (!playlist?.tracks) return;

		const playerTracks = playlist.tracks.map((t) => ({
			id: t.id,
			title: t.title,
			duration: t.duration,
			trackNumber: t.trackNumber,
			format: t.format,
			albumId: t.albumId,
			artistId: t.artistId
		}));

		playTracks(shuffleArray(playerTracks));
		getAudioController().play();
	}

	async function handleSaveEdit() {
		if (!playlist || !editName.trim()) return;

		try {
			await updatePlaylist(playlist.id, {
				name: editName.trim(),
				description: editDescription.trim() || undefined
			});
			playlist.name = editName.trim();
			playlist.description = editDescription.trim();
			showEditModal = false;
		} catch (e) {
			console.error('Failed to update playlist:', e);
		}
	}

	async function handleDelete() {
		if (!playlist) return;

		try {
			await deletePlaylist(playlist.id);
			goto('/library?tab=playlists');
		} catch (e) {
			console.error('Failed to delete playlist:', e);
		}
	}

	async function handleRemoveTrack(trackId: string) {
		if (!playlist) return;

		try {
			await removeTrackFromPlaylist(playlist.id, trackId);
			if (playlist.tracks) {
				playlist.tracks = playlist.tracks.filter((t) => t.id !== trackId);
				playlist.trackCount--;
			}
		} catch (e) {
			console.error('Failed to remove track:', e);
		}
	}

	function handleDragStart(index: number) {
		draggedIndex = index;
	}

	function handleDragOver(e: DragEvent, index: number) {
		e.preventDefault();
		dragOverIndex = index;
	}

	function handleDragEnd() {
		if (draggedIndex !== null && dragOverIndex !== null && playlist?.tracks) {
			const tracks = [...playlist.tracks];
			const [draggedItem] = tracks.splice(draggedIndex, 1);
			tracks.splice(dragOverIndex, 0, draggedItem);
			playlist.tracks = tracks;
			// TODO: Save new order to backend
		}
		draggedIndex = null;
		dragOverIndex = null;
	}

	const totalDuration = $derived(
		playlist?.tracks?.reduce((sum, t) => sum + t.duration, 0) || 0
	);

	const menuItems = [
		{ label: 'Edit Details', icon: Pencil, action: () => (showEditModal = true) },
		{ label: 'Delete Playlist', icon: Trash2, action: () => (showDeleteModal = true) }
	];
</script>

<svelte:head>
	<title>{playlist?.name || 'Playlist'} - Harmony</title>
</svelte:head>

{#if loading}
	<div class="p-6">
		<!-- Header Skeleton -->
		<div class="flex gap-6 mb-8">
			<Skeleton class="w-48 h-48 rounded-lg flex-shrink-0" />
			<div class="flex flex-col justify-end">
				<Skeleton class="h-4 w-16 mb-2" />
				<Skeleton class="h-10 w-64 mb-4" />
				<Skeleton class="h-4 w-48" />
			</div>
		</div>
		<!-- Tracks Skeleton -->
		<div class="space-y-2">
			{#each Array(8) as _}
				<Skeleton class="h-14 w-full rounded" />
			{/each}
		</div>
	</div>
{:else if error}
	<div class="flex flex-col items-center justify-center py-20 text-center">
		<p class="text-text-secondary mb-4">{error}</p>
		<Button variant="secondary" onclick={loadPlaylist}>Try Again</Button>
	</div>
{:else if playlist}
	<!-- Playlist Header -->
	<div class="relative">
		<!-- Gradient Background -->
		<div
			class="absolute inset-0 h-80 bg-gradient-to-b from-purple-900/40 to-surface -z-10"
		></div>

		<div class="p-6 pt-8">
			<div class="flex flex-col md:flex-row gap-6 mb-8">
				<!-- Playlist Cover -->
				<div class="w-48 h-48 md:w-56 md:h-56 flex-shrink-0 shadow-2xl rounded-lg overflow-hidden bg-surface-elevated">
					{#if playlist.coverImageUrl}
						<img
							src={playlist.coverImageUrl}
							alt={playlist.name}
							class="w-full h-full object-cover"
						/>
					{:else}
						<div class="w-full h-full flex items-center justify-center text-6xl">
							🎵
						</div>
					{/if}
				</div>

				<!-- Playlist Info -->
				<div class="flex flex-col justify-end">
					<span class="text-sm font-medium uppercase tracking-wider mb-2">Playlist</span>
					<h1 class="text-4xl md:text-5xl font-bold mb-4">{playlist.name}</h1>
					{#if playlist.description}
						<p class="text-text-secondary mb-2">{playlist.description}</p>
					{/if}
					<div class="flex items-center gap-2 text-sm text-text-secondary">
						<span>{playlist.trackCount} {pluralize(playlist.trackCount, 'song')}</span>
						<span>•</span>
						<span>{formatTotalDuration(totalDuration)}</span>
					</div>
				</div>
			</div>

			<!-- Actions -->
			<div class="flex items-center gap-4 mb-6">
				<Button
					variant="primary"
					size="lg"
					class="rounded-full px-8"
					onclick={handlePlayAll}
					disabled={!playlist.tracks || playlist.tracks.length === 0}
				>
					<Play size={20} class="mr-2" />
					Play
				</Button>
				<Button
					variant="secondary"
					size="lg"
					class="rounded-full"
					onclick={handleShuffle}
					disabled={!playlist.tracks || playlist.tracks.length === 0}
				>
					<Shuffle size={20} />
				</Button>
				<Dropdown items={menuItems}>
					<Button variant="ghost" size="lg">
						<MoreHorizontal size={24} />
					</Button>
				</Dropdown>
			</div>

			<!-- Track List -->
			{#if playlist.tracks && playlist.tracks.length > 0}
				<div class="bg-surface/50 rounded-lg overflow-hidden">
					<!-- Header -->
					<div class="flex items-center gap-4 px-4 py-2 border-b border-surface-border text-xs text-text-muted uppercase tracking-wider">
						<span class="w-8"></span>
						<span class="w-8 text-center">#</span>
						<span class="flex-1">Title</span>
						<Clock size={14} class="w-12 text-center" />
						<span class="w-8"></span>
					</div>

					<!-- Tracks -->
					{#each playlist.tracks as track, i (track.id)}
						<div
							class="flex items-center group {dragOverIndex === i ? 'border-t-2 border-accent' : ''}"
							draggable="true"
							ondragstart={() => handleDragStart(i)}
							ondragover={(e) => handleDragOver(e, i)}
							ondragend={handleDragEnd}
						>
							<div class="px-2 cursor-grab opacity-0 group-hover:opacity-100 transition-opacity">
								<GripVertical size={16} class="text-text-muted" />
							</div>
							<div class="flex-1">
								<TrackRow
									{track}
									index={i}
									showAlbum
									showArtist
									onRemove={() => handleRemoveTrack(track.id)}
								/>
							</div>
						</div>
					{/each}
				</div>
			{:else}
				<div class="text-center py-20 bg-surface/50 rounded-lg">
					<p class="text-text-secondary mb-2">This playlist is empty</p>
					<p class="text-sm text-text-muted">
						Search for songs and add them to this playlist
					</p>
				</div>
			{/if}
		</div>
	</div>
{/if}

<!-- Edit Modal -->
<Modal bind:open={showEditModal} title="Edit Playlist">
	<form onsubmit|preventDefault={handleSaveEdit} class="space-y-4">
		<div>
			<label class="block text-sm font-medium mb-1">Name</label>
			<Input
				type="text"
				bind:value={editName}
				placeholder="Playlist name"
			/>
		</div>
		<div>
			<label class="block text-sm font-medium mb-1">Description</label>
			<textarea
				bind:value={editDescription}
				placeholder="Add an optional description"
				class="w-full px-3 py-2 bg-surface-elevated rounded-md text-sm placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-accent resize-none"
				rows="3"
			></textarea>
		</div>
		<div class="flex justify-end gap-2">
			<Button variant="ghost" onclick={() => (showEditModal = false)}>
				Cancel
			</Button>
			<Button variant="primary" type="submit" disabled={!editName.trim()}>
				Save
			</Button>
		</div>
	</form>
</Modal>

<!-- Delete Confirmation Modal -->
<Modal bind:open={showDeleteModal} title="Delete Playlist">
	<p class="text-text-secondary mb-4">
		Are you sure you want to delete "{playlist?.name}"? This action cannot be undone.
	</p>
	<div class="flex justify-end gap-2">
		<Button variant="ghost" onclick={() => (showDeleteModal = false)}>
			Cancel
		</Button>
		<Button variant="primary" class="bg-red-600 hover:bg-red-700" onclick={handleDelete}>
			Delete
		</Button>
	</div>
</Modal>
