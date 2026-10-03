import { getApi } from '$lib/shared/clients/api';
import type {
	MusicAlbum,
	MusicCatalog,
	MusicLyricsResult,
	MusicStatus,
	MusicPlaylist,
	MusicSongPage
} from './types';

export const MUSIC_PAGE_SIZE = 10;

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
		query: { query, offset, limit: MUSIC_PAGE_SIZE },
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

export function browseMusic(query: string, offset: number, signal?: AbortSignal) {
	return getApi()<MusicCatalog>('/public/music/browse', {
		query: { query, offset, limit: MUSIC_PAGE_SIZE },
		signal
	});
}
export function browseMusicAlbum(id: string, signal?: AbortSignal) {
	return getApi()<MusicAlbum>(`/public/music/browse/albums/${encodeURIComponent(id)}`, { signal });
}
export function getMusicAccess(signal?: AbortSignal) {
	return getApi()<{ privateAccess: boolean }>('/music/access', { signal });
}
export function createMusicSession(signal?: AbortSignal) {
	return getApi()('/music/session', { method: 'POST', signal });
}
export function getMusicFavorites(offset = 0, signal?: AbortSignal) {
	return getApi()<MusicSongPage>('/music/favorites', {
		query: { offset, limit: MUSIC_PAGE_SIZE },
		signal
	});
}
export function checkMusicFavorites(ids: string[], signal?: AbortSignal) {
	return getApi()<{ ids: string[] }>('/music/favorites/check', {
		query: { ids: ids.join(',') },
		signal
	});
}
export function setMusicFavorite(id: string, add: boolean) {
	return getApi()(`/music/favorites/${encodeURIComponent(id)}`, { method: add ? 'PUT' : 'DELETE' });
}
export function getMusicPlaylists(offset = 0, signal?: AbortSignal) {
	return getApi()<{ items: MusicPlaylist[]; hasMore: boolean }>('/music/playlists', {
		query: { offset, limit: MUSIC_PAGE_SIZE },
		signal
	});
}
export function createMusicPlaylist(name: string) {
	return getApi()<MusicPlaylist>('/music/playlists', { method: 'POST', body: { name } });
}
export function renameMusicPlaylist(id: number, name: string) {
	return getApi()(`/music/playlists/${id}`, { method: 'PATCH', body: { name } });
}
export function deleteMusicPlaylist(id: number) {
	return getApi()(`/music/playlists/${id}`, { method: 'DELETE' });
}
export function getMusicPlaylist(id: number, offset = 0, signal?: AbortSignal) {
	return getApi()<MusicSongPage & { playlist: MusicPlaylist }>(`/music/playlists/${id}`, {
		query: { offset, limit: MUSIC_PAGE_SIZE },
		signal
	});
}
export function setMusicPlaylistSong(id: number, songId: string, add: boolean) {
	return getApi()(`/music/playlists/${id}/songs/${encodeURIComponent(songId)}`, {
		method: add ? 'PUT' : 'DELETE'
	});
}
export function reorderMusicPlaylist(id: number, songIds: string[]) {
	return getApi()(`/music/playlists/${id}/order`, { method: 'PUT', body: { songIds } });
}

export function formatMusicTime(seconds: number) {
	if (!Number.isFinite(seconds) || seconds < 0) return '0:00';
	return `${Math.floor(seconds / 60)}:${String(Math.floor(seconds % 60)).padStart(2, '0')}`;
}
