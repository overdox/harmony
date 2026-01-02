<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	interface Props {
		key?: string | number;
		type?: 'fade' | 'slide' | 'scale';
		duration?: number;
		children: import('svelte').Snippet;
	}

	let { key, type = 'fade', duration = 200, children }: Props = $props();
</script>

{#key key}
	<div
		class="contents"
		in:fade={{ duration, easing: cubicOut }}
		out:fade={{ duration: duration / 2, easing: cubicOut }}
	>
		{@render children()}
	</div>
{/key}
