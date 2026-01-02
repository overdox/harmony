import { api } from './client';
import type { SearchResults, LibraryStats, ScanProgress } from './types';

export async function search(query: string, limit = 10): Promise<SearchResults> {
	const response = await api.get<SearchResults>('/search', { q: query, limit });
	return response.data || { query, tracks: [], albums: [], artists: [] };
}

export async function getLibraryStats(): Promise<LibraryStats> {
	const response = await api.get<LibraryStats>('/library/stats');
	return response.data!;
}

export async function startLibraryScan(incremental = false): Promise<void> {
	await api.post('/library/scan', { incremental });
}

export async function getScanStatus(): Promise<ScanProgress> {
	const response = await api.get<ScanProgress>('/library/scan/status');
	return response.data!;
}

export async function cancelScan(): Promise<void> {
	await api.post('/library/scan/cancel');
}

export default {
	search,
	getLibraryStats,
	startLibraryScan,
	getScanStatus,
	cancelScan
};
