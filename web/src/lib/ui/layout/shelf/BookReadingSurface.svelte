<script lang="ts">
	import { page } from '$app/state';
	import { SHELF_BOOKS } from '$lib/shared/nav/nav-items';
	import { bookTransition } from '$lib/shared/stores/book-transition.svelte';
	import type { Snippet } from 'svelte';

	let { children } = $props<{ children: Snippet }>();
	const turnLeaves = Array.from({ length: 7 }, (_, index) => index);
	const activeBook = $derived(
		SHELF_BOOKS.find(
			(book) =>
				page.url.pathname === book.url ||
				(book.url !== '/' && page.url.pathname.startsWith(book.url + '/'))
		) ?? SHELF_BOOKS[1]
	);
</script>

<div
	class="reading-surface"
	class:reading={bookTransition.reading}
	class:closing={bookTransition.closing}
>
	{#if bookTransition.reading}
		<div class="page-stack stack-left" aria-hidden="true"></div>
		<div class="page-stack stack-right" aria-hidden="true"></div>
		<div class="page-stack stack-bottom" aria-hidden="true"></div>
		<div class="chapter-page" aria-hidden="true">
			<span>CHAPTER</span>
			<strong>{activeBook?.name}</strong>
			<i></i>
		</div>
		<div class="page-ribbon" aria-hidden="true"></div>
		{#if bookTransition.pageTurning}
			<div class="page-turn-layer" aria-hidden="true">
				{#each turnLeaves as index (index)}
					<span class="turn-leaf" style="--turn-index:{index};"></span>
				{/each}
			</div>
		{/if}
		{#if bookTransition.closing}
			<div class="closing-cover" aria-hidden="true">
				<span>{activeBook?.name}</span>
			</div>
		{/if}
	{/if}
	<div class="reading-content">
		{@render children()}
	</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";

	.reading-surface {
		position: relative;
	}

	.reading-content {
		position: relative;
		z-index: 3;
	}

	.reading-surface.reading {
		width: calc(100vw - 52px);
		min-height: calc(100vh - 58px);
		margin: 28px auto 30px;
		padding: 10px 12px 17px;
		border: 2px solid #3e2a20;
		border-radius: 20px 27px 27px 20px;
		background:
			repeating-linear-gradient(90deg, rgb(255 255 255 / 0.025) 0 2px, transparent 2px 6px),
			linear-gradient(135deg, #76523d, #4c3328 48%, #6a4938);
		box-shadow:
			0 5px 0 #2d1d17,
			0 18px 0 -8px #b5a288,
			0 24px 0 -10px #6c503d,
			0 34px 58px rgb(35 22 14 / 0.42),
			inset 0 0 0 2px rgb(228 191 139 / 0.18);
	}

	.reading .reading-content {
		min-height: calc(100vh - 88px);
		padding: 12rem clamp(1.4rem, 3.4vw, 4rem) 2rem calc(50% + clamp(1.4rem, 3.4vw, 4rem));
		overflow: clip;
		border: 1px solid #a99578;
		border-radius: 11px 17px 17px 11px;
		background:
			linear-gradient(
				90deg,
				#d5c6b0 0,
				#fffaf0 2.2%,
				#f6eddd 47.8%,
				#d2c1a8 49.2%,
				#a18d70 49.9%,
				#8e795e 50%,
				#b5a187 50.2%,
				#eee2cf 51.2%,
				#fffaf0 97.7%,
				#d1c0a8 100%
			),
			repeating-linear-gradient(0deg, transparent 0 27px, rgb(102 79 53 / 0.038) 27px 28px);
		box-shadow:
			inset 24px 0 30px rgb(80 56 34 / 0.1),
			inset -24px 0 30px rgb(80 56 34 / 0.1),
			0 2px 8px rgb(28 18 11 / 0.28);
	}

	.reading-surface.closing {
		transform-origin: center;
		animation: close-reading-surface 650ms cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	.page-stack {
		position: absolute;
		z-index: 2;
		border: 1px solid #9b876c;
		background: repeating-linear-gradient(0deg, #efe6d7 0 1px, #cbbda7 1px 2px, #f7efe2 2px 3px);
		pointer-events: none;
	}

	.stack-bottom {
		right: 18px;
		bottom: 5px;
		left: 18px;
		height: 15px;
		border-radius: 0 0 13px 13px;
	}

	.stack-left,
	.stack-right {
		top: 18px;
		bottom: 20px;
		width: 12px;
		background: repeating-linear-gradient(90deg, #e9dfd0 0 1px, #c5b69f 1px 2px, #f5ecde 2px 3px);
	}

	.stack-left {
		left: 5px;
		border-radius: 10px 0 0 10px;
	}

	.stack-right {
		right: 5px;
		border-radius: 0 13px 13px 0;
	}

	.chapter-page {
		position: absolute;
		top: 31%;
		left: 7%;
		z-index: 5;
		display: flex;
		width: 36%;
		flex-direction: column;
		align-items: center;
		color: #78634b;
		font-family: var(--font-serif);
		text-align: center;
		pointer-events: none;
	}

	.chapter-page span {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		letter-spacing: 0.38em;
		opacity: 0.58;
	}

	.chapter-page strong {
		margin-top: 1.1rem;
		font-size: clamp(2rem, 4vw, 4.5rem);
		font-weight: 500;
		letter-spacing: 0.22em;
		text-indent: 0.22em;
	}

	.chapter-page i {
		width: 54px;
		height: 1px;
		margin-top: 1.3rem;
		background: currentColor;
		opacity: 0.45;
	}

	.page-ribbon {
		position: absolute;
		top: 0;
		right: clamp(76px, 8vw, 136px);
		z-index: 7;
		width: 20px;
		height: 88px;
		background: linear-gradient(90deg, #6b3340, #8a4653 52%, #5d2935);
		clip-path: polygon(0 0, 100% 0, 100% 100%, 50% 82%, 0 100%);
		box-shadow: 2px 4px 6px rgb(58 31 28 / 0.2);
	}

	.page-turn-layer {
		position: absolute;
		inset: 11px 13px 18px;
		z-index: 8;
		overflow: hidden;
		border-radius: 10px 16px 16px 10px;
		perspective: 1800px;
		pointer-events: auto;
	}

	.turn-leaf {
		position: absolute;
		top: 0;
		right: 0;
		bottom: 0;
		width: 50%;
		transform-origin: left center;
		border: 1px solid rgb(151 126 94 / 0.5);
		border-radius: 0 12px 12px 0;
		background:
			repeating-linear-gradient(0deg, transparent 0 25px, rgb(103 78 50 / 0.05) 25px 26px),
			linear-gradient(90deg, #d9cbb5, #fffaf0 7% 94%, #d1c1aa);
		box-shadow: 10px 0 20px rgb(60 39 23 / 0.14);
		backface-visibility: hidden;
		animation: turn-reading-page 570ms cubic-bezier(0.55, 0.08, 0.45, 0.95) forwards;
		animation-delay: calc(var(--turn-index) * 32ms);
	}

	.closing-cover {
		position: absolute;
		top: 11px;
		right: 13px;
		bottom: 18px;
		z-index: 10;
		display: grid;
		width: calc(50% - 13px);
		transform-origin: left center;
		place-items: center;
		border: 3px solid #3e2c31;
		border-radius: 7px 18px 18px 7px;
		background:
			linear-gradient(90deg, rgb(255 255 255 / 0.1), transparent 12% 85%, rgb(0 0 0 / 0.22)),
			repeating-linear-gradient(0deg, rgb(255 255 255 / 0.018) 0 1px, transparent 1px 5px), #655263;
		box-shadow:
			inset 12px 0 rgb(0 0 0 / 0.13),
			14px 20px 30px rgb(33 21 14 / 0.24);
		color: #ead9b7;
		font-family: var(--font-serif);
		font-size: 2rem;
		letter-spacing: 0.28em;
		animation: close-cover 470ms cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	:global(.dark) .reading .reading-content {
		background:
			linear-gradient(
				90deg,
				#c8b89f,
				#eee3d1 3%,
				#e7dbc8 48%,
				#8f795e 50%,
				#e7dbc8 52%,
				#eee3d1 97%,
				#c8b89f
			),
			repeating-linear-gradient(0deg, transparent 0 27px, rgb(78 57 37 / 0.05) 27px 28px);
	}

	@media (max-width: 767px) {
		.reading-surface.reading {
			width: 100%;
			margin: 0;
			padding: 0;
			border: 0;
			border-radius: 0;
			background: transparent;
			box-shadow: none;
		}

		.reading .reading-content {
			padding: 0;
			border: 0;
			background: transparent;
			box-shadow: none;
		}

		.page-stack,
		.chapter-page,
		.page-ribbon {
			display: none;
		}
	}

	@keyframes turn-reading-page {
		0% {
			transform: rotateY(0deg) translateZ(calc(var(--turn-index) * 0.5px));
		}
		48% {
			transform: rotateY(-88deg) translateZ(16px) skewY(-1deg);
		}
		100% {
			transform: rotateY(-178deg) translateZ(calc(var(--turn-index) * 0.35px));
		}
	}

	@keyframes close-cover {
		from {
			transform: rotateY(-178deg);
		}
		to {
			transform: rotateY(0deg);
		}
	}

	@keyframes close-reading-surface {
		0%,
		72% {
			transform: scale(1);
			opacity: 1;
		}
		100% {
			transform: translate(35vw, -37vh) scale(0.14);
			opacity: 0;
		}
	}
</style>
