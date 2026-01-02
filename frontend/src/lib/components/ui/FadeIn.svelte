<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	interface Props {
		delay?: number;
		duration?: number;
		y?: number;
		once?: boolean;
		class?: string;
		children: import('svelte').Snippet;
	}

	let {
		delay = 0,
		duration = 300,
		y = 10,
		once = true,
		class: className = '',
		children
	}: Props = $props();

	let visible = $state(false);

	onMount(() => {
		// Small delay to ensure DOM is ready
		requestAnimationFrame(() => {
			visible = true;
		});
	});
</script>

{#if visible}
	<div
		class={className}
		in:fly={{ y, duration, delay, easing: cubicOut }}
	>
		{@render children()}
	</div>
{:else}
	<div class="{className} opacity-0">
		{@render children()}
	</div>
{/if}
