<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import Search from 'lucide-svelte/icons/search';
	import Play from 'lucide-svelte/icons/play';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { browseMusic, browseMusicAlbum, MUSIC_PAGE_SIZE } from '../api';
	import { playableMusic } from '../playback';
	import type { MusicTrackProps } from '../types';
	import MusicAlbums from './MusicAlbums.svelte';
	import MusicTracks from './MusicTracks.svelte';
	let { searchMode = false, ...tracks } = $props<MusicTrackProps & { searchMode?: boolean }>();
	let input = $state('');
	let query = $state('');
	let offset = $state(0);
	let albumId = $state('');
	const catalog = createQuery(() => ({
		queryKey: ['music', tracks.viewer, 'browse', query, offset],
		queryFn: ({ signal }) => browseMusic(query, offset, signal),
		retry: false
	}));
	const album = createQuery(() => ({
		queryKey: ['music', tracks.viewer, 'browse-album', albumId],
		queryFn: ({ signal }) => browseMusicAlbum(albumId, signal),
		enabled: !!albumId,
		retry: false
	}));
	const songs = $derived(albumId ? (album.data?.song ?? []) : (catalog.data?.songs ?? []));
	const pending = $derived(albumId ? album.isFetching : catalog.isFetching);
	const error = $derived(albumId ? album.error : catalog.error);
	const playable = $derived(playableMusic(songs, tracks.privateAccess));
	function search(event: SubmitEvent) {
		event.preventDefault();
		query = input.trim();
		offset = 0;
		albumId = '';
	}
</script>

<section class="space-y-7">
	{#if !searchMode && !albumId}<div class="flex flex-wrap items-end justify-between gap-4">
			<div>
				<p class="mb-2 text-xs font-medium tracking-widest text-rose-600">音乐室</p>
				<h2 class="text-2xl font-semibold">发现音乐</h2>
			</div>
			<span class="text-sm text-stone-500">本站曲库</span>
		</div>{/if}
	<form onsubmit={search} class="flex flex-wrap items-center gap-3">
		<label
			class="flex min-w-0 flex-1 items-center gap-2 rounded-lg border border-stone-200 bg-white px-3"
			><Search size={18} /><input
				class="min-w-0 flex-1 bg-transparent py-3 text-sm outline-none"
				bind:value={input}
				type="search"
				maxlength="100"
				placeholder="搜索歌曲、艺术家或专辑"
				aria-label="搜索音乐"
			/></label
		>
	</form>
	{#if albumId}<Button variant="secondary" class="music-secondary" onclick={() => (albumId = '')}
			>返回曲库</Button
		>{/if}
	{#if pending}<p role="status" class="py-12 text-center text-sm text-stone-500">正在加载曲库…</p>
	{:else if error}<div role="alert" class="py-10 text-sm text-rose-700">
			{error.message}<Button
				class="ml-3"
				variant="secondary"
				onclick={() => (albumId ? album.refetch() : catalog.refetch())}>重试</Button
			>
		</div>
	{:else}
		{#if !albumId && catalog.data?.albums.length}<section class="space-y-4">
				<h3 class="text-base font-medium">专辑</h3>
				<MusicAlbums
					albums={catalog.data.albums}
					publicAccess={true}
					select={(id) => (albumId = id)}
				/>
			</section>{/if}
		<section class="space-y-4">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<h3 class="text-base font-medium">
					{albumId ? album.data?.name : query ? '搜索结果' : '全部歌曲'}
				</h3>
				<Button
					class="music-primary"
					disabled={!playable.length}
					onclick={() => tracks.play(playable[0], songs)}><Play size={15} />播放本页</Button
				>
			</div>
			{#if songs.length}<MusicTracks
					{...tracks}
					{songs}
					startIndex={albumId ? 0 : offset}
				/>{:else}<p class="py-12 text-center text-sm text-stone-500">
					{query ? '没有找到匹配的歌曲' : '曲库还没有歌曲'}
				</p>{/if}
		</section>
		{#if !albumId && (songs.length || offset > 0)}<div
				class="flex items-center justify-center gap-4"
			>
				<Button
					variant="secondary"
					class="music-secondary"
					disabled={offset === 0}
					onclick={() => (offset -= MUSIC_PAGE_SIZE)}>上一页</Button
				><span class="text-xs text-stone-500">第 {offset / MUSIC_PAGE_SIZE + 1} 页</span><Button
					variant="secondary"
					class="music-secondary"
					disabled={!catalog.data?.hasMore}
					onclick={() => (offset += MUSIC_PAGE_SIZE)}>下一页</Button
				>
			</div>{/if}
	{/if}
</section>
