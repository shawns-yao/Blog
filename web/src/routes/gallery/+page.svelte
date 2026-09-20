<script lang="ts">
	import { goto } from '$app/navigation';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { resolveHref, resolvePath } from '$lib/shared/utils/resolve-path';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	const totalPages = $derived(Math.max(1, Math.ceil(data.moments.total / data.moments.size)));
	function onPageChange(page: number) {
		const query = new URLSearchParams();
		if (page > 1) query.set('page', String(page));
		if (data.columnId) query.set('column', String(data.columnId));
		goto(resolvePath(`/gallery/${query.size ? `?${query}` : ''}`));
	}
</script>

<div class="mx-auto w-full max-w-5xl px-6 py-16 md:px-0">
	<header class="mb-8 flex items-baseline justify-between gap-4">
		<h1 class="font-serif text-2xl font-medium text-ink-900 dark:text-ink-100">图书馆</h1>
		<span class="text-sm text-ink-500 dark:text-ink-400">{data.moments.total} 篇文章</span>
	</header>
	{#if data.columns.length}
		<div class="mb-4 flex items-center gap-3 text-sm">
			<label for="library-column" class="text-ink-600 dark:text-ink-300">主题</label>
			<select
				id="library-column"
				value={data.columnId ?? ''}
				onchange={(event) =>
					goto(
						resolvePath(
							`/gallery/${event.currentTarget.value ? `?column=${event.currentTarget.value}` : ''}`
						)
					)}
				class="max-w-full rounded border border-ink-300 bg-transparent px-3 py-2 text-ink-900 dark:border-ink-700 dark:bg-ink-900 dark:text-ink-100"
			>
				<option value="">全部主题</option>
				{#each data.columns as column (column.id)}
					<option value={column.id}>{column.name}</option>
				{/each}
			</select>
		</div>
	{/if}

	{#if data.moments.items.length === 0}
		<p class="py-12 text-center text-sm text-ink-500 dark:text-ink-400">暂无文章</p>
	{:else}
		<div class="divide-y divide-ink-200 dark:divide-ink-800">
			{#each data.moments.items as article (article.id)}
				<a
					href={resolveHref(buildMomentPath(article.shortUrl, article.createdAt))}
					class="group flex gap-5 py-6 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jade-500"
				>
					<div class="min-w-0 flex-1">
						<div class="mb-2 flex flex-wrap gap-3 text-xs text-ink-500 dark:text-ink-400">
							<time datetime={article.createdAt}>{article.createdAt.slice(0, 10)}</time>
							{#if article.columnName}<span>{article.columnName}</span>{/if}
						</div>
						<h2
							class="break-words font-serif text-lg font-medium text-ink-900 group-hover:text-jade-600 dark:text-ink-100"
						>
							{article.title}
						</h2>
						{#if article.summary}
							<p
								class="mt-2 line-clamp-3 break-words text-sm leading-relaxed text-ink-600 dark:text-ink-400"
							>
								{article.summary}
							</p>
						{/if}
					</div>
					{#if article.cover}
						<img
							src={article.cover}
							alt=""
							loading="lazy"
							class="h-24 w-24 shrink-0 rounded object-cover sm:h-28 sm:w-40"
						/>
					{/if}
				</a>
			{/each}
		</div>
		{#if totalPages > 1}
			<div class="flex justify-center pt-8">
				<Pagination current={data.moments.page} total={totalPages} {onPageChange} />
			</div>
		{/if}
	{/if}
</div>
