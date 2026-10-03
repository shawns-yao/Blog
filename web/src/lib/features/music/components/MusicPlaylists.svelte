<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import ListMusic from 'lucide-svelte/icons/list-music';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import {
		MUSIC_PAGE_SIZE,
		createMusicPlaylist,
		deleteMusicPlaylist,
		getMusicPlaylists,
		renameMusicPlaylist
	} from '../api';
	import type { MusicPlaylist, MusicTrackProps } from '../types';
	import MusicDialog from './MusicDialog.svelte';
	import MusicPlaylistDetail from './MusicPlaylistDetail.svelte';
	let tracks: MusicTrackProps = $props();
	const client = useQueryClient();
	let offset = $state(0);
	let activeId = $state(0);
	let open = $state(false);
	let action = $state<'create' | 'rename' | 'delete'>('create');
	let selected = $state<MusicPlaylist | null>(null);
	let name = $state('');
	let busy = $state(false);
	let error = $state('');
	const lists = createQuery(() => ({
		queryKey: ['music', tracks.viewer, 'personal', 'playlists', offset],
		queryFn: ({ signal }) => getMusicPlaylists(offset, signal),
		enabled: !!tracks.viewer,
		retry: false
	}));
	function edit(kind: typeof action, list: MusicPlaylist | null = null) {
		action = kind;
		selected = list;
		name = list?.name ?? '';
		error = '';
		open = true;
	}
	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (busy) return;
		busy = true;
		error = '';
		try {
			if (action === 'create') {
				await createMusicPlaylist(name.trim());
				offset = 0;
			} else if (selected && action === 'rename')
				await renameMusicPlaylist(selected.id, name.trim());
			else if (selected) await deleteMusicPlaylist(selected.id);
			await client.invalidateQueries({ queryKey: ['music', tracks.viewer, 'personal'] });
			open = false;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : '保存失败，请重试';
		} finally {
			busy = false;
		}
	}
</script>

{#if !tracks.viewer}<div class="space-y-4 py-12 text-center">
		<p class="text-sm text-stone-500">登录后管理自己的歌单</p>
		<Button class="music-primary" onclick={tracks.login}>登录</Button>
	</div>
{:else if activeId}
	{#key activeId}<MusicPlaylistDetail {...tracks} id={activeId} back={() => (activeId = 0)} />{/key}
{:else}
	<div class="mb-6 flex items-center justify-between gap-3">
		<h2 class="text-xl font-medium">我的歌单</h2>
		<Button class="music-primary" onclick={() => edit('create')}>新建歌单</Button>
	</div>
	{#if lists.isFetching}<p role="status" class="py-12 text-center text-sm text-stone-500">
			正在加载歌单…
		</p>
	{:else if lists.isError}<p role="alert" class="mb-4 text-sm text-rose-700">
			{lists.error.message}
		</p>
		<Button variant="secondary" onclick={() => lists.refetch()}>重试</Button>
	{:else}<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
			{#each lists.data?.items ?? [] as list (list.id)}<article
					class="min-w-0 rounded-xl border border-stone-200 bg-white p-5"
				>
					<button type="button" class="mb-4 w-full text-left" onclick={() => (activeId = list.id)}
						><div
							class="mb-4 flex aspect-[2/1] items-center justify-center rounded-lg bg-stone-100 text-stone-400"
						>
							<ListMusic size={36} />
						</div>
						<h3 class="truncate font-medium">{list.name}</h3>
						<p class="mt-1 text-sm text-stone-500">{list.songCount} 首歌曲</p></button
					>
					<div class="flex flex-wrap gap-2">
						<Button variant="secondary" onclick={() => edit('rename', list)}>重命名</Button><Button
							variant="ghost"
							onclick={() => edit('delete', list)}>删除</Button
						>
					</div>
				</article>{:else}<p class="py-12 text-sm text-stone-500">还没有歌单</p>{/each}
		</div>{/if}
	{#if lists.data?.items.length || offset > 0}<div
			class="mt-6 flex items-center justify-center gap-4"
		>
			<Button
				variant="secondary"
				disabled={offset === 0 || lists.isFetching}
				onclick={() => (offset -= MUSIC_PAGE_SIZE)}>上一页</Button
			><span class="text-xs">第 {offset / MUSIC_PAGE_SIZE + 1} 页</span><Button
				variant="secondary"
				disabled={!lists.data?.hasMore || lists.isFetching}
				onclick={() => (offset += MUSIC_PAGE_SIZE)}>下一页</Button
			>
		</div>{/if}
{/if}

<MusicDialog
	bind:open
	title={action === 'create' ? '新建歌单' : action === 'rename' ? '重命名歌单' : '删除歌单'}
>
	<form onsubmit={submit} class="space-y-4">
		{#if action === 'delete'}<p class="break-words text-sm text-stone-600">
				确认删除“{selected?.name}”？仅删除歌单，不删除音乐文件。
			</p>
		{:else}<label class="block text-sm"
				>歌单名称<input
					class="mt-2 w-full rounded-lg border border-stone-300 p-3"
					bind:value={name}
					maxlength="100"
					required
					disabled={busy}
				/></label
			>{/if}
		{#if error}<p role="alert" class="text-sm text-red-600">{error}</p>{/if}
		<div class="flex justify-end gap-3">
			<Button variant="secondary" disabled={busy} onclick={() => (open = false)}>取消</Button
			><Button type="submit" disabled={busy || (action !== 'delete' && !name.trim())}
				>{busy ? '正在保存…' : action === 'delete' ? '确认删除' : '保存'}</Button
			>
		</div>
	</form>
</MusicDialog>
