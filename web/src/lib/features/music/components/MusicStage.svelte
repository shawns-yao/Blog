<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { createQuery } from '@tanstack/svelte-query';
	import ChevronDown from 'lucide-svelte/icons/chevron-down';
	import Music2 from 'lucide-svelte/icons/music-2';
	import Play from 'lucide-svelte/icons/play';
	import Pause from 'lucide-svelte/icons/pause';
	import SkipBack from 'lucide-svelte/icons/skip-back';
	import SkipForward from 'lucide-svelte/icons/skip-forward';
	import Volume2 from 'lucide-svelte/icons/volume-2';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { followMusicLyric } from '$lib/shared/actions/music-lyrics';
	import type { AudioState } from '$lib/shared/actions/music-audio';
	import { formatMusicTime, getMusicLyrics, musicCoverURL } from '../api';
	import type { MusicSong } from '../types';

	let {
		open = $bindable(false),
		song,
		viewer,
		publicAccess,
		playback,
		volume = $bindable(1),
		toggle,
		seek,
		previous,
		next,
		canPrevious,
		canNext
	} = $props<{
		open: boolean;
		song: MusicSong | null;
		viewer: number;
		publicAccess: boolean;
		playback: AudioState;
		volume: number;
		toggle: () => Promise<void>;
		seek: (seconds: number) => void;
		previous: () => void;
		next: () => void;
		canPrevious: boolean;
		canNext: boolean;
	}>();
	const lyrics = createQuery(() => ({
		queryKey: ['music', viewer, 'lyrics', song?.id ?? ''],
		queryFn: ({ signal }) => getMusicLyrics(song?.id ?? '', signal, publicAccess),
		enabled: open && !!song,
		retry: false,
		staleTime: 0
	}));
	const selected = $derived(
		lyrics.data?.lyrics.find((item) => item.synced) ?? lyrics.data?.lyrics[0]
	);
	const lines = $derived(selected?.line ?? []);
	const offset = $derived(selected?.offset ?? 0);
	const activeIndex = $derived.by(() => {
		if (!selected?.synced) return -1;
		let active = -1;
		for (let i = 0; i < lines.length; i++) {
			const start = lines[i].start;
			if (typeof start === 'number' && start <= playback.time * 1000 + offset) active = i;
		}
		return active;
	});
	const duration = $derived(playback.duration || song?.duration || 0);
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay class="music-stage-overlay" />
		<Dialog.Content class="music-stage">
			<header class="stage-header">
				<Dialog.Title class="stage-title">正在播放</Dialog.Title>
				<Dialog.Close class="stage-close" aria-label="收起歌词页面"
					><ChevronDown size={24} /></Dialog.Close
				>
			</header>
			<Dialog.Description class="sr-only"
				>展示歌曲封面和歌词。点击带有时间的歌词可以跳转播放，收起页面后音乐继续播放。</Dialog.Description
			>
			<main class="stage-body">
				<div class="stage-record">
					<div class="stage-artwork">
						{#if song?.coverArt}<img
								src={musicCoverURL(song.coverArt, publicAccess)}
								alt={`${song.title}的封面`}
							/>{:else}<Music2 size={76} strokeWidth={0.8} />{/if}
					</div>
					<div class="stage-meta">
						<h2>{song?.title || '暂无播放'}</h2>
						<p>{song?.artist || '未知艺术家'}</p>
						{#if song?.album}<p class="stage-album">{song.album}</p>{/if}
					</div>
				</div>
				<div
					class="stage-lyrics"
					aria-label="歌曲歌词"
					use:followMusicLyric={{ index: activeIndex, key: `${song?.id}-${lyrics.dataUpdatedAt}` }}
				>
					{#if lyrics.isPending}
						<p class="lyrics-state" role="status">正在加载歌词…</p>
					{:else if lyrics.isError}
						<div class="lyrics-state" role="alert">
							<p>歌词暂时无法加载</p>
							<Button variant="ghost" class="stage-button" onclick={() => lyrics.refetch()}
								>重新加载</Button
							>
						</div>
					{:else if lines.length}
						<div class="lyrics-lines">
							{#each lines as line, index (index)}
								{#if selected?.synced && typeof line.start === 'number'}
									<button
										type="button"
										class="lyric-line"
										class:active={index === activeIndex}
										data-lyric-index={index}
										aria-current={index === activeIndex ? 'true' : undefined}
										aria-label={`跳转到 ${formatMusicTime(Math.max(0, (line.start - offset) / 1000))}，${line.value || '间奏'}`}
										onclick={() => seek(Math.max(0, ((line.start ?? 0) - offset) / 1000))}
										>{line.value || '· · ·'}</button
									>
								{:else}<p class="lyric-line static-line" data-lyric-index={index}>
										{line.value}
									</p>{/if}
							{/each}
						</div>
					{:else}<p class="lyrics-state">暂无歌词</p>{/if}
				</div>
			</main>
			<footer class="stage-footer">
				<div class="stage-progress">
					<span>{formatMusicTime(playback.time)}</span><input
						type="range"
						aria-label="歌词页面播放进度"
						min="0"
						max={duration}
						step="0.1"
						value={playback.time}
						disabled={!playback.duration}
						oninput={(event) => seek(Number(event.currentTarget.value))}
					/><span>{formatMusicTime(duration)}</span>
				</div>
				<div class="stage-controls">
					<div class="stage-playback-state" role="status">
						{playback.error || (playback.loading ? '正在缓冲…' : '')}
					</div>
					<div class="stage-transport">
						<Button
							variant="icon"
							class="stage-button"
							aria-label="上一首"
							disabled={!canPrevious}
							onclick={() => (playback.time > 3 ? seek(0) : previous())}
							><SkipBack size={20} /></Button
						>
						<Button
							class="stage-play"
							aria-label={playback.paused ? '播放' : '暂停'}
							disabled={!song}
							onclick={toggle}
							>{#if playback.paused}<Play size={22} fill="currentColor" />{:else}<Pause
									size={22}
									fill="currentColor"
								/>{/if}</Button
						>
						<Button
							variant="icon"
							class="stage-button"
							aria-label="下一首"
							disabled={!canNext}
							onclick={next}><SkipForward size={20} /></Button
						>
					</div>
					<label class="stage-volume"
						><Volume2 size={18} /><span class="sr-only">歌词页面音量</span><input
							type="range"
							min="0"
							max="1"
							step="0.01"
							bind:value={volume}
						/></label
					>
				</div>
			</footer>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style lang="postcss">
	@reference '../../../../routes/layout.css';
	:global(.music-stage-overlay) {
		position: fixed;
		inset: 0;
		z-index: 2000;
		background: #fff;
	}
	:global(.music-stage) {
		position: fixed;
		inset: 0;
		z-index: 2001;
		display: grid;
		grid-template-rows: auto minmax(0, 1fr) auto;
		height: 100dvh;
		background: #fff;
		color: #292524;
		color-scheme: light;
		overflow: hidden;
	}
	.stage-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 22px 40px;
	}
	:global(.stage-title) {
		font-size: 14px;
		font-weight: 500;
		color: #78716c;
	}
	:global(.stage-close) {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		border-radius: 50%;
		color: #78716c;
		background: #fff;
	}
	.stage-body {
		width: min(1160px, 100%);
		margin: 0 auto;
		display: grid;
		grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr);
		gap: clamp(40px, 7vw, 110px);
		min-height: 0;
		padding: 16px 40px 32px;
	}
	.stage-record {
		min-width: 0;
		min-height: 0;
		display: flex;
		flex-direction: column;
		justify-content: center;
	}
	.stage-artwork {
		display: flex;
		align-items: center;
		justify-content: center;
		width: min(100%, 400px, max(96px, calc(100dvh - 330px)));
		aspect-ratio: 1;
		align-self: center;
		overflow: hidden;
		border-radius: 8px;
		background: #f5f5f4;
		color: #a8a29e;
		box-shadow: 0 12px 32px rgb(28 25 23 / 8%);
	}
	.stage-artwork img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.stage-meta {
		margin-top: 28px;
		text-align: center;
	}
	h2 {
		font-family: 'Noto Serif SC', serif;
		font-size: clamp(22px, 2.5vw, 30px);
		line-height: 1.4;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.stage-meta p {
		margin-top: 12px;
		font-size: 14px;
		color: #78716c;
	}
	.stage-meta .stage-album {
		font-size: 12px;
		color: #78716c;
		margin-top: 6px;
	}
	.stage-lyrics {
		position: relative;
		min-height: 0;
		overflow-y: auto;
		overscroll-behavior: contain;
		scrollbar-width: thin;
		scrollbar-color: #e7e5e4 transparent;
	}
	.lyrics-lines {
		padding-block: min(30dvh, 180px);
	}
	.lyric-line {
		display: block;
		width: 100%;
		padding: 14px 8px;
		text-align: left;
		font-size: clamp(20px, 2.2vw, 27px);
		line-height: 1.65;
		color: #78716c;
		font-weight: 500;
		overflow-wrap: anywhere;
		transition: color 180ms;
	}
	.lyric-line:hover {
		color: #57534e;
	}
	.lyric-line.active {
		color: #e11d48;
		font-weight: 600;
	}
	.static-line {
		color: #78716c;
	}
	.lyrics-state {
		height: 100%;
		min-height: 120px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		align-items: center;
		justify-content: center;
		color: #78716c;
		font-size: 15px;
	}
	.stage-footer {
		width: min(960px, 100%);
		margin: 0 auto;
		padding: 16px 40px calc(24px + env(safe-area-inset-bottom, 0px));
	}
	.stage-progress {
		display: flex;
		align-items: center;
		gap: 14px;
		font-family: monospace;
		color: #78716c;
		font-size: 12px;
	}
	input[type='range'] {
		min-width: 0;
		flex: 1;
		width: 100%;
		height: 4px;
		accent-color: #f43f5e;
	}
	.stage-controls {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
		align-items: center;
		gap: 20px;
		margin-top: 18px;
	}
	.stage-transport {
		display: flex;
		align-items: center;
		gap: 24px;
	}
	.stage-volume {
		display: flex;
		align-items: center;
		justify-self: end;
		gap: 10px;
		width: 120px;
		color: #78716c;
	}
	.stage-playback-state {
		color: #e11d48;
		font-size: 12px;
	}
	:global(.music-stage .stage-button) {
		color: #78716c !important;
		background: #fff !important;
	}
	:global(.music-stage .stage-button:hover),
	:global(.stage-close:hover) {
		color: #e11d48 !important;
		background: #fff1f2 !important;
	}
	:global(.music-stage .stage-play) {
		width: 52px;
		height: 52px;
		padding: 0;
		border: 0;
		border-radius: 50%;
		background: #f43f5e !important;
		color: #fff !important;
	}
	:global(.music-stage .stage-play:hover) {
		background: #e11d48 !important;
	}
	:global(.music-stage button:focus-visible),
	:global(.music-stage input:focus-visible) {
		outline: 2px solid #f43f5e;
		outline-offset: 4px;
	}
	@media (prefers-reduced-motion: reduce) {
		.lyric-line {
			transition: none;
		}
	}
	@media (max-width: 640px) {
		.stage-header {
			padding: 8px 18px;
		}
		.stage-body {
			grid-template-columns: minmax(0, 1fr);
			grid-template-rows: auto minmax(0, 1fr);
			gap: 12px;
			padding: 0 24px 8px;
		}
		.stage-record {
			flex-direction: row;
			justify-content: flex-start;
			align-items: center;
			gap: 18px;
		}
		.stage-artwork {
			flex: 0 0 96px;
			width: 96px;
			height: 96px;
			max-height: none;
		}
		.stage-meta {
			margin-top: 0;
			text-align: left;
			min-width: 0;
		}
		h2 {
			font-size: 20px;
		}
		.stage-meta p {
			margin-top: 6px;
			font-size: 12px;
		}
		.lyric-line {
			font-size: 21px;
			text-align: center;
			padding: 13px 6px;
		}
		.lyrics-lines {
			padding-block: 100px;
		}
		.stage-footer {
			padding: 12px 24px calc(18px + env(safe-area-inset-bottom, 0px));
		}
		.stage-controls {
			grid-template-columns: 1fr;
			gap: 8px;
			margin-top: 14px;
		}
		.stage-transport {
			justify-content: center;
		}
		.stage-volume {
			display: none;
		}
		.stage-playback-state:empty {
			display: none;
		}
		.stage-playback-state {
			text-align: center;
		}
	}
</style>
