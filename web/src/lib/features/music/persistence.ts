import { readMusicStorage, writeMusicStorage } from '$lib/shared/dom/music-storage';
import type { MusicDensity, MusicPlaybackMode, MusicSong } from './types';

export const musicModes: MusicPlaybackMode[] = ['sequence', 'single', 'repeat', 'shuffle'];
export const musicModeLabels: Record<MusicPlaybackMode, string> = {
	sequence: '顺序播放',
	single: '单曲循环',
	repeat: '列表循环',
	shuffle: '随机播放'
};

export interface MusicPreferences {
	volume: number;
	density: MusicDensity;
	showCovers: boolean;
	mode: MusicPlaybackMode;
}

export function loadMusicPreferences(): MusicPreferences {
	const defaults: MusicPreferences = {
		volume: 1,
		density: 'comfortable',
		showCovers: true,
		mode: 'sequence'
	};
	const saved = readMusicStorage('preferences') as Partial<MusicPreferences> | null;
	if (!saved || typeof saved !== 'object') return defaults;
	return {
		volume:
			typeof saved.volume === 'number' && Number.isFinite(saved.volume)
				? Math.min(1, Math.max(0, saved.volume))
				: defaults.volume,
		density: saved.density === 'compact' ? 'compact' : defaults.density,
		showCovers: typeof saved.showCovers === 'boolean' ? saved.showCovers : defaults.showCovers,
		mode: musicModes.includes(saved.mode as MusicPlaybackMode) ? saved.mode! : defaults.mode
	};
}

export function saveMusicPreferences(preferences: MusicPreferences) {
	return writeMusicStorage('preferences', preferences);
}

export function loadMusicQueue(viewer: number): { songs: MusicSong[]; index: number } {
	const empty = { songs: [], index: -1 };
	const saved = readMusicStorage('queue') as {
		version?: number;
		viewer?: number;
		songs?: unknown[];
		currentId?: string;
	} | null;
	if (!saved || saved.version !== 1 || saved.viewer !== viewer || !Array.isArray(saved.songs))
		return empty;
	const seen = new Set<string>();
	const songs = saved.songs.slice(0, 500).filter((value): value is MusicSong => {
		if (!value || typeof value !== 'object') return false;
		const song = value as MusicSong;
		if (
			typeof song.id !== 'string' ||
			!song.id ||
			seen.has(song.id) ||
			typeof song.title !== 'string' ||
			typeof song.artist !== 'string' ||
			typeof song.album !== 'string' ||
			typeof song.albumId !== 'string' ||
			typeof song.public !== 'boolean' ||
			(song.coverArt !== undefined && typeof song.coverArt !== 'string') ||
			!['duration', 'track', 'year'].every(
				(key) =>
					typeof song[key as 'duration' | 'track' | 'year'] === 'number' &&
					Number.isFinite(song[key as 'duration' | 'track' | 'year'])
			)
		)
			return false;
		seen.add(song.id);
		return viewer !== 0 || song.public;
	});
	return { songs, index: songs.findIndex((song) => song.id === saved.currentId) };
}

export function saveMusicQueue(viewer: number, songs: MusicSong[], index: number) {
	return writeMusicStorage(
		'queue',
		songs.length
			? {
					version: 1,
					viewer,
					songs,
					currentId: songs[index]?.id
				}
			: null
	);
}
