<script lang="ts">
	import type { MomentSummary } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { ArrowUpRight, Pin } from 'lucide-svelte';

	let { moment }: { moment: MomentSummary } = $props();
	const href = $derived(resolvePath(buildMomentPath(moment.shortUrl, moment.createdAt)));
	const date = $derived(moment.createdAt.slice(0, 10));
</script>

<article
	class="grid min-w-0 gap-4 border-b border-ink-200 py-8 sm:grid-cols-[104px_minmax(0,1fr)] sm:gap-8 dark:border-ink-800"
>
	<div
		class="flex items-center gap-3 text-xs text-ink-500 sm:flex-col sm:items-start sm:pt-1 dark:text-ink-400"
	>
		<time datetime={moment.createdAt} class="font-mono">{date}</time>
		{#if moment.isTop}<span class="inline-flex items-center gap-1 text-jade-700 dark:text-jade-400"
				><Pin size={12} />置顶</span
			>{/if}
	</div>
	<div class="min-w-0">
		{#if moment.title}
			<h2
				class="font-serif text-xl font-medium leading-relaxed break-words text-ink-900 dark:text-ink-100"
			>
				<a
					{href}
					class="hover:text-jade-700 focus-visible:outline-2 focus-visible:outline-jade-500 dark:hover:text-jade-400"
					>{moment.title}</a
				>
			</h2>
		{/if}
		{#if moment.summary}
			<p
				class="whitespace-pre-line break-words text-base leading-8 text-ink-700 dark:text-ink-300"
				class:mt-3={!!moment.title}
			>
				{moment.summary}
			</p>
		{/if}
		{#if moment.cover}
			<a
				{href}
				class="mt-5 block max-w-xl focus-visible:outline-2 focus-visible:outline-jade-500"
				aria-label={moment.title || '查看图片手记'}
			>
				<img
					src={moment.cover}
					alt={moment.title || '手记图片'}
					loading="lazy"
					class="aspect-[4/3] w-full rounded-sm bg-ink-100 object-contain dark:bg-ink-900"
				/>
			</a>
		{/if}
		<div class="mt-5 flex flex-wrap items-center justify-between gap-3">
			<div class="flex min-w-0 flex-wrap gap-x-3 gap-y-1 text-xs text-ink-500 dark:text-ink-400">
				{#each moment.topics ?? [] as topic}<span class="break-all">#{topic}</span>{/each}
			</div>
			<a
				{href}
				class="inline-flex shrink-0 items-center gap-1 text-xs text-jade-700 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-jade-500 dark:text-jade-400"
			>
				查看手记<ArrowUpRight size={14} />
			</a>
		</div>
	</div>
</article>
