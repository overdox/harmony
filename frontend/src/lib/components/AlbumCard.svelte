<script lang="ts">
	import { Play } from 'lucide-svelte';
	import { Card, Button } from '$lib/components/ui';
	import { getAlbumArtworkUrl } from '$lib/api/albums';
	import { playTracks } from '$lib/stores/player';
	import { getAudioController } from '$lib/audio';
	import type { Album } from '$lib/api/types';

	interface Props {
		album: Album;
		showArtist?: boolean;
		size?: 'sm' | 'md' | 'lg';
	}

	let { album, showArtist = true, size = 'md' }: Props = $props();

	const sizeClasses = {
		sm: 'w-32',
		md: 'w-40',
		lg: 'w-48'
	};

	let isHovered = $state(false);

	async function handlePlay(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();

		if (album.tracks && album.tracks.length > 0) {
			// Map tracks to player format
			const playerTracks = album.tracks.map((t) => ({
				id: t.id,
				title: t.title,
				duration: t.duration,
				trackNumber: t.trackNumber,
				format: t.format,
				albumId: album.id,
				albumTitle: album.title,
				artistId: album.artistId,
				artistName: album.artistName
			}));
			playTracks(playerTracks);
			getAudioController().play();
		}
	}
</script>

<Card
	href="/album/{album.id}"
	class="{sizeClasses[size]} flex-shrink-0"
	onmouseenter={() => (isHovered = true)}
	onmouseleave={() => (isHovered = false)}
>
	<div class="relative aspect-square mb-3 rounded-md overflow-hidden bg-surface-hover shadow-lg">
		<img
			src={getAlbumArtworkUrl(album.id, 'medium')}
			alt={album.title}
			class="w-full h-full object-cover"
			loading="lazy"
		/>
		{#if isHovered}
			<div class="absolute inset-0 bg-black/40 flex items-center justify-center transition-opacity">
				<Button
					variant="primary"
					size="lg"
					class="w-12 h-12 rounded-full shadow-xl"
					onclick={handlePlay}
				>
					<Play size={24} class="ml-0.5" />
				</Button>
			</div>
		{/if}
	</div>
	<h3 class="font-medium text-sm truncate">{album.title}</h3>
	{#if showArtist}
		<p class="text-xs text-text-secondary truncate mt-0.5">
			{album.year ? `${album.year} • ` : ''}{album.artistName || 'Unknown Artist'}
		</p>
	{:else if album.year}
		<p class="text-xs text-text-secondary truncate mt-0.5">{album.year}</p>
	{/if}
</Card>
