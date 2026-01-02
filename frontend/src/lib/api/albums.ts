import { api, getArtworkUrl } from './client';
import type { Album, AlbumFilters, Pagination } from './types';

export interface AlbumsResponse {
	albums: Album[];
	pagination: Pagination;
}

export async function getAlbums(filters: AlbumFilters = {}): Promise<AlbumsResponse> {
	const response = await api.get<Album[]>('/albums', filters);
	return {
		albums: response.data || [],
		pagination: response.meta?.pagination || {
			page: 1,
			limit: 20,
			total: 0,
			totalPages: 0,
			hasMore: false
		}
	};
}

export async function getAlbum(id: string): Promise<Album> {
	const response = await api.get<Album>(`/albums/${id}`);
	return response.data!;
}

export async function getRecentAlbums(limit = 20): Promise<Album[]> {
	const response = await api.get<Album[]>('/recent', { type: 'albums', limit });
	return response.data || [];
}

export async function getRandomAlbums(limit = 20): Promise<Album[]> {
	const response = await api.get<Album[]>('/random', { type: 'albums', limit });
	return response.data || [];
}

export function getAlbumArtworkUrl(
	albumId: string,
	size: 'thumbnail' | 'small' | 'medium' | 'large' = 'medium'
): string {
	return getArtworkUrl('album', albumId, size);
}

export default {
	getAlbums,
	getAlbum,
	getRecentAlbums,
	getRandomAlbums,
	getAlbumArtworkUrl
};
