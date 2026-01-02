<script lang="ts">
	import { onMount } from 'svelte';
	import { Skeleton } from '$lib/components/ui';
	import AlbumCard from '$lib/components/AlbumCard.svelte';
	import TrackRow from '$lib/components/TrackRow.svelte';
	import SectionHeader from '$lib/components/SectionHeader.svelte';
	import { getRecentAlbums, getRandomAlbums } from '$lib/api/albums';
	import { getRecentTracks, getRandomTracks } from '$lib/api/tracks';
	import { getLibraryStats } from '$lib/api/search';
	import type { Album, Track, LibraryStats } from '$lib/api/types';

	let recentAlbums = $state<Album[]>([]);
	let randomAlbums = $state<Album[]>([]);
	let recentTracks = $state<Track[]>([]);
	let randomTracks = $state<Track[]>([]);
	let stats = $state<LibraryStats | null>(null);

	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			const [recent, random, recTracks, randTracks, libStats] = await Promise.all([
				getRecentAlbums(10),
				getRandomAlbums(10),
				getRecentTracks(5),
				getRandomTracks(5),
				getLibraryStats()
			]);

			recentAlbums = recent;
			randomAlbums = random;
			recentTracks = recTracks;
			randomTracks = randTracks;
			stats = libStats;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load library';
		} finally {
			loading = false;
		}
	});

	function getGreeting(): string {
		const hour = new Date().getHours();
		if (hour < 12) return 'Good morning';
		if (hour < 18) return 'Good afternoon';
		return 'Good evening';
	}
</script>

<svelte:head>
	<title>Home - Harmony</title>
</svelte:head>

<div class="p-6 space-y-8">
	<!-- Header -->
	<header>
		<h1 class="text-3xl font-bold mb-2">{getGreeting()}</h1>
		{#if stats}
			<p class="text-text-secondary">
				{stats.totalTracks.toLocaleString()} tracks • {stats.totalAlbums.toLocaleString()} albums • {stats.totalArtists.toLocaleString()} artists
			</p>
		{/if}
	</header>

	{#if loading}
		<!-- Loading Skeletons -->
		<section>
			<Skeleton class="h-7 w-48 mb-4" />
			<div class="flex gap-4 overflow-hidden">
				{#each Array(6) as _}
					<div class="w-40 flex-shrink-0">
						<Skeleton class="aspect-square rounded-md mb-3" />
						<Skeleton class="h-4 w-32 mb-1" />
						<Skeleton class="h-3 w-24" />
					</div>
				{/each}
			</div>
		</section>

		<section>
			<Skeleton class="h-7 w-48 mb-4" />
			<div class="flex gap-4 overflow-hidden">
				{#each Array(6) as _}
					<div class="w-40 flex-shrink-0">
						<Skeleton class="aspect-square rounded-md mb-3" />
						<Skeleton class="h-4 w-32 mb-1" />
						<Skeleton class="h-3 w-24" />
					</div>
				{/each}
			</div>
		</section>
	{:else if error}
		<div class="flex flex-col items-center justify-center py-20 text-center">
			<p class="text-text-secondary mb-4">{error}</p>
			<p class="text-sm text-text-muted">
				Make sure the backend is running and your library has been scanned.
			</p>
		</div>
	{:else}
		<!-- Recently Added -->
		{#if recentAlbums.length > 0}
			<section>
				<SectionHeader title="Recently Added" href="/library?tab=albums&sort=recent" showMore />
				<div class="flex gap-4 overflow-x-auto pb-4 scrollbar-thin">
					{#each recentAlbums as album (album.id)}
						<AlbumCard {album} />
					{/each}
				</div>
			</section>
		{/if}

		<!-- Quick Picks -->
		{#if randomAlbums.length > 0}
			<section>
				<SectionHeader title="Quick Picks" />
				<div class="flex gap-4 overflow-x-auto pb-4 scrollbar-thin">
					{#each randomAlbums as album (album.id)}
						<AlbumCard {album} />
					{/each}
				</div>
			</section>
		{/if}

		<!-- Recent Tracks -->
		{#if recentTracks.length > 0}
			<section>
				<SectionHeader title="Recently Played" />
				<div class="bg-surface-elevated rounded-lg overflow-hidden">
					{#each recentTracks as track, i (track.id)}
						<TrackRow {track} index={i} showAlbum showArtist />
					{/each}
				</div>
			</section>
		{/if}

		<!-- Random Tracks -->
		{#if randomTracks.length > 0}
			<section>
				<SectionHeader title="Discover" />
				<div class="bg-surface-elevated rounded-lg overflow-hidden">
					{#each randomTracks as track, i (track.id)}
						<TrackRow {track} index={i} showAlbum showArtist />
					{/each}
				</div>
			</section>
		{/if}

		<!-- Empty State -->
		{#if recentAlbums.length === 0 && randomAlbums.length === 0}
			<div class="flex flex-col items-center justify-center py-20 text-center">
				<h2 class="text-xl font-semibold mb-2">Your library is empty</h2>
				<p class="text-text-secondary mb-4">
					Scan your music folder to get started.
				</p>
				<a
					href="/library"
					class="text-accent hover:underline"
				>
					Go to Library Settings
				</a>
			</div>
		{/if}
	{/if}
</div>
