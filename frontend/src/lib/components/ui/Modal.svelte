<script lang="ts">
	import { clsx } from 'clsx';
	import { X } from 'lucide-svelte';
	import Button from './Button.svelte';

	interface Props {
		open?: boolean;
		title?: string;
		size?: 'sm' | 'md' | 'lg';
		closeOnOverlay?: boolean;
		showClose?: boolean;
		class?: string;
		onclose?: () => void;
	}

	let {
		open = $bindable(false),
		title = '',
		size = 'md',
		closeOnOverlay = true,
		showClose = true,
		class: className = '',
		onclose,
		children
	}: Props & { children?: any } = $props();

	const sizeClasses = {
		sm: 'max-w-sm',
		md: 'max-w-md',
		lg: 'max-w-lg'
	};

	function handleClose() {
		open = false;
		onclose?.();
	}

	function handleOverlayClick(e: MouseEvent) {
		if (closeOnOverlay && e.target === e.currentTarget) {
			handleClose();
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			handleClose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<!-- Overlay -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-fade-in"
		onclick={handleOverlayClick}
		role="dialog"
		aria-modal="true"
		aria-labelledby={title ? 'modal-title' : undefined}
	>
		<!-- Modal content -->
		<div
			class={clsx(
				'w-full bg-surface-elevated rounded-lg shadow-lg animate-slide-up',
				sizeClasses[size],
				className
			)}
		>
			<!-- Header -->
			{#if title || showClose}
				<div class="flex items-center justify-between px-6 py-4 border-b border-surface-border">
					{#if title}
						<h2 id="modal-title" class="text-lg font-semibold">
							{title}
						</h2>
					{:else}
						<div></div>
					{/if}

					{#if showClose}
						<Button variant="icon" size="sm" onclick={handleClose}>
							<X size={18} />
						</Button>
					{/if}
				</div>
			{/if}

			<!-- Body -->
			<div class="px-6 py-4">
				{@render children?.()}
			</div>
		</div>
	</div>
{/if}
