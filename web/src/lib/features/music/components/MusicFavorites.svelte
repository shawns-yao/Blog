<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { getMusicFavorites, MUSIC_PAGE_SIZE } from '../api';
	import { playableMusic } from '../playback';
	import type { MusicTrackProps } from '../types';
	import MusicTracks from './MusicTracks.svelte';
	let tracks: MusicTrackProps = $props();
	let offset = $state(0);
	const favorites = createQuery(() => ({
		queryKey: ['music', tracks.viewer, 'personal', 'favorites', offset],
		queryFn: ({ signal }) => getMusicFavorites(offset, signal),
		enabled: !!tracks.viewer,
		retry: false
	}));
	const songs = $derived(favorites.data?.songs ?? []);
	const playable = $derived(playableMusic(songs, tracks.privateAccess));
</script>

{#if !tracks.viewer}<div class="space-y-4 py-12 text-center">
		<p class="text-sm text-stone-500">登录后查看自己的收藏</p>
		<Button class="music-primary" onclick={tracks.login}>登录</Button>
	</div>
{:else}
	<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
		<h2 class="text-xl font-medium">我的收藏</h2>
		<Button
			class="music-primary"
			disabled={!playable.length || favorites.isFetching}
			onclick={() => tracks.play(playable[0], songs)}>播放本页</Button
		>
	</div>
	{#if favorites.isFetching}<p role="status" class="py-12 text-center text-sm text-stone-500">
			正在加载收藏…
		</p>
	{:else if favorites.isError}<p role="alert" class="mb-4 text-sm text-rose-700">
			{favorites.error.message}
		</p>
		<Button variant="secondary" onclick={() => favorites.refetch()}>重试</Button>
	{:else if songs.length}<MusicTracks {...tracks} {songs} startIndex={offset} />
	{:else}<p class="py-12 text-center text-sm text-stone-500">还没有收藏的歌曲</p>{/if}
	{#if songs.length || offset > 0}<div class="mt-6 flex items-center justify-center gap-4">
			<Button
				variant="secondary"
				disabled={offset === 0 || favorites.isFetching}
				onclick={() => (offset -= MUSIC_PAGE_SIZE)}>上一页</Button
			><span class="text-xs">第 {offset / MUSIC_PAGE_SIZE + 1} 页</span><Button
				variant="secondary"
				disabled={!favorites.data?.hasMore || favorites.isFetching}
				onclick={() => (offset += MUSIC_PAGE_SIZE)}>下一页</Button
			>
		</div>{/if}
{/if}
