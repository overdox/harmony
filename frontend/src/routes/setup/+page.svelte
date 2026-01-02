<script lang="ts">
	import { goto } from '$app/navigation';
	import { Folder, Play, ArrowRight, Shield } from 'lucide-svelte';
	import { Button } from '$lib/components/ui';
	import logo from '$lib/assets/logo.svg';
</script>

<div class="text-center space-y-8 animate-fade-in">
	<!-- Logo/Icon -->
	<div class="flex justify-center">
		<img src={logo} alt="Harmony" class="w-24 h-24" />
	</div>

	<!-- Welcome Text -->
	<div class="space-y-3">
		<h1 class="text-4xl font-bold text-text-primary">Welcome to Harmony</h1>
		<p class="text-lg text-text-secondary max-w-md mx-auto">
			Your personal, self-hosted music streaming server. Let's get you set up in just a few steps.
		</p>
	</div>

	<!-- Features -->
	<div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-left">
		<div class="bg-surface-elevated rounded-lg p-4 space-y-2">
			<div class="w-10 h-10 rounded-lg bg-accent/10 flex items-center justify-center">
				<Folder size={20} class="text-accent" />
			</div>
			<h3 class="font-medium text-text-primary">Your Music</h3>
			<p class="text-sm text-text-secondary">Stream your personal music collection from anywhere.</p>
		</div>
		<div class="bg-surface-elevated rounded-lg p-4 space-y-2">
			<div class="w-10 h-10 rounded-lg bg-accent/10 flex items-center justify-center">
				<Play size={20} class="text-accent" />
			</div>
			<h3 class="font-medium text-text-primary">High Quality</h3>
			<p class="text-sm text-text-secondary">Listen in original quality or transcode on the fly.</p>
		</div>
		<div class="bg-surface-elevated rounded-lg p-4 space-y-2">
			<div class="w-10 h-10 rounded-lg bg-accent/10 flex items-center justify-center">
				<Shield size={20} class="text-accent" />
			</div>
			<h3 class="font-medium text-text-primary">Full Control</h3>
			<p class="text-sm text-text-secondary">Your data stays on your server. No subscriptions.</p>
		</div>
	</div>

	<!-- CTA Button -->
	<div class="pt-4">
		<Button variant="primary" size="lg" onclick={() => goto('/setup/library')}>
			Get Started
			<ArrowRight size={20} class="ml-2" />
		</Button>
	</div>
</div>

<style>
	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.animate-fade-in {
		animation: fade-in 0.5s ease-out;
	}
</style>
