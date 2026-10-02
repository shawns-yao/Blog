<script lang="ts">
	import { goto } from '$app/navigation';
	import { ArrowRight, BookOpen, FileText, Search } from 'lucide-svelte';
	import type { Column } from '$lib/features/taxonomy/types';
	import type { MomentListResponse } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import { libraryPath } from './paths';

	let {
		columns,
		root,
		secondary,
		moments,
		query,
		reading
	}: {
		columns: Column[];
		root: Column | null;
		secondary: Column | null;
		moments: MomentListResponse;
		query: string;
		reading: boolean;
	} = $props();
	const children = $derived(columns.filter((item) => root && item.parentId === root.id));
	const roots = $derived(columns.filter((item) => !item.parentId));
	const colors = ['#396558', '#4b6079', '#87565b', '#69614d'];
	const pages = $derived(Math.max(1, Math.ceil(moments.total / moments.size)));
</script>

<section class="hierarchy" aria-labelledby="library-title">
	<header class="heading">
		<div>
			<p class="eyebrow">图书馆</p>
			<h1 id="library-title">{root?.name || '全部文章'}</h1>
		</div>
		<form action={resolvePath('/gallery/')} method="GET" role="search">
			{#if root}<input type="hidden" name="column" value={root.id} />{/if}
			<input
				name="q"
				type="search"
				value={query}
				aria-label="搜索馆内文章"
				placeholder="搜索馆内文章"
			/>
			<button aria-label="搜索" title="搜索" type="submit"><Search size={18} /></button>
		</form>
	</header>
	<nav
		aria-label="一级分类"
		class="mb-8 flex flex-wrap gap-x-5 gap-y-3 border-b border-ink-200 pb-5 text-sm dark:border-ink-700"
	>
		<a
			href={resolvePath('/gallery/')}
			aria-current={!root ? 'page' : undefined}
			class="rounded-default px-1 py-2 text-ink-600 hover:text-jade-700 focus-visible:outline-2 focus-visible:outline-jade-600 aria-[current=page]:text-jade-700 dark:text-ink-300 dark:aria-[current=page]:text-jade-400"
			>全部</a
		>
		{#each roots as category (category.id)}
			<a
				href={resolvePath(libraryPath({ column: category.id }))}
				aria-current={root?.id === category.id ? 'page' : undefined}
				class="rounded-default px-1 py-2 text-ink-600 hover:text-jade-700 focus-visible:outline-2 focus-visible:outline-jade-600 aria-[current=page]:text-jade-700 dark:text-ink-300 dark:aria-[current=page]:text-jade-400"
				>{category.name}</a
			>
		{/each}
	</nav>
	{#if root}
		<nav
			class="secondary-shelf"
			aria-label="二级分类"
			style:--wood={`url("${resolvePath('/library-shelf.svg')}")`}
		>
			{#each children as child, index (child.id)}
				<a
					class="category-book"
					style:--cover={colors[index % colors.length]}
					href={resolvePath(libraryPath({ column: child.id }))}
					aria-current={secondary?.id === child.id ? 'page' : undefined}
				>
					<span class="binding" aria-hidden="true"></span>
					<span class="book-name">{child.name}</span>
					<BookOpen size={24} strokeWidth={1.2} aria-hidden="true" />
					<span class="book-state">{secondary?.id === child.id ? '正在阅读' : root.name}</span>
				</a>
			{:else}
				<p class="no-children">暂无二级分类</p>
			{/each}
		</nav>
	{/if}
	{#if !reading}
		<section class="article-list" aria-label="按时间排序的文章">
			<header>
				<h2>{query ? '搜索结果' : secondary?.name || '最近收录'}</h2>
				<span>共 {moments.total} 篇</span>
			</header>
			{#if query}
				<p class="query">
					“{query}” <a href={resolvePath(libraryPath({ column: root?.id }))}>清除搜索</a>
				</p>
			{/if}
			{#each moments.items as article (article.id)}
				<a class="article-row" href={resolvePath(libraryPath({ read: article.shortUrl }))}>
					<FileText size={18} aria-hidden="true" />
					<div>
						<h3>{article.title}</h3>
						{#if article.summary}<p>{article.summary}</p>{/if}
					</div>
					<time datetime={article.createdAt}>{article.createdAt.slice(0, 10)}</time>
					<ArrowRight size={16} aria-hidden="true" />
				</a>
			{:else}
				<p class="empty">{query ? '没有找到相关文章' : '暂无公开文章'}</p>
			{/each}
			{#if pages > 1}
				<div class="pagination">
					<Pagination
						current={moments.page}
						total={pages}
						onPageChange={(page) =>
							goto(
								resolvePath(
									libraryPath({ column: secondary?.id ?? root?.id, page, q: query, view: 'list' })
								)
							)}
					/>
				</div>
			{/if}
		</section>
	{/if}
</section>

<style>
	.hierarchy {
		color: var(--color-ink-900);
	}
	.heading {
		display: flex;
		align-items: end;
		justify-content: space-between;
		gap: 28px;
		margin-bottom: 30px;
	}
	.eyebrow {
		font-size: 12px;
		color: var(--color-ink-500);
		margin-bottom: 10px;
	}
	h1 {
		font: 30px/1.5 var(--font-serif);
		overflow-wrap: anywhere;
	}
	form {
		display: flex;
		width: 270px;
		flex-shrink: 0;
		border-bottom: 1px solid var(--color-ink-400);
	}
	input {
		padding: 10px 0;
		min-width: 0;
		flex: 1;
		background: transparent;
		font-size: 13px;
	}
	button {
		padding: 10px;
	}
	.secondary-shelf {
		display: flex;
		gap: 30px;
		align-items: end;
		min-height: 240px;
		padding: 15px 28px 30px;
		margin-bottom: 34px;
		overflow-x: auto;
		position: relative;
		background-image: var(--wood);
		background-position: bottom;
		background-size: 100% 30px;
		background-repeat: no-repeat;
	}
	.category-book {
		position: relative;
		flex: 0 0 144px;
		height: 194px;
		padding: 26px 16px 18px 24px;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		align-items: center;
		background: var(--cover);
		color: #faf7ed;
		border-radius: 2px 4px 4px 2px;
		box-shadow:
			inset 6px 0 10px #0005,
			3px 1px 0 #cdc8bc,
			6px 2px 0 #645b4f,
			10px 8px 14px #0002;
		transition: transform 180ms ease;
	}
	.binding {
		position: absolute;
		inset: 10px 8px 10px 14px;
		border: 1px solid #eedcaa60;
		pointer-events: none;
	}
	.book-name {
		font: 19px/1.7 var(--font-serif);
		text-align: center;
		overflow-wrap: anywhere;
	}
	.book-state {
		font-size: 11px;
	}
	.category-book[aria-current],
	.category-book:hover {
		transform: translateY(-5px);
	}
	.category-book[aria-current] {
		outline: 2px solid var(--color-jade-600);
		outline-offset: 7px;
	}
	.no-children {
		align-self: center;
		font-size: 14px;
		color: var(--color-ink-500);
	}
	.article-list > header {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		padding-bottom: 18px;
		border-bottom: 1px solid var(--color-ink-200);
		gap: 20px;
	}
	h2 {
		font: 21px var(--font-serif);
	}
	.article-list header span,
	time {
		font-size: 12px;
		color: var(--color-ink-500);
	}
	.article-row {
		display: flex;
		align-items: center;
		gap: 18px;
		padding: 20px 0;
		border-bottom: 1px solid var(--color-ink-200);
	}
	.article-row > div {
		flex: 1;
		min-width: 0;
	}
	.article-row :global(svg),
	time {
		flex-shrink: 0;
	}
	h3 {
		font: 16px/1.7 var(--font-serif);
		overflow-wrap: anywhere;
	}
	.article-row p {
		font-size: 13px;
		color: var(--color-ink-500);
		margin-top: 5px;
		overflow-wrap: anywhere;
	}
	.article-row:hover h3,
	.query a {
		color: var(--color-jade-700);
	}
	.empty {
		padding: 60px 0;
		text-align: center;
		color: var(--color-ink-500);
	}
	.query {
		font-size: 14px;
		margin: 16px 0;
		overflow-wrap: anywhere;
	}
	.query a {
		text-decoration: underline;
		margin-left: 15px;
	}
	.pagination {
		display: flex;
		justify-content: center;
		padding: 28px 0;
	}
	a:focus-visible,
	input:focus-visible,
	button:focus-visible {
		outline: 2px solid var(--color-jade-500);
		outline-offset: 4px;
	}
	:global(.dark) .hierarchy {
		color: var(--color-ink-100);
	}
	:global(.dark) .article-list > header,
	:global(.dark) .article-row {
		border-color: var(--color-ink-800);
	}
	@media (max-width: 640px) {
		.heading {
			flex-direction: column;
			align-items: stretch;
		}
		form {
			width: 100%;
		}
		.article-row {
			flex-wrap: wrap;
			gap: 12px;
		}
		.article-row > div {
			flex-basis: calc(100% - 50px);
		}
		time {
			margin-left: 30px;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.category-book {
			transition: none;
		}
		.category-book:hover,
		.category-book[aria-current] {
			transform: none;
		}
	}
</style>
