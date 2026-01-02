<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import {
		Folder,
		FolderOpen,
		Music,
		ChevronRight,
		ChevronUp,
		Check,
		ArrowLeft,
		ArrowRight,
		Loader2
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui';
	import {
		browseFolders,
		setSelectedFolders,
		getSelectedFolders,
		type FolderInfo,
		type BrowseFoldersResponse
	} from '$lib/api/setup';

	let loading = $state(true);
	let saving = $state(false);
	let error = $state<string | null>(null);

	let currentPath = $state('');
	let parentPath = $state('');
	let isMediaRoot = $state(true);
	let hasAudioFiles = $state(false);
	let folders = $state<FolderInfo[]>([]);
	let selectedPaths = $state<Set<string>>(new Set());

	onMount(async () => {
		await loadFolders();
		await loadSelectedFolders();
	});

	async function loadFolders(path?: string) {
		loading = true;
		error = null;

		try {
			const response = await browseFolders(path);
			currentPath = response.currentPath;
			parentPath = response.parentPath;
			isMediaRoot = response.isMediaRoot;
			hasAudioFiles = response.hasAudioFiles;
			folders = response.folders;
		} catch (e) {
			console.error('Failed to load folders:', e);
			error = 'Failed to load folders';
		} finally {
			loading = false;
		}
	}

	async function loadSelectedFolders() {
		try {
			const paths = await getSelectedFolders();
			selectedPaths = new Set(paths);
		} catch (e) {
			console.error('Failed to load selected folders:', e);
		}
	}

	function navigateToFolder(path: string) {
		loadFolders(path);
	}

	function navigateUp() {
		if (!isMediaRoot) {
			loadFolders(parentPath);
		}
	}

	function toggleSelection(path: string) {
		const newSelected = new Set(selectedPaths);
		if (newSelected.has(path)) {
			newSelected.delete(path);
		} else {
			newSelected.add(path);
		}
		selectedPaths = newSelected;
	}

	function selectCurrentFolder() {
		const newSelected = new Set(selectedPaths);
		newSelected.add(currentPath);
		selectedPaths = newSelected;
	}

	function isSelected(path: string): boolean {
		return selectedPaths.has(path);
	}

	async function handleNext() {
		if (selectedPaths.size === 0) {
			error = 'Please select at least one folder';
			return;
		}

		saving = true;
		error = null;

		try {
			await setSelectedFolders(Array.from(selectedPaths));
			goto('/setup/complete');
		} catch (e) {
			console.error('Failed to save folders:', e);
			error = 'Failed to save folder selection';
		} finally {
			saving = false;
		}
	}

	const selectedCount = $derived(selectedPaths.size);
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="text-center space-y-2">
		<h1 class="text-2xl font-bold text-text-primary">Select Your Music Folders</h1>
		<p class="text-text-secondary">
			Browse and select the folders containing your music library.
		</p>
	</div>

	<!-- Breadcrumb / Current Path -->
	<div class="bg-surface-elevated rounded-lg p-3 flex items-center gap-2">
		<Folder size={18} class="text-text-muted flex-shrink-0" />
		<span class="text-sm text-text-secondary truncate flex-1">{currentPath}</span>
		{#if !isMediaRoot}
			<Button variant="ghost" size="sm" onclick={navigateUp}>
				<ChevronUp size={16} />
				Up
			</Button>
		{/if}
	</div>

	<!-- Add Current Folder Button -->
	{#if hasAudioFiles && !isSelected(currentPath)}
		<Button
			variant="secondary"
			class="w-full justify-start"
			onclick={selectCurrentFolder}
		>
			<Music size={18} class="mr-2 text-accent" />
			Add this folder ({currentPath.split('/').pop()})
		</Button>
	{/if}

	<!-- Folder Browser -->
	<div class="bg-surface-elevated rounded-lg overflow-hidden">
		{#if loading}
			<div class="p-8 flex items-center justify-center">
				<Loader2 size={24} class="animate-spin text-accent" />
			</div>
		{:else if folders.length === 0}
			<div class="p-8 text-center text-text-muted">
				<Folder size={32} class="mx-auto mb-2 opacity-50" />
				<p>No subfolders found</p>
			</div>
		{:else}
			<div class="divide-y divide-surface-border max-h-80 overflow-y-auto">
				{#each folders as folder}
					<div
						class="flex items-center gap-3 p-3 hover:bg-surface-hover transition-colors"
					>
						<!-- Selection Checkbox -->
						<button
							class="w-5 h-5 rounded border-2 flex items-center justify-center transition-colors {isSelected(folder.path) ? 'bg-accent border-accent' : 'border-surface-border hover:border-accent'}"
							onclick={() => toggleSelection(folder.path)}
						>
							{#if isSelected(folder.path)}
								<Check size={12} class="text-black" />
							{/if}
						</button>

						<!-- Folder Info -->
						<button
							class="flex-1 flex items-center gap-3 text-left"
							onclick={() => navigateToFolder(folder.path)}
						>
							{#if folder.hasAudio}
								<FolderOpen size={20} class="text-accent" />
							{:else}
								<Folder size={20} class="text-text-muted" />
							{/if}
							<div class="flex-1 min-w-0">
								<p class="font-medium text-text-primary truncate">{folder.name}</p>
								<p class="text-xs text-text-muted">
									{#if folder.hasAudio}
										<span class="text-accent">Contains audio</span>
									{:else if folder.children > 0}
										{folder.children} subfolder{folder.children !== 1 ? 's' : ''}
									{:else}
										Empty
									{/if}
								</p>
							</div>
						</button>

						<!-- Navigate Arrow -->
						{#if folder.children > 0}
							<button
								class="p-1 hover:bg-surface-active rounded transition-colors"
								onclick={() => navigateToFolder(folder.path)}
							>
								<ChevronRight size={18} class="text-text-muted" />
							</button>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Selected Folders Summary -->
	{#if selectedCount > 0}
		<div class="bg-accent/10 rounded-lg p-4">
			<h3 class="font-medium text-accent mb-2">
				{selectedCount} folder{selectedCount !== 1 ? 's' : ''} selected
			</h3>
			<div class="space-y-1 max-h-32 overflow-y-auto">
				{#each Array.from(selectedPaths) as path}
					<div class="flex items-center gap-2 text-sm">
						<Check size={14} class="text-accent flex-shrink-0" />
						<span class="text-text-secondary truncate">{path}</span>
						<button
							class="ml-auto text-text-muted hover:text-error"
							onclick={() => toggleSelection(path)}
						>
							&times;
						</button>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Error Message -->
	{#if error}
		<div class="bg-error/10 border border-error/20 rounded-lg p-3 text-error text-sm">
			{error}
		</div>
	{/if}

	<!-- Navigation Buttons -->
	<div class="flex items-center justify-between pt-4">
		<Button variant="ghost" onclick={() => goto('/setup')}>
			<ArrowLeft size={18} class="mr-2" />
			Back
		</Button>
		<Button
			variant="primary"
			onclick={handleNext}
			disabled={selectedCount === 0 || saving}
		>
			{#if saving}
				<Loader2 size={18} class="mr-2 animate-spin" />
				Saving...
			{:else}
				Next
				<ArrowRight size={18} class="ml-2" />
			{/if}
		</Button>
	</div>
</div>
