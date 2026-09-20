<script lang="ts">
	import { goto } from '$app/navigation';
	import { ArrowLeft, ArrowRight, BookOpen, Search } from 'lucide-svelte';
	import type { Column } from '$lib/features/taxonomy/types';
	import type { MomentListResponse } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import LibraryBook from './LibraryBook.svelte';
	import { libraryPath } from './paths';

	let {
		columns,
		column,
		moments,
		query
	}: {
		columns: Column[];
		column: Column | null;
		moments: MomentListResponse;
		query: string;
	} = $props();
	const overview = $derived(!column && !query);
	const pages = $derived(Math.max(1, Math.ceil(moments.total / moments.size)));
	const rows = $derived.by(() => {
		const groups = new Map<
			string,
			{ name: string; columnId?: number; items: typeof moments.items }
		>();
		for (const item of moments.items) {
			const match = columns.find((candidate) => candidate.shortUrl === item.columnShortUrl);
			const key = match ? String(match.id) : item.columnShortUrl || '';
			if (!groups.has(key))
				groups.set(key, {
					name: match?.name || item.columnName || '未分类',
					columnId: match?.id,
					items: []
				});
			groups.get(key)!.items.push(item);
		}
		return [...groups.values()];
	});
</script>

<section class="collection" class:overview>
	{#if overview}
		<header class="library-scene">
			<img class="day" src={resolvePath('/bg.png')} alt="" />
			<img class="night" src={resolvePath('/bg-night.png')} alt="" />
			<div class="scene-title">
				<BookOpen size={34} strokeWidth={1.3} />
				<h1>图书馆</h1>
				<p>一本书，一个新的世界。</p>
			</div>
		</header>
	{:else}
		<a class="back" href={resolvePath('/gallery/')}><ArrowLeft size={16} />返回图书馆</a>
		<header class="collection-heading">
			<h1>{column?.name || '搜索结果'}</h1>
			<p>共 {moments.total} 篇文章</p>
		</header>
	{/if}

	<div class="collection-tools">
		<nav aria-label="分类书架">
			<a href={resolvePath('/gallery/')} aria-current={!column ? 'page' : undefined}>全部书架</a>
			{#each columns as item (item.id)}
				<a
					href={resolvePath(libraryPath({ column: item.id }))}
					aria-current={column?.id === item.id ? 'page' : undefined}>{item.name}</a
				>
			{/each}
		</nav>
		<form action={resolvePath('/gallery/')} method="GET" role="search">
			{#if column}<input type="hidden" name="column" value={column.id} />{/if}
			<input
				name="q"
				type="search"
				value={query}
				aria-label={column ? '搜索本书架文章' : '搜索馆内文章'}
				placeholder={column ? '搜索本书架文章' : '搜索馆内文章'}
			/>
			<button type="submit" aria-label="搜索" title="搜索"><Search size={18} /></button>
		</form>
	</div>
	{#if query}<p class="query">
			“{query}” <a href={resolvePath(libraryPath({ column: column?.id }))}>清除搜索</a>
		</p>{/if}

	{#if overview}
		<div class="section-heading">
			<h2>最近收录</h2>
			<span>{moments.total} 篇文章</span>
		</div>
		{#each rows as row}
			<section class="shelf-row" style:--shelf-art={`url("${resolvePath('/library-shelf.svg')}")`}>
				<header>
					<h2><BookOpen size={21} />{row.name}</h2>
					{#if row.columnId}<a href={resolvePath(libraryPath({ column: row.columnId }))}
							>查看全部<ArrowRight size={15} /></a
						>{/if}
				</header>
				<div class="book-grid" role="region" aria-label={`${row.name}书籍`}>
					{#each row.items as article (article.id)}<LibraryBook
							{article}
							columnId={row.columnId}
							onShelf
						/>{/each}
				</div>
			</section>
		{/each}
	{:else if moments.items.length}
		<div class="book-grid catalog">
			{#each moments.items as article (article.id)}<LibraryBook
					{article}
					columnId={column?.id}
				/>{/each}
		</div>
	{/if}
	{#if !moments.items.length}
		<div class="empty">
			<BookOpen size={38} strokeWidth={1} />
			<h2>{query ? '没有找到相关文章' : '书架还没有公开文章'}</h2>
			{#if !columns.length && overview}<p>暂无主题分类</p>{/if}
		</div>
	{/if}
	{#if pages > 1}
		<div class="pagination">
			<Pagination
				current={moments.page}
				total={pages}
				onPageChange={(page) =>
					goto(resolvePath(libraryPath({ column: column?.id, page, q: query })))}
			/>
		</div>
	{/if}
</section>

<style>
	.collection {
		max-width: 1180px;
		margin: auto;
		color: var(--color-ink-900);
	}
	.library-scene {
		position: relative;
		min-height: 280px;
		display: flex;
		align-items: center;
		overflow: hidden;
	}
	.library-scene > img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: center 42%;
	}
	.library-scene .night {
		display: none;
	}
	.scene-title {
		position: relative;
		padding: 44px;
		color: white;
		text-shadow:
			0 2px 5px #000,
			0 0 20px #000;
	}
	h1 {
		font: 34px/1.4 var(--font-serif);
		margin: 10px 0;
	}
	.scene-title p {
		font: 15px var(--font-serif);
	}
	.collection-tools {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		padding: 26px 0;
		border-bottom: 1px solid var(--color-ink-200);
	}
	nav {
		display: flex;
		flex-wrap: wrap;
		gap: 8px 22px;
		min-width: 0;
	}
	nav a {
		font-size: 14px;
		padding: 6px 0;
		border-bottom: 2px solid transparent;
		overflow-wrap: anywhere;
	}
	nav a[aria-current] {
		color: var(--color-jade-700);
		border-color: var(--color-jade-600);
	}
	form {
		display: flex;
		width: 250px;
		flex-shrink: 0;
		align-items: center;
		border: 1px solid var(--color-ink-300);
		border-radius: 4px;
	}
	input {
		min-width: 0;
		width: 100%;
		background: transparent;
		padding: 10px;
		font-size: 13px;
	}
	button {
		padding: 10px;
	}
	.section-heading,
	.shelf-row header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
	}
	.section-heading {
		padding: 30px 0 12px;
	}
	h2 {
		font: 21px/1.5 var(--font-serif);
	}
	.section-heading span,
	.collection-heading p {
		font-size: 13px;
		color: var(--color-ink-500);
	}
	.shelf-row {
		position: relative;
		isolation: isolate;
		margin: 0;
		padding: 26px 38px 20px;
		color: #fffaf2;
		box-shadow: 0 14px 18px -14px #0009;
	}
	.shelf-row::before {
		content: '';
		position: absolute;
		inset: 0;
		z-index: -1;
		background-image: var(--shelf-art);
		background-size: 100% 100%;
		pointer-events: none;
	}
	.shelf-row .book-grid {
		display: flex;
		align-items: end;
		gap: 32px;
		overflow-x: auto;
		padding: 8px 12px 15px 2px;
		scrollbar-width: thin;
		scrollbar-color: #bd9c76 #3a2d23;
	}
	.shelf-row .book-grid :global(.library-book) {
		flex: 0 0 158px;
	}
	.shelf-row .book-grid:focus-visible {
		outline: 2px solid #f4dfb4;
		outline-offset: 3px;
	}
	.shelf-row header {
		text-shadow: 0 2px 4px #000a;
	}
	.shelf-row header {
		margin-bottom: 25px;
	}
	.shelf-row h2,
	.shelf-row header a,
	.back {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.shelf-row header a,
	.back {
		font-size: 13px;
	}
	.book-grid {
		display: grid;
		grid-template-columns: repeat(6, minmax(0, 1fr));
		gap: 32px 26px;
	}
	.catalog {
		padding: 34px 0;
		grid-template-columns: repeat(5, minmax(0, 1fr));
		gap: 36px;
	}
	.back {
		margin-bottom: 24px;
	}
	.collection-heading {
		padding-bottom: 12px;
	}
	.query {
		padding-top: 20px;
		overflow-wrap: anywhere;
		font-size: 14px;
	}
	.query a {
		margin-left: 14px;
		text-decoration: underline;
	}
	.empty {
		min-height: 230px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 15px;
		color: var(--color-ink-500);
	}
	.empty h2 {
		font-size: 18px;
	}
	.empty p {
		font-size: 13px;
	}
	.pagination {
		display: flex;
		justify-content: center;
		padding: 28px 0;
	}
	a:focus-visible,
	button:focus-visible,
	input:focus-visible {
		outline: 2px solid var(--color-jade-500);
		outline-offset: 4px;
	}
	:global(.dark) .collection {
		color: var(--color-ink-100);
	}
	:global(.dark) .shelf-row::before {
		filter: brightness(0.62) saturate(0.8);
	}
	:global(.dark) .collection-tools {
		border-color: var(--color-ink-800);
	}
	:global(.dark) nav a[aria-current] {
		color: var(--color-jade-400);
	}
	:global(.dark) .library-scene .day {
		display: none;
	}
	:global(.dark) .library-scene .night {
		display: block;
	}
	@media (max-width: 1000px) {
		.book-grid,
		.catalog {
			grid-template-columns: repeat(4, minmax(0, 1fr));
		}
	}
	@media (max-width: 640px) {
		.collection-tools {
			align-items: stretch;
			flex-direction: column;
		}
		form {
			width: 100%;
		}
		.book-grid,
		.catalog {
			grid-template-columns: repeat(2, minmax(0, 1fr));
			gap: 24px;
		}
		.shelf-row {
			padding: 20px 24px 20px;
		}
		.scene-title {
			padding: 26px;
		}
	}
</style>
