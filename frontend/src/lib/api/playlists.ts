import { api, getArtworkUrl } from './client';
import type {
	Playlist,
	PlaylistFilters,
	Pagination,
	CreatePlaylistRequest,
	UpdatePlaylistRequest
} from './types';

export interface PlaylistsResponse {
	playlists: Playlist[];
	pagination: Pagination;
}

export async function getPlaylists(filters: PlaylistFilters = {}): Promise<PlaylistsResponse> {
	const response = await api.get<Playlist[]>('/playlists', filters);
	return {
		playlists: response.data || [],
		pagination: response.meta?.pagination || {
			page: 1,
			limit: 20,
			total: 0,
			totalPages: 0,
			hasMore: false
		}
	};
}

export async function getPlaylist(id: string): Promise<Playlist> {
	const response = await api.get<Playlist>(`/playlists/${id}`);
	return response.data!;
}

export async function createPlaylist(data: CreatePlaylistRequest): Promise<Playlist> {
	const response = await api.post<Playlist>('/playlists', data);
	return response.data!;
}

export async function updatePlaylist(id: string, data: UpdatePlaylistRequest): Promise<Playlist> {
	const response = await api.put<Playlist>(`/playlists/${id}`, data);
	return response.data!;
}

export async function deletePlaylist(id: string): Promise<void> {
	await api.delete(`/playlists/${id}`);
}

export async function addTrackToPlaylist(playlistId: string, trackId: string): Promise<void> {
	await api.post(`/playlists/${playlistId}/tracks`, { trackId });
}

export async function removeTrackFromPlaylist(
	playlistId: string,
	trackId: string
): Promise<void> {
	await api.delete(`/playlists/${playlistId}/tracks/${trackId}`);
}

export function getPlaylistCoverUrl(
	playlistId: string,
	size: 'thumbnail' | 'small' | 'medium' | 'large' = 'medium'
): string {
	return getArtworkUrl('playlist', playlistId, size);
}

export default {
	getPlaylists,
	getPlaylist,
	createPlaylist,
	updatePlaylist,
	deletePlaylist,
	addTrackToPlaylist,
	removeTrackFromPlaylist,
	getPlaylistCoverUrl
};
