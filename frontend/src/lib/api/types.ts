// API Response Types

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	error?: ApiError;
	meta?: ApiMeta;
}

export interface ApiError {
	code: string;
	message: string;
	details?: string;
}

export interface ApiMeta {
	pagination?: Pagination;
}

export interface Pagination {
	page: number;
	limit: number;
	total: number;
	totalPages: number;
	hasMore: boolean;
}

export interface Link {
	href: string;
	rel: string;
}

// Entity Types

export interface Track {
	id: string;
	title: string;
	duration: number;
	trackNumber: number;
	discNumber: number;
	format: string;
	bitrate?: number;
	albumId?: string;
	artistId?: string;
	genre?: string;
	year?: number;
	links?: Link[];
}

export interface Album {
	id: string;
	title: string;
	year?: number;
	artistId: string;
	artistName?: string;
	trackCount?: number;
	duration?: number;
	coverArtUrl?: string;
	links?: Link[];
	tracks?: Track[];
}

export interface Artist {
	id: string;
	name: string;
	bio?: string;
	imageUrl?: string;
	albumCount?: number;
	trackCount?: number;
	links?: Link[];
	albums?: Album[];
	popularTracks?: Track[];
}

export interface Playlist {
	id: string;
	name: string;
	description?: string;
	isPublic: boolean;
	trackCount: number;
	duration: number;
	userId: string;
	coverImageUrl?: string;
	createdAt: string;
	updatedAt: string;
	tracks?: Track[];
}

export interface User {
	id: string;
	username: string;
	email: string;
	createdAt: string;
}

// Search Types

export interface SearchResults {
	query: string;
	tracks: Track[];
	albums: Album[];
	artists: Artist[];
}

// Library Types

export interface LibraryStats {
	totalTracks: number;
	totalAlbums: number;
	totalArtists: number;
	totalDuration: number;
	totalSize: number;
	lastScanAt?: string;
}

export interface ScanProgress {
	status: 'idle' | 'scanning' | 'processing' | 'completed' | 'failed' | 'cancelled';
	totalFiles: number;
	processedFiles: number;
	newTracks: number;
	updatedTracks: number;
	deletedTracks: number;
	errorCount: number;
	currentFile?: string;
	startedAt?: string;
	completedAt?: string;
	duration?: string;
}

// Request Types

export interface PaginationParams {
	page?: number;
	limit?: number;
}

export interface SortParams {
	sortBy?: string;
	order?: 'asc' | 'desc';
}

export interface TrackFilters extends PaginationParams, SortParams {
	albumId?: string;
	artistId?: string;
	genre?: string;
	year?: number;
	q?: string;
}

export interface AlbumFilters extends PaginationParams, SortParams {
	artistId?: string;
	year?: number;
	q?: string;
}

export interface ArtistFilters extends PaginationParams, SortParams {
	q?: string;
}

export interface PlaylistFilters extends PaginationParams, SortParams {
	userId?: string;
	q?: string;
}

export interface CreatePlaylistRequest {
	name: string;
	description?: string;
	isPublic?: boolean;
}

export interface UpdatePlaylistRequest {
	name?: string;
	description?: string;
	isPublic?: boolean;
}
