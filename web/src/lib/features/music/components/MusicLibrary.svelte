<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { userStore } from '$lib/shared/stores/userStore';
	import { authModalStore } from '$lib/shared/stores/authModalStore';
	import { getMusicAccess, getMusicStatus } from '../api';
	import { getMusicContext } from '../context.svelte';
	import type { MusicView } from '../types';
	import MusicShell from './MusicShell.svelte';
	import MusicSettings from './MusicSettings.svelte';
	import MusicHome from './MusicHome.svelte';
	import MusicFavorites from './MusicFavorites.svelte';
	import MusicPlaylists from './MusicPlaylists.svelte';

	const music = getMusicContext();
	const playbackState = music.state;
	let view = $state<MusicView>('home');
	const titles = {
		home: '音乐首页',
		search: '搜索音乐',
		playlists: '我的歌单',
		favorites: '我的收藏',
		settings: '设置'
	};
	const viewer = $derived($userStore.isLogin ? ($userStore.userInfo?.id ?? 0) : 0);
	const status = createQuery(() => ({
		queryKey: ['music-status'],
		queryFn: ({ signal }) => getMusicStatus(signal),
		retry: false,
		staleTime: 30_000
	}));
	const access = createQuery(() => ({
		queryKey: ['music', viewer, 'access'],
		queryFn: ({ signal }) => getMusicAccess(signal),
		enabled: !!viewer,
		retry: false,
		staleTime: 0
	}));
	const privateAccess = $derived(playbackState.privateAccess);
	const current = $derived(music.current());
	const trackProps = $derived({
		viewer,
		privateAccess,
		currentId: current?.id,
		density: playbackState.density,
		showCovers: playbackState.showCovers,
		play: music.play,
		enqueue: music.enqueue,
		login: () => authModalStore.open('music')
	});
	function openAccount() {
		if (viewer) view = 'settings';
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
	{#if access.isError && viewer}<div
			role="alert"
			class="mb-5 flex flex-wrap items-center gap-3 text-sm text-rose-700"
		>
			<p>播放权限查询失败，公开歌曲仍可试听。</p>
			<Button variant="secondary" onclick={() => access.refetch()}>重新加载权限</Button>
		</div>{/if}
	{#key viewer}
		{#if view === 'settings'}<MusicSettings
				loggedIn={!!viewer}
				bind:density={playbackState.density}
				bind:showCovers={playbackState.showCovers}
				bind:volume={playbackState.volume}
			/>
		{:else if view === 'favorites'}<MusicFavorites {...trackProps} />
		{:else if view === 'playlists'}<MusicPlaylists {...trackProps} />
		{:else if status.isPending}<p role="status" class="py-24 text-center text-sm text-stone-500">
				正在打开音乐室…
			</p>
		{:else if status.isError || !status.data?.available}<div class="space-y-4 py-24 text-center">
				<p role="alert" class="text-sm text-stone-500">
					{status.error?.message || status.data?.message || '暂时无法连接音乐服务'}
				</p>
				<Button class="music-primary" onclick={() => status.refetch()}>重新加载</Button>
			</div>
		{:else}{#key view}<MusicHome {...trackProps} searchMode={view === 'search'} />{/key}{/if}
	{/key}
	{#if playbackState.preparing || playbackState.notice}
		<p role="status" aria-live="polite" class="mt-4 text-sm text-rose-700">
			{playbackState.preparing ? '正在准备播放…' : playbackState.notice}
		</p>
	{/if}
</MusicShell>
