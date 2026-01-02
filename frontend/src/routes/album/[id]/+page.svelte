<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Play, Shuffle, Clock, MoreHorizontal } from 'lucide-svelte';
	import { Button, Skeleton, Dropdown } from '$lib/components/ui';
	import TrackRow from '$lib/components/TrackRow.svelte';
	import { getAlbum, getAlbumArtworkUrl } from '$lib/api/albums';
	import { formatDuration, formatTotalDuration, pluralize } from '$lib/utils';
	import { playTracks } from '$lib/stores/player';
	import { getAudioController } from '$lib/audio';
	import { shuffleArray } from '$lib/utils';
	import type { Album } from '$lib/api/types';

	let album = $state<Album | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const albumId = $derived($page.params.id);

	onMount(() => {
		loadAlbum();
	});

	$effect(() => {
		if (albumId) {
			loadAlbum();
		}
	});

	async function loadAlbum() {
		loading = true;
		error = null;

		try {
			album = await getAlbum(albumId);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load album';
		} finally {
			loading = false;
		}
	}

	function handlePlayAll() {
		if (!album?.tracks) return;

		const playerTracks = album.tracks.map((t) => ({
			id: t.id,
			title: t.title,
			duration: t.duration,
			trackNumber: t.trackNumber,
			format: t.format,
			albumId: album!.id,
			albumTitle: album!.title,
			artistId: album!.artistId,
			artistName: album!.artistName
		}));

		playTracks(playerTracks);
		getAudioController().play();
	}

	function handleShuffle() {
		if (!album?.tracks) return;

		const playerTracks = album.tracks.map((t) => ({
			id: t.id,
			title: t.title,
			duration: t.duration,
			trackNumber: t.trackNumber,
			format: t.format,
			albumId: album!.id,
			albumTitle: album!.title,
			artistId: album!.artistId,
			artistName: album!.artistName
		}));

		playTracks(shuffleArray(playerTracks));
		getAudioController().play();
	}

	const totalDuration = $derived(
		album?.tracks?.reduce((sum, t) => sum + t.duration, 0) || 0
	);

	const menuItems = [
		{ label: 'Add to Queue', action: () => {} },
		{ label: 'Add to Playlist', action: () => {} },
		{ label: 'Share', action: () => {} }
	];
</script>

<svelte:head>
	<title>{album?.title || 'Album'} - Harmony</title>
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
		<Button variant="secondary" onclick={loadAlbum}>Try Again</Button>
	</div>
{:else if album}
	<!-- Album Header -->
	<div class="relative">
		<!-- Gradient Background -->
		<div
			class="absolute inset-0 h-80 bg-gradient-to-b from-accent/20 to-surface -z-10"
		></div>

		<div class="p-6 pt-8">
			<div class="flex flex-col md:flex-row gap-6 mb-8">
				<!-- Album Artwork -->
				<div class="w-48 h-48 md:w-56 md:h-56 flex-shrink-0 shadow-2xl rounded-lg overflow-hidden">
					<img
						src={getAlbumArtworkUrl(album.id, 'large')}
						alt={album.title}
						class="w-full h-full object-cover"
					/>
				</div>

				<!-- Album Info -->
				<div class="flex flex-col justify-end">
					<span class="text-sm font-medium uppercase tracking-wider mb-2">Album</span>
					<h1 class="text-4xl md:text-5xl font-bold mb-4">{album.title}</h1>
					<div class="flex items-center gap-2 text-sm text-text-secondary">
						<a
							href="/artist/{album.artistId}"
							class="font-medium text-text-primary hover:underline"
						>
							{album.artistName || 'Unknown Artist'}
						</a>
						{#if album.year}
							<span>•</span>
							<span>{album.year}</span>
						{/if}
						{#if album.tracks}
							<span>•</span>
							<span>{album.tracks.length} {pluralize(album.tracks.length, 'song')}</span>
							<span>•</span>
							<span>{formatTotalDuration(totalDuration)}</span>
						{/if}
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
				>
					<Play size={20} class="mr-2" />
					Play
				</Button>
				<Button
					variant="secondary"
					size="lg"
					class="rounded-full"
					onclick={handleShuffle}
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
			{#if album.tracks && album.tracks.length > 0}
				<div class="bg-surface/50 rounded-lg overflow-hidden">
					<!-- Header -->
					<div class="flex items-center gap-4 px-4 py-2 border-b border-surface-border text-xs text-text-muted uppercase tracking-wider">
						<span class="w-8 text-center">#</span>
						<span class="flex-1">Title</span>
						<Clock size={14} class="w-12 text-center" />
						<span class="w-8"></span>
					</div>

					<!-- Tracks -->
					{#each album.tracks as track (track.id)}
						<TrackRow
							{track}
							albumTitle={album.title}
							artistName={album.artistName}
							albumId={album.id}
							artistId={album.artistId}
							showArtist={false}
						/>
					{/each}
				</div>
			{/if}

			<!-- Album Info Footer -->
			<div class="mt-8 text-sm text-text-muted">
				{#if album.year}
					<p>{album.year}</p>
				{/if}
			</div>
		</div>
	</div>
{/if}
