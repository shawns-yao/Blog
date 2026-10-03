<script lang="ts">
	import type { Snippet } from 'svelte';
	import Plus from 'lucide-svelte/icons/plus';
	import Play from 'lucide-svelte/icons/play';
	import Music2 from 'lucide-svelte/icons/music-2';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { formatMusicTime, musicCoverURL } from '../api';
	import type { MusicDensity, MusicSong } from '../types';
	let {
		songs,
		publicAccess = false,
		currentId,
		play,
		enqueue,
		density = 'comfortable',
		showCovers = true,
		startIndex = 0,
		actions,
		statusLabel = (song: MusicSong) =>
			song.unavailable ? '歌曲已失效' : song.public ? '公开试听' : '需授权播放'
	} = $props<{
		songs: MusicSong[];
		publicAccess?: boolean;
		currentId?: string;
		play: (song: MusicSong) => void;
		enqueue: (song: MusicSong) => void;
		density?: MusicDensity;
		showCovers?: boolean;
		startIndex?: number;
		actions?: Snippet<[MusicSong]>;
		statusLabel?: (song: MusicSong) => string;
	}>();
</script>

<table class="song-table" class:compact={density === 'compact'} aria-label="歌曲列表">
	<thead
		><tr
			><th scope="col" class="sequence">#</th><th scope="col">歌曲</th><th
				scope="col"
				class="artist-column">艺术家</th
			><th scope="col" class="album-column">专辑</th><th scope="col" class="duration">时长</th><th
				scope="col"
				class="action-column"><span class="sr-only">操作</span></th
			></tr
		></thead
	>
	<tbody
		>{#each songs as song, index (song.id)}<tr class:current={currentId === song.id}
				><td class="sequence"><span>{String(startIndex + index + 1).padStart(2, '0')}</span></td><td
					><button
						type="button"
						class="song-select"
						onclick={() => play(song)}
						aria-label={`播放 ${song.title}`}
						>{#if showCovers}<span class="song-cover"
								>{#if song.coverArt}<img
										src={musicCoverURL(song.coverArt, publicAccess)}
										alt=""
										loading="lazy"
									/>{:else}<Music2 size={19} />{/if}<span class="cover-play"
									><Play size={17} fill="currentColor" /></span
								></span
							>{/if}<span class="song-label"
							><span class="song-title">{song.title}</span><span class="metadata"
								>{statusLabel(song)}</span
							><span class="mobile-artist">{song.artist || '未知艺术家'}</span></span
						></button
					></td
				><td class="artist-column"><span class="metadata">{song.artist || '未知艺术家'}</span></td
				><td class="album-column"><span class="metadata">{song.album || '未归属专辑'}</span></td><td
					class="duration">{formatMusicTime(song.duration)}</td
				><td class="action-column"
					>{#if actions}{@render actions(song)}{:else}<Button
							variant="icon"
							class="music-icon"
							aria-label={`将 ${song.title} 加入队列`}
							onclick={() => enqueue(song)}><Plus size={17} /></Button
						>{/if}</td
				></tr
			>{/each}</tbody
	>
</table>

<style lang="postcss">
	@reference '../../../../routes/layout.css';
	.song-table {
		width: 100%;
		table-layout: fixed;
		text-align: left;
		border-collapse: collapse;
	}
	th {
		padding: 12px 12px;
		border-bottom: 1px solid var(--music-border);
		color: var(--music-muted);
		font-size: 12px;
		font-weight: 400;
	}
	td {
		padding: 14px 12px;
		border-bottom: 1px solid var(--music-border);
	}
	.sequence {
		width: 46px;
		text-align: center;
		color: #a8a29e;
	}
	td.sequence {
		@apply font-mono text-xs;
	}
	.artist-column {
		width: 20%;
	}
	.album-column {
		width: 24%;
	}
	.duration {
		width: 70px;
		color: var(--music-muted);
		white-space: nowrap;
	}
	td.duration {
		@apply font-mono text-[11px];
	}
	.action-column {
		width: 52px;
		text-align: right;
	}
	.song-select {
		display: flex;
		align-items: center;
		gap: 14px;
		width: 100%;
		min-width: 0;
		min-height: 44px;
		text-align: left;
	}
	.song-cover {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
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
	.cover-play {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgb(28 25 23 / 40%);
		color: #fff;
		opacity: 0;
		transition: opacity 0.15s ease;
	}
	.song-select:hover .cover-play,
	.song-select:focus-visible .cover-play {
		opacity: 1;
	}
	.song-label {
		min-width: 0;
		flex: 1;
	}
	.song-title,
	.metadata {
		display: block;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
	.song-title {
		font-size: 13px;
		font-weight: 500;
	}
	.metadata {
		color: var(--music-muted);
		font-size: 12px;
	}
	.mobile-artist {
		display: none;
		margin-top: 4px;
		color: var(--music-muted);
		font-size: 11px;
	}
	.current {
		background: var(--music-accent-soft);
	}
	.current .song-title {
		color: var(--music-accent-strong);
	}
	.compact td {
		padding-top: 7px;
		padding-bottom: 7px;
	}
	.compact .song-cover {
		width: 36px;
		height: 36px;
	}
	@media (max-width: 1100px) {
		.album-column {
			display: none;
		}
		.artist-column {
			width: 25%;
		}
	}
	@media (max-width: 850px) {
		.artist-column {
			display: none;
		}
		.mobile-artist {
			display: block;
			overflow: hidden;
			white-space: nowrap;
			text-overflow: ellipsis;
		}
	}
	@media (max-width: 640px) {
		.sequence {
			display: none;
		}
		th,
		td {
			padding-left: 6px;
			padding-right: 6px;
		}
		.duration {
			width: 48px;
		}
		.action-column {
			width: 44px;
		}
		.song-select {
			gap: 10px;
		}
		.song-cover {
			width: 38px;
			height: 38px;
		}
	}
</style>
