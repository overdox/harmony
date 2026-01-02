<script lang="ts">
	import { onMount } from 'svelte';
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import NowPlaying from '$lib/components/layout/NowPlaying.svelte';
	import MobileNav from '$lib/components/layout/MobileNav.svelte';
	import MobilePlayer from '$lib/components/layout/MobilePlayer.svelte';
	import QueuePanel from '$lib/components/QueuePanel.svelte';
	import { Toast } from '$lib/components/ui';
	import {
		showQueuePanel,
		closeQueuePanel,
		showMobilePlayer,
		closeMobilePlayer
	} from '$lib/stores/ui';
	import { initializeTheme } from '$lib/stores/theme';

	let { children } = $props();

	let sidebarCollapsed = $state(false);

	onMount(() => {
		initializeTheme();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap"
		rel="stylesheet"
	/>
	<title>Harmony</title>
</svelte:head>

<div class="flex h-screen overflow-hidden">
	<!-- Sidebar (hidden on mobile) -->
	<div class="hidden md:block">
		<Sidebar bind:collapsed={sidebarCollapsed} />
	</div>

	<!-- Main Content Area -->
	<main
		class="flex-1 overflow-y-auto pb-[calc(var(--spacing-nowplaying)+3.5rem)] md:pb-[var(--spacing-nowplaying)]"
		style="background: linear-gradient(180deg, #1a1a2e 0%, #121212 300px)"
	>
		{@render children()}
	</main>
</div>

<!-- Mobile Navigation -->
<MobileNav />

<!-- Now Playing Bar -->
<NowPlaying />

<!-- Mobile Full-Screen Player -->
<MobilePlayer open={$showMobilePlayer} onclose={closeMobilePlayer} />

<!-- Queue Panel -->
<QueuePanel open={$showQueuePanel} onclose={closeQueuePanel} />

<!-- Toast Notifications -->
<Toast />
