<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import Ellipsis from 'lucide-svelte/icons/ellipsis';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import {
		MUSIC_PAGE_SIZE,
		checkMusicFavorites,
		getMusicPlaylists,
		setMusicFavorite,
		setMusicPlaylistSong
	} from '../api';
	import type { MusicSong, MusicTrackProps } from '../types';
	import SongList from './SongList.svelte';
	import MusicDialog from './MusicDialog.svelte';
	let {
		songs,
		viewer,
		privateAccess,
		currentId,
		density,
		showCovers,
		play,
		enqueue,
		login,
		startIndex = 0,
		removeSong,
		moveSong,
		hasPrevious = false,
		hasNext = false
	}: MusicTrackProps & {
		songs: MusicSong[];
		startIndex?: number;
		hasPrevious?: boolean;
		hasNext?: boolean;
		removeSong?: (song: MusicSong) => Promise<unknown>;
		moveSong?: (song: MusicSong, direction: number) => Promise<unknown>;
	} = $props();
	const client = useQueryClient();
	let selected = $state<MusicSong | null>(null);
	let open = $state(false);
	let busy = $state(false);
	let error = $state('');
	let notice = $state('');
	let choosingPlaylist = $state(false);
	let listOffset = $state(0);
	const favorite = createQuery(() => ({
		queryKey: ['music', viewer, 'personal', 'favorite-state', selected?.id],
		queryFn: ({ signal }) => checkMusicFavorites(selected ? [selected.id] : [], signal),
		enabled: !!viewer && !!selected && open,
		retry: false
	}));
	const lists = createQuery(() => ({
		queryKey: ['music', viewer, 'personal', 'playlists', listOffset],
		queryFn: ({ signal }) => getMusicPlaylists(listOffset, signal),
		enabled: !!viewer && open && choosingPlaylist,
		retry: false
	}));
	const isFavorite = $derived(!!selected && !!favorite.data?.ids.includes(selected.id));
	async function perform(action: () => Promise<unknown>, message: string) {
		if (busy) return;
		busy = true;
		error = '';
		notice = '';
		try {
			await action();
			await client.invalidateQueries({ queryKey: ['music', viewer, 'personal'] });
			notice = message;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : '操作失败，请重试';
		} finally {
			busy = false;
		}
	}
	function openActions(song: MusicSong) {
		selected = song;
		open = true;
		error = '';
		notice = '';
		choosingPlaylist = false;
		listOffset = 0;
	}
</script>

<SongList
	{songs}
	{currentId}
	{density}
	{showCovers}
	{startIndex}
	publicAccess={true}
	play={(song) => play(song, songs)}
	{enqueue}
	statusLabel={(song) =>
		song.unavailable
			? '歌曲已失效'
			: song.public
				? '公开试听'
				: privateAccess
					? '私有歌曲'
					: viewer
						? '暂无播放权限'
						: '登录后授权播放'}
>
	{#snippet actions(song)}<Button
			variant="icon"
			class="music-icon"
			aria-label={`${song.title} 的操作`}
			onclick={() => openActions(song)}><Ellipsis size={18} /></Button
		>{/snippet}
</SongList>

<MusicDialog bind:open title={selected?.title || '歌曲操作'}>
	{#if selected}
		{@const song = selected}
		<div class="flex flex-col gap-3">
			<Button
				variant="secondary"
				onclick={() => {
					enqueue(song);
					open = false;
				}}>加入播放队列</Button
			>
			{#if !viewer}
				<Button
					onclick={() => {
						open = false;
						login();
					}}>登录后收藏或加入歌单</Button
				>
			{:else}
				{#if favorite.isError}<p role="alert" class="text-sm text-red-600">
						{favorite.error.message}
					</p>
					<Button variant="secondary" onclick={() => favorite.refetch()}>重新加载收藏状态</Button
					>{/if}
				<Button
					variant="secondary"
					disabled={busy ||
						favorite.isPending ||
						favorite.isError ||
						(song.unavailable && !isFavorite)}
					onclick={() =>
						perform(
							() => setMusicFavorite(song.id, !isFavorite),
							isFavorite ? '已取消收藏' : '已收藏'
						)}
				>
					{isFavorite ? '取消收藏' : '收藏歌曲'}
				</Button>
				<Button
					variant="secondary"
					disabled={busy || song.unavailable}
					onclick={() => (choosingPlaylist = !choosingPlaylist)}>加入歌单</Button
				>
				{#if removeSong}<Button
						variant="secondary"
						disabled={busy}
						onclick={() => perform(() => removeSong!(song), '已从歌单移除')}>从歌单移除</Button
					>{/if}
				{#if moveSong}
					<div class="flex gap-2">
						<Button
							variant="secondary"
							disabled={busy ||
								(!hasPrevious && songs.findIndex((item) => item.id === song.id) <= 0)}
							onclick={() => perform(() => moveSong!(song, -1), '顺序已更新')}>上移</Button
						><Button
							variant="secondary"
							disabled={busy ||
								(!hasNext && songs.findIndex((item) => item.id === song.id) >= songs.length - 1)}
							onclick={() => perform(() => moveSong!(song, 1), '顺序已更新')}>下移</Button
						>
					</div>
				{/if}
				{#if choosingPlaylist}
					{#if lists.isPending}<p role="status">正在加载歌单…</p>
					{:else if lists.isError}<p role="alert" class="text-sm text-red-600">
							{lists.error.message}
						</p>
						<Button variant="secondary" onclick={() => lists.refetch()}>重试</Button>
					{:else}
						{#each lists.data?.items ?? [] as list (list.id)}<Button
								variant="secondary"
								disabled={busy}
								onclick={() =>
									perform(() => setMusicPlaylistSong(list.id, song.id, true), '已加入歌单')}
								><span class="truncate">{list.name}</span></Button
							>{:else}<p class="text-sm text-stone-500">
								还没有歌单，请先在“我的歌单”创建。
							</p>{/each}
						<div class="flex gap-2">
							<Button
								variant="secondary"
								disabled={listOffset === 0 || busy}
								onclick={() => (listOffset -= MUSIC_PAGE_SIZE)}>上一页</Button
							><Button
								variant="secondary"
								disabled={!lists.data?.hasMore || busy}
								onclick={() => (listOffset += MUSIC_PAGE_SIZE)}>下一页</Button
							>
						</div>
					{/if}
				{/if}
			{/if}
			{#if error}<p role="alert" class="text-sm text-red-600">{error}</p>{/if}
			{#if notice}<p role="status" class="text-sm text-stone-600">{notice}</p>{/if}
		</div>
	{/if}
</MusicDialog>
