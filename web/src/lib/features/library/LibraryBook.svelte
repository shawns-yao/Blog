<script lang="ts">
	import type { MomentSummary } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { libraryPath } from './paths';
	import { BookOpen } from 'lucide-svelte';

	let {
		article,
		columnId,
		onShelf = false
	}: { article: MomentSummary; columnId?: number; onShelf?: boolean } = $props();
	const colors = ['#315c54', '#4a526b', '#854e59', '#546375', '#625679'];
</script>

<a
	class="library-book"
	class:on-shelf={onShelf}
	title={article.title}
	aria-label={article.title || '查看文章'}
	href={resolvePath(libraryPath({ column: columnId, read: article.shortUrl }))}
>
	<div class="cover" style:--book-color={colors[article.id % colors.length]}>
		{#if article.cover}
			<img src={article.cover} alt="" loading="lazy" />
		{:else}
			<span class="cover-title">{article.title || '未命名文章'}</span>
			<BookOpen size={26} strokeWidth={1} aria-hidden="true" />
			<span class="author">{article.authorName || 'shawn-blog'}</span>
		{/if}
	</div>
	{#if !onShelf}
		<h3>{article.title || '未命名文章'}</h3>
		<time datetime={article.createdAt}>{article.createdAt.slice(0, 10)}</time>
	{:else if article.cover}
		<span class="shelf-caption">{article.title}</span>
	{/if}
</a>

<style>
	.library-book {
		display: block;
		min-width: 0;
		color: inherit;
	}
	.on-shelf {
		position: relative;
		width: 158px;
		align-self: end;
	}
	.on-shelf .cover {
		box-shadow:
			inset 7px 0 9px #0005,
			3px 1px 0 #ccc7bc,
			6px 2px 0 #51473d,
			10px 5px 12px #0007;
	}
	.shelf-caption {
		position: absolute;
		bottom: 8px;
		left: 10px;
		right: 6px;
		padding: 6px;
		background: #171b20e6;
		color: #fff;
		font-size: 12px;
		line-height: 1.5;
		overflow-wrap: anywhere;
	}
	.cover {
		position: relative;
		aspect-ratio: 0.7;
		background: var(--book-color);
		color: #fff;
		border-radius: 2px 4px 4px 2px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: space-between;
		padding: 22px 16px 18px 23px;
		box-shadow:
			inset 7px 0 9px #0005,
			3px 2px 0 #c9c8c1,
			7px 7px 13px #0003;
		transition: transform 180ms ease;
		overflow: hidden;
	}
	.cover::after {
		content: '';
		position: absolute;
		inset: 0;
		border-left: 8px solid #0002;
		border-right: 2px solid #ffffff40;
		pointer-events: none;
	}
	img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.cover-title {
		font: 18px/1.7 var(--font-serif);
		overflow-wrap: anywhere;
		display: -webkit-box;
		-webkit-line-clamp: 4;
		line-clamp: 4;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.author {
		font-size: 10px;
		opacity: 0.75;
		overflow-wrap: anywhere;
	}
	h3 {
		font: 14px/1.7 var(--font-serif);
		margin-top: 16px;
		overflow-wrap: anywhere;
	}
	time {
		display: block;
		margin-top: 5px;
		font-size: 11px;
		opacity: 0.65;
	}
	a:hover .cover {
		transform: translateY(-5px);
	}
	a:focus-visible {
		outline: 2px solid var(--color-jade-500);
		outline-offset: 7px;
	}
	@media (prefers-reduced-motion: reduce) {
		.cover {
			transition: none;
		}
		a:hover .cover {
			transform: none;
		}
	}
</style>
