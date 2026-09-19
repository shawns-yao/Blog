<script lang="ts">
	import type { AlbumSummary } from '$lib/features/album/types';
	import type { MomentSummary } from '$lib/features/moment/types';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { resolveHref } from '$lib/shared/utils/resolve-path';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	type GalleryItem =
		| { kind: 'moment'; at: string; moment: MomentSummary }
		| { kind: 'album'; at: string; album: AlbumSummary };

	const items = $derived.by(() => {
		const merged: GalleryItem[] = [
			...data.moments.items.map(
				(moment): GalleryItem => ({ kind: 'moment', at: moment.createdAt, moment })
			),
			...data.albums.items.map(
				(album): GalleryItem => ({ kind: 'album', at: album.createdAt, album })
			)
		];
		return merged.sort((a, b) => new Date(b.at).getTime() - new Date(a.at).getTime());
	});
</script>

<div class="mx-auto max-w-5xl">
	<header class="mb-10">
		<h1 class="font-serif text-2xl font-medium text-ink-900 dark:text-ink-100">图库</h1>
		<p class="mt-2 text-sm text-ink-500 dark:text-ink-400">全部内容 · 手记与相册</p>
	</header>

	{#if items.length === 0}
		<p class="py-20 text-center text-sm text-ink-400 dark:text-ink-500">暂无内容</p>
	{:else}
		<div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
			{#each items as item (item.kind === 'moment' ? `m-${item.moment.id}` : `a-${item.album.id}`)}
				{#if item.kind === 'moment'}
					<a
						href={resolveHref(buildMomentPath(item.moment.shortUrl, item.moment.createdAt))}
						class="group flex flex-col overflow-hidden rounded-default border border-ink-200 bg-white transition-colors hover:border-ink-300 dark:border-ink-800 dark:bg-ink-900 dark:hover:border-ink-700"
					>
						{#if item.moment.cover}
							<div class="aspect-[16/10] overflow-hidden bg-ink-100 dark:bg-ink-800">
								<img
									src={item.moment.cover}
									alt={item.moment.title}
									loading="lazy"
									class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-[1.03]"
								/>
							</div>
						{/if}
						<div class="flex flex-1 flex-col p-4">
							<span class="text-xs text-ink-400 dark:text-ink-500">手记</span>
							<h2 class="mt-1 font-serif text-base font-medium text-ink-900 dark:text-ink-100">
								{item.moment.title}
							</h2>
							{#if item.moment.summary}
								<p class="mt-2 line-clamp-2 text-xs leading-relaxed text-ink-500 dark:text-ink-400">
									{item.moment.summary}
								</p>
							{/if}
						</div>
					</a>
				{:else}
					<a
						href={resolveHref(`/albums/${item.album.shortUrl}`)}
						class="group flex flex-col overflow-hidden rounded-default border border-ink-200 bg-white transition-colors hover:border-ink-300 dark:border-ink-800 dark:bg-ink-900 dark:hover:border-ink-700"
					>
						{#if item.album.cover}
							<div class="aspect-[16/10] overflow-hidden bg-ink-100 dark:bg-ink-800">
								<img
									src={item.album.cover}
									alt={item.album.title}
									loading="lazy"
									class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-[1.03]"
								/>
							</div>
						{/if}
						<div class="flex flex-1 flex-col p-4">
							<span class="text-xs text-ink-400 dark:text-ink-500">相册</span>
							<h2 class="mt-1 font-serif text-base font-medium text-ink-900 dark:text-ink-100">
								{item.album.title}
							</h2>
							<p class="mt-2 text-xs text-ink-500 dark:text-ink-400">
								{item.album.photoCount} 张照片
							</p>
						</div>
					</a>
				{/if}
			{/each}
		</div>
	{/if}
</div>
