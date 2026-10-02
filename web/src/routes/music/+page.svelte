<script lang="ts">
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
</script>

<svelte:head><title>音乐</title><meta name="robots" content="noindex" /></svelte:head>

<div class="music-route">
	<QueryRoot loader={() => import('$lib/features/music/components/MusicLibrary.svelte')}>
		{#snippet fallback()}<p role="status" class="py-24 text-center text-sm text-ink-500">
				正在打开音乐室…
			</p>{/snippet}
	</QueryRoot>
</div>

<style>
	.music-route {
		min-height: 100dvh;
		background: #fff;
		color: #292524;
	}
</style>
