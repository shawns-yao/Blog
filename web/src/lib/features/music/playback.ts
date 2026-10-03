import type { MusicPlaybackMode, MusicSong } from './types';

export function canPlayMusic(song: MusicSong, privateAccess: boolean) {
	return !song.unavailable && (song.public || privateAccess);
}

export function playableMusic(songs: MusicSong[], privateAccess: boolean) {
	return songs.filter((song) => canPlayMusic(song, privateAccess));
}

export function nextMusicIndex(
	songs: MusicSong[],
	index: number,
	privateAccess: boolean,
	mode: MusicPlaybackMode,
	direction: 1 | -1,
	ended = false
) {
	const available = songs.flatMap((song, i) => (canPlayMusic(song, privateAccess) ? [i] : []));
	if (!available.length || index < 0) return -1;
	if (ended && mode === 'single' && available.includes(index)) return index;
	if (mode === 'shuffle') {
		const choices = available.filter((i) => i !== index);
		return choices.length ? choices[Math.floor(Math.random() * choices.length)] : available[0];
	}
	const next =
		direction === 1 ? available.find((i) => i > index) : available.findLast((i) => i < index);
	if (next !== undefined) return next;
	if (mode === 'repeat' || mode === 'single')
		return direction === 1 ? available[0] : available[available.length - 1];
	return -1;
}
