import { describe, expect, it } from 'vitest';
import { canPlayMusic, playableMusic } from './playback';
import type { MusicSong } from './types';

const song = (id: string, isPublic: boolean, unavailable = false) =>
	({ id, public: isPublic, unavailable }) as MusicSong;
describe('music playback permissions', () => {
	it('keeps public songs playable without private access', () => {
		expect(canPlayMusic(song('public', true), false)).toBe(true);
		expect(canPlayMusic(song('private', false), false)).toBe(false);
	});
	it('requires private access and rejects missing tracks', () => {
		expect(canPlayMusic(song('private', false), true)).toBe(true);
		expect(canPlayMusic(song('missing', true, true), true)).toBe(false);
	});
	it('filters inaccessible queue entries without changing order', () => {
		expect(
			playableMusic([song('a', true), song('b', false), song('c', true)], false).map((s) => s.id)
		).toEqual(['a', 'c']);
	});
});
