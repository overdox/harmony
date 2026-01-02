<script lang="ts">
	import { clsx } from 'clsx';
	import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-svelte';
	import { toasts, dismissToast, type Toast } from '$lib/stores/toast';

	const icons = {
		success: CheckCircle,
		error: AlertCircle,
		warning: AlertTriangle,
		info: Info
	};

	const typeClasses = {
		success: 'border-success/30 bg-success/10',
		error: 'border-error/30 bg-error/10',
		warning: 'border-warning/30 bg-warning/10',
		info: 'border-surface-border bg-surface-elevated'
	};

	const iconClasses = {
		success: 'text-success',
		error: 'text-error',
		warning: 'text-warning',
		info: 'text-text-secondary'
	};
</script>

<div class="fixed bottom-24 right-4 z-50 flex flex-col gap-2 max-w-sm">
	{#each $toasts as toast (toast.id)}
		<div
			class={clsx(
				'flex items-start gap-3 px-4 py-3 rounded-lg border shadow-lg animate-slide-up',
				typeClasses[toast.type]
			)}
			role="alert"
		>
			<svelte:component
				this={icons[toast.type]}
				size={20}
				class={iconClasses[toast.type]}
			/>

			<div class="flex-1 min-w-0">
				{#if toast.title}
					<p class="font-medium text-white">{toast.title}</p>
				{/if}
				<p class="text-sm text-text-secondary">{toast.message}</p>
			</div>

			<button
				type="button"
				class="text-text-muted hover:text-white transition-colors"
				onclick={() => dismissToast(toast.id)}
			>
				<X size={16} />
			</button>
		</div>
	{/each}
</div>
