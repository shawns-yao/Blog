<script lang="ts">
	import ContentLikeButton from '$lib/features/analytics/components/ContentLikeButton.svelte';
	import type { MomentDetail } from '$lib/features/moment/types';
	import TagList from '$lib/features/tag/components/TagList.svelte';
	import { formatDateCN, isDifferentDay } from '$lib/shared/utils/date';
	import { RollingNumber } from '$lib/ui/animation';
	import DetailActionBar from '$lib/ui/detail/DetailActionBar.svelte';
	import DetailAiSummary from '$lib/ui/detail/DetailAiSummary.svelte';
	import DetailCommentSection from '$lib/ui/detail/DetailCommentSection.svelte';
	import DetailMarkdownContent from '$lib/ui/detail/DetailMarkdownContent.svelte';
	import MomentAtmosphere from './MomentAtmosphere.svelte';

	interface Props {
		moment: MomentDetail;
		dateStr: string;
		onActiveAnchorChange: (anchor: string | null) => void;
		onContentRootChange: (node: HTMLElement | null) => void;
	}

	let {
		moment,
		dateStr,
		onActiveAnchorChange,
		onContentRootChange
	}: Props = $props();
	const showUpdated = $derived(isDifferentDay(moment.createdAt, moment.contentUpdatedAt));
	const monthLabel = $derived(
		[
			'JAN.',
			'FEB.',
			'MAR.',
			'APR.',
			'MAY.',
			'JUN.',
			'JUL.',
			'AUG.',
			'SEP.',
			'OCT.',
			'NOV.',
			'DEC.'
		][Number(moment.createdAt.slice(5, 7)) - 1] ?? ''
	);
</script>

<div
	class="moment-detail-paper moment-vt"
	class:short-with-photo={!!moment.cover && moment.content.length < 240}
	style:view-transition-name={`moment-${moment.id}`}
>
	<header class="article-header">
		<div class="article-topline">
			<div class="article-date" aria-label={dateStr}>
				<span>{monthLabel}</span>
				<strong>{Number(moment.createdAt.slice(8, 10))}</strong>
				<small>{moment.createdAt.slice(0, 4)}</small>
			</div>
			<div class="article-weather">
				<MomentAtmosphere atmosphere={moment.extInfo?.moment} />
				<time datetime={moment.createdAt}>{moment.createdAt.slice(11, 16)}</time>
				{#if showUpdated}<small>更新于 {formatDateCN(moment.contentUpdatedAt)}</small>{/if}
			</div>
		</div>

		{#if moment.title}
			<h1>{moment.title}</h1>
		{/if}

		{#if moment.summary && moment.contentKind === 'article'}
			<p class="article-deck">{moment.summary}</p>
		{/if}

		{#if moment.contentKind === 'article'}
			<div class="article-facts">
				<span>浏览 <RollingNumber value={moment.metrics?.views ?? 0} /></span>
				<i aria-hidden="true"></i>
				<ContentLikeButton
					contentType="moment"
					contentId={moment.id}
					likes={moment.metrics?.likes ?? 0}
					className="inline-flex items-center gap-1.5"
				/>
				<i aria-hidden="true"></i>
				<span>评论 <RollingNumber value={moment.metrics?.comments ?? 0} /></span>
			</div>
		{/if}
	</header>

	{#if moment.aiSummary}
		<div class="article-summary"><DetailAiSummary summary={moment.aiSummary} /></div>
	{/if}

	<div class:has-photo={!!moment.cover} class="article-body">
		<div class="article-content">
			<DetailMarkdownContent
				content={moment.content}
				toc={moment.toc}
				className="max-w-none font-serif text-justify text-[15px]"
				{onContentRootChange}
				{onActiveAnchorChange}
			/>
		</div>
		{#if moment.cover}
			<figure class="article-photo">
				<img src={moment.cover} alt={moment.title || '手记照片'} loading="lazy" />
			</figure>
		{/if}
	</div>

	{#if moment.topics?.length}
		<div class="article-tags"><TagList tags={moment.topics} /></div>
	{/if}

	<div class="article-actions">
			<DetailActionBar
				contentType="moment"
				contentId={moment.id}
				likes={moment.metrics?.likes ?? 0}
				comments={moment.metrics?.comments ?? 0}
				tone="cinnabar"
				shareTitle={moment.title}
				shareDescription={moment.summary}
				shareImageUrl={moment.cover ?? ''}
			/>
		</div>

	<div class="article-seal" aria-hidden="true">
		<span>{moment.contentKind === 'article' ? '图书馆' : '手记'}</span>
		<strong>{moment.contentKind === 'article' ? '文' : '记'}</strong>
		<small>{dateStr}</small>
	</div>

	<div class="article-comments">
			<DetailCommentSection
				commentAreaId={moment.commentAreaId}
				commentsCount={moment.metrics?.comments ?? 0}
				fediverseObjectUrl={moment.activityPubObjectId}
				containerClass="mt-16 pt-10 border-t"
				fallbackText="正在展开评论……"
				fallbackSize="w-6 h-6"
				fallbackContainerClass="flex justify-center py-20"
			/>
		</div>
</div>

<style>
	.moment-detail-paper {
		position: relative;
		flex: 1 0 auto;
		max-width: 37rem;
		width: 100%;
		margin-inline: auto;
		color: var(--book-ink);
	}
	.short-with-photo {
		min-height: 25rem;
	}

	.article-header {
		padding-bottom: 0.15rem;
	}

	.article-topline {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.article-date {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		line-height: 1;
		color: var(--book-ink);
	}
	.article-date span,
	.article-date small {
		font-family: var(--font-mono);
		font-size: 0.72rem;
		letter-spacing: 0.18em;
		color: var(--book-faint);
	}
	.article-date strong {
		margin: 0.12rem 0 0.22rem;
		font-family: var(--font-serif);
		font-size: clamp(2.6rem, 4vw, 3.4rem);
		font-weight: 400;
		line-height: 0.95;
	}
	.article-weather {
		display: flex;
		align-items: flex-end;
		flex-direction: column;
		gap: 0.45rem;
		padding-top: 0.3rem;
		font-family: var(--font-mono);
		font-size: 0.67rem;
		letter-spacing: 0.08em;
		color: var(--book-muted);
	}
	.article-weather small {
		font-size: 0.6rem;
	}
	.article-weather :global(span.inline-flex) {
		border: 0;
		border-radius: 0;
		padding: 0;
	}

	.article-facts i {
		display: inline-block;
		width: 1.2rem;
		height: 1px;
		background: var(--book-rule);
	}

	.article-header h1 {
		max-width: 55rem;
		margin-top: clamp(0.8rem, 1.8vw, 1.35rem);
		font-family: var(--font-serif);
		font-size: clamp(1.8rem, 2.8vw, 2.55rem);
		font-weight: 600;
		line-height: 1.32;
		letter-spacing: 0.045em;
		text-wrap: balance;
	}

	.article-deck {
		max-width: 46rem;
		margin-top: 0.8rem;
		font-family: var(--font-serif);
		font-size: clamp(0.9rem, 1.2vw, 1.04rem);
		line-height: 2;
		color: var(--book-muted);
	}

	.article-facts {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.65rem;
		margin-top: 0.9rem;
		font-family: var(--font-mono);
		font-size: 0.57rem;
		letter-spacing: 0.11em;
		color: var(--book-faint);
	}

	.article-tags {
		margin: 1.5rem 0 1rem;
	}

	.article-tags :global(*) {
		color: var(--book-muted) !important;
		border-color: var(--book-rule) !important;
		background: transparent !important;
	}

	.article-summary {
		margin-top: 2.5rem;
	}

	.article-summary :global(> div) {
		border-color: rgba(74, 103, 79, 0.22) !important;
		background: rgba(122, 139, 102, 0.07) !important;
		box-shadow: none !important;
	}

	.article-summary :global(.markdown-preview) {
		color: var(--book-muted) !important;
	}

	.article-body {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 1rem;
		margin-top: 1rem;
	}
	.article-body.has-photo {
		grid-template-columns: minmax(0, 1fr) clamp(9rem, 22%, 13rem);
		align-items: end;
		gap: clamp(1rem, 3vw, 1.8rem);
	}
	.article-content {
		min-width: 0;
		max-width: 54rem;
		margin: 0 auto;
	}

	.article-content :global(.markdown-preview) {
		color: var(--book-ink) !important;
		font-size: clamp(0.94rem, 1.2vw, 1.04rem);
		line-height: 1.72;
	}
	.article-content :global(.markdown-preview p) {
		margin-bottom: 0.7rem;
	}
	.article-content :global(.markdown-preview p:last-child) {
		margin-bottom: 0;
	}
	.article-photo {
		width: 100%;
		max-width: 13rem;
		margin: 0 0 0 auto;
		padding: 0.4rem 0.4rem 1.2rem;
		background: #ead7b7;
		box-shadow: 0 0.55rem 1.1rem rgba(63, 42, 25, 0.25);
		transform: rotate(4deg);
	}
	.article-photo img {
		display: block;
		width: 100%;
		aspect-ratio: 1 / 1;
		object-fit: cover;
	}

	.article-content :global(.markdown-preview :where(p, li, strong, em, h1, h2, h3, h4, h5, h6)) {
		color: inherit !important;
	}

	.article-content :global(.markdown-preview :where(h1, h2, h3, h4)) {
		font-family: var(--font-serif);
		letter-spacing: 0.04em;
	}

	.article-content :global(.markdown-preview a) {
		color: var(--book-accent) !important;
		text-decoration-color: rgba(141, 56, 45, 0.35);
	}

	.article-content :global(.markdown-preview blockquote) {
		color: var(--book-muted) !important;
		border-color: var(--book-accent) !important;
		background: rgba(124, 86, 52, 0.045) !important;
	}

	.article-actions {
		max-width: 54rem;
		margin: 0 auto;
	}

	.article-actions :global(> div) {
		border-color: var(--book-rule) !important;
	}

	.article-actions :global(button),
	.article-actions :global(a) {
		color: var(--book-muted) !important;
	}

	.article-actions :global(button:hover),
	.article-actions :global(a:hover) {
		color: var(--book-accent) !important;
	}

	.article-seal {
		display: inline-flex;
		width: fit-content;
		align-items: center;
		gap: 0.45rem;
		margin: 1.2rem 0;
		color: rgba(117, 50, 42, 0.63);
		transform: rotate(-3deg);
	}

	.article-seal span,
	.article-seal small {
		font-family: var(--font-mono);
		font-size: 0.48rem;
		letter-spacing: 0.1em;
	}

	.article-seal strong {
		font-family: var(--font-serif);
		font-size: 1rem;
	}

	.article-comments :global([data-comment-area]) {
		border-color: var(--book-rule) !important;
	}

	@media (max-width: 767px) {
		.short-with-photo .article-photo {
			position: static;
		}
		.article-topline {
			padding-right: 2.8rem;
		}

		.article-header h1 {
			font-size: clamp(2rem, 10.5vw, 3.15rem);
		}

		.article-content {
			margin-top: 1rem;
		}
		.article-body.has-photo {
			grid-template-columns: minmax(0, 1fr);
		}
		.article-photo {
			width: min(13rem, 48%);
			margin-top: 0.2rem;
		}
	}
</style>
