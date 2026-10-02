import { getApi } from '$lib/shared/clients/api';
import type { MusicAlbum, MusicCatalog, MusicLyricsResult, MusicStatus } from './types';

export function getMusicStatus(signal?: AbortSignal) {
	return getApi()<MusicStatus>('/public/music/status', { signal });
}

export async function getMusicCatalog(
	query: string,
	offset: number,
	signal?: AbortSignal,
	publicAccess = false
) {
	if (!publicAccess) await getApi()('/music/session', { method: 'POST', signal });
	return getApi()<MusicCatalog>(`${publicAccess ? '/public' : ''}/music/catalog`, {
		query: { query, offset, limit: 30 },
		signal
	});
}

export async function getMusicAlbum(id: string, signal?: AbortSignal, publicAccess = false) {
	if (!publicAccess) await getApi()('/music/session', { method: 'POST', signal });
	return getApi()<MusicAlbum>(
		`${publicAccess ? '/public' : ''}/music/albums/${encodeURIComponent(id)}`,
		{ signal }
	);
}

export function clearMusicSession() {
	return getApi()('/music/session', { method: 'DELETE' });
}

export function getMusicLyrics(id: string, signal?: AbortSignal, publicAccess = false) {
	return getApi()<MusicLyricsResult>(
		`${publicAccess ? '/public' : ''}/music/lyrics/${encodeURIComponent(id)}`,
		{ signal }
	);
}

export function musicStreamURL(id: string, publicAccess = false) {
	return `/api/v2${publicAccess ? '/public' : ''}/music/stream/${encodeURIComponent(id)}`;
}

export function musicCoverURL(id: string, publicAccess = false) {
	return `/api/v2${publicAccess ? '/public' : ''}/music/cover/${encodeURIComponent(id)}`;
}

export function formatMusicTime(seconds: number) {
	if (!Number.isFinite(seconds) || seconds < 0) return '0:00';
	return `${Math.floor(seconds / 60)}:${String(Math.floor(seconds % 60)).padStart(2, '0')}`;
}
