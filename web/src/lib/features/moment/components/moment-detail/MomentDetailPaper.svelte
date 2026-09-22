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
		dateNo: string;
		onActiveAnchorChange: (anchor: string | null) => void;
		onContentRootChange: (node: HTMLElement | null) => void;
	}

	let { moment, dateStr, dateNo, onActiveAnchorChange, onContentRootChange }: Props = $props();
	const showUpdated = $derived(isDifferentDay(moment.createdAt, moment.contentUpdatedAt));
</script>

<div class="moment-detail-paper moment-vt" style:view-transition-name={`moment-${moment.id}`}>
	<header class="article-header">
		<div class="article-meta-line">
			<div class="article-meta">
				<span>NO. {dateNo}</span>
				<i aria-hidden="true"></i>
				<strong>{moment.contentKind === 'article' ? '图书馆' : '手记'}</strong>
				<i aria-hidden="true"></i>
				<time datetime={moment.createdAt}>{dateStr}</time>
				{#if showUpdated}
					<small>更新于 {formatDateCN(moment.contentUpdatedAt)}</small>
				{/if}
			</div>
			<MomentAtmosphere atmosphere={moment.extInfo?.moment} />
		</div>

		{#if moment.title}
			<h1>{moment.title}</h1>
		{/if}

		{#if moment.summary}
			<p class="article-deck">{moment.summary}</p>
		{/if}

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

		<div class="article-tags"><TagList tags={moment.topics ?? []} /></div>
	</header>

	{#if moment.aiSummary}
		<div class="article-summary"><DetailAiSummary summary={moment.aiSummary} /></div>
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
		min-height: 70vh;
		color: var(--book-ink);
	}

	.article-header {
		padding-bottom: clamp(2.4rem, 5vw, 4.5rem);
		border-bottom: 1px solid var(--book-rule);
	}

	.article-meta-line {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--book-rule);
	}

	.article-meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.55rem;
		font-family: var(--font-mono);
		font-size: 0.57rem;
		letter-spacing: 0.08em;
		color: var(--book-faint);
	}

	.article-meta i,
	.article-facts i {
		display: inline-block;
		width: 1.2rem;
		height: 1px;
		background: var(--book-rule);
	}

	.article-meta strong {
		font-family: var(--font-serif);
		font-weight: 600;
		color: var(--book-accent);
	}

	.article-meta small {
		font-size: inherit;
	}

	.article-header h1 {
		max-width: 55rem;
		margin-top: clamp(2.4rem, 5vw, 4.6rem);
		font-family: var(--font-serif);
		font-size: clamp(2.1rem, 4.7vw, 4.7rem);
		font-weight: 600;
		line-height: 1.32;
		letter-spacing: 0.045em;
		text-wrap: balance;
	}

	.article-deck {
		max-width: 46rem;
		margin-top: 1.5rem;
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
		margin-top: 1.6rem;
		font-family: var(--font-mono);
		font-size: 0.57rem;
		letter-spacing: 0.11em;
		color: var(--book-faint);
	}

	.article-tags {
		margin-top: 1rem;
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

	.article-content {
		max-width: 54rem;
		margin: clamp(2.8rem, 6vw, 5.6rem) auto 0;
	}

	.article-content :global(.markdown-preview) {
		color: var(--book-ink) !important;
		font-size: clamp(0.94rem, 1.2vw, 1.04rem);
		line-height: 2.15;
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
		display: flex;
		width: 5.4rem;
		height: 5.4rem;
		margin: 5rem auto 0;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		color: rgba(117, 50, 42, 0.63);
		border: 1px solid currentColor;
		border-radius: 50%;
		outline: 1px dashed rgba(117, 50, 42, 0.38);
		outline-offset: -0.35rem;
		transform: rotate(8deg);
	}

	.article-seal span,
	.article-seal small {
		font-family: var(--font-mono);
		font-size: 0.48rem;
		letter-spacing: 0.1em;
	}

	.article-seal strong {
		margin: 0.18rem 0;
		font-family: var(--font-serif);
		font-size: 1.25rem;
	}

	.article-comments :global([data-comment-area]) {
		border-color: var(--book-rule) !important;
	}

	@media (max-width: 767px) {
		.article-meta-line {
			padding-right: 2.8rem;
		}

		.article-header h1 {
			font-size: clamp(2rem, 10.5vw, 3.15rem);
		}

		.article-content {
			margin-top: 3rem;
		}
	}
</style>
