import { browser } from '$app/environment';

const prefix = 'music-player:';

export function readMusicStorage(key: string): unknown {
	if (!browser) return null;
	try {
		return JSON.parse(localStorage.getItem(prefix + key) ?? 'null');
	} catch {
		return null;
	}
}

export function writeMusicStorage(key: string, value: unknown): boolean {
	if (!browser) return false;
	try {
		if (value === null) localStorage.removeItem(prefix + key);
		else localStorage.setItem(prefix + key, JSON.stringify(value));
		return true;
	} catch {
		return false;
	}
}
