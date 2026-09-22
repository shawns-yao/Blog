<script lang="ts">
	import { momentDetailCtx } from '$lib/features/moment/context';
	import type { MomentRelatedMoment } from '$lib/features/moment/types';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { ArrowRight } from 'lucide-svelte';

	const sameRelatedMoments = (
		a: MomentRelatedMoment[] | null | undefined,
		b: MomentRelatedMoment[] | null | undefined
	): boolean => {
		if (a === b) return true;
		if (!a?.length && !b?.length) return true;
		if (!a || !b || a.length !== b.length) return false;

		for (let i = 0; i < a.length; i += 1) {
			const left = a[i];
			const right = b[i];
			if (
				left.id !== right.id ||
				left.title !== right.title ||
				left.shortUrl !== right.shortUrl ||
				left.summary !== right.summary ||
				left.cover !== right.cover ||
				left.createdAt !== right.createdAt
			) {
				return false;
			}
		}

		return true;
	};

	const relatedMomentsStore = momentDetailCtx.selectModelData(
		(data) => data?.relatedMoments ?? [],
		{ equals: sameRelatedMoments }
	);

	function formatDate(dateStr: string) {
		const date = new Date(dateStr);
		return `${date.getMonth() + 1}.${String(date.getDate()).padStart(2, '0')}`;
	}
</script>

<section class="related-moments" aria-labelledby="related-moments-heading">
	<div class="related-head">
		<h3 id="related-moments-heading">同期手记</h3>
		<a href={resolvePath('/moments')} aria-label="查看全部手记"><ArrowRight size={12} /></a>
	</div>

	{#if $relatedMomentsStore.length === 0}
		<p class="related-empty">暂无同期手记</p>
	{:else}
		<ol>
			{#each $relatedMomentsStore as moment (moment.id)}
				<li>
					<a href={resolvePath(buildMomentPath(moment.shortUrl, moment.createdAt))}>
						<time datetime={moment.createdAt}>{formatDate(moment.createdAt)}</time>
						<strong>{moment.title}</strong>
						<p>{moment.summary}</p>
					</a>
				</li>
			{/each}
		</ol>
	{/if}
</section>

<style>
	.related-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding-bottom: 0.65rem;
		border-bottom: 1px solid var(--book-rule);
	}

	.related-head h3 {
		font-family: var(--font-serif);
		font-size: 0.62rem;
		letter-spacing: 0.17em;
		color: var(--book-faint);
	}

	.related-head a {
		color: var(--book-faint);
	}

	.related-head a:hover,
	.related-head a:focus-visible {
		color: var(--book-accent);
	}

	.related-moments ol {
		margin-top: 0.4rem;
	}

	.related-moments li a {
		display: grid;
		grid-template-columns: 2.4rem minmax(0, 1fr);
		gap: 0.3rem 0.55rem;
		padding: 0.72rem 0;
		border-bottom: 1px solid rgba(76, 58, 43, 0.1);
	}

	.related-moments time {
		grid-row: 1 / span 2;
		font-family: var(--font-mono);
		font-size: 0.5rem;
		color: var(--book-faint);
	}

	.related-moments strong {
		overflow: hidden;
		font-family: var(--font-serif);
		font-size: 0.69rem;
		font-weight: 600;
		line-height: 1.55;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.related-moments p {
		display: -webkit-box;
		overflow: hidden;
		font-family: var(--font-serif);
		font-size: 0.58rem;
		line-height: 1.55;
		color: var(--book-muted);
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 2;
		line-clamp: 2;
	}

	.related-moments li a:hover strong,
	.related-moments li a:focus-visible strong {
		color: var(--book-accent);
	}

	.related-empty {
		padding: 0.8rem 0;
		font-family: var(--font-serif);
		font-size: 0.64rem;
		color: var(--book-faint);
	}
</style>
