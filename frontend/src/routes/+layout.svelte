<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import NowPlaying from '$lib/components/layout/NowPlaying.svelte';
	import QueuePanel from '$lib/components/QueuePanel.svelte';
	import { Toast } from '$lib/components/ui';
	import { showQueuePanel, closeQueuePanel } from '$lib/stores/ui';

	let { children } = $props();

	let sidebarCollapsed = $state(false);
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
	<!-- Sidebar -->
	<Sidebar bind:collapsed={sidebarCollapsed} />

	<!-- Main Content Area -->
	<main
		class="flex-1 overflow-y-auto pb-[var(--spacing-nowplaying)]"
		style="background: linear-gradient(180deg, #1a1a2e 0%, #121212 300px)"
	>
		{@render children()}
	</main>
</div>

<!-- Now Playing Bar -->
<NowPlaying />

<!-- Queue Panel -->
<QueuePanel open={$showQueuePanel} onclose={closeQueuePanel} />

<!-- Toast Notifications -->
<Toast />
