import type { ApiResponse, ApiError } from './types';

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1';

export class ApiClientError extends Error {
	public status: number;
	public code: string;
	public details?: string;

	constructor(status: number, error: ApiError) {
		super(error.message);
		this.name = 'ApiClientError';
		this.status = status;
		this.code = error.code;
		this.details = error.details;
	}
}

interface RequestOptions extends RequestInit {
	params?: Record<string, string | number | boolean | undefined>;
}

async function request<T>(
	endpoint: string,
	options: RequestOptions = {}
): Promise<ApiResponse<T>> {
	const { params, ...fetchOptions } = options;

	// Build URL with query parameters
	let url = `${API_BASE}${endpoint}`;
	if (params) {
		const searchParams = new URLSearchParams();
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined && value !== null && value !== '') {
				searchParams.set(key, String(value));
			}
		});
		const queryString = searchParams.toString();
		if (queryString) {
			url += `?${queryString}`;
		}
	}

	// Default headers
	const headers = new Headers(fetchOptions.headers);
	if (!headers.has('Content-Type') && fetchOptions.body) {
		headers.set('Content-Type', 'application/json');
	}

	const response = await fetch(url, {
		...fetchOptions,
		headers
	});

	// Handle non-JSON responses
	const contentType = response.headers.get('content-type');
	if (!contentType?.includes('application/json')) {
		if (!response.ok) {
			throw new ApiClientError(response.status, {
				code: 'REQUEST_FAILED',
				message: `Request failed with status ${response.status}`
			});
		}
		return { success: true, data: undefined as T };
	}

	const data: ApiResponse<T> = await response.json();

	if (!response.ok || !data.success) {
		throw new ApiClientError(response.status, data.error || {
			code: 'UNKNOWN_ERROR',
			message: 'An unknown error occurred'
		});
	}

	return data;
}

export const api = {
	get: <T>(endpoint: string, params?: Record<string, any>) =>
		request<T>(endpoint, { method: 'GET', params }),

	post: <T>(endpoint: string, body?: unknown, params?: Record<string, any>) =>
		request<T>(endpoint, {
			method: 'POST',
			body: body ? JSON.stringify(body) : undefined,
			params
		}),

	put: <T>(endpoint: string, body?: unknown, params?: Record<string, any>) =>
		request<T>(endpoint, {
			method: 'PUT',
			body: body ? JSON.stringify(body) : undefined,
			params
		}),

	patch: <T>(endpoint: string, body?: unknown, params?: Record<string, any>) =>
		request<T>(endpoint, {
			method: 'PATCH',
			body: body ? JSON.stringify(body) : undefined,
			params
		}),

	delete: <T>(endpoint: string, params?: Record<string, any>) =>
		request<T>(endpoint, { method: 'DELETE', params })
};

// Stream URL helper
export function getStreamUrl(trackId: string, quality?: string): string {
	let url = `${API_BASE}/tracks/${trackId}/stream`;
	if (quality && quality !== 'original') {
		url += `?quality=${quality}`;
	}
	return url;
}

// Artwork URL helper
export function getArtworkUrl(
	type: 'album' | 'artist' | 'playlist',
	id: string,
	size: 'thumbnail' | 'small' | 'medium' | 'large' = 'medium'
): string {
	return `${API_BASE}/artwork/${type}/${id}?size=${size}`;
}

export default api;
