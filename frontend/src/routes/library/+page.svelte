<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		LayoutGrid,
		List,
		Plus,
		SortAsc,
		SortDesc,
		RefreshCw,
		FolderSync
	} from 'lucide-svelte';
	import { Button, Input, Skeleton, Modal } from '$lib/components/ui';
	import AlbumCard from '$lib/components/AlbumCard.svelte';
	import ArtistCard from '$lib/components/ArtistCard.svelte';
	import TrackRow from '$lib/components/TrackRow.svelte';
	import { getAlbums } from '$lib/api/albums';
	import { getArtists } from '$lib/api/artists';
	import { getTracks } from '$lib/api/tracks';
	import { getPlaylists, createPlaylist } from '$lib/api/playlists';
	import { startLibraryScan, getScanStatus, getLibraryStats } from '$lib/api/search';
	import { debounce, storage } from '$lib/utils';
	import type { Album, Artist, Track, Playlist, LibraryStats, ScanProgress } from '$lib/api/types';

	type Tab = 'playlists' | 'albums' | 'artists' | 'tracks';
	type ViewMode = 'grid' | 'list';
	type SortOrder = 'asc' | 'desc';

	let activeTab = $state<Tab>('albums');
	let viewMode = $state<ViewMode>('grid');
	let sortBy = $state('name');
	let sortOrder = $state<SortOrder>('asc');
	let searchQuery = $state('');
	let currentPage = $state(1);

	let albums = $state<Album[]>([]);
	let artists = $state<Artist[]>([]);
	let tracks = $state<Track[]>([]);
	let playlists = $state<Playlist[]>([]);
	let stats = $state<LibraryStats | null>(null);
	let scanProgress = $state<ScanProgress | null>(null);

	let loading = $state(true);
	let scanning = $state(false);
	let showCreatePlaylist = $state(false);
	let newPlaylistName = $state('');

	const VIEW_MODE_KEY = 'harmony:library-view';

	onMount(() => {
		// Restore view mode preference
		viewMode = storage.get<ViewMode>(VIEW_MODE_KEY, 'grid');

		// Check URL params for initial tab
		const urlTab = $page.url.searchParams.get('tab') as Tab | null;
		if (urlTab && ['playlists', 'albums', 'artists', 'tracks'].includes(urlTab)) {
			activeTab = urlTab;
		}

		loadData();
		loadStats();
		checkScanStatus();
	});

	$effect(() => {
		// Update URL when tab changes
		const url = new URL(window.location.href);
		url.searchParams.set('tab', activeTab);
		goto(url.toString(), { replaceState: true, keepFocus: true });
	});

	$effect(() => {
		// Save view mode preference
		storage.set(VIEW_MODE_KEY, viewMode);
	});

	async function loadData() {
		loading = true;
		currentPage = 1;

		try {
			const params = {
				page: currentPage,
				limit: 50,
				sortBy,
				order: sortOrder,
				q: searchQuery || undefined
			};

			switch (activeTab) {
				case 'albums':
					const albumsRes = await getAlbums(params);
					albums = albumsRes.albums;
					break;
				case 'artists':
					const artistsRes = await getArtists(params);
					artists = artistsRes.artists;
					break;
				case 'tracks':
					const tracksRes = await getTracks(params);
					tracks = tracksRes.tracks;
					break;
				case 'playlists':
					const playlistsRes = await getPlaylists(params);
					playlists = playlistsRes.playlists;
					break;
			}
		} catch (e) {
			console.error('Failed to load data:', e);
		} finally {
			loading = false;
		}
	}

	async function loadStats() {
		try {
			stats = await getLibraryStats();
		} catch (e) {
			console.error('Failed to load stats:', e);
		}
	}

	async function checkScanStatus() {
		try {
			scanProgress = await getScanStatus();
			scanning = scanProgress.status === 'scanning' || scanProgress.status === 'processing';

			if (scanning) {
				// Poll for updates
				setTimeout(checkScanStatus, 2000);
			}
		} catch (e) {
			scanning = false;
		}
	}

	async function handleScan(incremental = false) {
		scanning = true;
		try {
			await startLibraryScan(incremental);
			checkScanStatus();
		} catch (e) {
			console.error('Failed to start scan:', e);
			scanning = false;
		}
	}

	async function handleCreatePlaylist() {
		if (!newPlaylistName.trim()) return;

		try {
			await createPlaylist({ name: newPlaylistName.trim() });
			newPlaylistName = '';
			showCreatePlaylist = false;
			loadData();
		} catch (e) {
			console.error('Failed to create playlist:', e);
		}
	}

	const debouncedSearch = debounce(() => {
		loadData();
	}, 300);

	function handleSearch(e: Event) {
		searchQuery = (e.target as HTMLInputElement).value;
		debouncedSearch();
	}

	function toggleSort() {
		sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
		loadData();
	}

	function handleTabChange(tab: Tab) {
		activeTab = tab;
		searchQuery = '';
		loadData();
	}

	const tabs: { id: Tab; label: string; count?: number }[] = $derived([
		{ id: 'playlists', label: 'Playlists', count: playlists.length },
		{ id: 'albums', label: 'Albums', count: stats?.totalAlbums },
		{ id: 'artists', label: 'Artists', count: stats?.totalArtists },
		{ id: 'tracks', label: 'Tracks', count: stats?.totalTracks }
	]);

	const sortOptions: Record<Tab, { value: string; label: string }[]> = {
		playlists: [
			{ value: 'name', label: 'Name' },
			{ value: 'created', label: 'Date Created' },
			{ value: 'updated', label: 'Date Modified' }
		],
		albums: [
			{ value: 'name', label: 'Name' },
			{ value: 'artist', label: 'Artist' },
			{ value: 'year', label: 'Year' },
			{ value: 'created', label: 'Date Added' }
		],
		artists: [
			{ value: 'name', label: 'Name' },
			{ value: 'albums', label: 'Album Count' }
		],
		tracks: [
			{ value: 'name', label: 'Name' },
			{ value: 'artist', label: 'Artist' },
			{ value: 'album', label: 'Album' },
			{ value: 'duration', label: 'Duration' }
		]
	};
</script>

<svelte:head>
	<title>Library - Harmony</title>
</svelte:head>

<div class="p-6">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-bold">Your Library</h1>
		<div class="flex items-center gap-2">
			{#if scanning}
				<div class="flex items-center gap-2 text-sm text-text-secondary">
					<RefreshCw size={16} class="animate-spin" />
					<span>
						{scanProgress?.status === 'processing'
							? `Processing... ${scanProgress.processedFiles}/${scanProgress.totalFiles}`
							: 'Scanning...'}
					</span>
				</div>
			{:else}
				<Button variant="ghost" onclick={() => handleScan(true)} title="Quick scan (new files only)">
					<FolderSync size={18} />
				</Button>
				<Button variant="ghost" onclick={() => handleScan(false)} title="Full scan">
					<RefreshCw size={18} />
				</Button>
			{/if}
		</div>
	</div>

	<!-- Tabs -->
	<div class="flex items-center gap-2 mb-6 border-b border-surface-border overflow-x-auto">
		{#each tabs as tab}
			<button
				class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px whitespace-nowrap {activeTab === tab.id
					? 'border-accent text-accent'
					: 'border-transparent text-text-secondary hover:text-text-primary'}"
				onclick={() => handleTabChange(tab.id)}
			>
				{tab.label}
				{#if tab.count !== undefined}
					<span class="ml-1 text-text-muted">({tab.count.toLocaleString()})</span>
				{/if}
			</button>
		{/each}
	</div>

	<!-- Toolbar -->
	<div class="flex flex-wrap items-center justify-between gap-4 mb-6">
		<div class="flex items-center gap-4">
			<!-- Search -->
			<Input
				type="search"
				placeholder="Filter {activeTab}..."
				value={searchQuery}
				oninput={handleSearch}
				class="w-64"
			/>

			<!-- Sort -->
			<select
				class="px-3 py-2 bg-surface-elevated rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-accent"
				bind:value={sortBy}
				onchange={() => loadData()}
			>
				{#each sortOptions[activeTab] as option}
					<option value={option.value}>{option.label}</option>
				{/each}
			</select>

			<Button variant="ghost" size="sm" onclick={toggleSort}>
				{#if sortOrder === 'asc'}
					<SortAsc size={18} />
				{:else}
					<SortDesc size={18} />
				{/if}
			</Button>
		</div>

		<div class="flex items-center gap-2">
			<!-- Create Playlist (only on playlists tab) -->
			{#if activeTab === 'playlists'}
				<Button variant="secondary" size="sm" onclick={() => (showCreatePlaylist = true)}>
					<Plus size={16} class="mr-1" />
					New Playlist
				</Button>
			{/if}

			<!-- View Toggle -->
			{#if activeTab !== 'tracks'}
				<div class="flex rounded-md overflow-hidden">
					<Button
						variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
						size="sm"
						class="rounded-none"
						onclick={() => (viewMode = 'grid')}
					>
						<LayoutGrid size={18} />
					</Button>
					<Button
						variant={viewMode === 'list' ? 'secondary' : 'ghost'}
						size="sm"
						class="rounded-none"
						onclick={() => (viewMode = 'list')}
					>
						<List size={18} />
					</Button>
				</div>
			{/if}
		</div>
	</div>

	<!-- Content -->
	{#if loading}
		<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
			{#each Array(12) as _}
				<div>
					<Skeleton class="aspect-square rounded-md mb-3" />
					<Skeleton class="h-4 w-3/4 mb-1" />
					<Skeleton class="h-3 w-1/2" />
				</div>
			{/each}
		</div>
	{:else if activeTab === 'albums'}
		{#if albums.length === 0}
			<div class="text-center py-20 text-text-secondary">
				<p>No albums found</p>
			</div>
		{:else if viewMode === 'grid'}
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
				{#each albums as album (album.id)}
					<AlbumCard {album} size="lg" />
				{/each}
			</div>
		{:else}
			<div class="space-y-2">
				{#each albums as album (album.id)}
					<a
						href="/album/{album.id}"
						class="flex items-center gap-4 p-3 rounded-lg hover:bg-surface-hover transition-colors"
					>
						<img
							src={`/api/v1/artwork/album/${album.id}?size=thumbnail`}
							alt={album.title}
							class="w-12 h-12 rounded object-cover"
						/>
						<div class="flex-1 min-w-0">
							<p class="font-medium truncate">{album.title}</p>
							<p class="text-sm text-text-secondary truncate">{album.artistName}</p>
						</div>
						<span class="text-sm text-text-muted">{album.year || '-'}</span>
					</a>
				{/each}
			</div>
		{/if}
	{:else if activeTab === 'artists'}
		{#if artists.length === 0}
			<div class="text-center py-20 text-text-secondary">
				<p>No artists found</p>
			</div>
		{:else if viewMode === 'grid'}
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
				{#each artists as artist (artist.id)}
					<ArtistCard {artist} size="lg" />
				{/each}
			</div>
		{:else}
			<div class="space-y-2">
				{#each artists as artist (artist.id)}
					<a
						href="/artist/{artist.id}"
						class="flex items-center gap-4 p-3 rounded-lg hover:bg-surface-hover transition-colors"
					>
						<img
							src={`/api/v1/artwork/artist/${artist.id}?size=thumbnail`}
							alt={artist.name}
							class="w-12 h-12 rounded-full object-cover"
						/>
						<div class="flex-1 min-w-0">
							<p class="font-medium truncate">{artist.name}</p>
							<p class="text-sm text-text-secondary">
								{artist.albumCount || 0} albums • {artist.trackCount || 0} tracks
							</p>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	{:else if activeTab === 'tracks'}
		{#if tracks.length === 0}
			<div class="text-center py-20 text-text-secondary">
				<p>No tracks found</p>
			</div>
		{:else}
			<div class="bg-surface-elevated rounded-lg overflow-hidden">
				{#each tracks as track, i (track.id)}
					<TrackRow {track} index={i} showAlbum showArtist />
				{/each}
			</div>
		{/if}
	{:else if activeTab === 'playlists'}
		{#if playlists.length === 0}
			<div class="text-center py-20">
				<p class="text-text-secondary mb-4">No playlists yet</p>
				<Button variant="primary" onclick={() => (showCreatePlaylist = true)}>
					<Plus size={16} class="mr-1" />
					Create Playlist
				</Button>
			</div>
		{:else if viewMode === 'grid'}
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
				{#each playlists as playlist (playlist.id)}
					<a
						href="/playlist/{playlist.id}"
						class="group p-4 bg-surface-elevated rounded-lg hover:bg-surface-hover transition-colors"
					>
						<div class="aspect-square mb-3 rounded-md overflow-hidden bg-surface-hover">
							{#if playlist.coverImageUrl}
								<img
									src={playlist.coverImageUrl}
									alt={playlist.name}
									class="w-full h-full object-cover"
								/>
							{:else}
								<div class="w-full h-full flex items-center justify-center text-4xl">
									🎵
								</div>
							{/if}
						</div>
						<h3 class="font-medium truncate">{playlist.name}</h3>
						<p class="text-sm text-text-secondary truncate">
							{playlist.trackCount} tracks
						</p>
					</a>
				{/each}
			</div>
		{:else}
			<div class="space-y-2">
				{#each playlists as playlist (playlist.id)}
					<a
						href="/playlist/{playlist.id}"
						class="flex items-center gap-4 p-3 rounded-lg hover:bg-surface-hover transition-colors"
					>
						<div class="w-12 h-12 rounded bg-surface-hover flex items-center justify-center">
							{#if playlist.coverImageUrl}
								<img
									src={playlist.coverImageUrl}
									alt={playlist.name}
									class="w-full h-full rounded object-cover"
								/>
							{:else}
								<span class="text-xl">🎵</span>
							{/if}
						</div>
						<div class="flex-1 min-w-0">
							<p class="font-medium truncate">{playlist.name}</p>
							<p class="text-sm text-text-secondary">{playlist.trackCount} tracks</p>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	{/if}
</div>

<!-- Create Playlist Modal -->
<Modal bind:open={showCreatePlaylist} title="Create Playlist">
	<form onsubmit|preventDefault={handleCreatePlaylist} class="space-y-4">
		<Input
			type="text"
			placeholder="Playlist name"
			bind:value={newPlaylistName}
			autofocus
		/>
		<div class="flex justify-end gap-2">
			<Button variant="ghost" onclick={() => (showCreatePlaylist = false)}>
				Cancel
			</Button>
			<Button variant="primary" type="submit" disabled={!newPlaylistName.trim()}>
				Create
			</Button>
		</div>
	</form>
</Modal>
