<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { CheckCircle, Music, Loader2, ArrowLeft, Play } from 'lucide-svelte';
	import { Button } from '$lib/components/ui';
	import { completeSetup, getSelectedFolders } from '$lib/api/setup';
	import { markSetupCompleted } from '$lib/stores/setup';

	let selectedFolders = $state<string[]>([]);
	let startScan = $state(true);
	let completing = $state(false);
	let error = $state<string | null>(null);
	let completed = $state(false);

	onMount(async () => {
		try {
			selectedFolders = await getSelectedFolders();
		} catch (e) {
			console.error('Failed to load selected folders:', e);
		}
	});

	async function handleComplete() {
		completing = true;
		error = null;

		try {
			await completeSetup(startScan);
			markSetupCompleted();
			completed = true;

			// Wait a moment then redirect
			setTimeout(() => {
				goto('/');
			}, 2000);
		} catch (e) {
			console.error('Failed to complete setup:', e);
			error = 'Failed to complete setup';
			completing = false;
		}
	}
</script>

<div class="space-y-6">
	{#if completed}
		<!-- Success State -->
		<div class="text-center space-y-6 py-8">
			<div class="flex justify-center">
				<div class="w-20 h-20 rounded-full bg-accent/20 flex items-center justify-center animate-bounce-in">
					<CheckCircle size={48} class="text-accent" />
				</div>
			</div>
			<div class="space-y-2">
				<h1 class="text-2xl font-bold text-text-primary">You're All Set!</h1>
				<p class="text-text-secondary">
					{#if startScan}
						Your music library is now being scanned. Redirecting...
					{:else}
						Harmony is ready to use. Redirecting...
					{/if}
				</p>
			</div>
			<div class="flex justify-center">
				<Loader2 size={24} class="animate-spin text-accent" />
			</div>
		</div>
	{:else}
		<!-- Configuration Review -->
		<div class="text-center space-y-2">
			<h1 class="text-2xl font-bold text-text-primary">Ready to Go!</h1>
			<p class="text-text-secondary">
				Review your settings and complete the setup.
			</p>
		</div>

		<!-- Selected Folders -->
		<div class="bg-surface-elevated rounded-lg p-4 space-y-3">
			<h3 class="font-medium text-text-primary flex items-center gap-2">
				<Music size={18} class="text-accent" />
				Music Folders
			</h3>
			<div class="space-y-2">
				{#each selectedFolders as folder}
					<div class="flex items-center gap-2 text-sm bg-surface-hover rounded px-3 py-2">
						<CheckCircle size={14} class="text-accent flex-shrink-0" />
						<span class="text-text-secondary truncate">{folder}</span>
					</div>
				{:else}
					<p class="text-text-muted text-sm">No folders selected</p>
				{/each}
			</div>
		</div>

		<!-- Scan Option -->
		<div class="bg-surface-elevated rounded-lg p-4">
			<label class="flex items-start gap-3 cursor-pointer">
				<input
					type="checkbox"
					bind:checked={startScan}
					class="mt-1 w-4 h-4 rounded border-surface-border text-accent focus:ring-accent"
				/>
				<div>
					<p class="font-medium text-text-primary">Start scanning immediately</p>
					<p class="text-sm text-text-secondary">
						Begin scanning your music library as soon as setup is complete.
						This may take a while for large libraries.
					</p>
				</div>
			</label>
		</div>

		<!-- Error Message -->
		{#if error}
			<div class="bg-error/10 border border-error/20 rounded-lg p-3 text-error text-sm">
				{error}
			</div>
		{/if}

		<!-- Navigation Buttons -->
		<div class="flex items-center justify-between pt-4">
			<Button variant="ghost" onclick={() => goto('/setup/library')}>
				<ArrowLeft size={18} class="mr-2" />
				Back
			</Button>
			<Button
				variant="primary"
				onclick={handleComplete}
				disabled={completing || selectedFolders.length === 0}
			>
				{#if completing}
					<Loader2 size={18} class="mr-2 animate-spin" />
					Completing...
				{:else}
					<Play size={18} class="mr-2" />
					Complete Setup
				{/if}
			</Button>
		</div>
	{/if}
</div>

<style>
	@keyframes bounce-in {
		0% {
			transform: scale(0);
			opacity: 0;
		}
		50% {
			transform: scale(1.1);
		}
		100% {
			transform: scale(1);
			opacity: 1;
		}
	}

	.animate-bounce-in {
		animation: bounce-in 0.5s ease-out;
	}
</style>
