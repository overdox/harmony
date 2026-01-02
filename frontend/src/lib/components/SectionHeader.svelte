<script lang="ts">
	import { ChevronRight } from 'lucide-svelte';

	interface Props {
		title: string;
		href?: string;
		showMore?: boolean;
	}

	let { title, href, showMore = false }: Props = $props();
</script>

<div class="flex items-center justify-between mb-4">
	<h2 class="text-xl font-bold">{title}</h2>
	{#if showMore && href}
		<a
			{href}
			class="text-sm text-text-secondary hover:text-text-primary flex items-center gap-1 transition-colors"
		>
			Show all
			<ChevronRight size={16} />
		</a>
	{/if}
</div>
