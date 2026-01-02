<script lang="ts">
	import { page } from '$app/stores';
	import { Home, Search, Library, ListMusic } from 'lucide-svelte';
	import { toggleQueuePanel } from '$lib/stores/ui';

	const navItems = [
		{ href: '/', icon: Home, label: 'Home' },
		{ href: '/search', icon: Search, label: 'Search' },
		{ href: '/library', icon: Library, label: 'Library' }
	];

	function isActive(href: string): boolean {
		if (href === '/') {
			return $page.url.pathname === '/';
		}
		return $page.url.pathname.startsWith(href);
	}
</script>

<nav class="fixed bottom-[var(--spacing-nowplaying)] left-0 right-0 bg-surface-elevated border-t border-surface-border md:hidden z-30">
	<div class="flex items-center justify-around h-14">
		{#each navItems as item}
			<a
				href={item.href}
				class="flex flex-col items-center justify-center gap-0.5 px-4 py-2 transition-colors {isActive(item.href)
					? 'text-accent'
					: 'text-text-secondary'}"
			>
				<svelte:component this={item.icon} size={22} />
				<span class="text-[10px] font-medium">{item.label}</span>
			</a>
		{/each}
		<button
			class="flex flex-col items-center justify-center gap-0.5 px-4 py-2 text-text-secondary transition-colors"
			onclick={toggleQueuePanel}
		>
			<ListMusic size={22} />
			<span class="text-[10px] font-medium">Queue</span>
		</button>
	</div>
</nav>
