<script lang="ts">
	import { clsx } from 'clsx';

	interface Props {
		href?: string;
		class?: string;
		padding?: 'none' | 'sm' | 'md' | 'lg';
		hoverable?: boolean;
		onclick?: (e: MouseEvent) => void;
	}

	let {
		href,
		class: className = '',
		padding = 'md',
		hoverable = true,
		onclick,
		children
	}: Props & { children?: any } = $props();

	const paddingClasses = {
		none: 'p-0',
		sm: 'p-2',
		md: 'p-4',
		lg: 'p-6'
	};
</script>

{#if href}
	<a
		{href}
		class={clsx(
			'block bg-surface-elevated rounded-lg transition-all duration-200',
			paddingClasses[padding],
			hoverable && 'hover:bg-surface-hover cursor-pointer',
			className
		)}
	>
		{@render children?.()}
	</a>
{:else if onclick}
	<button
		type="button"
		class={clsx(
			'block w-full text-left bg-surface-elevated rounded-lg transition-all duration-200',
			paddingClasses[padding],
			hoverable && 'hover:bg-surface-hover cursor-pointer',
			className
		)}
		onclick={onclick}
	>
		{@render children?.()}
	</button>
{:else}
	<div
		class={clsx(
			'bg-surface-elevated rounded-lg',
			paddingClasses[padding],
			hoverable && 'hover:bg-surface-hover transition-colors duration-200',
			className
		)}
	>
		{@render children?.()}
	</div>
{/if}
