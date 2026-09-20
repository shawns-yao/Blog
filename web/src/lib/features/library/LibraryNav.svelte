<script lang="ts">
	import { NAV_ITEMS } from '$lib/shared/nav/nav-items';
	import { brand } from '$lib/shared/brand/brand';
	import { resolveHref } from '$lib/shared/utils/resolve-path';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import ThemeIcon from '$lib/ui/layout/sidebar/ThemeIcon.svelte';
	import { Search, ArrowLeft, Menu, Sun, Moon } from 'lucide-svelte';
	import { themeManager } from '$lib/shared/theme/theme.svelte';
	let { activePath = '/gallery' }: { activePath?: string } = $props();
</script>

<header class="library-nav">
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
	}
</style>
