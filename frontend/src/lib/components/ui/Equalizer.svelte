<script lang="ts">
	import { clsx } from 'clsx';

	interface Props {
		playing?: boolean;
		size?: 'sm' | 'md' | 'lg';
		class?: string;
	}

	let { playing = true, size = 'md', class: className }: Props = $props();

	const sizeMap = {
		sm: { height: 12, barWidth: 2, gap: 1 },
		md: { height: 16, barWidth: 3, gap: 1 },
		lg: { height: 24, barWidth: 4, gap: 2 }
	};

	const config = $derived(sizeMap[size]);
</script>

<div
	class={clsx('flex items-end', className)}
	style="height: {config.height}px; gap: {config.gap}px;"
	aria-label={playing ? 'Playing' : 'Paused'}
>
	{#each [0, 1, 2, 3] as i}
		<div
			class={clsx(
				'bg-accent rounded-full transition-all',
				playing ? 'animate-equalizer' : 'h-1'
			)}
			style="
				width: {config.barWidth}px;
				animation-delay: {i * 0.1}s;
				{!playing ? `height: ${config.barWidth}px;` : ''}
			"
		></div>
	{/each}
</div>

<style>
	@keyframes equalizer {
		0%, 100% {
			height: 20%;
		}
		25% {
			height: 80%;
		}
		50% {
			height: 40%;
		}
		75% {
			height: 100%;
		}
	}

	.animate-equalizer {
		animation: equalizer 0.8s ease-in-out infinite;
	}
</style>
