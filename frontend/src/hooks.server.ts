import type { Handle } from '@sveltejs/kit';

// Backend base URL (without /api/v1 suffix)
const BACKEND_HOST = process.env.BACKEND_HOST || 'http://backend:8080';

export const handle: Handle = async ({ event, resolve }) => {
	// Proxy API requests to backend
	if (event.url.pathname.startsWith('/api/')) {
		const backendUrl = `${BACKEND_HOST}${event.url.pathname}${event.url.search}`;

		try {
			const response = await fetch(backendUrl, {
				method: event.request.method,
				headers: event.request.headers,
				body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
					? await event.request.text()
					: undefined,
				// Don't follow redirects, let the client handle them
				redirect: 'manual'
			});

			// Clone headers but remove some that shouldn't be forwarded
			const headers = new Headers(response.headers);
			headers.delete('transfer-encoding');

			return new Response(response.body, {
				status: response.status,
				statusText: response.statusText,
				headers
			});
		} catch (error) {
			console.error('API proxy error:', error);
			return new Response(JSON.stringify({
				error: 'Backend unavailable',
				message: 'Could not connect to backend API'
			}), {
				status: 502,
				headers: { 'Content-Type': 'application/json' }
			});
		}
	}

	return resolve(event);
};
