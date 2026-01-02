<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import '../app.css';
	import logo from '$lib/assets/logo.svg';
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
	import { checkSetupStatus, setupRequired, setupLoading } from '$lib/stores/setup';

	let { children } = $props();

	let sidebarCollapsed = $state(false);
	let checkingSetup = $state(true);

	// Check if current path is a setup route
	const isSetupRoute = $derived($page.url.pathname.startsWith('/setup'));

	onMount(async () => {
		initializeTheme();

		// Check setup status
		const status = await checkSetupStatus();

		// If setup not completed and not already on setup page, redirect
		if (status && !status.completed && !isSetupRoute) {
			goto('/setup');
		}

		checkingSetup = false;
	});

	// Also watch for setup requirement changes
	$effect(() => {
		if (!$setupLoading && $setupRequired && !isSetupRoute) {
			goto('/setup');
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={logo} />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap"
		rel="stylesheet"
	/>
	<title>Harmony</title>
</svelte:head>

{#if isSetupRoute}
	<!-- Setup routes have their own layout -->
	{@render children()}
{:else if checkingSetup}
	<!-- Loading State (only for non-setup routes) -->
	<div class="min-h-screen bg-surface flex items-center justify-center">
		<div class="text-center space-y-4">
			<div class="w-12 h-12 border-4 border-accent border-t-transparent rounded-full animate-spin mx-auto"></div>
			<p class="text-text-muted">Loading...</p>
		</div>
	</div>
{:else}
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
{/if}
