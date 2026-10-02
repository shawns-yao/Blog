<script lang="ts">
	import ListMusic from 'lucide-svelte/icons/list-music';
	import ChartNoAxesColumn from 'lucide-svelte/icons/chart-no-axes-column';
	import Heart from 'lucide-svelte/icons/heart';
	type Feature = 'playlists' | 'charts' | 'favorites';
	let { feature }: { feature: Feature } = $props();
	const content = {
		playlists: {
			icon: ListMusic,
			title: '歌单尚未开放',
			description: '歌单功能将在后续版本开放。'
		},
		charts: {
			icon: ChartNoAxesColumn,
			title: '排行榜尚未开放',
			description: '排行榜功能将在后续版本开放。'
		},
		favorites: { icon: Heart, title: '收藏尚未开放', description: '收藏功能将在后续版本开放。' }
	};
	const selected = $derived(content[feature]);
</script>

<section class="planned-view" aria-label={selected.title}>
	<selected.icon size={34} strokeWidth={1.4} />
	<h2>{selected.title}</h2>
	<p>{selected.description}</p>
</section>

<style>
	.planned-view {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 100px 20px;
		text-align: center;
		color: var(--music-muted);
	}
	h2 {
		margin-top: 20px;
		color: var(--music-ink);
		font-size: 18px;
		font-weight: 500;
	}
	p {
		margin-top: 10px;
		font-size: 13px;
		line-height: 1.7;
	}
</style>
