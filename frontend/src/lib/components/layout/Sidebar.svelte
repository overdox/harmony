<script lang="ts">
	import { clsx } from 'clsx';
	import { page } from '$app/stores';
	import { Home, Search, Library, Plus, Music, ChevronLeft, ChevronRight, Settings } from 'lucide-svelte';
	import Button from '../ui/Button.svelte';

	interface Props {
		collapsed?: boolean;
		oncollapse?: (collapsed: boolean) => void;
	}

	let {
		collapsed = $bindable(false),
		oncollapse
	}: Props = $props();

	// Mock playlists - will be replaced with real data
	const playlists = [
		{ id: '1', name: 'Liked Songs', trackCount: 42 },
		{ id: '2', name: 'Discover Weekly', trackCount: 30 },
		{ id: '3', name: 'Chill Vibes', trackCount: 25 }
	];

	const navItems = [
		{ href: '/', icon: Home, label: 'Home' },
		{ href: '/search', icon: Search, label: 'Search' },
		{ href: '/library', icon: Library, label: 'Your Library' },
		{ href: '/settings', icon: Settings, label: 'Settings' }
	];

	function toggleCollapse() {
		collapsed = !collapsed;
		oncollapse?.(collapsed);
	}

	function isActive(href: string, currentPath: string): boolean {
		if (href === '/') return currentPath === '/';
		return currentPath.startsWith(href);
	}
</script>

<aside
	class={clsx(
		'flex flex-col h-full bg-black transition-all duration-300',
		collapsed ? 'w-[var(--spacing-sidebar-collapsed)]' : 'w-[var(--spacing-sidebar)]'
	)}
>
	<!-- Logo -->
	<div class="flex items-center gap-3 px-4 py-4">
		<div class="flex items-center justify-center w-10 h-10 bg-accent rounded-lg">
			<Music class="w-6 h-6 text-black" />
		</div>
		{#if !collapsed}
			<span class="text-xl font-bold tracking-tight">Harmony</span>
		{/if}
	</div>

	<!-- Main Navigation -->
	<nav class="px-2 py-2">
		<ul class="space-y-1">
			{#each navItems as item}
				<li>
					<a
						href={item.href}
						class={clsx(
							'flex items-center gap-4 px-3 py-3 rounded-md transition-colors',
							isActive(item.href, $page.url.pathname)
								? 'bg-surface-elevated text-white'
								: 'text-text-secondary hover:text-white'
						)}
					>
						<svelte:component this={item.icon} size={24} />
						{#if !collapsed}
							<span class="font-medium">{item.label}</span>
						{/if}
					</a>
				</li>
			{/each}
		</ul>
	</nav>

	<!-- Playlists Section -->
	{#if !collapsed}
		<div class="flex-1 flex flex-col mt-4 mx-2 bg-surface-elevated rounded-lg overflow-hidden">
			<div class="flex items-center justify-between px-4 py-3">
				<span class="text-sm font-semibold text-text-secondary">Playlists</span>
				<Button variant="icon" size="sm">
					<Plus size={18} />
				</Button>
			</div>

			<div class="flex-1 overflow-y-auto px-2 pb-2">
				<ul class="space-y-0.5">
					{#each playlists as playlist}
						<li>
							<a
								href="/playlist/{playlist.id}"
								class={clsx(
									'flex items-center gap-3 px-2 py-2 rounded-md transition-colors',
									$page.url.pathname === `/playlist/${playlist.id}`
										? 'bg-surface-hover text-white'
										: 'text-text-secondary hover:text-white hover:bg-surface-hover/50'
								)}
							>
								<div class="w-10 h-10 bg-surface-hover rounded flex items-center justify-center flex-shrink-0">
									<Music size={16} class="text-text-muted" />
								</div>
								<div class="min-w-0 flex-1">
									<p class="text-sm font-medium truncate">{playlist.name}</p>
									<p class="text-xs text-text-muted">{playlist.trackCount} tracks</p>
								</div>
							</a>
						</li>
					{/each}
				</ul>
			</div>
		</div>
	{/if}

	<!-- Collapse Toggle -->
	<div class="px-2 py-2">
		<Button
			variant="ghost"
			class="w-full justify-center"
			onclick={toggleCollapse}
		>
			{#if collapsed}
				<ChevronRight size={20} />
			{:else}
				<ChevronLeft size={20} />
				<span>Collapse</span>
			{/if}
		</Button>
	</div>
</aside>
