import { api, getStreamUrl } from './client';
import type { Track, TrackFilters, Pagination } from './types';

export interface TracksResponse {
	tracks: Track[];
	pagination: Pagination;
}

export async function getTracks(filters: TrackFilters = {}): Promise<TracksResponse> {
	const response = await api.get<Track[]>('/tracks', filters);
	return {
		tracks: response.data || [],
		pagination: response.meta?.pagination || {
			page: 1,
			limit: 20,
			total: 0,
			totalPages: 0,
			hasMore: false
		}
	};
}

export async function getTrack(id: string): Promise<Track> {
	const response = await api.get<Track>(`/tracks/${id}`);
	return response.data!;
}

export async function getRecentTracks(limit = 20): Promise<Track[]> {
	const response = await api.get<Track[]>('/recent', { type: 'tracks', limit });
	return response.data || [];
}

export async function getRandomTracks(limit = 20): Promise<Track[]> {
	const response = await api.get<Track[]>('/random', { type: 'tracks', limit });
	return response.data || [];
}

export { getStreamUrl };

export default {
	getTracks,
	getTrack,
	getRecentTracks,
	getRandomTracks,
	getStreamUrl
};
