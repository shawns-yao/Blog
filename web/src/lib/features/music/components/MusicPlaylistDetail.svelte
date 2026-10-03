<script lang="ts">
	import { onDestroy } from 'svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import {
		getMusicPlaylist,
		reorderMusicPlaylist,
		setMusicPlaylistSong,
		MUSIC_PAGE_SIZE
	} from '../api';
	import { playableMusic } from '../playback';
	import type { MusicSong, MusicTrackProps } from '../types';
	import MusicTracks from './MusicTracks.svelte';
	let { id, back, ...tracks } = $props<MusicTrackProps & { id: number; back: () => void }>();
	let offset = $state(0);
	let loadingQueue = $state(false);
	let error = $state('');
	let controller: AbortController | undefined;
	onDestroy(() => controller?.abort());
	const detail = createQuery(() => ({
		queryKey: ['music', tracks.viewer, 'personal', 'playlist', id, offset],
		queryFn: ({ signal }) => getMusicPlaylist(id, offset, signal),
		retry: false
	}));
	const songs = $derived(detail.data?.songs ?? []);
	async function move(song: MusicSong, direction: number) {
		const index = songs.findIndex((item) => item.id === song.id);
		if (index < 0) throw new Error('歌单内容已变化，请刷新后重试');
		let neighbor = songs[index + direction];
		if (!neighbor && offset + index + direction >= 0) {
			const page = await getMusicPlaylist(id, offset + index + direction);
			neighbor = page.songs[0];
		}
		if (!neighbor) throw new Error('歌单内容已变化，请刷新后重试');
		return reorderMusicPlaylist(
			id,
			direction < 0 ? [song.id, neighbor.id] : [neighbor.id, song.id]
		);
	}
	async function playPlaylist() {
		if (loadingQueue) return;
		loadingQueue = true;
		error = '';
		controller = new AbortController();
		try {
			const all: MusicSong[] = [];
			for (let start = 0; start < 500; start += MUSIC_PAGE_SIZE) {
				const page = await getMusicPlaylist(id, start, controller.signal);
				all.push(...page.songs);
				if (!page.hasMore) break;
			}
			if (controller.signal.aborted) return;
			const available = playableMusic(
				[...new Map(all.map((song) => [song.id, song])).values()],
				tracks.privateAccess
			);
			if (!available.length) {
				error = '歌单中暂时没有可播放的歌曲';
				return;
			}
			tracks.play(available[0], available);
		} catch (cause) {
			if (!controller.signal.aborted)
				error = cause instanceof Error ? cause.message : '歌单加载失败';
		} finally {
			loadingQueue = false;
		}
	}
</script>

<section class="space-y-5">
	<Button variant="secondary" class="music-secondary" onclick={back}>返回我的歌单</Button>
	<div class="flex flex-wrap items-center justify-between gap-3">
		<div class="min-w-0">
			<h2 class="break-words text-xl font-medium">{detail.data?.playlist.name || '歌单'}</h2>
			<p class="mt-2 text-sm text-stone-500">{detail.data?.playlist.songCount ?? 0} 首歌曲</p>
		</div>
		<Button
			class="music-primary"
			disabled={loadingQueue || detail.isFetching || !detail.data?.playlist.songCount}
			onclick={playPlaylist}>{loadingQueue ? '正在准备…' : '播放歌单'}</Button
		>
	</div>
	{#if error}<p role="alert" class="text-sm text-rose-700">{error}</p>{/if}
	{#if detail.isFetching}<p role="status" class="py-12 text-center text-sm text-stone-500">
			正在加载歌曲…
		</p>
	{:else if detail.isError}<p role="alert" class="text-sm text-rose-700">{detail.error.message}</p>
		<Button variant="secondary" onclick={() => detail.refetch()}>重试</Button>
	{:else if songs.length}<MusicTracks
			{...tracks}
			{songs}
			startIndex={offset}
			hasPrevious={offset > 0}
			hasNext={!!detail.data?.hasMore}
			removeSong={(song) => setMusicPlaylistSong(id, song.id, false)}
			moveSong={move}
		/>
	{:else}<p class="py-12 text-center text-sm text-stone-500">歌单中还没有歌曲</p>{/if}
	{#if songs.length || offset > 0}<div class="flex items-center justify-center gap-4">
			<Button
				variant="secondary"
				disabled={offset === 0 || detail.isFetching}
				onclick={() => (offset -= MUSIC_PAGE_SIZE)}>上一页</Button
			><span class="text-xs">第 {offset / MUSIC_PAGE_SIZE + 1} 页</span><Button
				variant="secondary"
				disabled={!detail.data?.hasMore || detail.isFetching}
				onclick={() => (offset += MUSIC_PAGE_SIZE)}>下一页</Button
			>
		</div>{/if}
</section>
