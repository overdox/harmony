<script lang="ts">
	import { clsx } from 'clsx';

	type Variant = 'text' | 'circular' | 'rectangular';

	interface Props {
		variant?: Variant;
		width?: string;
		height?: string;
		class?: string;
		animate?: boolean;
	}

	let {
		variant = 'text',
		width,
		height,
		class: className = '',
		animate = true
	}: Props = $props();

	const variantClasses: Record<Variant, string> = {
		text: 'rounded h-4',
		circular: 'rounded-full',
		rectangular: 'rounded-md'
	};
</script>

<div
	class={clsx(
		'bg-surface-hover',
		variantClasses[variant],
		animate && 'animate-pulse',
		className
	)}
	style:width={width}
	style:height={height}
></div>
