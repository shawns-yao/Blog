import { browser } from '$app/environment';

export type Theme = 'light' | 'dark';
export type ResolvedTheme = 'light' | 'dark';

class ThemeManager {
	current = $state<Theme>('light');

	set(theme: Theme) {
		this.current = theme;
	}
}

export const themeManager = new ThemeManager();

export const resolveTheme = (theme: Theme): ResolvedTheme => theme;

export const initTheme = (manager: ThemeManager): void => {
	if (!browser) return;

	const saved = localStorage.getItem('theme');
	manager.set(saved === 'dark' ? 'dark' : 'light');
};

export const startThemeSync = (manager: ThemeManager): void => {
	$effect(() => {
		if (!browser) return;

		document.documentElement.classList.toggle('dark', manager.current === 'dark');
		localStorage.setItem('theme', manager.current);
	});
};
