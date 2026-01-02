import { api, getArtworkUrl } from './client';
import type { Artist, ArtistFilters, Pagination } from './types';

export interface ArtistsResponse {
	artists: Artist[];
	pagination: Pagination;
}

export async function getArtists(filters: ArtistFilters = {}): Promise<ArtistsResponse> {
	const response = await api.get<Artist[]>('/artists', filters);
	return {
		artists: response.data || [],
		pagination: response.meta?.pagination || {
			page: 1,
			limit: 20,
			total: 0,
			totalPages: 0,
			hasMore: false
		}
	};
}

export async function getArtist(id: string): Promise<Artist> {
	const response = await api.get<Artist>(`/artists/${id}`);
	return response.data!;
}

export function getArtistImageUrl(
	artistId: string,
	size: 'thumbnail' | 'small' | 'medium' | 'large' = 'medium'
): string {
	return getArtworkUrl('artist', artistId, size);
}

export default {
	getArtists,
	getArtist,
	getArtistImageUrl
};
