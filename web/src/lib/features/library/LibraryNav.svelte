<script lang="ts">
	import { NAV_ITEMS } from '$lib/shared/nav/nav-items';
	import { brand } from '$lib/shared/brand/brand';
	import { resolveHref } from '$lib/shared/utils/resolve-path';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import ThemeIcon from '$lib/ui/layout/sidebar/ThemeIcon.svelte';
	import { Search, ArrowLeft, Menu, Sun, Moon } from 'lucide-svelte';
	import { themeManager } from '$lib/shared/theme/theme.svelte';
	import { page } from '$app/state';
	import type { Column } from '$lib/features/taxonomy/types';
	import { libraryPath } from './paths';
	let { activePath = '/gallery' }: { activePath?: string } = $props();
	const roots = $derived(
		((page.data as { columns?: Column[] }).columns ?? []).filter((item) => !item.parentId)
	);
	const activeRoot = $derived((page.data as { root?: Column | null }).root?.id);
	const bookColors = ['#3e6154', '#4a5b73', '#815159', '#70604a'];
</script>

<header class="library-nav" class:library-categories={activePath === '/gallery'}>
	{#if activePath === '/gallery'}
		<a class="back-home" href={resolveHref('/')}><ArrowLeft size={16} />返回首页</a>
		<a class="brand" href={resolveHref('/gallery/')}>图书馆</a>
		<div class="library-actions">
			<button type="button" aria-label="搜索" title="搜索" onclick={() => uiState.openSearch()}
				><Search size={19} /></button
			>
			<button
				type="button"
				aria-label={themeManager.current === 'dark' ? '切换浅色' : '切换深色'}
				title="切换主题"
				onclick={() => themeManager.set(themeManager.current === 'dark' ? 'light' : 'dark')}
			>
				{#if themeManager.current === 'dark'}<Sun size={19} />{:else}<Moon size={19} />{/if}
			</button>
			<details class="site-directory">
				<summary aria-label="站点目录" title="站点目录"><Menu size={20} /></summary>
				<nav aria-label="主导航">
					{#each NAV_ITEMS.filter((item) => item.url !== '/search') as item}
						<a
							href={resolveHref(item.url)}
							aria-current={item.url === activePath ? 'page' : undefined}>{item.name}</a
						>
					{/each}
				</nav>
			</details>
		</div>
		<nav class="primary-shelf" aria-label="一级分类书架">
			<a
				class="all-categories"
				href={resolveHref('/gallery/')}
				aria-current={!activeRoot ? 'page' : undefined}>全部</a
			>
			{#each roots as root, index (root.id)}
				<a
					class="category-spine"
					style:--spine={bookColors[index % bookColors.length]}
					href={resolveHref(libraryPath({ column: root.id }))}
					aria-current={activeRoot === root.id ? 'page' : undefined}
					title={root.name}
				>
					<span>{root.name}</span>
				</a>
			{/each}
		</nav>
	{:else}
		<a class="brand" href={resolveHref('/')}>{brand.name}</a>
		<nav aria-label="主导航">
			{#each NAV_ITEMS as item (item.url)}
				{#if item.url === '/search'}
					<button type="button" onclick={() => uiState.openSearch()}
						><Search size={15} />搜索</button
					>
				{:else}
					<a
						href={resolveHref(item.url)}
						aria-current={item.url === activePath ? 'page' : undefined}>{item.name}</a
					>
				{/if}
			{/each}
		</nav>
		<div class="clock"><ThemeIcon compact /></div>
	{/if}
</header>

<style>
	.library-nav.library-categories {
		position: relative;
		height: auto;
		min-height: 180px;
		padding-top: 18px;
		padding-bottom: 22px;
		align-items: flex-start;
		gap: 24px;
		background: transparent;
		box-shadow: none;
		border: 0;
		flex-wrap: wrap;
	}
	.library-categories .brand {
		margin-right: 0;
		padding-top: 6px;
	}
	.library-categories .back-home {
		padding-top: 14px;
	}
	.library-categories .library-actions {
		padding-top: 3px;
	}
	.library-categories .primary-shelf {
		margin-left: auto;
		display: flex;
		align-items: end;
		gap: 6px;
		height: 154px;
		max-width: min(600px, 100%);
		overflow-x: auto;
		padding: 8px 14px 12px;
		border-bottom: 9px solid #826447;
		box-shadow: 0 9px 10px -8px #0008;
	}
	.primary-shelf .category-spine {
		display: flex;
		justify-content: center;
		align-items: center;
		width: 49px;
		flex-shrink: 0;
		height: 130px;
		padding: 9px 0;
		background: var(--spine);
		color: #f6ecd7;
		border: 1px solid #ead5a94d;
		border-radius: 3px;
		box-shadow:
			inset 4px 0 5px #0005,
			inset -3px 0 2px #fff2,
			2px 2px 2px #0003;
		transition: transform 180ms ease;
	}
	.category-spine span {
		writing-mode: vertical-rl;
		font: 15px/1.5 var(--font-serif);
	}
	.primary-shelf .category-spine[aria-current],
	.primary-shelf .category-spine:hover {
		transform: translateY(-5px);
		outline: 1px solid #aa8959;
	}
	.primary-shelf .all-categories {
		height: auto;
		font-size: 12px;
		padding: 8px;
		white-space: nowrap;
	}
	@media (prefers-reduced-motion: reduce) {
		.primary-shelf .category-spine {
			transition: none;
		}
		.primary-shelf .category-spine[aria-current],
		.primary-shelf .category-spine:hover {
			transform: none;
		}
	}
	.back-home {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
	}
	.library-actions {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.library-actions > button,
	.site-directory summary {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 38px;
		height: 38px;
		cursor: pointer;
		list-style: none;
	}
	.site-directory {
		position: relative;
	}
	.site-directory nav {
		position: absolute;
		right: 0;
		top: 48px;
		display: flex;
		flex-direction: column;
		align-items: stretch;
		gap: 0;
		width: 170px;
		height: auto;
		padding: 8px;
		background: var(--color-ink-50);
		border: 1px solid var(--color-ink-200);
		border-radius: 4px;
		box-shadow: 0 8px 24px #0002;
	}
	.site-directory nav a {
		padding: 10px 14px;
	}
	:global(.dark) .site-directory nav {
		background: var(--color-ink-950);
		border-color: var(--color-ink-800);
	}
	.library-nav {
		position: fixed;
		inset: 0 0 auto;
		z-index: 1000;
		height: 88px;
		display: flex;
		align-items: center;
		gap: 40px;
		padding: 0 max(32px, calc((100vw - 1280px) / 2));
		background: var(--color-ink-50);
		color: var(--color-ink-900);
		border-bottom: 1px solid var(--color-ink-200);
		box-shadow: 0 3px 12px #00000008;
	}
	.brand {
		font-family: var(--font-serif);
		font-size: 22px;
		margin-right: auto;
	}
	nav {
		display: flex;
		align-items: center;
		gap: 28px;
		height: 100%;
	}
	nav a,
	nav button {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		height: 100%;
		font-size: 14px;
		border-bottom: 2px solid transparent;
	}
	nav a[aria-current],
	nav a:hover,
	nav button:hover {
		color: var(--color-jade-700);
		border-bottom-color: var(--color-jade-600);
	}
	.clock {
		width: 68px;
		height: 77px;
		flex-shrink: 0;
	}
	a:focus-visible,
	button:focus-visible {
		outline: 2px solid var(--color-jade-600);
		outline-offset: 5px;
	}
	:global(.dark) .library-nav {
		background: var(--color-ink-950);
		color: var(--color-ink-100);
		border-color: var(--color-ink-800);
	}
	@media (max-width: 767px) {
		.library-nav {
			display: none;
		}
		.library-nav.library-categories {
			display: flex;
			padding: 24px 20px;
		}
		.library-categories .primary-shelf {
			width: 100%;
		}
	}
</style>
