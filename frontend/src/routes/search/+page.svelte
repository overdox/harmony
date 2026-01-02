<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Search, X, Clock, TrendingUp } from 'lucide-svelte';
	import { Input, Button, Skeleton } from '$lib/components/ui';
	import AlbumCard from '$lib/components/AlbumCard.svelte';
	import ArtistCard from '$lib/components/ArtistCard.svelte';
	import TrackRow from '$lib/components/TrackRow.svelte';
	import SectionHeader from '$lib/components/SectionHeader.svelte';
	import { search as searchApi } from '$lib/api/search';
	import { debounce, storage } from '$lib/utils';
	import type { SearchResults } from '$lib/api/types';

	type Tab = 'all' | 'tracks' | 'albums' | 'artists';

	let query = $state('');
	let activeTab = $state<Tab>('all');
	let results = $state<SearchResults | null>(null);
	let loading = $state(false);
	let searchHistory = $state<string[]>([]);
	let inputElement: HTMLInputElement;
	let selectedIndex = $state(-1);

	const HISTORY_KEY = 'harmony:search-history';
	const MAX_HISTORY = 10;

	onMount(() => {
		// Load search history
		searchHistory = storage.get<string[]>(HISTORY_KEY, []);

		// Check for query in URL
		const urlQuery = $page.url.searchParams.get('q');
		if (urlQuery) {
			query = urlQuery;
			performSearch(urlQuery);
		}

		// Focus input
		inputElement?.focus();
	});

	const debouncedSearch = debounce((q: string) => {
		performSearch(q);
	}, 300);

	async function performSearch(q: string) {
		if (!q.trim()) {
			results = null;
			return;
		}

		loading = true;
		try {
			results = await searchApi(q, 50);

			// Update URL
			const url = new URL(window.location.href);
			url.searchParams.set('q', q);
			goto(url.toString(), { replaceState: true, keepFocus: true });

			// Add to history
			addToHistory(q);
		} catch (e) {
			console.error('Search failed:', e);
			results = null;
		} finally {
			loading = false;
		}
	}

	function addToHistory(q: string) {
		const trimmed = q.trim();
		if (!trimmed) return;

		// Remove if exists, add to front
		searchHistory = [trimmed, ...searchHistory.filter((h) => h !== trimmed)].slice(0, MAX_HISTORY);
		storage.set(HISTORY_KEY, searchHistory);
	}

	function removeFromHistory(q: string) {
		searchHistory = searchHistory.filter((h) => h !== q);
		storage.set(HISTORY_KEY, searchHistory);
	}

	function clearHistory() {
		searchHistory = [];
		storage.remove(HISTORY_KEY);
	}

	function handleInput(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		query = value;
		selectedIndex = -1;
		if (value.trim()) {
			debouncedSearch(value);
		} else {
			results = null;
		}
	}

	function handleHistoryClick(q: string) {
		query = q;
		performSearch(q);
	}

	function handleKeydown(e: KeyboardEvent) {
		// Global search focus
		if (e.key === '/' && !e.ctrlKey && !e.metaKey) {
			if (document.activeElement !== inputElement) {
				e.preventDefault();
				inputElement?.focus();
			}
		}

		// Escape to clear
		if (e.key === 'Escape') {
			if (query) {
				query = '';
				results = null;
			} else {
				inputElement?.blur();
			}
		}
	}

	function clearSearch() {
		query = '';
		results = null;
		inputElement?.focus();
	}

	const tabs: { id: Tab; label: string }[] = [
		{ id: 'all', label: 'All' },
		{ id: 'tracks', label: 'Tracks' },
		{ id: 'albums', label: 'Albums' },
		{ id: 'artists', label: 'Artists' }
	];

	const hasResults = $derived(
		results && (results.tracks.length > 0 || results.albums.length > 0 || results.artists.length > 0)
	);
</script>

<svelte:head>
	<title>{query ? `${query} - Search` : 'Search'} - Harmony</title>
</svelte:head>

<svelte:window on:keydown={handleKeydown} />

<div class="p-6">
	<!-- Search Input -->
	<div class="max-w-2xl mx-auto mb-8">
		<div class="relative">
			<Search class="absolute left-4 top-1/2 -translate-y-1/2 text-text-muted" size={20} />
			<input
				bind:this={inputElement}
				type="text"
				value={query}
				oninput={handleInput}
				placeholder="What do you want to listen to?"
				class="w-full pl-12 pr-12 py-4 bg-surface-elevated rounded-full text-lg placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-accent"
			/>
			{#if query}
				<Button
					variant="ghost"
					size="sm"
					class="absolute right-2 top-1/2 -translate-y-1/2"
					onclick={clearSearch}
				>
					<X size={20} />
				</Button>
			{/if}
		</div>
	</div>

	{#if loading}
		<!-- Loading State -->
		<div class="space-y-8">
			<section>
				<Skeleton class="h-7 w-32 mb-4" />
				<div class="flex gap-4">
					{#each Array(5) as _}
						<div class="w-40 flex-shrink-0">
							<Skeleton class="aspect-square rounded-md mb-3" />
							<Skeleton class="h-4 w-32 mb-1" />
							<Skeleton class="h-3 w-24" />
						</div>
					{/each}
				</div>
			</section>
		</div>
	{:else if results && hasResults}
		<!-- Tabs -->
		<div class="flex gap-2 mb-6 border-b border-surface-border">
			{#each tabs as tab}
				<button
					class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px {activeTab === tab.id
						? 'border-accent text-accent'
						: 'border-transparent text-text-secondary hover:text-text-primary'}"
					onclick={() => (activeTab = tab.id)}
				>
					{tab.label}
					{#if tab.id === 'tracks' && results.tracks.length > 0}
						<span class="ml-1 text-text-muted">({results.tracks.length})</span>
					{:else if tab.id === 'albums' && results.albums.length > 0}
						<span class="ml-1 text-text-muted">({results.albums.length})</span>
					{:else if tab.id === 'artists' && results.artists.length > 0}
						<span class="ml-1 text-text-muted">({results.artists.length})</span>
					{/if}
				</button>
			{/each}
		</div>

		<!-- Results -->
		<div class="space-y-8">
			<!-- Artists -->
			{#if (activeTab === 'all' || activeTab === 'artists') && results.artists.length > 0}
				<section>
					<SectionHeader title="Artists" />
					<div class="flex gap-4 overflow-x-auto pb-4 scrollbar-thin">
						{#each results.artists as artist (artist.id)}
							<ArtistCard {artist} />
						{/each}
					</div>
				</section>
			{/if}

			<!-- Albums -->
			{#if (activeTab === 'all' || activeTab === 'albums') && results.albums.length > 0}
				<section>
					<SectionHeader title="Albums" />
					<div class="flex gap-4 overflow-x-auto pb-4 scrollbar-thin">
						{#each results.albums as album (album.id)}
							<AlbumCard {album} />
						{/each}
					</div>
				</section>
			{/if}

			<!-- Tracks -->
			{#if (activeTab === 'all' || activeTab === 'tracks') && results.tracks.length > 0}
				<section>
					<SectionHeader title="Tracks" />
					<div class="bg-surface-elevated rounded-lg overflow-hidden">
						{#each results.tracks as track, i (track.id)}
							<TrackRow {track} index={i} showAlbum showArtist />
						{/each}
					</div>
				</section>
			{/if}
		</div>
	{:else if query && !loading}
		<!-- No Results -->
		<div class="text-center py-20">
			<p class="text-text-secondary mb-2">No results found for "{query}"</p>
			<p class="text-sm text-text-muted">Try different keywords or check your spelling</p>
		</div>
	{:else}
		<!-- Search History / Browse -->
		<div class="max-w-2xl mx-auto">
			{#if searchHistory.length > 0}
				<div class="mb-8">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold flex items-center gap-2">
							<Clock size={20} />
							Recent Searches
						</h2>
						<Button variant="ghost" size="sm" onclick={clearHistory}>
							Clear all
						</Button>
					</div>
					<div class="flex flex-wrap gap-2">
						{#each searchHistory as historyItem}
							<button
								class="group flex items-center gap-2 px-4 py-2 bg-surface-elevated rounded-full hover:bg-surface-hover transition-colors"
								onclick={() => handleHistoryClick(historyItem)}
							>
								<span>{historyItem}</span>
								<button
									class="opacity-0 group-hover:opacity-100 hover:text-accent transition-opacity"
									onclick|stopPropagation={() => removeFromHistory(historyItem)}
								>
									<X size={14} />
								</button>
							</button>
						{/each}
					</div>
				</div>
			{/if}

			<div>
				<h2 class="text-lg font-semibold flex items-center gap-2 mb-4">
					<TrendingUp size={20} />
					Browse
				</h2>
				<div class="grid grid-cols-2 sm:grid-cols-3 gap-4">
					<a
						href="/library?tab=albums"
						class="p-6 bg-gradient-to-br from-purple-600 to-purple-800 rounded-lg hover:scale-105 transition-transform"
					>
						<span class="font-bold text-lg">Albums</span>
					</a>
					<a
						href="/library?tab=artists"
						class="p-6 bg-gradient-to-br from-blue-600 to-blue-800 rounded-lg hover:scale-105 transition-transform"
					>
						<span class="font-bold text-lg">Artists</span>
					</a>
					<a
						href="/library?tab=playlists"
						class="p-6 bg-gradient-to-br from-green-600 to-green-800 rounded-lg hover:scale-105 transition-transform"
					>
						<span class="font-bold text-lg">Playlists</span>
					</a>
				</div>
			</div>
		</div>
	{/if}
</div>
