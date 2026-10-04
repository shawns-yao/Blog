<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { authModalStore } from '$lib/shared/stores/authModalStore';
	import { checkMusicFavorites, getMusicAccess, setMusicFavorite } from '../api';
	import { getMusicContext } from '../context.svelte';
	import MusicPlayer from './MusicPlayer.svelte';
	import MusicQueue from './MusicQueue.svelte';

	const music = getMusicContext();
	const playbackState = music.state;
	const client = useQueryClient();
	let previousViewer = playbackState.viewer;
	const access = createQuery(() => ({
		queryKey: ['music', playbackState.viewer, 'access'],
		queryFn: ({ signal }) => getMusicAccess(signal),
		enabled: playbackState.authReady && playbackState.viewer > 0,
		retry: false,
		staleTime: 0
	}));
	const onMusicPage = $derived(/^\/music(?:\/|$)/.test(page.url.pathname));
	const current = $derived(music.current());
	let favoriteBusy = $state(false);
	const favorite = createQuery(() => ({
		queryKey: ['music', playbackState.viewer, 'personal', 'favorite-state', current?.id],
		queryFn: ({ signal }) => checkMusicFavorites(current ? [current.id] : [], signal),
		enabled: playbackState.authReady && playbackState.viewer > 0 && !!current,
		retry: false
	}));
	const isFavorite = $derived(!!current && !!favorite.data?.ids.includes(current.id));
	$effect(() => {
		const notice = playbackState.notice;
		if (!notice || !/^(已收藏|已取消收藏|已加入队列|这首歌曲已在队列中)/.test(notice)) return;
		const timeout = setTimeout(() => {
			if (playbackState.notice === notice) playbackState.notice = '';
		}, 3000);
		return () => clearTimeout(timeout);
	});
	async function toggleFavorite() {
		if (!current) return;
		if (!playbackState.viewer) {
			authModalStore.open('music');
			return;
		}
		if (favorite.isError) {
			await favorite.refetch();
			return;
		}
		if (favoriteBusy || favorite.isPending) return;
		const viewer = playbackState.viewer;
		const song = current;
		const next = !isFavorite;
		favoriteBusy = true;
		try {
			await setMusicFavorite(song.id, next);
			await client.invalidateQueries({ queryKey: ['music', viewer, 'personal'] });
			if (playbackState.viewer === viewer)
				playbackState.notice = `${next ? '已收藏' : '已取消收藏'}：${song.title}`;
		} catch (cause) {
			if (playbackState.viewer === viewer)
				playbackState.notice = cause instanceof Error ? cause.message : '收藏操作失败，请重试';
		} finally {
			favoriteBusy = false;
		}
	}
	$effect(() => {
		if (previousViewer !== playbackState.viewer) {
			void client.cancelQueries({ queryKey: ['music', previousViewer] });
			client.removeQueries({ queryKey: ['music', previousViewer] });
			previousViewer = playbackState.viewer;
		}
	});
	$effect(() => {
		playbackState.privateAccess =
			playbackState.viewer > 0 && !!access.data?.privateAccess && !access.isError;
		playbackState.accessPending = playbackState.viewer > 0 && access.isPending;
		playbackState.accessError = playbackState.viewer > 0 && access.isError;
		if (current && !current.public && !playbackState.privateAccess && !playbackState.accessPending)
			playbackState.sourceReady = false;
	});
</script>

<div
	class="music-dock"
	class:home-player={page.url.pathname === '/'}
	hidden={!onMusicPage && !playbackState.queue.length}
>
	<MusicPlayer
		song={current}
		playRequest={playbackState.playRequest}
		publicAccess={!!current?.public}
		viewer={Math.max(0, playbackState.viewer)}
		sourceReady={playbackState.sourceReady}
		autoplay={playbackState.autoplay}
		beforePlay={music.beforePlay}
		onEnded={() => music.advance(true)}
		mode={playbackState.mode}
		cycleMode={music.cycleMode}
		bind:volume={playbackState.volume}
		previous={music.previous}
		next={() => music.advance()}
		canPrevious={!!current}
		canNext={music.canNext()}
		queueCount={playbackState.queue.length}
		queueOpen={playbackState.queueOpen}
		toggleQueue={() => (playbackState.queueOpen = !playbackState.queueOpen)}
		notice={playbackState.preparing ? '正在准备播放…' : playbackState.notice}
		{isFavorite}
		favoritePending={favoriteBusy || (playbackState.viewer > 0 && favorite.isPending)}
		favoriteError={playbackState.viewer > 0 && favorite.isError}
		{toggleFavorite}
	/>
	<MusicQueue
		bind:open={playbackState.queueOpen}
		songs={playbackState.queue}
		currentIndex={playbackState.queueIndex}
		select={music.chooseQueue}
		remove={music.remove}
		clear={music.clear}
	/>
</div>

<style lang="postcss">
	@reference '../../../../routes/layout.css';
	.music-dock {
		--music-accent: #f43f5e;
		--music-accent-strong: #e11d48;
		--music-border: #ece8e5;
		--music-ink: #292524;
		--music-muted: #78716c;
		color-scheme: light;
		@apply font-sans;
	}
	.music-dock :global(.music-primary) {
		background: var(--music-accent) !important;
		color: #fff !important;
		box-shadow: none;
	}
	.music-dock :global(.music-primary:hover) {
		background: var(--music-accent-strong) !important;
	}
	.music-dock :global(.music-icon) {
		color: #57534e !important;
		background: #fff !important;
		border-color: var(--music-border) !important;
		box-shadow: none;
	}
	.music-dock :global(.music-icon:hover),
	.music-dock :global(.player-mode[data-active='true']) {
		background: #fff1f2 !important;
		color: var(--music-accent-strong) !important;
	}
	.music-dock :global(button:focus-visible),
	.music-dock :global(input:focus-visible) {
		outline: 2px solid var(--music-accent);
		outline-offset: 3px;
	}
	.music-dock :global(input[type='range']) {
		accent-color: var(--music-accent);
	}
	.home-player {
		--music-ink: #fff;
		--music-muted: #f5f5f4;
	}
	.home-player :global(.music-player) {
		border-top: 0;
		background: transparent;
		text-shadow: 0 1px 3px rgb(0 0 0 / 65%);
	}
	.home-player :global(.music-icon) {
		background: transparent !important;
		color: #fff !important;
	}
	.home-player :global(.music-icon:hover) {
		background: rgb(255 255 255 / 14%) !important;
	}
	.home-player :global(.player-mode[data-active='true']) {
		background: transparent !important;
		color: var(--music-accent) !important;
	}
	.home-player :global(.player-cover) {
		background: rgb(255 255 255 / 14%);
		color: #fff;
	}
	.home-player :global(.player-note) {
		border-top: 0;
		background: rgb(28 25 23 / 85%);
		color: #fff;
	}
	.music-dock :global(.player-favorite),
	.home-player :global(.player-favorite) {
		color: var(--music-accent) !important;
	}
</style>
