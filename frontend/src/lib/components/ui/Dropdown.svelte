<script lang="ts">
	import { clsx } from 'clsx';

	interface DropdownItem {
		label: string;
		value: string;
		icon?: any;
		disabled?: boolean;
		danger?: boolean;
	}

	interface Props {
		items: DropdownItem[];
		open?: boolean;
		align?: 'left' | 'right';
		class?: string;
		onselect?: (item: DropdownItem) => void;
		onclose?: () => void;
	}

	let {
		items,
		open = $bindable(false),
		align = 'left',
		class: className = '',
		onselect,
		onclose,
		children
	}: Props & { children?: any } = $props();

	let menuRef: HTMLDivElement;

	function handleSelect(item: DropdownItem) {
		if (item.disabled) return;
		onselect?.(item);
		open = false;
		onclose?.();
	}

	function handleClickOutside(e: MouseEvent) {
		if (menuRef && !menuRef.contains(e.target as Node)) {
			open = false;
			onclose?.();
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
			onclose?.();
		}
	}
</script>

<svelte:window onclick={handleClickOutside} onkeydown={handleKeydown} />

<div class="relative inline-block" bind:this={menuRef}>
	<!-- Trigger -->
	<div onclick={() => (open = !open)}>
		{@render children?.()}
	</div>

	<!-- Menu -->
	{#if open}
		<div
			class={clsx(
				'absolute z-50 mt-2 py-1 min-w-[180px] bg-surface-elevated rounded-md shadow-lg border border-surface-border animate-fade-in',
				align === 'right' ? 'right-0' : 'left-0',
				className
			)}
			role="menu"
		>
			{#each items as item}
				<button
					type="button"
					class={clsx(
						'w-full flex items-center gap-3 px-3 py-2 text-sm text-left transition-colors',
						item.disabled
							? 'opacity-50 cursor-not-allowed'
							: item.danger
								? 'text-error hover:bg-error/10'
								: 'text-text-secondary hover:text-white hover:bg-surface-hover',
					)}
					disabled={item.disabled}
					onclick={() => handleSelect(item)}
					role="menuitem"
				>
					{#if item.icon}
						<svelte:component this={item.icon} size={16} />
					{/if}
					{item.label}
				</button>
			{/each}
		</div>
	{/if}
</div>
