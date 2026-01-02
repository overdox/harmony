<script lang="ts">
	import { Sun, Moon, Monitor, Check } from 'lucide-svelte';
	import { Button } from '$lib/components/ui';
	import {
		theme,
		accentColor,
		accentColors,
		setTheme,
		setAccentColor,
		type Theme,
		type AccentColor
	} from '$lib/stores/theme';

	const themes: { id: Theme; label: string; icon: typeof Sun }[] = [
		{ id: 'dark', label: 'Dark', icon: Moon },
		{ id: 'light', label: 'Light', icon: Sun },
		{ id: 'system', label: 'System', icon: Monitor }
	];

	const accents: { id: AccentColor; label: string }[] = [
		{ id: 'green', label: 'Green' },
		{ id: 'blue', label: 'Blue' },
		{ id: 'purple', label: 'Purple' },
		{ id: 'pink', label: 'Pink' },
		{ id: 'orange', label: 'Orange' },
		{ id: 'red', label: 'Red' }
	];
</script>

<svelte:head>
	<title>Settings - Harmony</title>
</svelte:head>

<div class="p-6 max-w-2xl mx-auto">
	<h1 class="text-3xl font-bold mb-8">Settings</h1>

	<!-- Theme Section -->
	<section class="mb-8">
		<h2 class="text-lg font-semibold mb-4">Theme</h2>
		<div class="grid grid-cols-3 gap-3">
			{#each themes as t}
				<button
					class="flex flex-col items-center gap-2 p-4 rounded-lg border-2 transition-colors {$theme === t.id
						? 'border-accent bg-accent/10'
						: 'border-surface-border hover:border-surface-active'}"
					onclick={() => setTheme(t.id)}
				>
					<svelte:component this={t.icon} size={24} />
					<span class="text-sm font-medium">{t.label}</span>
					{#if $theme === t.id}
						<Check size={16} class="text-accent" />
					{/if}
				</button>
			{/each}
		</div>
	</section>

	<!-- Accent Color Section -->
	<section class="mb-8">
		<h2 class="text-lg font-semibold mb-4">Accent Color</h2>
		<div class="grid grid-cols-3 sm:grid-cols-6 gap-3">
			{#each accents as accent}
				{@const colors = accentColors[accent.id]}
				<button
					class="flex flex-col items-center gap-2 p-3 rounded-lg border-2 transition-colors {$accentColor === accent.id
						? 'border-accent'
						: 'border-transparent hover:border-surface-border'}"
					onclick={() => setAccentColor(accent.id)}
				>
					<div
						class="w-8 h-8 rounded-full flex items-center justify-center"
						style="background-color: {colors.primary}"
					>
						{#if $accentColor === accent.id}
							<Check size={16} class="text-white" />
						{/if}
					</div>
					<span class="text-xs">{accent.label}</span>
				</button>
			{/each}
		</div>
	</section>

	<!-- About Section -->
	<section class="mb-8">
		<h2 class="text-lg font-semibold mb-4">About</h2>
		<div class="bg-surface-elevated rounded-lg p-4 space-y-2">
			<p class="text-sm">
				<span class="text-text-secondary">Version:</span> 1.0.0
			</p>
			<p class="text-sm">
				<span class="text-text-secondary">Made with:</span> SvelteKit, Tailwind CSS, Go
			</p>
			<p class="text-sm text-text-secondary">
				Harmony is a self-hosted music streaming application.
			</p>
		</div>
	</section>

	<!-- Keyboard Shortcuts Section -->
	<section>
		<h2 class="text-lg font-semibold mb-4">Keyboard Shortcuts</h2>
		<div class="bg-surface-elevated rounded-lg divide-y divide-surface-border">
			{#each [
				{ key: 'Space', action: 'Play / Pause' },
				{ key: 'N', action: 'Next track' },
				{ key: 'P', action: 'Previous track' },
				{ key: 'M', action: 'Mute / Unmute' },
				{ key: 'S', action: 'Toggle shuffle' },
				{ key: 'R', action: 'Toggle repeat' },
				{ key: 'Q', action: 'Toggle queue' },
				{ key: 'Shift + ←/→', action: 'Seek backward/forward' },
				{ key: 'Shift + ↑/↓', action: 'Volume up/down' },
				{ key: '/', action: 'Focus search' }
			] as shortcut}
				<div class="flex items-center justify-between p-3">
					<span class="text-sm">{shortcut.action}</span>
					<kbd class="px-2 py-1 bg-surface-hover rounded text-xs font-mono">
						{shortcut.key}
					</kbd>
				</div>
			{/each}
		</div>
	</section>
</div>
