<script lang="ts">
	import type { ShelfBook } from '$lib/shared/nav/nav-items';
	import { onDestroy } from 'svelte';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';
	import ThemeIcon from '$lib/ui/layout/sidebar/ThemeIcon.svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import { page } from '$app/state';
	import { resolveHref } from '$lib/shared/utils/resolve-path';

	let { books = [] }: { books: ShelfBook[] } = $props();
	const aboutBook = $derived(books.find((book) => book.url === '/about'));

	const isActive = (href: string) =>
		page.url.pathname === href || (href !== '/' && page.url.pathname.startsWith(href + '/'));

	let bookFeedback: Animation | undefined;
	onDestroy(() => bookFeedback?.cancel());

	function handleBookClick(event: MouseEvent, book: ShelfBook) {
		if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey)
			return;
		bookFeedback?.cancel();
		if (!window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
			bookFeedback = (event.currentTarget as HTMLAnchorElement).animate(
				[{ translate: '0 0' }, { translate: '0 -4px', offset: 0.4 }, { translate: '0 0' }],
				{ duration: 260, easing: 'ease-out' }
			);
		}
		if (book.url === '/search') {
			event.preventDefault();
			uiState.openSearch();
		}
	}
</script>

<nav class="shelf-scene" class:home-shelf-scene={page.url.pathname === '/'} aria-label="主导航">
	<div class="window-light" aria-hidden="true"></div>
	<p class="wall-note" aria-hidden="true">
		<span>Good books,</span>
		<span>better you.</span>
	</p>

	<div class="shelf-stage">
		{#if page.url.pathname === '/'}
			<img class="shelf-plant" src="/shelf-plant.svg?v=4" alt="" aria-hidden="true" />
		{/if}

		<div class="shelf-books">
			{#each books as book, index (book.url)}
				{@const active = isActive(book.url)}
				{@const href = /^(https?:|\/\/)/i.test(book.url) ? book.url : resolveHref(book.url)}
				<a
					{href}
					aria-label={book.name}
					aria-current={active ? 'page' : undefined}
					data-sveltekit-preload-data="hover"
					onclick={(event) => handleBookClick(event, book)}
					class="book"
					class:active
					class:flat={book.placement === 'flat'}
					style="--tilt:{book.tilt}deg; --h:{book.height}px; --w:{book.width}px; --lift:{book.lift ??
						0}px; --c:{book.color}; --e:{book.edge}; --i:{index};"
				>
					<span class="book-pages" aria-hidden="true"></span>
					<span class="book-edge" aria-hidden="true"></span>
					<span class="book-frame" aria-hidden="true"></span>
					<span class="book-name">{book.name}</span>
					<span class="book-ornament" aria-hidden="true">
						<span></span>
						<DynamicLucideIcon name={book.icon} className="book-icon" />
						<span></span>
					</span>
				</a>
			{/each}
			{#if aboutBook}
				<div
					class="shelf-theme-toggle"
					style="--support-width:{aboutBook.width}px; --support-top:{aboutBook.height +
						(aboutBook.lift ?? 0)}px;"
				>
					<ThemeIcon compact interactive={page.url.pathname !== '/'} />
				</div>
			{/if}
		</div>

		<div class="shelf-board" aria-hidden="true">
			<span class="shelf-support support-left"></span>
			<span class="shelf-support support-right"></span>
		</div>
	</div>
</nav>

<style lang="postcss">
	@reference "$routes/layout.css";

	.shelf-scene {
		position: relative;
		flex: 0 1 520px;
		width: min(520px, calc(100vw - 170px));
		height: 240px;
		margin-left: auto;
		margin-right: 0.75rem;
		overflow: hidden;
		isolation: isolate;
		background: transparent;
		animation: scene-enter 700ms ease-out both;
	}

	.shelf-scene::before {
		content: '';
		position: absolute;
		inset: 0;
		z-index: -1;
		background: transparent;
		pointer-events: none;
	}

	.shelf-scene.home-shelf-scene {
		overflow: visible;
	}

	.shelf-scene::after {
		content: '';
		position: absolute;
		top: 0;
		bottom: 0;
		left: 82%;
		width: 1px;
		display: none;
	}

	.window-light {
		display: none;
	}

	.wall-note {
		display: none;
		position: absolute;
		flex-direction: column;
		margin: 0;
		transform: rotate(-6deg);
		color: rgb(105 91 77 / 0.68);
		font-family: var(--font-mono);
		font-size: clamp(0.72rem, 1.25vw, 1.02rem);
		font-style: italic;
		line-height: 1.5;
		letter-spacing: -0.04em;
	}

	.wall-note::after {
		content: '';
		width: 3.6rem;
		height: 1px;
		margin-top: 0.45rem;
		margin-left: 0.45rem;
		transform: rotate(-12deg);
		background: currentColor;
		opacity: 0.72;
	}

	.shelf-stage {
		position: absolute;
		inset: 0;
		margin: auto;
		width: min(1120px, 100%);
	}

	.shelf-books {
		position: absolute;
		bottom: 64px;
		left: 145px;
		z-index: 2;
		display: flex;
		align-items: flex-end;
		gap: 0.45rem;
		width: auto;
		height: 150px;
	}

	.shelf-plant {
		position: absolute;
		top: 0;
		left: -27px;
		z-index: 4;
		width: 247.5px;
		height: 300px;
		object-fit: contain;
		object-position: center bottom;
		pointer-events: none;
		filter: drop-shadow(3px 7px 6px rgb(43 27 17 / 0.22));
	}

	.book {
		position: relative;
		display: flex;
		flex: 0 0 var(--w);
		align-items: center;
		justify-content: center;
		width: var(--w);
		height: var(--h);
		transform: rotate(var(--tilt));
		transform-origin: 50% 100%;
		border: 1px solid color-mix(in srgb, var(--e) 88%, black);
		border-radius: 4px 6px 3px 3px;
		background:
			linear-gradient(90deg, rgb(255 255 255 / 0.12), transparent 11% 84%, rgb(0 0 0 / 0.2)),
			repeating-linear-gradient(0deg, rgb(255 255 255 / 0.018) 0 1px, transparent 1px 4px), var(--c);
		box-shadow:
			2px 2px 2px rgb(42 27 17 / 0.18),
			7px 7px 12px rgb(42 27 17 / 0.12),
			inset 2px 0 rgb(255 255 255 / 0.1),
			inset -3px 0 rgb(0 0 0 / 0.12);
		color: #e9d7ae;
		text-decoration: none;
		transition:
			transform 300ms cubic-bezier(0.2, 0.8, 0.2, 1),
			filter 300ms ease,
			box-shadow 300ms ease;
		animation: book-reveal 500ms ease-out both;
		animation-delay: calc(90ms + var(--i) * 65ms);
	}

	.book:nth-of-type(1) {
		margin-left: 0;
	}

	.book:nth-of-type(2) {
		margin-left: -0.5rem;
	}

	.book:nth-of-type(3) {
		margin-left: -0.4rem;
	}

	.book:nth-of-type(4) {
		z-index: 1;
		margin-left: -0.35rem;
	}

	.book:nth-of-type(5) {
		z-index: 2;
		margin-left: -0.65rem;
	}

	.book:nth-of-type(6) {
		z-index: 4;
		margin-left: calc(-94px + 0.5rem);
		box-shadow:
			2px 7px 7px rgb(42 27 17 / 0.2),
			8px 10px 16px rgb(42 27 17 / 0.12),
			inset 2px 0 rgb(255 255 255 / 0.1),
			inset -3px 0 rgb(0 0 0 / 0.12);
	}

	.book:hover,
	.book:focus-visible {
		z-index: 8;
		transform: rotate(0deg);
		outline: none;
		filter: saturate(1.08) brightness(1.04);
		box-shadow:
			2px 3px 3px rgb(42 27 17 / 0.16),
			9px 15px 20px rgb(42 27 17 / 0.18),
			inset 2px 0 rgb(255 255 255 / 0.12),
			inset -3px 0 rgb(0 0 0 / 0.1);
	}

	.book:focus-visible {
		box-shadow:
			0 0 0 3px rgb(236 206 146 / 0.72),
			9px 15px 20px rgb(42 27 17 / 0.18);
	}

	.book.active {
		z-index: 6;
		filter: saturate(1.05) brightness(1.06);
		box-shadow:
			0 0 0 1px rgb(239 210 155 / 0.72),
			4px 7px 12px rgb(42 27 17 / 0.18),
			inset 2px 0 rgb(255 255 255 / 0.12),
			inset -3px 0 rgb(0 0 0 / 0.1);
	}

	.book.active:not(.flat) {
		transform: rotate(var(--tilt));
	}

	.book.active:not(.flat):hover,
	.book.active:not(.flat):focus-visible {
		transform: rotate(0deg);
	}

	.book::before {
		content: '';
		position: absolute;
		inset: 1px;
		border-radius: inherit;
		background-image: var(--texture-noise);
		opacity: 0.11;
		mix-blend-mode: soft-light;
		pointer-events: none;
	}

	.book-pages {
		position: absolute;
		top: -6px;
		left: 5px;
		right: -2px;
		height: 7px;
		transform: skewX(-18deg);
		border: 1px solid rgb(109 81 54 / 0.34);
		border-radius: 2px 4px 1px 1px;
		background: repeating-linear-gradient(0deg, #e9dfce 0 1px, #c9bca7 1px 2px, #f1e8da 2px 3px);
		box-shadow: 1px -1px 2px rgb(62 41 25 / 0.12);
	}

	.book-edge {
		position: absolute;
		inset: 0 auto 0 5px;
		width: 6px;
		border-right: 1px solid rgb(236 207 150 / 0.28);
		background: linear-gradient(90deg, var(--e), rgb(255 255 255 / 0.08), var(--e));
		opacity: 0.82;
	}

	.book-frame {
		position: absolute;
		inset: 8px 6px;
		border: 1px solid rgb(231 201 143 / 0.52);
		border-radius: 1px;
		box-shadow: inset 0 0 0 2px rgb(38 23 15 / 0.08);
		pointer-events: none;
	}

	.book-frame::before,
	.book-frame::after {
		content: '';
		position: absolute;
		left: 50%;
		width: 18px;
		height: 1px;
		transform: translateX(-50%);
		background: rgb(231 201 143 / 0.66);
		box-shadow: 0 3px 0 -0.2px rgb(231 201 143 / 0.35);
	}

	.book-frame::before {
		top: 8px;
	}

	.book-frame::after {
		bottom: 8px;
	}

	.book-name {
		position: relative;
		z-index: 2;
		margin-top: -0.8rem;
		font-family: var(--font-serif);
		font-size: 0.98rem;
		font-weight: 500;
		line-height: 1.15;
		letter-spacing: 0.16em;
		text-shadow: 0 1px 1px rgb(42 27 17 / 0.65);
		writing-mode: vertical-rl;
		text-orientation: upright;
	}

	.book-ornament {
		position: absolute;
		bottom: 16px;
		left: 50%;
		z-index: 2;
		display: flex;
		align-items: center;
		gap: 3px;
		transform: translateX(-50%);
		opacity: 0.8;
	}

	.book-ornament > span {
		width: 6px;
		height: 1px;
		background: currentColor;
	}

	:global(.book-icon) {
		width: 10px;
		height: 10px;
		stroke-width: 1.4;
	}

	.book.flat {
		align-self: flex-end;
		transform: perspective(260px) translateY(calc(-1px - var(--lift))) rotateX(-3deg)
			rotate(var(--tilt));
		transform-origin: center;
		border-radius: 5px 2px 2px 5px;
	}

	.book.flat:hover,
	.book.flat:focus-visible {
		transform: perspective(260px) translateY(calc(-1px - var(--lift))) rotateX(-3deg)
			rotate(var(--tilt));
	}

	.book.flat.active {
		transform: perspective(260px) translateY(calc(-1px - var(--lift))) rotateX(-3deg)
			rotate(var(--tilt));
	}

	.book.flat.active:hover,
	.book.flat.active:focus-visible {
		transform: perspective(260px) translateY(calc(-1px - var(--lift))) rotateX(-3deg)
			rotate(var(--tilt));
	}

	.book.flat .book-pages {
		top: 5px;
		right: -7px;
		bottom: 4px;
		left: auto;
		width: 8px;
		height: auto;
		transform: skewY(-12deg);
		background: repeating-linear-gradient(90deg, #e9dfce 0 1px, #c9bca7 1px 2px, #f1e8da 2px 3px);
	}

	.book.flat .book-edge {
		inset: 5px auto 4px 0;
		width: 8px;
		border-right: 1px solid rgb(236 207 150 / 0.24);
		border-radius: 4px 0 0 4px;
	}

	.book.flat .book-frame {
		inset: 6px 9px;
	}

	.book.flat .book-frame::before,
	.book.flat .book-frame::after {
		top: 50%;
		bottom: auto;
		width: 1px;
		height: 14px;
		transform: translateY(-50%);
		box-shadow: 3px 0 0 -0.2px rgb(231 201 143 / 0.35);
	}

	.book.flat .book-frame::before {
		left: 9px;
	}

	.book.flat .book-frame::after {
		right: 9px;
		left: auto;
	}

	.book.flat .book-name {
		margin: 0;
		font-size: 0.78rem;
		letter-spacing: 0.32em;
		writing-mode: horizontal-tb;
	}

	.book.flat .book-ornament {
		top: 50%;
		right: 11px;
		bottom: auto;
		left: auto;
		transform: translateY(-50%);
	}

	.book.flat .book-ornament > span {
		display: none;
	}

	.shelf-board {
		position: absolute;
		right: 60px;
		bottom: 50px;
		left: 28px;
		z-index: 3;
		height: 13px;
		border: 1px solid #68452d;
		border-radius: 2px 3px 2px 2px;
		background:
			linear-gradient(180deg, rgb(255 255 255 / 0.2), transparent 30%),
			repeating-linear-gradient(88deg, rgb(60 34 17 / 0.14) 0 2px, transparent 2px 37px),
			linear-gradient(180deg, #9c623b 0 30%, #704329 31% 100%);
		box-shadow:
			0 4px 5px rgb(45 27 15 / 0.34),
			0 12px 20px rgb(45 27 15 / 0.18),
			inset 0 -3px rgb(57 33 19 / 0.2);
	}

	.shelf-theme-toggle {
		position: absolute;
		z-index: 10;
		right: 0;
		bottom: var(--support-top);
		display: flex;
		justify-content: center;
		width: var(--support-width);
	}

	.shelf-board::before {
		content: '';
		position: absolute;
		top: -5px;
		right: -1px;
		left: -1px;
		height: 6px;
		transform: perspective(300px) rotateX(28deg);
		transform-origin: bottom;
		border: 1px solid #7d5335;
		border-radius: 3px 3px 0 0;
		background: #ad7548;
	}

	.shelf-support {
		position: absolute;
		top: 12px;
		width: 31px;
		height: 44px;
		clip-path: polygon(0 0, 100% 0, 78% 16%, 63% 100%, 25% 100%, 35% 18%, 0 18%);
		background: linear-gradient(
			105deg,
			#3b251a 0 16%,
			#8d5e3f 18% 48%,
			#5d3b2a 50% 72%,
			#2c1b14 74% 100%
		);
		box-shadow:
			inset 2px 0 rgb(234 196 145 / 0.2),
			inset -3px 0 rgb(0 0 0 / 0.24);
		filter: drop-shadow(3px 5px 3px rgb(35 20 12 / 0.34));
	}

	.shelf-support::before {
		content: '';
		position: absolute;
		inset: 3px 6px 4px 7px;
		clip-path: polygon(0 0, 100% 0, 70% 15%, 54% 100%, 22% 100%, 34% 18%, 0 18%);
		background: linear-gradient(100deg, rgb(231 191 139 / 0.22), transparent 48%, rgb(0 0 0 / 0.2));
	}

	.shelf-support::after {
		content: '';
		position: absolute;
		top: 0;
		left: 1px;
		width: 29px;
		height: 7px;
		clip-path: polygon(0 0, 100% 0, 82% 100%, 5% 100%);
		background: linear-gradient(180deg, #a97852, #5c3927);
		box-shadow: inset 0 1px rgb(247 216 173 / 0.22);
	}

	.support-left {
		left: 12%;
	}

	.support-right {
		right: 12%;
		transform: scaleX(-1);
	}

	:global(.dark) .shelf-scene {
		background: transparent;
	}

	:global(.dark) .window-light {
		opacity: 0.12;
	}

	:global(.dark) .wall-note {
		color: rgb(222 207 188 / 0.44);
	}

	:global(.dark) .shelf-plant {
		filter: brightness(0.68) saturate(0.82) sepia(0.1) drop-shadow(3px 7px 6px rgb(20 13 10 / 0.36));
	}

	@keyframes scene-enter {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@keyframes book-reveal {
		from {
			opacity: 0;
			filter: blur(3px);
		}
		to {
			opacity: 1;
			filter: blur(0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.shelf-scene,
		.book {
			animation: none;
		}

		.book {
			transition-duration: 0.01ms;
		}
	}

	@media (min-width: 768px) and (max-width: 1919px) {
		.shelf-scene.home-shelf-scene {
			transform: scale(clamp(0.4, calc(100vw / 1920px), 1));
			transform-origin: top right;
		}
	}
</style>
