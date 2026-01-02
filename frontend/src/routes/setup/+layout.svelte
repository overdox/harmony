<script lang="ts">
	import '../../app.css';
	import logo from '$lib/assets/logo.svg';
	import { onMount } from 'svelte';
	import { initializeTheme } from '$lib/stores/theme';

	let { children } = $props();

	onMount(() => {
		initializeTheme();
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
	<title>Setup - Harmony</title>
</svelte:head>

<div class="min-h-screen bg-surface flex items-center justify-center p-4">
	<div class="w-full max-w-2xl">
		{@render children()}
	</div>
</div>
