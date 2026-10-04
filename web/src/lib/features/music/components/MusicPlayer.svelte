<script lang="ts">
	import { tick } from 'svelte';
	import Music2 from 'lucide-svelte/icons/music-2';
	import Play from 'lucide-svelte/icons/play';
	import Pause from 'lucide-svelte/icons/pause';
	import SkipBack from 'lucide-svelte/icons/skip-back';
	import SkipForward from 'lucide-svelte/icons/skip-forward';
	import Volume2 from 'lucide-svelte/icons/volume-2';
	import ListMusic from 'lucide-svelte/icons/list-music';
	import Captions from 'lucide-svelte/icons/captions';
	import Repeat from 'lucide-svelte/icons/repeat';
	import Repeat1 from 'lucide-svelte/icons/repeat-1';
	import Shuffle from 'lucide-svelte/icons/shuffle';
	import ArrowRight from 'lucide-svelte/icons/arrow-right';
	import Heart from 'lucide-svelte/icons/heart';
	import MusicStage from './MusicStage.svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import {
		musicAudio,
		toggleMusicAudio,
		seekMusicAudio,
		setMusicVolume,
		type AudioState
	} from '$lib/shared/actions/music-audio';
	import { formatMusicTime, musicCoverURL, musicStreamURL } from '../api';
	import type { MusicPlaybackMode, MusicSong } from '../types';
	import { musicModeLabels } from '../persistence';

	let {
		song,
		playRequest,
		publicAccess = false,
		viewer,
		previous,
		next,
		canPrevious,
		canNext,
		volume = $bindable(1),
		queueCount,
		queueOpen,
		toggleQueue,
		sourceReady,
		autoplay,
		beforePlay,
		onEnded,
		mode,
		cycleMode,
		isFavorite,
		favoritePending,
		favoriteError,
		toggleFavorite,
		notice = ''
	}: {
		playRequest: number;
		publicAccess?: boolean;
		viewer: number;
		song: MusicSong | null;
		previous: () => void;
		next: () => void;
		canPrevious: boolean;
		canNext: boolean;
		volume: number;
		queueCount: number;
		queueOpen: boolean;
		toggleQueue: () => void;
		sourceReady: boolean;
		autoplay: boolean;
		beforePlay: () => Promise<boolean>;
		onEnded: () => void;
		mode: MusicPlaybackMode;
		cycleMode: () => void;
		isFavorite: boolean;
		favoritePending: boolean;
		favoriteError: boolean;
		toggleFavorite: () => void;
		notice?: string;
	} = $props();
	let audio = $state<HTMLAudioElement>();
	let expanded = $state(false);
	let playback = $state<AudioState>({
		paused: true,
		loading: false,
		time: 0,
		duration: 0,
		volume: 1,
		error: ''
	});
	const duration = $derived(playback.duration || song?.duration || 0);
	$effect(() => setMusicVolume(audio, volume));
	$effect(() => {
		if (!song) expanded = false;
	});
	async function toggle() {
		try {
			if (playback.paused) {
				if (!(await beforePlay())) return;
				await tick();
			}
			await toggleMusicAudio(audio);
			playback.error = '';
		} catch {
			playback.error = '暂时无法播放，请稍后重试';
		}
	}
</script>

<audio
	bind:this={audio}
	preload="metadata"
	use:musicAudio={{
		src: song && sourceReady ? musicStreamURL(song.id, publicAccess) : '',
		duration: song?.duration,
		key: playRequest,
		autoplay,
		onstate: (value) => (playback = value),
		onend: onEnded
	}}
></audio>
<section class="music-player" aria-label="音乐播放器">
	<button
		type="button"
		class="player-song"
		disabled={!song}
		aria-label={song ? `展开 ${song.title} 的歌词页面` : '暂无播放'}
		aria-expanded={expanded}
		onclick={() => (expanded = true)}
	>
		<div class="player-cover">
			{#if song?.coverArt}<img
					src={musicCoverURL(song.coverArt, publicAccess)}
					alt=""
				/>{:else}<Music2 size={24} strokeWidth={1.4} />{/if}
		</div>
		<div class="min-w-0">
			<p class="song-name">{song?.title || '暂无播放'}</p>
			<p class="song-artist">{song?.artist || '选择一首歌曲开始播放'}</p>
		</div>
	</button>
	<div class="player-controls">
		<div class="player-progress">
			<span>{formatMusicTime(playback.time)}</span><input
				aria-label="播放进度"
				type="range"
				min="0"
				max={duration}
				step="0.1"
				value={playback.time}
				disabled={!song || !sourceReady || duration <= 0}
				oninput={(event) => seekMusicAudio(audio, Number(event.currentTarget.value), duration)}
			/><span>{formatMusicTime(duration)}</span>
		</div>
		<div class="player-transport">
			<Button
				variant="icon"
				class="music-icon player-mode"
				data-active={mode !== 'sequence'}
				aria-label={`播放模式：${musicModeLabels[mode]}，点击切换`}
				title={musicModeLabels[mode]}
				onclick={cycleMode}
				>{#if mode === 'single'}<Repeat1 size={18} />
				{:else if mode === 'repeat'}<Repeat size={18} />
				{:else if mode === 'shuffle'}<Shuffle size={18} />
				{:else}<ArrowRight size={18} />{/if}</Button
			>
			<Button
				variant="icon"
				class="music-icon"
				aria-label="上一首"
				disabled={!canPrevious}
				onclick={() => {
					if (playback.time > 3) seekMusicAudio(audio, 0, duration);
					else previous();
				}}><SkipBack size={18} /></Button
			><Button
				class="music-primary music-play"
				aria-label={playback.paused ? '播放' : '暂停'}
				disabled={!song}
				onclick={toggle}
				>{#if playback.paused}<Play size={24} fill="currentColor" />{:else}<Pause
						size={24}
						fill="currentColor"
					/>{/if}</Button
			><Button
				variant="icon"
				class="music-icon"
				aria-label="下一首"
				disabled={!canNext}
				onclick={next}><SkipForward size={18} /></Button
			>
			<Button
				variant="icon"
				class="music-icon player-favorite"
				aria-label={favoriteError
					? '重新加载收藏状态'
					: isFavorite
						? '取消收藏当前歌曲'
						: '收藏当前歌曲'}
				aria-pressed={isFavorite}
				aria-busy={favoritePending}
				disabled={!song || favoritePending}
				onclick={toggleFavorite}
				><Heart size={21} fill={isFavorite ? 'currentColor' : 'none'} /></Button
			>
		</div>
	</div>
	<div class="player-options">
		<Button
			variant="icon"
			class="music-icon"
			aria-label="展开歌词"
			disabled={!song}
			onclick={() => (expanded = true)}><Captions size={20} /></Button
		>
		<Button
			variant="icon"
			class="music-icon player-queue"
			aria-label={`播放队列，${queueCount} 首歌曲`}
			aria-expanded={queueOpen}
			onclick={toggleQueue}><ListMusic size={19} /><span>{queueCount}</span></Button
		><label class="player-volume"
			><Volume2 size={18} /><span class="sr-only">音量</span><input
				type="range"
				min="0"
				max="1"
				step="0.01"
				bind:value={volume}
			/></label
		>
	</div>
	{#if playback.error}<p class="player-note error" role="alert">
			{playback.error}
		</p>{:else if notice}<p class="player-note" role="status">{notice}</p>
	{:else if playback.loading && !playback.paused}<p class="player-note" role="status">
			正在缓冲…
		</p>{/if}
</section>

<MusicStage
	bind:open={expanded}
	{song}
	{viewer}
	{publicAccess}
	{playback}
	bind:volume
	{toggle}
	seek={(seconds) => seekMusicAudio(audio, seconds, duration)}
	{previous}
	{next}
	{canPrevious}
	{canNext}
/>

<style lang="postcss">
	@reference '../../../../routes/layout.css';
	.music-player {
		position: fixed;
		inset: auto 0 0;
		z-index: 40;
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(260px, 1.5fr) minmax(150px, 1fr);
		align-items: center;
		gap: 28px;
		min-height: 116px;
		padding: 18px 28px;
		border-top: 1px solid var(--music-border);
		background: #fff;
		color: var(--music-ink);
	}
	.player-song {
		display: flex;
		align-items: center;
		gap: 13px;
		min-width: 0;
		text-align: left;
		border-radius: 5px;
	}
	.player-song:focus-visible {
		outline: 2px solid #f43f5e;
		outline-offset: 5px;
	}
	.player-song:not(:disabled):hover .song-name {
		color: #e11d48;
	}
	.player-cover {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 50px;
		height: 50px;
		flex-shrink: 0;
		overflow: hidden;
		border-radius: 5px;
		background: #f5f5f4;
		color: #a8a29e;
	}
	img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.song-name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-size: 14px;
		font-weight: 500;
	}
	.song-artist {
		margin-top: 4px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-size: 12px;
		color: var(--music-muted);
	}
	.player-progress {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;
		color: var(--music-muted);
		@apply font-mono text-[11px];
	}
	input[type='range'] {
		min-width: 0;
		width: 100%;
		flex: 1;
		height: 24px;
		cursor: pointer;
	}
	.player-transport {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 19px;
	}
	.music-player :global(.music-icon) {
		width: 44px;
		height: 44px;
		min-width: 44px;
		padding: 0;
		flex-shrink: 0;
	}
	.music-player :global(svg) {
		flex-shrink: 0;
	}
	:global(.music-play) {
		width: 56px;
		height: 56px;
		padding: 0;
		border-radius: 50%;
	}
	.player-options {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 17px;
	}
	:global(.player-queue) {
		width: auto;
		min-width: 58px;
		gap: 8px;
		padding: 6px;
	}
	:global(.player-queue span) {
		@apply font-mono text-xs;
	}
	.player-volume {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 116px;
		color: var(--music-muted);
	}
	.player-note {
		position: absolute;
		bottom: 100%;
		left: 0;
		right: 0;
		padding: 9px 28px;
		background: #fff;
		border-top: 1px solid var(--music-border);
		color: var(--music-muted);
		font-size: 12px;
	}
	.player-note.error {
		color: var(--music-accent-strong);
	}
	@media (max-width: 1000px) {
		.music-player {
			grid-template-columns: minmax(0, 1fr) minmax(220px, 1.4fr) auto;
			gap: 16px;
			padding-left: 22px;
			padding-right: 22px;
		}
		.player-volume {
			display: none;
		}
	}
	@media (max-width: 640px) {
		.music-player {
			grid-template-columns: minmax(0, 1fr) auto;
			gap: 12px;
			padding: 14px 18px calc(14px + env(safe-area-inset-bottom, 0px));
		}
		.player-controls {
			grid-column: 1 / -1;
			grid-row: 2;
		}
		.player-options {
			grid-column: 2;
			grid-row: 1;
		}
		.player-cover {
			width: 42px;
			height: 42px;
		}
		.song-name {
			font-size: 13px;
		}
		.song-artist {
			font-size: 11px;
		}
		.player-progress {
			margin-bottom: 8px;
		}
		.player-transport {
			gap: 18px;
		}
	}
</style>
