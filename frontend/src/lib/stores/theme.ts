import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'dark' | 'light' | 'system';
export type AccentColor = 'green' | 'blue' | 'purple' | 'pink' | 'orange' | 'red';

const THEME_KEY = 'harmony:theme';
const ACCENT_KEY = 'harmony:accent';

// Accent color definitions
export const accentColors: Record<AccentColor, { primary: string; hover: string; muted: string }> = {
	green: { primary: '#1db954', hover: '#1ed760', muted: '#1a7f3e' },
	blue: { primary: '#3b82f6', hover: '#60a5fa', muted: '#2563eb' },
	purple: { primary: '#8b5cf6', hover: '#a78bfa', muted: '#7c3aed' },
	pink: { primary: '#ec4899', hover: '#f472b6', muted: '#db2777' },
	orange: { primary: '#f97316', hover: '#fb923c', muted: '#ea580c' },
	red: { primary: '#ef4444', hover: '#f87171', muted: '#dc2626' }
};

function getInitialTheme(): Theme {
	if (!browser) return 'dark';
	const stored = localStorage.getItem(THEME_KEY);
	if (stored && ['dark', 'light', 'system'].includes(stored)) {
		return stored as Theme;
	}
	return 'dark';
}

function getInitialAccent(): AccentColor {
	if (!browser) return 'green';
	const stored = localStorage.getItem(ACCENT_KEY);
	if (stored && Object.keys(accentColors).includes(stored)) {
		return stored as AccentColor;
	}
	return 'green';
}

// Stores
export const theme = writable<Theme>(getInitialTheme());
export const accentColor = writable<AccentColor>(getInitialAccent());

// Derived store for resolved theme (handles 'system')
export const resolvedTheme = derived(theme, ($theme) => {
	if ($theme === 'system') {
		if (browser) {
			return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
		}
		return 'dark';
	}
	return $theme;
});

// Apply theme to document
export function applyTheme(isDark: boolean): void {
	if (!browser) return;

	const root = document.documentElement;

	if (isDark) {
		root.classList.add('dark');
		root.classList.remove('light');
		root.style.colorScheme = 'dark';

		// Dark theme colors
		root.style.setProperty('--color-surface', '#121212');
		root.style.setProperty('--color-surface-elevated', '#181818');
		root.style.setProperty('--color-surface-hover', '#282828');
		root.style.setProperty('--color-surface-active', '#333333');
		root.style.setProperty('--color-surface-border', '#404040');
		root.style.setProperty('--color-text-primary', '#ffffff');
		root.style.setProperty('--color-text-secondary', '#b3b3b3');
		root.style.setProperty('--color-text-muted', '#6a6a6a');
	} else {
		root.classList.add('light');
		root.classList.remove('dark');
		root.style.colorScheme = 'light';

		// Light theme colors
		root.style.setProperty('--color-surface', '#ffffff');
		root.style.setProperty('--color-surface-elevated', '#f5f5f5');
		root.style.setProperty('--color-surface-hover', '#e5e5e5');
		root.style.setProperty('--color-surface-active', '#d4d4d4');
		root.style.setProperty('--color-surface-border', '#e5e5e5');
		root.style.setProperty('--color-text-primary', '#171717');
		root.style.setProperty('--color-text-secondary', '#525252');
		root.style.setProperty('--color-text-muted', '#a3a3a3');
	}
}

// Apply accent color to document
export function applyAccent(color: AccentColor): void {
	if (!browser) return;

	const root = document.documentElement;
	const colors = accentColors[color];

	root.style.setProperty('--color-accent', colors.primary);
	root.style.setProperty('--color-accent-hover', colors.hover);
	root.style.setProperty('--color-accent-muted', colors.muted);
}

// Theme actions
export function setTheme(newTheme: Theme): void {
	theme.set(newTheme);
	if (browser) {
		localStorage.setItem(THEME_KEY, newTheme);
	}
}

export function setAccentColor(color: AccentColor): void {
	accentColor.set(color);
	if (browser) {
		localStorage.setItem(ACCENT_KEY, color);
	}
}

export function toggleTheme(): void {
	theme.update((t) => {
		const newTheme = t === 'dark' ? 'light' : 'dark';
		if (browser) {
			localStorage.setItem(THEME_KEY, newTheme);
		}
		return newTheme;
	});
}

// Initialize theme on load
export function initializeTheme(): void {
	if (!browser) return;

	// Subscribe to resolved theme changes
	resolvedTheme.subscribe((resolved) => {
		applyTheme(resolved === 'dark');
	});

	// Subscribe to accent color changes
	accentColor.subscribe((color) => {
		applyAccent(color);
	});

	// Listen for system theme changes
	const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
	mediaQuery.addEventListener('change', (e) => {
		theme.update((t) => {
			if (t === 'system') {
				applyTheme(e.matches);
			}
			return t;
		});
	});
}
