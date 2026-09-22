<script lang="ts">
	import type { MomentDetail } from '$lib/features/moment/types';
	import { detailHeroBgSrc } from '$lib/shared/stores/detailHeroBg';
	import { formatDateCompact, formatDateDotted } from '$lib/shared/utils/date';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import DetailTocNavList from '$lib/ui/detail/DetailTocNavList.svelte';
	import { ArrowLeft } from 'lucide-svelte';
	import { onDestroy } from 'svelte';
	import MomentBookShell from './MomentBookShell.svelte';
	import MomentDetailPaper from './moment-detail/MomentDetailPaper.svelte';
	import MomentDetailRelatedMoments from './moment-detail/MomentDetailRelatedMoments.svelte';

	let { moment }: { moment: MomentDetail } = $props();

	const dateStr = $derived(formatDateDotted(moment.createdAt));
	const dateNo = $derived(formatDateCompact(moment.createdAt));
	const columnLabel = $derived((moment.columnName || '').trim() || '未分类手记');
	const toc = $derived(moment.toc ?? []);

	$effect(() => {
		detailHeroBgSrc.set(moment.cover ?? '');
	});
	onDestroy(() => detailHeroBgSrc.set(''));

	let contentRoot: HTMLElement | null = $state(null);
	let activeAnchor: string | null = $state(null);

	const handleContentRootChange = (node: HTMLElement | null) => {
		contentRoot = node;
	};

	const handleActiveAnchorChange = (anchor: string | null) => {
		activeAnchor = anchor;
	};
</script>

<MomentBookShell pageLabel={moment.title || '手记详情'}>
	{#snippet directory()}
		<nav class="detail-directory" aria-label="本篇手记目录">
			<p class="directory-kicker">CONTENTS</p>
			<h2>目录</h2>

			<a class="back-to-index" href={resolvePath('/moments')}>
				<ArrowLeft size={14} strokeWidth={1.6} aria-hidden="true" />
				<span>返回全部手记</span>
			</a>

			<div class="current-entry">
				<span>正在阅读</span>
				<strong>{moment.title || '无题手记'}</strong>
				<small>{dateStr} · {columnLabel}</small>
			</div>

			{#if toc.length > 0}
				<section class="toc-section" aria-labelledby="moment-toc-heading">
					<h3 id="moment-toc-heading">本页段落</h3>
					<DetailTocNavList
						{toc}
						{contentRoot}
						{activeAnchor}
						onAnchorChange={handleActiveAnchorChange}
						tone="ink"
						size="md"
					/>
				</section>
			{/if}

			<div class="related-section">
				<MomentDetailRelatedMoments />
			</div>
		</nav>
	{/snippet}

	<article class="moment-detail-sheet" style="view-transition-name: moment-sheet">
		<div class="page-ribbon" aria-label={`所属栏目：${columnLabel}`}>
			<span>{columnLabel}</span>
		</div>
		<MomentDetailPaper
			{moment}
			{dateStr}
			{dateNo}
			onContentRootChange={handleContentRootChange}
			onActiveAnchorChange={handleActiveAnchorChange}
		/>
	</article>
</MomentBookShell>

<style>
	.detail-directory {
		font-family: var(--font-serif);
	}

	.directory-kicker {
		font-family: var(--font-mono);
		font-size: 0.58rem;
		letter-spacing: 0.28em;
		color: var(--book-faint);
	}

	.detail-directory h2 {
		margin-top: 0.8rem;
		font-size: clamp(1.8rem, 3vw, 2.45rem);
		font-weight: 500;
		letter-spacing: 0.08em;
	}

	.back-to-index {
		display: flex;
		align-items: center;
		gap: 0.45rem;
		margin-top: 2rem;
		padding: 0.8rem 0;
		font-size: 0.72rem;
		letter-spacing: 0.08em;
		color: var(--book-accent);
		border-top: 1px solid var(--book-rule);
		border-bottom: 1px solid var(--book-rule);
	}

	.back-to-index:hover :global(svg) {
		transform: translateX(-0.18rem);
	}

	.back-to-index :global(svg) {
		transition: transform 160ms ease;
	}

	.current-entry {
		display: flex;
		flex-direction: column;
		gap: 0.55rem;
		margin-top: 2rem;
		padding-left: 0.85rem;
		border-left: 2px solid var(--book-accent);
	}

	.current-entry span,
	.toc-section h3 {
		font-size: 0.62rem;
		letter-spacing: 0.17em;
		color: var(--book-faint);
	}

	.current-entry strong {
		font-size: 0.84rem;
		font-weight: 600;
		line-height: 1.7;
		color: var(--book-ink);
	}

	.current-entry small {
		font-family: var(--font-mono);
		font-size: 0.52rem;
		line-height: 1.7;
		color: var(--book-faint);
	}

	.toc-section,
	.related-section {
		margin-top: 2.25rem;
		padding-top: 1.2rem;
		border-top: 1px solid var(--book-rule);
	}

	.toc-section h3 {
		margin-bottom: 1rem;
	}

	.detail-directory :global(.custom-scrollbar) {
		max-height: 19rem;
	}

	.detail-directory :global(.custom-scrollbar a) {
		font-family: var(--font-serif);
		color: var(--book-muted) !important;
	}

	.detail-directory :global(.custom-scrollbar a:hover),
	.detail-directory :global(.custom-scrollbar a.font-bold) {
		color: var(--book-accent) !important;
	}

	.moment-detail-sheet {
		position: relative;
		min-width: 0;
	}

	.page-ribbon {
		position: absolute;
		top: calc(clamp(2.4rem, 5vw, 4.2rem) * -1);
		right: clamp(0.5rem, 2vw, 2rem);
		z-index: 4;
		display: flex;
		width: 2.6rem;
		min-height: 6.7rem;
		align-items: center;
		justify-content: center;
		padding: 0.8rem 0.4rem 1rem;
		color: #efe2c8;
		background: #7c332c;
		box-shadow: 0 0.55rem 1.2rem rgba(69, 36, 27, 0.2);
		clip-path: polygon(0 0, 100% 0, 100% 88%, 50% 100%, 0 88%);
	}

	.page-ribbon span {
		font-family: var(--font-serif);
		font-size: 0.64rem;
		line-height: 1.4;
		letter-spacing: 0.16em;
		writing-mode: vertical-rl;
	}

	@media (max-width: 767px) {
		.page-ribbon {
			top: -2.5rem;
			right: 0;
			width: 2.25rem;
			min-height: 5.8rem;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.back-to-index :global(svg) {
			transition: none;
		}
	}
</style>
