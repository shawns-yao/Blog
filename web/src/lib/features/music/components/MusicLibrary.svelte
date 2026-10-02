<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import Search from 'lucide-svelte/icons/search';
	import Play from 'lucide-svelte/icons/play';
	import RotateCw from 'lucide-svelte/icons/rotate-cw';
	import Music2 from 'lucide-svelte/icons/music-2';
	import ArrowLeft from 'lucide-svelte/icons/arrow-left';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { userStore } from '$lib/shared/stores/userStore';
	import { authModalStore } from '$lib/shared/stores/authModalStore';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { clearMusicSession, getMusicAlbum, getMusicCatalog, getMusicStatus } from '../api';
	import type { MusicDensity, MusicSong, MusicView } from '../types';
	import MusicShell from './MusicShell.svelte';
	import MusicSettings from './MusicSettings.svelte';
	import MusicPlannedView from './MusicPlannedView.svelte';
	import MusicAlbums from './MusicAlbums.svelte';
	import MusicQueue from './MusicQueue.svelte';
	import SongList from './SongList.svelte';
	import MusicPlayer from './MusicPlayer.svelte';

	const client = useQueryClient();
	let view = $state<MusicView>('search');
	let contentKind = $state<'songs' | 'albums'>('songs');
	let density = $state<MusicDensity>('comfortable');
	let showCovers = $state(true);
	let volume = $state(1);
	let search = $state('');
	let query = $state('');
	let offset = $state(0);
	let albumId = $state('');
	let queue = $state<MusicSong[]>([]);
	let queueIndex = $state(-1);
	let queueOpen = $state(false);
	let playRequest = $state(0);
	let notice = $state('');
	const titles = {
		search: '搜索音乐',
		playlists: '歌单',
		charts: '排行榜',
		favorites: '我的收藏',
		settings: '设置'
	};
	const viewer = $derived($userStore.isLogin ? ($userStore.userInfo?.id ?? 0) : 0);
	const publicAccess = $derived(!viewer);
	const status = createQuery(() => ({
		queryKey: ['music-status'],
		queryFn: ({ signal }) => getMusicStatus(signal),
		retry: false,
		staleTime: 30_000
	}));
	const catalog = createQuery(() => ({
		queryKey: ['music', viewer, 'catalog', query, offset],
		queryFn: ({ signal }) => getMusicCatalog(query, offset, signal, publicAccess),
		enabled: !!status.data?.available,
		retry: false,
		staleTime: 30_000
	}));
	const album = createQuery(() => ({
		queryKey: ['music', viewer, 'album', albumId],
		queryFn: ({ signal }) => getMusicAlbum(albumId, signal, publicAccess),
		enabled: !!albumId && !!status.data?.available,
		retry: false
	}));
	const songs = $derived(albumId ? (album.data?.song ?? []) : (catalog.data?.songs ?? []));
	const current = $derived(queue[queueIndex] ?? null);
	let previousViewer = $state(0);
	$effect(() => {
		if (previousViewer === viewer) return;
		queue = [];
		queueIndex = -1;
		queueOpen = false;
		albumId = '';
		offset = 0;
		search = '';
		query = '';
		notice = '';
		if (previousViewer) {
			void client.cancelQueries({ queryKey: ['music', previousViewer] });
			client.removeQueries({ queryKey: ['music', previousViewer] });
			if (!viewer) void clearMusicSession().catch(() => {});
		}
		previousViewer = viewer;
	});
	function play(song: MusicSong) {
		queue = [...songs];
		queueIndex = queue.findIndex((item) => item.id === song.id);
		playRequest++;
		notice = '';
	}
	function chooseQueue(index: number) {
		queueIndex = index;
		playRequest++;
	}
	function enqueue(song: MusicSong) {
		if (queue.some((item) => item.id === song.id)) {
			notice = '这首歌曲已在队列中';
			return;
		}
		queue = [...queue, song];
		notice = `已加入队列：${song.title}`;
	}
	function remove(index: number) {
		queue = queue.filter((_, i) => i !== index);
		if (index < queueIndex) queueIndex--;
		else if (index === queueIndex && queueIndex >= queue.length) queueIndex = queue.length - 1;
	}
	function submitSearch(event: SubmitEvent) {
		event.preventDefault();
		query = search.trim();
		offset = 0;
		albumId = '';
	}
	function openAccount() {
		if (viewer) windowStore.open('我的账号', null, 'user-center');
		else authModalStore.open('music');
	}
</script>

<MusicShell
	{view}
	title={titles[view]}
	loggedIn={!!viewer}
	selectView={(next) => (view = next)}
	{openAccount}
>
	{#if view === 'settings'}
		<MusicSettings bind:density bind:showCovers bind:volume />
	{:else if view !== 'search'}
		<MusicPlannedView feature={view} />
	{:else if status.isPending}
		<p role="status" class="music-state">正在打开音乐室…</p>
	{:else if status.isError || !status.data?.available}
		<div class="music-state">
			<Music2 size={34} />
			<h2>暂时无法连接音乐服务</h2>
			<p role="status">{status.data?.message || '请稍后重新加载'}</p>
			<Button class="music-primary mt-6" onclick={() => status.refetch()}>重新加载</Button>
		</div>
	{:else}
		<div class="music-toolbar">
			<form onsubmit={submitSearch} class="music-search">
				<Search size={18} /><label for="music-search" class="sr-only">搜索歌曲、艺术家或专辑</label
				><input
					id="music-search"
					type="search"
					bind:value={search}
					placeholder="搜索歌曲、艺术家或专辑"
					maxlength="100"
				/><Button variant="icon" class="music-icon" type="submit" aria-label="搜索"
					><Search size={17} /></Button
				>
			</form>
			<Button
				variant="icon"
				class="music-icon"
				aria-label="刷新曲库"
				disabled={albumId ? album.isFetching : catalog.isFetching}
				onclick={() => (albumId ? album.refetch() : catalog.refetch())}
				><RotateCw size={17} /></Button
			>
			{#if contentKind === 'songs'}<Button
					class="music-primary play-all"
					disabled={!songs.length || (albumId ? album.isPending : catalog.isPending)}
					onclick={() => {
						if (songs[0]) play(songs[0]);
					}}><Play size={15} /><span>播放全部</span></Button
				>{/if}
		</div>
		<div class="music-library-heading">
			{#if albumId}<div class="album-heading">
					<Button
						variant="icon"
						class="music-icon"
						aria-label="返回专辑列表"
						onclick={() => {
							albumId = '';
							contentKind = 'albums';
						}}><ArrowLeft size={17} /></Button
					>
					<h2>{album.data?.name || '专辑'}</h2>
				</div>
			{:else}<div class="music-tabs" role="group" aria-label="曲库视图">
					<button
						type="button"
						class:active={contentKind === 'songs'}
						aria-pressed={contentKind === 'songs'}
						onclick={() => (contentKind = 'songs')}>全部歌曲</button
					><button
						type="button"
						class:active={contentKind === 'albums'}
						aria-pressed={contentKind === 'albums'}
						onclick={() => (contentKind = 'albums')}>专辑</button
					>
				</div>{/if}
			<span class="music-count"
				>{query
					? `“${query}”的搜索结果`
					: contentKind === 'songs'
						? publicAccess
							? '公开试听'
							: '本站曲库'
						: '专辑列表'}</span
			>
		</div>
		{#if albumId ? album.isPending : catalog.isPending}<p role="status" class="music-loading">
				正在加载音乐…
			</p>
		{:else if albumId ? album.isError : catalog.isError}<p role="alert" class="music-error">
				{(albumId ? album.error : catalog.error)?.message || '曲库加载失败'}
			</p>
			<Button
				variant="secondary"
				class="music-secondary"
				onclick={() => (albumId ? album.refetch() : catalog.refetch())}>重新加载</Button
			>
		{:else if contentKind === 'albums' && !albumId}
			{#if catalog.data?.albums.length}<MusicAlbums
					albums={catalog.data.albums}
					{publicAccess}
					select={(id) => {
						albumId = id;
						contentKind = 'songs';
					}}
				/>{:else}<p class="music-empty">{query ? '没有找到匹配的专辑' : '曲库还没有专辑'}</p>{/if}
		{:else if songs.length}<SongList
				{songs}
				currentId={current?.id}
				{play}
				{enqueue}
				{density}
				{showCovers}
				{publicAccess}
				startIndex={albumId ? 0 : offset}
			/>
		{:else}<p class="music-empty">
				{query ? '没有找到匹配的音乐' : publicAccess ? '暂时没有公开试听的歌曲' : '曲库还没有歌曲'}
			</p>{/if}
		{#if !albumId && contentKind === 'songs' && (offset > 0 || catalog.data?.hasMore)}<div
				class="music-pagination"
			>
				<Button
					variant="secondary"
					class="music-secondary"
					size="sm"
					disabled={offset === 0 || catalog.isFetching}
					onclick={() => (offset = Math.max(0, offset - 30))}>上一页</Button
				><span class="font-mono text-xs">第 {Math.floor(offset / 30) + 1} 页</span><Button
					variant="secondary"
					class="music-secondary"
					size="sm"
					disabled={!catalog.data?.hasMore || catalog.isFetching}
					onclick={() => (offset += 30)}>下一页</Button
				>
			</div>{/if}
	{/if}
	<p role="status" aria-live="polite" class="music-notice">{notice}</p>
	{#snippet player()}
		<MusicPlayer
			song={current}
			{playRequest}
			{publicAccess}
			{viewer}
			bind:volume
			previous={() => chooseQueue(Math.max(0, queueIndex - 1))}
			next={() => chooseQueue(Math.min(queue.length - 1, queueIndex + 1))}
			canPrevious={queueIndex >= 0}
			canNext={queueIndex >= 0 && queueIndex < queue.length - 1}
			queueCount={queue.length}
			{queueOpen}
			toggleQueue={() => (queueOpen = !queueOpen)}
		/>
		<MusicQueue
			bind:open={queueOpen}
			songs={queue}
			currentIndex={queueIndex}
			select={chooseQueue}
			{remove}
			clear={() => {
				queue = [];
				queueIndex = -1;
			}}
		/>
	{/snippet}
</MusicShell>

<style lang="postcss">
	@reference '../../../../routes/layout.css';
	.music-toolbar {
		display: flex;
		align-items: center;
		gap: 14px;
		margin-bottom: 26px;
	}
	.music-search {
		display: flex;
		align-items: center;
		gap: 11px;
		flex: 1;
		min-width: 0;
		min-height: 46px;
		padding: 4px 7px 4px 15px;
		border: 1px solid var(--music-border);
		border-radius: 6px;
		color: var(--music-muted);
		background: #fff;
	}
	.music-search:focus-within {
		border-color: var(--music-accent);
	}
	input {
		min-width: 0;
		width: 100%;
		flex: 1;
		padding: 8px 0;
		background: transparent;
		color: var(--music-ink);
		font-size: 13px;
	}
	input::placeholder {
		color: #a8a29e;
	}
	.music-library-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		margin-bottom: 18px;
	}
	.music-tabs {
		display: flex;
		gap: 24px;
	}
	.music-tabs button {
		min-height: 40px;
		padding: 4px 0;
		border-bottom: 2px solid transparent;
		color: var(--music-muted);
		font-size: 14px;
	}
	.music-tabs button.active {
		border-bottom-color: var(--music-accent);
		color: var(--music-accent-strong);
	}
	.music-count {
		color: var(--music-muted);
		font-size: 12px;
	}
	.album-heading {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	h2 {
		font-size: 16px;
		font-weight: 500;
	}
	.music-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 360px;
		padding: 36px 16px;
		text-align: center;
		color: var(--music-muted);
		font-size: 13px;
	}
	.music-state h2 {
		margin-top: 20px;
		color: var(--music-ink);
		font-size: 19px;
	}
	.music-state p {
		margin-top: 12px;
	}
	.music-loading,
	.music-empty {
		padding: 64px 0;
		color: var(--music-muted);
		text-align: center;
		font-size: 13px;
	}
	.music-error {
		padding: 28px 0;
		color: var(--music-accent-strong);
		font-size: 13px;
	}
	.music-pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 18px;
		padding: 24px 0;
		color: var(--music-muted);
	}
	.music-notice {
		min-height: 20px;
		margin-top: 14px;
		color: var(--music-accent-strong);
		font-size: 12px;
	}
	@media (max-width: 640px) {
		.music-toolbar {
			gap: 8px;
			margin-bottom: 18px;
		}
		.music-search {
			padding-left: 11px;
			gap: 8px;
		}
		.music-search > :global(svg) {
			display: none;
		}
		input {
			font-size: 16px;
		}
		:global(.play-all) {
			padding: 9px;
			min-width: 38px;
		}
		:global(.play-all span) {
			display: none;
		}
		.music-count {
			font-size: 11px;
		}
		.music-tabs {
			gap: 18px;
		}
	}
</style>
