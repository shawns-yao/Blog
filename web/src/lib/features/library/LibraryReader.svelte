<script lang="ts">
	import { goto } from '$app/navigation';
	import { ArrowLeft, BookOpen, FileText } from 'lucide-svelte';
	import type { Column } from '$lib/features/taxonomy/types';
	import type { MomentDetail, MomentListResponse } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { libraryPath } from './paths';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import MarkdownView from '$lib/shared/markdown/MarkdownView.svelte';
	import { flattenTOC } from '$lib/shared/types/toc';
	import DetailActionBar from '$lib/ui/detail/DetailActionBar.svelte';
	import DetailCommentSection from '$lib/ui/detail/DetailCommentSection.svelte';
	import ContentViewTracker from '$lib/features/analytics/components/ContentViewTracker.svelte';

	let {
		columns,
		column,
		moments,
		selected
	}: {
		columns: Column[];
		column: Column | null;
		moments: MomentListResponse;
		selected: MomentDetail | null;
	} = $props();
	const title = $derived(column?.name ?? '全部文章');
	const pages = $derived(Math.max(1, Math.ceil(moments.total / moments.size)));
</script>

<section class="library-reader">
	<header class="reader-heading">
		<a href={resolvePath(libraryPath({ column: column?.parentId ?? column?.id, view: 'list' }))}
			><ArrowLeft size={16} />返回时间列表</a
		>
		<label
			>切换分类
			<select
				value={column?.id ?? ''}
				onchange={(event) =>
					goto(
						resolvePath(libraryPath({ column: Number(event.currentTarget.value) || undefined }))
					)}
			>
				<option value="">分类书架</option>
				{#each columns as item (item.id)}<option value={item.id}>{item.name}</option>{/each}
			</select>
		</label>
	</header>
	<div class="open-book">
		<aside aria-label="分类文章目录">
			<div class="chapter-heading">
				<BookOpen size={22} strokeWidth={1.4} />
				<h1>{title}</h1>
				<span>{moments.total} 篇文章</span>
			</div>
			<nav aria-label={`${title}文章列表`}>
				{#if selected && !moments.items.some((item) => item.id === selected.id)}
					<a
						class="chapter"
						aria-current="page"
						href={resolvePath(
								libraryPath({ read: selected.shortUrl, page: moments.page })
						)}
					>
						<FileText size={16} /><span>{selected.title}</span>
					</a>
				{/if}
				{#each moments.items as article (article.id)}
					<a
						class="chapter"
						aria-current={selected?.id === article.id ? 'page' : undefined}
						href={resolvePath(
								libraryPath({ read: article.shortUrl, page: moments.page })
						)}
					>
						<FileText size={16} /><span>{article.title}</span>
					</a>
				{/each}
			</nav>
			{#if pages > 1}
				<div class="chapter-pagination">
					<Pagination
						current={moments.page}
						total={pages}
						onPageChange={(page) =>
							goto(
								resolvePath(
									libraryPath({
										column: column?.id,
										page,
										read: column ? undefined : selected?.shortUrl
									})
								)
							)}
					/>
				</div>
			{/if}
		</aside>
		<article class="reading-page">
			{#if selected}
				{#key selected.id}
					<div class="page-running">
						<span>{title}</span><time datetime={selected.createdAt}
							>{selected.createdAt.slice(0, 10)}</time
						>
					</div>
					<h2>{selected.title}</h2>
					{#if selected.summary}<p class="summary">{selected.summary}</p>{/if}
					{#if selected.toc?.length}
						<details class="article-toc">
							<summary>本文目录</summary>
							{#each selected.toc as heading}
								<a href={`#${heading.anchor}`}>{heading.name}</a>
							{/each}
						</details>
					{/if}
					{#if selected.cover}
						<img class="article-cover" src={selected.cover} alt={selected.title} loading="lazy" />
					{/if}
					<div class="prose-area">
						<MarkdownView
							content={selected.content}
							headingAnchors={flattenTOC(selected.toc ?? [])}
						/>
					</div>
					<DetailActionBar
						contentType="moment"
						contentId={selected.id}
						likes={selected.metrics?.likes ?? 0}
						comments={selected.metrics?.comments ?? 0}
						shareTitle={selected.title}
						shareDescription={selected.summary}
						shareImageUrl={selected.cover ?? ''}
					/>
					<DetailCommentSection
						commentAreaId={selected.commentAreaId}
						commentsCount={selected.metrics?.comments ?? 0}
						fediverseObjectUrl={selected.activityPubObjectId}
						containerClass="mt-12 border-t border-ink-200 pt-8 dark:border-ink-700"
					/>
					<ContentViewTracker contentType="moment" contentId={selected.id} />
				{/key}
			{:else}
				<div class="reader-empty">
					<BookOpen size={36} strokeWidth={1} />
					<h2>{title}</h2>
					<p>这个分类还没有公开文章</p>
				</div>
			{/if}
		</article>
	</div>
</section>

<style>
	.library-reader {
		max-width: 1280px;
		margin: 0 auto;
		color: var(--color-ink-900);
	}
	.reader-heading {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 20px;
		margin: 6px 0 26px;
		font-size: 13px;
	}
	.reader-heading a {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	label {
		display: flex;
		align-items: center;
		gap: 12px;
		color: var(--color-ink-500);
	}
	select {
		max-width: 240px;
		padding: 8px 32px 8px 12px;
		border: 1px solid var(--color-ink-300);
		border-radius: 3px;
		background: var(--color-ink-50);
		color: var(--color-ink-900);
	}
	.open-book {
		display: grid;
		grid-template-columns: 270px minmax(0, 1fr);
		border: 1px solid #c4baa5;
		border-bottom: 5px double #c4baa5;
		border-radius: 3px;
		background: #f5f0e5;
		box-shadow:
			0 12px 26px #00000012,
			3px 3px 0 #d8d0bf;
		min-height: 650px;
	}
	aside {
		align-self: start;
		position: sticky;
		top: 110px;
		max-height: calc(100vh - 135px);
		overflow: auto;
		padding: 30px 18px;
	}
	.chapter-heading {
		padding: 0 15px 24px;
		border-bottom: 1px solid #aaa3;
		margin-bottom: 16px;
	}
	h1 {
		font: 23px/1.5 var(--font-serif);
		margin: 14px 0 7px;
		overflow-wrap: anywhere;
	}
	.chapter-heading > span {
		color: var(--color-ink-500);
		font-size: 12px;
	}
	.chapter {
		display: flex;
		align-items: baseline;
		gap: 10px;
		padding: 12px 14px;
		border-left: 3px solid transparent;
		font-size: 14px;
		line-height: 1.7;
		overflow-wrap: anywhere;
	}
	.chapter :global(svg) {
		flex-shrink: 0;
	}
	.chapter:hover {
		background: #68847510;
	}
	.chapter[aria-current] {
		border-left-color: var(--color-jade-600);
		background: #68847520;
		color: var(--color-jade-800);
	}
	.chapter-pagination {
		padding-top: 24px;
		overflow-x: auto;
	}
	.reading-page {
		min-width: 0;
		padding: 28px 52px 56px;
		border-left: 1px solid #c2bdae;
		box-shadow: inset 12px 0 18px -14px #51402b80;
	}
	.page-running {
		display: flex;
		justify-content: space-between;
		gap: 20px;
		border-bottom: 1px solid #aaa4;
		padding-bottom: 14px;
		margin-bottom: 34px;
		color: var(--color-ink-500);
		font-size: 11px;
	}
	h2 {
		font: 30px/1.5 var(--font-serif);
		overflow-wrap: anywhere;
	}
	.summary {
		font-size: 15px;
		line-height: 1.9;
		color: var(--color-ink-500);
		margin-top: 14px;
	}
	.article-toc {
		border-block: 1px solid #aaa3;
		padding: 12px 0;
		margin-top: 24px;
		font-size: 13px;
	}
	.article-toc summary {
		cursor: pointer;
	}
	.article-toc a {
		display: block;
		margin: 10px 12px;
		color: var(--color-jade-700);
	}
	.prose-area {
		margin-top: 36px;
		font-size: 16px;
		line-height: 1.9;
	}
	.article-cover {
		display: block;
		width: 100%;
		max-height: 380px;
		object-fit: contain;
		margin-top: 28px;
	}
	.prose-area :global([id]) {
		scroll-margin-top: 110px;
	}
	.reader-empty {
		min-height: 500px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 22px;
	}
	.reader-empty p {
		font-size: 14px;
		color: var(--color-ink-500);
	}
	a:focus-visible,
	select:focus-visible,
	summary:focus-visible {
		outline: 2px solid var(--color-jade-600);
		outline-offset: 3px;
	}
	:global(.dark) .library-reader {
		color: var(--color-ink-100);
	}
	:global(.dark) .open-book {
		background: var(--color-ink-950);
		border-color: var(--color-ink-800);
	}
	:global(.dark) .reading-page {
		border-color: #555d56;
	}
	:global(.dark) select {
		background: var(--color-ink-900);
		color: var(--color-ink-100);
		border-color: var(--color-ink-700);
	}
	:global(.dark) .chapter[aria-current] {
		color: var(--color-jade-300);
	}
	@media (max-width: 900px) {
		.open-book {
			grid-template-columns: 220px minmax(0, 1fr);
		}
		.reading-page {
			padding: 28px;
		}
	}
	@media (max-width: 640px) {
		.open-book {
			grid-template-columns: minmax(0, 1fr);
		}
		aside {
			position: static;
			max-height: 230px;
			border-bottom: 1px solid #aaa4;
		}
		.reading-page {
			border-left: 0;
			padding: 24px 18px;
		}
		.reader-heading {
			flex-wrap: wrap;
		}
	}
</style>
