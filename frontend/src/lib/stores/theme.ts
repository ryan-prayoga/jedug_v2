import { browser } from '$app/environment';
import { onMount } from 'svelte';

export type ThemeMode = 'light' | 'dark' | 'system';

const THEME_KEY = 'jedug-theme';

function getSystemPreference(): 'light' | 'dark' {
	if (!browser) return 'light';
	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function getStoredTheme(): ThemeMode | null {
	if (!browser) return null;
	try {
		const stored = localStorage.getItem(THEME_KEY);
		if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
	} catch {
		// localStorage unavailable
	}
	return null;
}

function applyTheme(theme: ThemeMode) {
	if (!browser) return;
	const root = document.documentElement;
	const resolvedTheme = theme === 'system' ? getSystemPreference() : theme;

	root.setAttribute('data-theme', resolvedTheme);

	if (resolvedTheme === 'dark') {
		root.classList.add('dark');
		root.classList.remove('light');
	} else {
		root.classList.add('light');
		root.classList.remove('dark');
	}
}

export function getInitialTheme(): ThemeMode {
	return getStoredTheme() || 'system';
}

export function setTheme(theme: ThemeMode) {
	if (browser) {
		try {
			localStorage.setItem(THEME_KEY, theme);
		} catch {
			// localStorage unavailable
		}
	}
	applyTheme(theme);
}

export function getResolvedTheme(): 'light' | 'dark' {
	const stored = getStoredTheme() || 'system';
	return stored === 'system' ? getSystemPreference() : stored;
}

export function useThemeSync() {
	if (!browser) return;

	// Listen for system preference changes
	const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
	const handleChange = () => {
		const stored = getStoredTheme();
		if (stored === 'system') {
			applyTheme('system');
		}
	};

	mediaQuery.addEventListener('change', handleChange);

	return () => {
		mediaQuery.removeEventListener('change', handleChange);
	};
}

// Svelte 5 store-like API
export const theme = {
	get current(): ThemeMode {
		return getInitialTheme();
	},
	get resolved(): 'light' | 'dark' {
		return getResolvedTheme();
	},
	set(value: ThemeMode) {
		setTheme(value);
	},
	getIcon(): string {
		return getResolvedTheme() === 'dark' ? 'sun' : 'moon';
	},
	getLabel(): string {
		return getResolvedTheme() === 'dark' ? 'Mode Terang' : 'Mode Gelap';
	}
};

export { onMount };
