import { getContext, setContext } from 'svelte';
import { authModalStore } from '$lib/shared/stores/authModalStore';
import { clearMusicSession, createMusicSession } from './api';
import { canPlayMusic, nextMusicIndex, playableMusic } from './playback';
import { loadMusicPreferences, loadMusicQueue, musicModes } from './persistence';
import type { MusicDensity, MusicPlaybackMode, MusicSong } from './types';

const contextKey = Symbol('music-playback');

export function createMusicContext() {
	const state = $state({
		authReady: false,
		initialized: false,
		viewer: -1,
		privateAccess: false,
		accessPending: false,
		accessError: false,
		density: 'comfortable' as MusicDensity,
		showCovers: true,
		volume: 1,
		mode: 'sequence' as MusicPlaybackMode,
		queue: [] as MusicSong[],
		queueIndex: -1,
		queueOpen: false,
		playRequest: 0,
		sourceReady: false,
		autoplay: false,
		preparing: false,
		notice: ''
	});
	let generation = 0;
	let sessionController: AbortController | undefined;
	let sessionBarrier: Promise<unknown> = Promise.resolve();
	let history: string[] = [];
	const current = () => state.queue[state.queueIndex] ?? null;

	function cancel() {
		generation++;
		sessionController?.abort();
		state.preparing = false;
	}
	function clear() {
		cancel();
		state.sourceReady = false;
		state.autoplay = false;
		state.queue = [];
		state.queueIndex = -1;
		state.queueOpen = false;
		history = [];
	}
	function initialize() {
		Object.assign(state, loadMusicPreferences());
		state.initialized = true;
	}
	function setViewer(viewer: number) {
		if (state.viewer === viewer) return;
		const initial = state.viewer === -1;
		clear();
		state.viewer = viewer;
		state.privateAccess = false;
		state.accessError = false;
		state.accessPending = !!viewer;
		state.notice = '';
		if (initial) {
			const saved = loadMusicQueue(viewer);
			state.queue = saved.songs;
			state.queueIndex = saved.index;
			state.sourceReady = !!current()?.public;
		} else {
			const token = generation;
			// Serialize account cleanup before a new private playback session.
			sessionBarrier = sessionBarrier
				.then(() => clearMusicSession())
				.catch(() => {
					if (token === generation) state.notice = '音乐会话清理失败，请重试或刷新页面';
				});
		}
	}
	function allowed(song: MusicSong) {
		if (!state.authReady || state.viewer < 0) {
			state.notice = '正在确认登录状态，请稍后重试';
			return false;
		}
		if (canPlayMusic(song, state.privateAccess)) return true;
		if (song.unavailable) state.notice = '歌曲已不在曲库中';
		else if (!state.viewer) authModalStore.open('music');
		else
			state.notice = state.accessPending
				? '正在确认播放权限，请稍后重试'
				: state.accessError
					? '播放权限查询失败，请重新加载权限'
					: '当前账号没有私有音乐播放权限';
		return false;
	}
	async function prepare(song: MusicSong, apply: () => void) {
		const token = ++generation;
		sessionController?.abort();
		const controller = new AbortController();
		sessionController = controller;
		state.preparing = true;
		state.notice = '';
		try {
			if (!song.public) {
				await sessionBarrier;
				if (token !== generation) return false;
				await createMusicSession(controller.signal);
			}
			if (token !== generation) return false;
			apply();
			return true;
		} catch (cause) {
			if (token === generation)
				state.notice = cause instanceof Error ? cause.message : '播放会话创建失败';
			return false;
		} finally {
			if (token === generation) state.preparing = false;
		}
	}
	function play(song: MusicSong, songs: MusicSong[]) {
		if (!allowed(song)) return;
		const next = playableMusic(songs, state.privateAccess);
		const index = next.findIndex((item) => item.id === song.id);
		if (index < 0) return;
		void prepare(song, () => {
			state.queue = next;
			state.queueIndex = index;
			state.sourceReady = true;
			state.autoplay = true;
			state.playRequest++;
			history = [];
		});
	}
	function chooseQueue(index: number, remember = true) {
		const song = state.queue[index];
		if (!song || !allowed(song)) return;
		void prepare(song, () => {
			if (remember && current() && index !== state.queueIndex) history.push(current()!.id);
			state.queueIndex = index;
			state.sourceReady = true;
			state.autoplay = true;
			state.playRequest++;
		});
	}
	async function beforePlay() {
		const song = current();
		if (!song || !allowed(song)) return false;
		return prepare(song, () => {
			state.sourceReady = true;
		});
	}
	function target(direction: 1 | -1, ended = false) {
		return nextMusicIndex(
			state.queue,
			state.queueIndex,
			state.privateAccess,
			state.mode,
			direction,
			ended
		);
	}
	function advance(ended = false) {
		const index = target(1, ended);
		if (index >= 0) chooseQueue(index);
	}
	function previous() {
		if (state.mode === 'shuffle') {
			while (history.length) {
				const id = history.pop();
				const index = state.queue.findIndex(
					(song) => song.id === id && canPlayMusic(song, state.privateAccess)
				);
				if (index >= 0) {
					chooseQueue(index, false);
					return;
				}
			}
		}
		const index = target(-1);
		if (index >= 0) chooseQueue(index, false);
		else if (current()) chooseQueue(state.queueIndex, false);
	}
	function enqueue(song: MusicSong) {
		if (!allowed(song)) return;
		if (state.queue.some((item) => item.id === song.id)) {
			state.notice = '这首歌曲已在队列中';
			return;
		}
		if (state.queue.length >= 500) {
			state.notice = '播放队列最多支持 500 首歌曲';
			return;
		}
		state.queue = [...state.queue, song];
		state.notice = `已加入队列：${song.title}`;
	}
	function remove(index: number) {
		cancel();
		const wasCurrent = index === state.queueIndex;
		state.queue = state.queue.filter((_, i) => i !== index);
		if (index < state.queueIndex) state.queueIndex--;
		else if (wasCurrent) {
			state.queueIndex = -1;
			state.sourceReady = false;
			if (state.queue.length) chooseQueue(Math.min(index, state.queue.length - 1));
		}
	}
	function cycleMode() {
		state.mode = musicModes[(musicModes.indexOf(state.mode) + 1) % musicModes.length];
		history = [];
	}
	const context = {
		state,
		current,
		initialize,
		setViewer,
		play,
		chooseQueue,
		beforePlay,
		advance,
		previous,
		enqueue,
		remove,
		clear,
		cycleMode,
		canNext: () => target(1) >= 0,
		destroy: cancel
	};
	setContext(contextKey, context);
	return context;
}

export function getMusicContext() {
	return getContext<ReturnType<typeof createMusicContext>>(contextKey);
}
