import { browser } from '$app/environment';

export type Theme = 'light' | 'dark';
export type ResolvedTheme = 'light' | 'dark';

class ThemeManager {
	current = $state<Theme>('light');
	private overrideUntil = 0;

	set(theme: Theme) {
		this.current = theme;
		const next = new Date();
		const hour = next.getHours();
		next.setHours(hour < 6 ? 6 : hour < 18 ? 18 : 30, 0, 0, 0);
		this.overrideUntil = next.getTime();
		if (browser) {
			try {
				localStorage.setItem('theme', theme);
				localStorage.setItem('theme-override-until', String(this.overrideUntil));
			} catch { /* The in-memory override still works when storage is unavailable. */ }
		}
	}

	syncTime() {
		const now = new Date();
		if (now.getTime() < this.overrideUntil) return;
		this.current = now.getHours() >= 6 && now.getHours() < 18 ? 'light' : 'dark';
	}

	restore() {
		try {
			const until = Number(localStorage.getItem('theme-override-until'));
			const saved = localStorage.getItem('theme');
			if (Number.isFinite(until) && until > Date.now() && (saved === 'light' || saved === 'dark')) {
				this.current = saved;
				this.overrideUntil = until;
			}
		} catch { /* Fall back to the local clock. */ }
		this.syncTime();
	}
}

export const themeManager = new ThemeManager();

export const resolveTheme = (theme: Theme): ResolvedTheme => theme;

export const initTheme = (manager: ThemeManager): void => {
	if (!browser) return;

	manager.restore();
};

export const startThemeSync = (manager: ThemeManager): void => {
	$effect(() => {
		if (!browser) return;
		let timer: ReturnType<typeof setTimeout>;
		const sync = () => {
			clearTimeout(timer);
			manager.syncTime();
			const now = new Date();
			const boundary = new Date(now);
			boundary.setHours(now.getHours() < 6 ? 6 : now.getHours() < 18 ? 18 : 30, 0, 0, 0);
			// Recheck clock changes as well as the exact next day/night boundary.
			timer = setTimeout(sync, Math.min(60_000, Math.max(1, boundary.getTime() - now.getTime())));
		};
		sync();
		window.addEventListener('focus', sync);
		document.addEventListener('visibilitychange', sync);
		return () => {
			clearTimeout(timer);
			window.removeEventListener('focus', sync);
			document.removeEventListener('visibilitychange', sync);
		};
	});
	$effect(() => {
		if (!browser) return;

		document.documentElement.classList.toggle('dark', manager.current === 'dark');
	});
};
