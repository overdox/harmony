<script lang="ts">
	import { clsx } from 'clsx';

	interface Props {
		value?: number;
		min?: number;
		max?: number;
		step?: number;
		disabled?: boolean;
		class?: string;
		showTooltip?: boolean;
		formatTooltip?: (value: number) => string;
		onchange?: (value: number) => void;
		oninput?: (value: number) => void;
	}

	let {
		value = $bindable(0),
		min = 0,
		max = 100,
		step = 1,
		disabled = false,
		class: className = '',
		showTooltip = false,
		formatTooltip = (v: number) => String(v),
		onchange,
		oninput
	}: Props = $props();

	let isDragging = $state(false);
	let showTooltipState = $state(false);
	let sliderRef: HTMLDivElement;

	const percentage = $derived(((value - min) / (max - min)) * 100);

	function handlePointerDown(e: PointerEvent) {
		if (disabled) return;
		isDragging = true;
		showTooltipState = true;
		updateValue(e);
		(e.target as HTMLElement).setPointerCapture(e.pointerId);
	}

	function handlePointerMove(e: PointerEvent) {
		if (!isDragging) return;
		updateValue(e);
	}

	function handlePointerUp() {
		isDragging = false;
		showTooltipState = false;
		if (onchange) onchange(value);
	}

	function updateValue(e: PointerEvent) {
		if (!sliderRef) return;
		const rect = sliderRef.getBoundingClientRect();
		const percent = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
		const newValue = min + percent * (max - min);
		value = Math.round(newValue / step) * step;
		if (oninput) oninput(value);
	}
</script>

<div
	bind:this={sliderRef}
	class={clsx(
		'relative h-1 bg-white/20 rounded-full cursor-pointer group',
		disabled && 'opacity-50 cursor-not-allowed',
		className
	)}
	role="slider"
	aria-valuemin={min}
	aria-valuemax={max}
	aria-valuenow={value}
	tabindex={disabled ? -1 : 0}
	onpointerdown={handlePointerDown}
	onpointermove={handlePointerMove}
	onpointerup={handlePointerUp}
	onpointerleave={handlePointerUp}
	onmouseenter={() => showTooltip && (showTooltipState = true)}
	onmouseleave={() => !isDragging && (showTooltipState = false)}
>
	<!-- Progress track -->
	<div
		class="absolute inset-y-0 left-0 bg-white rounded-full group-hover:bg-accent transition-colors"
		style="width: {percentage}%"
	></div>

	<!-- Thumb -->
	<div
		class={clsx(
			'absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-3 h-3 bg-white rounded-full shadow-md',
			'opacity-0 group-hover:opacity-100 transition-opacity',
			isDragging && 'opacity-100'
		)}
		style="left: {percentage}%"
	></div>

	<!-- Tooltip -->
	{#if showTooltip && showTooltipState}
		<div
			class="absolute bottom-full mb-2 -translate-x-1/2 px-2 py-1 bg-surface-elevated rounded text-xs text-white whitespace-nowrap"
			style="left: {percentage}%"
		>
			{formatTooltip(value)}
		</div>
	{/if}
</div>
