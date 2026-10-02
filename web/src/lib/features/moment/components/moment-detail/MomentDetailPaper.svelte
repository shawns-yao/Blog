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
		preview?: boolean;
		dateStr: string;
		onActiveAnchorChange?: (anchor: string | null) => void;
		onContentRootChange?: (node: HTMLElement | null) => void;
	}

	let {
		moment,
		preview = false,
		dateStr,
		onActiveAnchorChange = () => {},
		onContentRootChange = () => {}
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

<div class="moment-detail-paper moment-vt" style:view-transition-name={`moment-${moment.id}`}>
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
				{#if preview}<small>版式预览</small>{/if}
			</div>
		</div>

		{#if moment.title}
			<h1>{moment.title}</h1>
		{/if}

		{#if moment.summary && moment.contentKind === 'article' && !preview}
			<p class="article-deck">{moment.summary}</p>
		{/if}

		{#if moment.contentKind === 'article' && !preview}
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

	<div class="article-body">
		{#if moment.cover}
			<figure class="article-photo">
				<img src={moment.cover} alt={moment.title || '手记照片'} loading="lazy" />
			</figure>
		{/if}
		<div class="article-content">
			<DetailMarkdownContent
				content={moment.content}
				toc={moment.toc}
				className="max-w-none font-serif text-justify text-[15px]"
				{onContentRootChange}
				{onActiveAnchorChange}
			/>
		</div>
	</div>

	{#if moment.topics?.length}
		<div class="article-tags"><TagList tags={moment.topics} /></div>
	{/if}

	{#if !preview}<div class="article-actions">
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
		</div>{/if}

	<div class="article-seal" aria-hidden="true">
		<span>{moment.contentKind === 'article' ? '图书馆' : '手记'}</span>
		<strong>{moment.contentKind === 'article' ? '文' : '记'}</strong>
		<small>{dateStr}</small>
	</div>

	{#if !preview}<div class="article-comments">
			<DetailCommentSection
				commentAreaId={moment.commentAreaId}
				commentsCount={moment.metrics?.comments ?? 0}
				fediverseObjectUrl={moment.activityPubObjectId}
				containerClass="mt-8 pt-6 border-t"
				fallbackText="正在展开评论……"
				fallbackSize="w-6 h-6"
				fallbackContainerClass="flex justify-center py-20"
			/>
		</div>{/if}
</div>

<style>
	.moment-detail-paper {
		position: relative;
		flex: 0 0 auto;
		max-width: 42rem;
		width: 100%;
		margin-inline: auto;
		color: var(--book-ink);
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
		align-items: baseline;
		gap: 0.45rem;
		line-height: 1;
		color: var(--book-ink);
	}
	.article-date span,
	.article-date small {
		font-family: var(--font-mono);
		font-size: 0.68rem;
		letter-spacing: 0.1em;
		color: var(--book-faint);
	}
	.article-date strong {
		margin: 0;
		font-family: var(--font-serif);
		font-size: 1.65rem;
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
		font-size: 0.72rem;
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
		font-size: clamp(1.6rem, 2.7vw, 2.15rem);
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
		display: flow-root;
		margin-top: 1.6rem;
		padding: 0.5rem 0.4rem 0.75rem 0;
	}
	.article-content {
		min-width: 0;
		width: 100%;
	}

	.article-content :global(.markdown-preview) {
		color: var(--book-ink) !important;
		font-size: clamp(0.94rem, 1.2vw, 1.04rem);
		line-height: 1.9;
	}
	.article-content :global(.markdown-preview p) {
		margin-bottom: 0.7rem;
	}
	.article-content :global(.markdown-preview p:last-child) {
		margin-bottom: 0;
	}
	.article-photo,
	.article-content :global(.md-figure) {
		float: right;
		width: min(48%, 19rem);
		margin: 0.25rem 0.6rem 1.25rem 1.5rem;
		padding: 0.45rem 0.45rem 0.9rem;
		border: 1px solid rgba(97, 82, 62, 0.1);
		background: #fffdf8;
		box-shadow: 0.2rem 0.45rem 0.9rem rgba(63, 42, 25, 0.14);
		transform: rotate(4deg);
	}
	.article-photo img {
		display: block;
		width: 100%;
		height: auto;
	}
	.article-content :global(.md-figure .md-caption) {
		font-size: 0.72rem;
		line-height: 1.6;
		text-align: left;
	}
	.article-content :global(.markdown-preview :where(pre, table)) {
		clear: both;
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
		width: 100%;
		margin-top: 1.25rem;
	}

	.article-actions :global(> div) {
		margin-top: 0;
		border-bottom: 0;
		padding-block: 0.75rem;
		border-color: var(--book-rule) !important;
	}

	.article-actions :global(button),
	.article-actions :global(a) {
		min-height: 2.75rem;
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
		margin: 0.5rem 0 0;
		color: rgba(117, 50, 42, 0.63);
		transform: rotate(-3deg);
	}

	.article-seal span,
	.article-seal small {
		font-family: var(--font-mono);
		font-size: 0.62rem;
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
		.article-topline {
			gap: 0.75rem;
		}

		.article-header h1 {
			font-size: clamp(1.5rem, 6vw, 1.9rem);
		}

		.article-body {
			margin-top: 1.25rem;
		}
		.article-photo,
		.article-content :global(.md-figure) {
			width: 46%;
			margin: 0.25rem 0.4rem 0.8rem 0.9rem;
			padding: 0.3rem 0.3rem 0.65rem;
			transform: rotate(3deg);
		}
	}
</style>
