<script lang="ts">
	import Disc3 from 'lucide-svelte/icons/disc-3';
	import { musicCoverURL } from '../api';
	import type { MusicAlbum } from '../types';
	let {
		albums,
		select,
		publicAccess = false
	} = $props<{ albums: MusicAlbum[]; select: (id: string) => void; publicAccess?: boolean }>();
</script>

<div class="album-grid">
	{#each albums as album (album.id)}<button
			type="button"
			class="album-tile"
			onclick={() => select(album.id)}
			><span class="album-cover"
				>{#if album.coverArt}<img
						src={musicCoverURL(album.coverArt, publicAccess)}
						alt=""
						loading="lazy"
					/>{:else}<Disc3 size={42} strokeWidth={1.2} />{/if}</span
			><span class="album-name">{album.name}</span><span class="album-artist"
				>{album.artist || '未知艺术家'}</span
			><span class="album-count">{album.songCount} 首歌曲</span></button
		>{/each}
</div>

<style>
	.album-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
		gap: 28px 22px;
	}
	.album-tile {
		min-width: 0;
		text-align: left;
	}
	.album-cover {
		display: flex;
		aspect-ratio: 1;
		overflow: hidden;
		align-items: center;
		justify-content: center;
		border-radius: 7px;
		background: var(--music-panel);
		color: #a8a29e;
	}
	img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.album-name,
	.album-artist,
	.album-count {
		display: block;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.album-name {
		margin-top: 12px;
		font-size: 13px;
		font-weight: 500;
	}
	.album-tile:hover .album-name {
		color: var(--music-accent-strong);
	}
	.album-artist {
		margin-top: 4px;
		font-size: 12px;
		color: var(--music-muted);
	}
	.album-count {
		margin-top: 3px;
		font-size: 11px;
		color: var(--music-muted);
	}
	@media (max-width: 640px) {
		.album-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
			gap: 22px 16px;
		}
	}
</style>
