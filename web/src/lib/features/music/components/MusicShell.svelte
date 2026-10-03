<script lang="ts">
	import type { Snippet } from 'svelte';
	import Search from 'lucide-svelte/icons/search';
	import ListMusic from 'lucide-svelte/icons/list-music';
	import House from 'lucide-svelte/icons/house';
	import Heart from 'lucide-svelte/icons/heart';
	import Settings2 from 'lucide-svelte/icons/settings-2';
	import Music2 from 'lucide-svelte/icons/music-2';
	import ArrowLeft from 'lucide-svelte/icons/arrow-left';
	import UserRound from 'lucide-svelte/icons/user-round';
	import type { MusicView } from '../types';

	let { view, title, loggedIn, selectView, openAccount, children } = $props<{
		view: MusicView;
		title: string;
		loggedIn: boolean;
		selectView: (view: MusicView) => void;
		openAccount: () => void;
		children: Snippet;
	}>();
	const navigation = [
		{ id: 'home', label: '首页', icon: House },
		{ id: 'search', label: '搜索', icon: Search },
		{ id: 'playlists', label: '我的歌单', icon: ListMusic },
		{ id: 'favorites', label: '我的收藏', icon: Heart },
		{ id: 'settings', label: '设置', icon: Settings2 }
	] as const;
</script>

<div class="music-shell">
	<aside class="music-sidebar" aria-label="音乐导航">
		<a class="music-brand" href="/music/" aria-label="音乐室首页"
			><Music2 size={27} /><span>音乐室</span></a
		>
		<nav aria-label="音乐功能">
			{#each navigation as item (item.id)}
				<button
					type="button"
					class="music-nav-item"
					class:selected={view === item.id}
					aria-current={view === item.id ? 'page' : undefined}
					onclick={() => selectView(item.id)}
					><item.icon size={18} /><span>{item.label}</span></button
				>
			{/each}
		</nav>
		<button type="button" class="music-account" onclick={openAccount}
			><UserRound size={18} />{loggedIn ? '我的账号' : '登录'}</button
		>
	</aside>
	<div class="music-workspace">
		<header class="music-header">
			<h1>{title}</h1>
			<a href="/" class="music-back"><ArrowLeft size={16} /><span>返回博客</span></a>
		</header>
		<div class="music-content">{@render children()}</div>
	</div>
</div>

<style lang="postcss">
	@reference '../../../../routes/layout.css';
	.music-shell {
		--music-accent: #f43f5e;
		--music-accent-strong: #e11d48;
		--music-accent-soft: #fff1f2;
		--music-ink: #292524;
		--music-muted: #78716c;
		--music-border: #ece8e5;
		--music-panel: #faf9f8;
		min-height: 100dvh;
		background: #fff;
		color: var(--music-ink);
		color-scheme: light;
		@apply font-sans;
	}
	.music-sidebar {
		position: fixed;
		inset: 0 auto 116px 0;
		z-index: 30;
		width: 220px;
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--music-border);
		background: #fff;
	}
	.music-brand {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 76px;
		padding: 0 28px;
		border-bottom: 1px solid var(--music-border);
		color: var(--music-accent);
	}
	.music-brand span {
		@apply font-serif text-2xl font-medium;
	}
	nav {
		padding: 22px 0;
	}
	.music-nav-item {
		display: flex;
		align-items: center;
		gap: 13px;
		min-height: 48px;
		width: 100%;
		padding: 12px 26px;
		border-right: 3px solid transparent;
		color: #57534e;
		text-align: left;
		font-size: 14px;
	}
	.music-nav-item:hover,
	.music-account:hover {
		background: var(--music-panel);
	}
	.music-nav-item.selected {
		border-right-color: var(--music-accent);
		background: var(--music-accent-soft);
		color: var(--music-accent-strong);
	}
	.music-account {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 20px 26px;
		margin-top: auto;
		border-top: 1px solid var(--music-border);
		color: var(--music-muted);
		font-size: 13px;
	}
	.music-workspace {
		margin-left: 220px;
		min-width: 0;
	}
	.music-header {
		min-height: 76px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		padding: 0 30px;
		border-bottom: 1px solid var(--music-border);
		background: #fff;
	}
	h1 {
		@apply text-lg font-medium;
	}
	.music-back {
		display: flex;
		align-items: center;
		gap: 7px;
		color: var(--music-muted);
		font-size: 12px;
	}
	.music-back:hover {
		color: var(--music-accent-strong);
	}
	.music-content {
		min-width: 0;
		padding: 26px 30px 158px;
	}
	.music-shell :global(.music-primary) {
		background: var(--music-accent-strong) !important;
		color: #fff !important;
		box-shadow: none;
	}
	.music-shell :global(.music-primary:hover) {
		background: var(--music-accent-strong) !important;
	}
	.music-shell :global(.music-play) {
		background: var(--music-accent) !important;
	}
	.music-shell :global(.music-play:hover) {
		background: var(--music-accent-strong) !important;
	}
	.music-shell :global(.music-secondary),
	.music-shell :global(.music-icon) {
		color: #57534e !important;
		background: #fff !important;
		border-color: var(--music-border) !important;
		box-shadow: none;
	}
	.music-shell :global(.music-secondary:hover),
	.music-shell :global(.music-icon:hover) {
		background: var(--music-accent-soft) !important;
		color: var(--music-accent-strong) !important;
	}
	.music-shell :global(button:focus-visible),
	.music-shell :global(a:focus-visible),
	.music-shell :global(input:focus-visible) {
		outline: 2px solid var(--music-accent);
		outline-offset: 3px;
	}
	.music-shell :global(input[type='range']),
	.music-shell :global(input[type='checkbox']) {
		accent-color: var(--music-accent);
	}
	@media (max-width: 1000px) {
		.music-sidebar {
			width: 184px;
		}
		.music-brand {
			padding: 0 22px;
		}
		.music-nav-item {
			padding-left: 22px;
		}
		.music-workspace {
			margin-left: 184px;
		}
		.music-content {
			padding-left: 22px;
			padding-right: 22px;
		}
	}
	@media (max-width: 640px) {
		.music-sidebar {
			position: static;
			width: auto;
			border-right: 0;
		}
		.music-brand {
			min-height: 64px;
			padding: 0 18px;
		}
		.music-brand span {
			font-size: 22px;
		}
		nav {
			display: grid;
			grid-template-columns: repeat(5, minmax(0, 1fr));
			padding: 0;
			border-bottom: 1px solid var(--music-border);
		}
		.music-nav-item {
			flex-direction: column;
			justify-content: center;
			gap: 5px;
			padding: 12px 2px;
			border-right: 0;
			border-bottom: 2px solid transparent;
			font-size: 11px;
		}
		.music-nav-item.selected {
			border-bottom-color: var(--music-accent);
		}
		.music-account {
			position: absolute;
			top: 0;
			right: 14px;
			min-height: 64px;
			border: 0;
			margin: 0;
			padding: 0 4px;
		}
		.music-workspace {
			margin-left: 0;
		}
		.music-header {
			min-height: 58px;
			padding: 0 18px;
		}
		.music-content {
			padding: 18px 14px calc(176px + env(safe-area-inset-bottom, 0px));
		}
	}
</style>
