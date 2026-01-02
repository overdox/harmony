<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Play, Shuffle, MoreHorizontal } from 'lucide-svelte';
	import { Button, Skeleton, Dropdown } from '$lib/components/ui';
	import AlbumCard from '$lib/components/AlbumCard.svelte';
	import TrackRow from '$lib/components/TrackRow.svelte';
	import SectionHeader from '$lib/components/SectionHeader.svelte';
	import { getArtist, getArtistImageUrl } from '$lib/api/artists';
	import { pluralize, shuffleArray } from '$lib/utils';
	import { playTracks, shuffle as shuffleStore } from '$lib/stores/player';
	import { getAudioController } from '$lib/audio';
	import type { Artist } from '$lib/api/types';

	let artist = $state<Artist | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const artistId = $derived($page.params.id);

	onMount(() => {
		loadArtist();
	});

	$effect(() => {
		if (artistId) {
			loadArtist();
		}
	});

	async function loadArtist() {
		loading = true;
		error = null;

		try {
			artist = await getArtist(artistId);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load artist';
		} finally {
			loading = false;
		}
	}

	function handlePlayArtist() {
		if (!artist?.popularTracks) return;

		const playerTracks = artist.popularTracks.map((t) => ({
			id: t.id,
			title: t.title,
			duration: t.duration,
			trackNumber: t.trackNumber,
			format: t.format,
			albumId: t.albumId,
			artistId: artist!.id,
			artistName: artist!.name
		}));

		// Shuffle by default when playing artist
		shuffleStore.set(true);
		playTracks(shuffleArray(playerTracks));
		getAudioController().play();
	}

	function handlePlayPopular() {
		if (!artist?.popularTracks) return;

		const playerTracks = artist.popularTracks.map((t) => ({
			id: t.id,
			title: t.title,
			duration: t.duration,
			trackNumber: t.trackNumber,
			format: t.format,
			albumId: t.albumId,
			artistId: artist!.id,
			artistName: artist!.name
		}));

		playTracks(playerTracks);
		getAudioController().play();
	}

	const menuItems = [
		{ label: 'Add to Queue', action: () => {} },
		{ label: 'Share', action: () => {} }
	];
</script>

<svelte:head>
	<title>{artist?.name || 'Artist'} - Harmony</title>
</svelte:head>

{#if loading}
	<div class="p-6">
		<!-- Header Skeleton -->
		<div class="flex gap-6 mb-8">
			<Skeleton class="w-48 h-48 rounded-full flex-shrink-0" />
			<div class="flex flex-col justify-end">
				<Skeleton class="h-4 w-16 mb-2" />
				<Skeleton class="h-12 w-64 mb-4" />
				<Skeleton class="h-4 w-32" />
			</div>
		</div>
		<!-- Content Skeleton -->
		<div class="space-y-2 mb-8">
			{#each Array(5) as _}
				<Skeleton class="h-14 w-full rounded" />
			{/each}
		</div>
		<Skeleton class="h-7 w-32 mb-4" />
		<div class="flex gap-4">
			{#each Array(6) as _}
				<div class="w-40 flex-shrink-0">
					<Skeleton class="aspect-square rounded-md mb-3" />
					<Skeleton class="h-4 w-32 mb-1" />
					<Skeleton class="h-3 w-24" />
				</div>
			{/each}
		</div>
	</div>
{:else if error}
	<div class="flex flex-col items-center justify-center py-20 text-center">
		<p class="text-text-secondary mb-4">{error}</p>
		<Button variant="secondary" onclick={loadArtist}>Try Again</Button>
	</div>
{:else if artist}
	<!-- Artist Header -->
	<div class="relative">
		<!-- Gradient Background -->
		<div
			class="absolute inset-0 h-96 bg-gradient-to-b from-accent/30 to-surface -z-10"
		></div>

		<div class="p-6 pt-12">
			<div class="flex flex-col md:flex-row items-center md:items-end gap-6 mb-8">
				<!-- Artist Image -->
				<div class="w-48 h-48 md:w-56 md:h-56 flex-shrink-0 shadow-2xl rounded-full overflow-hidden">
					<img
						src={getArtistImageUrl(artist.id, 'large')}
						alt={artist.name}
						class="w-full h-full object-cover"
					/>
				</div>

				<!-- Artist Info -->
				<div class="text-center md:text-left">
					<span class="text-sm font-medium uppercase tracking-wider mb-2 inline-block">Artist</span>
					<h1 class="text-4xl md:text-6xl font-bold mb-4">{artist.name}</h1>
					<p class="text-text-secondary">
						{#if artist.albumCount}
							{artist.albumCount} {pluralize(artist.albumCount, 'album')}
						{/if}
						{#if artist.trackCount}
							{#if artist.albumCount} • {/if}
							{artist.trackCount} {pluralize(artist.trackCount, 'track')}
						{/if}
					</p>
				</div>
			</div>

			<!-- Actions -->
			<div class="flex items-center gap-4 mb-8">
				<Button
					variant="primary"
					size="lg"
					class="rounded-full px-8"
					onclick={handlePlayArtist}
				>
					<Shuffle size={20} class="mr-2" />
					Shuffle
				</Button>
				<Dropdown items={menuItems}>
					<Button variant="ghost" size="lg">
						<MoreHorizontal size={24} />
					</Button>
				</Dropdown>
			</div>

			<!-- Popular Tracks -->
			{#if artist.popularTracks && artist.popularTracks.length > 0}
				<section class="mb-8">
					<SectionHeader title="Popular" />
					<div class="bg-surface/50 rounded-lg overflow-hidden">
						{#each artist.popularTracks.slice(0, 5) as track, i (track.id)}
							<TrackRow
								{track}
								index={i}
								artistName={artist.name}
								artistId={artist.id}
								showAlbum
								showArtist={false}
							/>
						{/each}
					</div>
					{#if artist.popularTracks.length > 5}
						<Button variant="ghost" class="mt-4">
							Show more
						</Button>
					{/if}
				</section>
			{/if}

			<!-- Discography -->
			{#if artist.albums && artist.albums.length > 0}
				<section>
					<SectionHeader title="Discography" href="/library?artist={artist.id}" showMore />
					<div class="flex gap-4 overflow-x-auto pb-4 scrollbar-thin">
						{#each artist.albums as album (album.id)}
							<AlbumCard {album} showArtist={false} />
						{/each}
					</div>
				</section>
			{/if}

			<!-- Bio -->
			{#if artist.bio}
				<section class="mt-8">
					<SectionHeader title="About" />
					<p class="text-text-secondary max-w-3xl leading-relaxed">{artist.bio}</p>
				</section>
			{/if}
		</div>
	</div>
{/if}
