<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import type { MomentSummary } from '$lib/features/moment/types';
	import { ArrowUpRight } from 'lucide-svelte';
	import { formatRelativeTime } from '$lib/shared/utils/date';
	import { buildMomentPath } from '$lib/shared/utils/content-path';

	let { moment }: { moment: MomentSummary } = $props();
</script>

<a
	href={resolvePath(buildMomentPath(moment.shortUrl, moment.createdAt))}
	class="moment-preview group"
	aria-label={moment.title || moment.summary || '查看图片手记'}
>
	<div class="min-w-0 flex-1">
		<div class="flex items-center justify-between gap-4 mb-3">
			<time datetime={moment.createdAt} class="text-xs text-ink-600 dark:text-ink-300">
				{formatRelativeTime(moment.createdAt)}
			</time>
			<ArrowUpRight size={16} class="shrink-0 text-jade-700 dark:text-jade-300" />
		</div>
		{#if moment.title}
			<h3
				class="font-serif text-lg font-medium leading-relaxed text-ink-900 dark:text-ink-100 group-hover:underline underline-offset-4 break-words"
			>
				{moment.title}
			</h3>
		{/if}
		{#if moment.summary}
			<p
				class="mt-2 leading-7 text-sm text-ink-700 dark:text-ink-200 break-words"
				class:line-clamp-2={!!moment.title}
			>
				{moment.summary}
			</p>
		{/if}
	</div>
	{#if moment.cover}
		<img src={moment.cover} alt={moment.title || '手记图片'} loading="lazy" class="moment-cover" />
	{/if}
</a>

<style lang="postcss">
	@reference "$routes/layout.css";
	.moment-preview {
		display: flex;
		align-items: flex-start;
		gap: 1.5rem;
		padding: 1.5rem 0;
		border-bottom: 1px solid var(--color-ink-300);
	}
	:global(.dark) .moment-preview {
		border-color: var(--color-ink-700);
	}
	.moment-preview:focus-visible {
		outline: 2px solid var(--color-jade-500);
		outline-offset: 4px;
	}
	.moment-cover {
		width: 160px;
		aspect-ratio: 4 / 3;
		object-fit: cover;
		border-radius: 3px;
		flex-shrink: 0;
	}
	@media (max-width: 639px) {
		.moment-preview {
			flex-direction: column;
			gap: 1rem;
		}
		.moment-cover {
			width: 100%;
			max-height: 240px;
		}
	}
</style>
