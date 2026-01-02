import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	resolve: {
		alias: {
			$components: '/src/lib/components',
			$stores: '/src/lib/stores',
			$api: '/src/lib/api',
			$utils: '/src/lib/utils',
			$audio: '/src/lib/audio'
		}
	},
	server: {
		port: 3000,
		proxy: {
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true
			}
		}
	}
});
