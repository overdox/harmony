<script lang="ts">
	import { clsx } from 'clsx';

	type Variant = 'primary' | 'secondary' | 'ghost' | 'icon';
	type Size = 'sm' | 'md' | 'lg';

	interface Props {
		variant?: Variant;
		size?: Size;
		disabled?: boolean;
		loading?: boolean;
		class?: string;
		href?: string;
		type?: 'button' | 'submit' | 'reset';
		onclick?: (e: MouseEvent) => void;
	}

	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		loading = false,
		class: className = '',
		href,
		type = 'button',
		onclick,
		children
	}: Props & { children?: any } = $props();

	const baseClasses = 'inline-flex items-center justify-center font-medium transition-all duration-150 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-surface disabled:opacity-50 disabled:cursor-not-allowed';

	const variants: Record<Variant, string> = {
		primary: 'bg-accent text-black hover:bg-accent-hover hover:scale-105 active:scale-100 rounded-full',
		secondary: 'bg-surface-elevated text-white hover:bg-surface-hover border border-surface-border rounded-full',
		ghost: 'bg-transparent text-text-secondary hover:text-white hover:bg-white/10 rounded-md',
		icon: 'bg-transparent text-text-secondary hover:text-white hover:bg-white/10 rounded-full'
	};

	const sizes: Record<Size, string> = {
		sm: 'px-3 py-1.5 text-sm gap-1.5',
		md: 'px-4 py-2 text-sm gap-2',
		lg: 'px-6 py-3 text-base gap-2'
	};

	const iconSizes: Record<Size, string> = {
		sm: 'w-8 h-8',
		md: 'w-10 h-10',
		lg: 'w-12 h-12'
	};

	$effect(() => {
		// Effect for loading state if needed
	});
</script>

{#if href && !disabled}
	<a
		{href}
		class={clsx(
			baseClasses,
			variants[variant],
			variant === 'icon' ? iconSizes[size] : sizes[size],
			className
		)}
	>
		{#if loading}
			<svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
			</svg>
		{/if}
		{@render children?.()}
	</a>
{:else}
	<button
		{type}
		{disabled}
		class={clsx(
			baseClasses,
			variants[variant],
			variant === 'icon' ? iconSizes[size] : sizes[size],
			className
		)}
		onclick={onclick}
	>
		{#if loading}
			<svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
			</svg>
		{/if}
		{@render children?.()}
	</button>
{/if}
