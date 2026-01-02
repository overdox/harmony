<script lang="ts">
	import { Play } from 'lucide-svelte';
	import { Card, Button, Avatar } from '$lib/components/ui';
	import { getArtistImageUrl } from '$lib/api/artists';
	import type { Artist } from '$lib/api/types';

	interface Props {
		artist: Artist;
		size?: 'sm' | 'md' | 'lg';
	}

	let { artist, size = 'md' }: Props = $props();

	const sizeClasses = {
		sm: 'w-32',
		md: 'w-40',
		lg: 'w-48'
	};

	const avatarSizes = {
		sm: 'lg' as const,
		md: 'xl' as const,
		lg: 'xl' as const
	};

	let isHovered = $state(false);
</script>

<Card
	href="/artist/{artist.id}"
	class="{sizeClasses[size]} flex-shrink-0 text-center"
	onmouseenter={() => (isHovered = true)}
	onmouseleave={() => (isHovered = false)}
>
	<div class="relative mb-3 flex justify-center">
		<div class="relative">
			<Avatar
				src={getArtistImageUrl(artist.id, 'medium')}
				alt={artist.name}
				fallback={artist.name}
				size={avatarSizes[size]}
				class="shadow-lg"
			/>
			{#if isHovered}
				<div class="absolute inset-0 rounded-full bg-black/40 flex items-center justify-center transition-opacity">
					<Button
						variant="primary"
						size="md"
						class="w-10 h-10 rounded-full shadow-xl"
					>
						<Play size={20} class="ml-0.5" />
					</Button>
				</div>
			{/if}
		</div>
	</div>
	<h3 class="font-medium text-sm truncate">{artist.name}</h3>
	<p class="text-xs text-text-secondary truncate mt-0.5">Artist</p>
</Card>
