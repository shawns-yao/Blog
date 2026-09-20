<script lang="ts">
	import { goto } from '$app/navigation';
	import { Search, ArrowUpRight, BookOpen } from 'lucide-svelte';
	import type { Column } from '$lib/features/taxonomy/types';
	import type { MomentListResponse } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import { libraryPath } from './paths';

	let {
		columns,
		moments,
		query
	}: { columns: Column[]; moments: MomentListResponse; query: string } = $props();
	const colors = ['#345d52', '#6f4346', '#3e566d', '#6a6250', '#544567'];
	const totalPages = $derived(Math.max(1, Math.ceil(moments.total / moments.size)));
</script>

<section class="library-shelf">
	<header class="shelf-heading">
		<div>
			<p class="eyebrow">LIBRARIUM</p>
			<h1>图书馆</h1>
		</div>
		<form action={resolvePath('/gallery/')} method="GET" role="search">
			<Search size={18} aria-hidden="true" />
			<input
				name="q"
				value={query}
				type="search"
				aria-label="搜索馆内文章"
				placeholder="搜索馆内文章"
			/>
			<button type="submit" aria-label="搜索" title="搜索"><ArrowUpRight size={19} /></button>
		</form>
	</header>
	{#if !query}
		{#if columns.length}
			<nav class="category-shelf" aria-label="分类书架">
				{#each columns as column, index (column.id)}
					<a
						class="category-book"
						style:--cover={colors[index % colors.length]}
						href={resolvePath(libraryPath({ column: column.id }))}
					>
						<span class="book-cover">
							<span class="book-number">{String(index + 1).padStart(2, '0')}</span>
							<span class="book-title">{column.name}</span>
							<BookOpen size={26} strokeWidth={1} />
							<span class="book-author">图书馆 · 主题文集</span>
						</span>
						<span class="book-foot">{column.name}<ArrowUpRight size={15} /></span>
					</a>
				{/each}
			</nav>
		{:else}
			<p class="empty">暂无主题分类</p>
		{/if}
	{/if}
	<section class="recent">
		<header>
			<h2>{query ? '搜索结果' : '最近收录'}</h2>
			<span>{moments.total} 篇</span>
		</header>
		{#if moments.items.length}
			{#each moments.items as article (article.id)}
				<a class="article-row" href={resolvePath(libraryPath({ read: article.shortUrl }))}>
					<div>
						<h3>{article.title}</h3>
						{#if article.summary}<p>{article.summary}</p>{/if}
					</div>
					<time datetime={article.createdAt}>{article.createdAt.slice(0, 10)}</time>
					<ArrowUpRight size={18} />
				</a>
			{/each}
			{#if totalPages > 1}
				<div class="pagination">
					<Pagination
						current={moments.page}
						total={totalPages}
						onPageChange={(page) => goto(resolvePath(libraryPath({ page, q: query })))}
					/>
				</div>
			{/if}
		{:else}
			<p class="empty">{query ? '没有找到相关文章' : '暂无文章'}</p>
		{/if}
		{#if query}<a class="clear" href={resolvePath('/gallery/')}>返回分类书架</a>{/if}
	</section>
</section>

<style>
	.library-shelf {
		max-width: 1180px;
		margin: 0 auto;
		color: var(--color-ink-900);
	}
	.shelf-heading {
		display: flex;
		align-items: end;
		justify-content: space-between;
		gap: 32px;
		margin: 12px 0 46px;
	}
	.eyebrow {
		font-size: 11px;
		color: var(--color-jade-700);
		margin-bottom: 10px;
	}
	h1 {
		font: 36px var(--font-serif);
	}
	h2 {
		font: 21px var(--font-serif);
	}
	form {
		display: flex;
		align-items: center;
		gap: 10px;
		border-bottom: 1px solid var(--color-ink-300);
		padding: 10px 0;
		width: 300px;
	}
	input {
		min-width: 0;
		width: 100%;
		background: transparent;
		border: 0;
		outline: none;
		font-size: 14px;
	}
	form:focus-within {
		border-color: var(--color-jade-600);
	}
	form button {
		padding: 4px;
	}
	.category-shelf {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(172px, 1fr));
		gap: 40px 34px;
		padding: 12px 16px 35px;
		border-bottom: 10px solid #685649;
		box-shadow: 0 9px 10px -8px #0008;
	}
	.category-book {
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.book-cover {
		position: relative;
		min-height: 250px;
		aspect-ratio: 0.72;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 24px 19px 24px 27px;
		background: var(--cover);
		color: #f2eee3;
		border-radius: 2px 5px 5px 2px;
		border: 1px solid #ffffff30;
		box-shadow:
			inset 7px 0 9px #0005,
			inset 10px 0 0 #ffffff12,
			5px 3px 0 #d8d4ca,
			8px 5px 0 #817970,
			10px 10px 16px #0002;
		transition: translate 180ms ease;
	}
	.book-cover::after {
		content: '';
		position: absolute;
		inset: 13px 11px 13px 19px;
		border: 1px solid #dfc58c70;
		pointer-events: none;
	}
	.category-book:hover .book-cover {
		translate: 0 -5px;
	}
	.book-number {
		font-size: 11px;
		opacity: 0.7;
	}
	.book-title {
		font: 23px/1.6 var(--font-serif);
		text-align: center;
		overflow-wrap: anywhere;
	}
	.book-author {
		font-size: 10px;
		opacity: 0.75;
	}
	.book-foot {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		font-size: 13px;
	}
	.book-foot :global(svg) {
		flex-shrink: 0;
	}
	.recent {
		margin-top: 52px;
	}
	.recent > header {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		padding-bottom: 16px;
		border-bottom: 1px solid var(--color-ink-300);
	}
	.recent > header span,
	time {
		font-size: 12px;
		color: var(--color-ink-500);
	}
	.article-row {
		display: flex;
		align-items: center;
		gap: 24px;
		padding: 23px 0;
		border-bottom: 1px solid var(--color-ink-200);
	}
	.article-row > div {
		flex: 1;
		min-width: 0;
	}
	h3 {
		font: 18px/1.7 var(--font-serif);
		overflow-wrap: anywhere;
	}
	.article-row p {
		margin-top: 7px;
		font-size: 13px;
		color: var(--color-ink-500);
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.article-row time,
	.article-row :global(svg) {
		flex-shrink: 0;
	}
	.article-row:hover h3,
	.clear {
		color: var(--color-jade-700);
	}
	.empty {
		padding: 48px 0;
		color: var(--color-ink-500);
		text-align: center;
		font-size: 14px;
	}
	.pagination {
		display: flex;
		justify-content: center;
		padding-top: 24px;
	}
	a:focus-visible,
	button:focus-visible {
		outline: 2px solid var(--color-jade-600);
		outline-offset: 6px;
	}
	:global(.dark) .library-shelf {
		color: var(--color-ink-100);
	}
	:global(.dark) .article-row,
	:global(.dark) .recent > header {
		border-color: var(--color-ink-700);
	}
	@media (prefers-reduced-motion: reduce) {
		.book-cover {
			transition: none;
		}
		.category-book:hover .book-cover {
			translate: none;
		}
	}
</style>
