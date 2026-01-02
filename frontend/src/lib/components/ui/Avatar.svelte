<script lang="ts">
	import { clsx } from 'clsx';
	import { User } from 'lucide-svelte';

	type Size = 'xs' | 'sm' | 'md' | 'lg' | 'xl';
	type Shape = 'circle' | 'square';

	interface Props {
		src?: string | null;
		alt?: string;
		size?: Size;
		shape?: Shape;
		class?: string;
		fallback?: string;
	}

	let {
		src = null,
		alt = '',
		size = 'md',
		shape = 'circle',
		class: className = '',
		fallback
	}: Props = $props();

	let hasError = $state(false);

	const sizeClasses: Record<Size, string> = {
		xs: 'w-6 h-6 text-xs',
		sm: 'w-8 h-8 text-sm',
		md: 'w-10 h-10 text-base',
		lg: 'w-16 h-16 text-xl',
		xl: 'w-24 h-24 text-3xl'
	};

	const iconSizes: Record<Size, number> = {
		xs: 12,
		sm: 16,
		md: 20,
		lg: 32,
		xl: 48
	};

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map(part => part[0])
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}
</script>

<div
	class={clsx(
		'relative flex items-center justify-center bg-surface-hover overflow-hidden',
		sizeClasses[size],
		shape === 'circle' ? 'rounded-full' : 'rounded-md',
		className
	)}
>
	{#if src && !hasError}
		<img
			{src}
			{alt}
			class="w-full h-full object-cover"
			onerror={() => (hasError = true)}
		/>
	{:else if fallback}
		<span class="font-semibold text-text-secondary">
			{getInitials(fallback)}
		</span>
	{:else}
		<User size={iconSizes[size]} class="text-text-muted" />
	{/if}
</div>
