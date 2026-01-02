<script lang="ts">
	import { clsx } from 'clsx';
	import { Search } from 'lucide-svelte';

	type Variant = 'default' | 'search';

	interface Props {
		type?: string;
		variant?: Variant;
		placeholder?: string;
		value?: string;
		disabled?: boolean;
		class?: string;
		id?: string;
		name?: string;
		oninput?: (e: Event) => void;
		onkeydown?: (e: KeyboardEvent) => void;
		onfocus?: (e: FocusEvent) => void;
		onblur?: (e: FocusEvent) => void;
	}

	let {
		type = 'text',
		variant = 'default',
		placeholder = '',
		value = $bindable(''),
		disabled = false,
		class: className = '',
		id,
		name,
		oninput,
		onkeydown,
		onfocus,
		onblur
	}: Props = $props();
</script>

<div class={clsx('relative', className)}>
	{#if variant === 'search'}
		<Search
			class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted pointer-events-none"
		/>
	{/if}

	<input
		{type}
		{id}
		{name}
		{placeholder}
		{disabled}
		bind:value
		{oninput}
		{onkeydown}
		{onfocus}
		{onblur}
		class={clsx(
			'w-full bg-surface-hover text-white placeholder-text-muted rounded-full',
			'border-none outline-none',
			'focus:ring-2 focus:ring-white/20',
			'transition-all duration-150',
			'disabled:opacity-50 disabled:cursor-not-allowed',
			variant === 'search' ? 'pl-10 pr-4 py-2 text-sm' : 'px-4 py-2',
		)}
	/>
</div>
