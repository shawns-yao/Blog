<script lang="ts">
	import { Dialog } from 'bits-ui';
	import X from 'lucide-svelte/icons/x';
	import ListMusic from 'lucide-svelte/icons/list-music';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import type { MusicSong } from '../types';
	let {
		open = $bindable(false),
		songs,
		currentIndex,
		select,
		remove,
		clear
	} = $props<{
		open: boolean;
		songs: MusicSong[];
		currentIndex: number;
		select: (index: number) => void;
		remove: (index: number) => void;
		clear: () => void;
	}>();
</script>

<Dialog.Root bind:open>
	<Dialog.Portal
		><Dialog.Overlay class="music-queue-overlay" /><Dialog.Content class="music-queue-panel">
			<div class="queue-heading">
				<Dialog.Title class="queue-title"
					><ListMusic size={19} />播放队列 <span>{songs.length}</span></Dialog.Title
				>
				<div class="flex items-center gap-2">
					{#if songs.length}<Button variant="ghost" class="queue-button" onclick={clear}
							>清空</Button
						>{/if}<Dialog.Close class="queue-close" aria-label="关闭播放队列"
						><X size={18} /></Dialog.Close
					>
				</div>
			</div>
			<Dialog.Description class="sr-only">选择歌曲进行播放，或移除队列中的歌曲。</Dialog.Description
			>
			{#if songs.length}<ol>
					{#each songs as song, index (song.id)}<li class:current={index === currentIndex}>
							<button
								type="button"
								class="queue-song"
								onclick={() => select(index)}
								aria-current={index === currentIndex ? 'true' : undefined}
								><span class="queue-number">{String(index + 1).padStart(2, '0')}</span><span
									class="min-w-0"
									><span class="block truncate text-sm">{song.title}</span><span
										class="block truncate text-xs text-stone-500"
										>{song.artist || '未知艺术家'}</span
									></span
								></button
							><Button
								variant="icon"
								class="queue-button"
								aria-label={`移除 ${song.title}`}
								onclick={() => remove(index)}><X size={15} /></Button
							>
						</li>{/each}
				</ol>{:else}<p class="queue-empty">播放一首歌曲，或将歌曲加入队列。</p>{/if}
		</Dialog.Content></Dialog.Portal
	>
</Dialog.Root>

<style lang="postcss">
	@reference '../../../../routes/layout.css';
	:global(.music-queue-overlay) {
		position: fixed;
		inset: 0;
		z-index: 2000;
		background: rgb(28 25 23 / 18%);
	}
	:global(.music-queue-panel) {
		position: fixed;
		right: 20px;
		bottom: 130px;
		z-index: 2001;
		width: min(390px, calc(100vw - 32px));
		max-height: min(540px, calc(100dvh - 180px));
		overflow-y: auto;
		padding: 22px;
		border: 1px solid #ece8e5;
		border-radius: 8px;
		background: #fff;
		color: #292524;
		box-shadow: 0 16px 50px rgb(28 25 23 / 12%);
		color-scheme: light;
	}
	.queue-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 12px;
	}
	:global(.queue-title) {
		display: flex;
		align-items: center;
		gap: 9px;
		font-size: 15px;
		font-weight: 500;
	}
	:global(.queue-title span) {
		color: #78716c;
		font-size: 12px;
	}
	:global(.queue-close) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		color: #78716c;
	}
	:global(.music-queue-panel .queue-button) {
		color: #78716c !important;
		background: #fff !important;
	}
	:global(.music-queue-panel .queue-button:hover),
	:global(.queue-close:hover) {
		color: #e11d48 !important;
		background: #fff1f2 !important;
	}
	:global(.music-queue-panel button:focus-visible) {
		outline: 2px solid #f43f5e;
		outline-offset: 2px;
	}
	li {
		display: flex;
		align-items: center;
		gap: 9px;
		border-top: 1px solid #ece8e5;
	}
	.queue-song {
		display: flex;
		align-items: center;
		gap: 12px;
		flex: 1;
		min-width: 0;
		min-height: 66px;
		padding: 10px 0;
		text-align: left;
	}
	.queue-number {
		width: 22px;
		flex-shrink: 0;
		color: #a8a29e;
		@apply font-mono text-xs;
	}
	.current .queue-song {
		color: #e11d48;
	}
	.queue-empty {
		padding: 24px 0;
		color: #78716c;
		font-size: 13px;
	}
	@media (max-width: 640px) {
		:global(.music-queue-panel) {
			right: 12px;
			bottom: calc(164px + env(safe-area-inset-bottom, 0px));
			width: calc(100vw - 24px);
			max-height: calc(100dvh - 190px);
			padding: 18px;
		}
	}
</style>
