<script lang="ts">
	import { bookTransition } from '$lib/shared/stores/book-transition.svelte';

	const pageLeaves = Array.from({ length: 7 }, (_, index) => index);
</script>

{#if bookTransition.active && bookTransition.book}
	<div
		class="book-transition-layer"
		class:opening={bookTransition.phase === 'opening'}
		class:revealing={bookTransition.phase === 'revealing'}
		aria-hidden="true"
	>
		<div class="transition-shade"></div>
		<div
			class="transition-book"
			style="--from-top:{bookTransition.book.from.top}px; --from-left:{bookTransition.book.from
				.left}px; --from-width:{bookTransition.book.from.width}px; --from-height:{bookTransition
				.book.from.height}px; --book-color:{bookTransition.book.color}; --book-edge:{bookTransition
				.book.edge};"
		>
			<div class="book-back"></div>
			<div class="paper-spread">
				<span class="paper-side paper-left"></span>
				<span class="paper-side paper-right"></span>
				<span class="paper-gutter"></span>
			</div>
			<div class="page-leaves">
				{#each pageLeaves as index (index)}
					<span class="page-leaf" style="--page-index:{index};"></span>
				{/each}
			</div>
			<div class="front-cover">
				<span class="cover-edge"></span>
				<span class="cover-frame"></span>
				<span class="cover-title">{bookTransition.book.title}</span>
				<span class="cover-mark"></span>
			</div>
		</div>
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";

	.book-transition-layer {
		position: fixed;
		inset: 0;
		z-index: 100000;
		pointer-events: auto;
		perspective: 1600px;
	}

	.transition-shade {
		position: absolute;
		inset: 0;
		background:
			radial-gradient(circle at 50% 48%, rgb(255 247 226 / 0.13), transparent 28%),
			rgb(27 22 18 / 0);
		transition:
			background-color 320ms ease,
			opacity 220ms ease;
	}

	.opening .transition-shade {
		background-color: rgb(27 22 18 / 0.58);
	}

	.revealing .transition-shade {
		opacity: 0;
	}

	.transition-book {
		position: absolute;
		top: var(--from-top);
		left: var(--from-left);
		width: var(--from-width);
		height: var(--from-height);
		transform-style: preserve-3d;
		filter: drop-shadow(10px 18px 18px rgb(15 10 7 / 0.2));
		transition:
			top 720ms cubic-bezier(0.16, 0.78, 0.18, 1),
			left 720ms cubic-bezier(0.16, 0.78, 0.18, 1),
			width 720ms cubic-bezier(0.16, 0.78, 0.18, 1),
			height 720ms cubic-bezier(0.16, 0.78, 0.18, 1),
			transform 720ms cubic-bezier(0.16, 0.78, 0.18, 1),
			filter 220ms ease,
			opacity 220ms ease;
	}

	.opening .transition-book {
		top: 30px;
		left: 50%;
		width: min(45vw, 720px);
		height: calc(100vh - 66px);
		transform: translate(0, 0) rotateX(0deg);
		filter: drop-shadow(26px 36px 34px rgb(10 7 5 / 0.46));
	}

	.revealing .transition-book {
		top: 30px;
		left: 50%;
		width: min(45vw, 720px);
		height: calc(100vh - 66px);
		transform: translate(0, 0) scale(1.01);
		filter: blur(4px) drop-shadow(22px 34px 32px rgb(10 7 5 / 0.3));
		opacity: 0;
	}

	.book-back,
	.front-cover,
	.paper-spread,
	.page-leaves {
		position: absolute;
		inset: 0;
		transform-style: preserve-3d;
	}

	.book-back {
		border: 2px solid var(--book-edge);
		border-radius: 5px 9px 9px 5px;
		background:
			linear-gradient(90deg, rgb(255 255 255 / 0.13), transparent 14% 86%, rgb(0 0 0 / 0.18)),
			var(--book-color);
		box-shadow: inset 4px 0 0 rgb(0 0 0 / 0.12);
	}

	.paper-spread {
		inset: 7px 6px 7px 8px;
		opacity: 0;
		transition: opacity 100ms ease 170ms;
	}

	.opening .paper-spread,
	.revealing .paper-spread {
		opacity: 1;
	}

	.paper-side {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 100%;
		border: 1px solid #c9bda8;
		background:
			repeating-linear-gradient(0deg, transparent 0 21px, rgb(118 94 67 / 0.06) 21px 22px),
			linear-gradient(90deg, #e9dfce, #fffaf0 12% 86%, #d8cbb7);
	}

	.paper-left {
		right: 100%;
		border-radius: 8px 3px 3px 8px;
	}

	.paper-right {
		left: 0;
		border-radius: 3px 8px 8px 3px;
	}

	.paper-gutter {
		position: absolute;
		top: 0;
		bottom: 0;
		left: -3px;
		width: 8px;
		background: linear-gradient(90deg, transparent, rgb(83 60 39 / 0.2), transparent);
	}

	.page-leaves {
		inset: 8px 7px 8px 8px;
	}

	.page-leaf {
		position: absolute;
		inset: 0;
		transform-origin: left center;
		border: 1px solid rgb(165 145 117 / 0.5);
		border-radius: 2px 8px 8px 2px;
		background:
			repeating-linear-gradient(0deg, transparent 0 20px, rgb(118 94 67 / 0.07) 20px 21px),
			linear-gradient(90deg, #ded2be, #fffaf0 9% 91%, #d6c7b0);
		backface-visibility: hidden;
		opacity: 0;
	}

	.opening .page-leaf {
		animation: page-turn 430ms cubic-bezier(0.55, 0.08, 0.45, 0.95) forwards;
		animation-delay: calc(90ms + var(--page-index) * 48ms);
	}

	.front-cover {
		transform-origin: left center;
		border: 2px solid var(--book-edge);
		border-radius: 5px 9px 9px 5px;
		background:
			linear-gradient(90deg, rgb(255 255 255 / 0.14), transparent 12% 84%, rgb(0 0 0 / 0.19)),
			repeating-linear-gradient(0deg, rgb(255 255 255 / 0.018) 0 1px, transparent 1px 4px),
			var(--book-color);
		box-shadow:
			inset 7px 0 0 rgb(0 0 0 / 0.12),
			inset -3px 0 rgb(0 0 0 / 0.1);
		transition: transform 680ms cubic-bezier(0.16, 0.78, 0.18, 1);
		backface-visibility: hidden;
	}

	.opening .front-cover,
	.revealing .front-cover {
		transform: rotateY(-168deg);
	}

	.cover-edge {
		position: absolute;
		top: -5px;
		right: -1px;
		left: 8px;
		height: 6px;
		transform: skewX(-14deg);
		border: 1px solid rgb(114 84 54 / 0.38);
		background: repeating-linear-gradient(0deg, #e8dece 0 1px, #cbbca7 1px 2px);
	}

	.cover-frame {
		position: absolute;
		inset: 18px 16px;
		border: 1px solid rgb(235 207 153 / 0.58);
		box-shadow: inset 0 0 0 3px rgb(30 20 13 / 0.07);
	}

	.cover-title {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -55%);
		color: #ead9b7;
		font-family: var(--font-serif);
		font-size: clamp(1.35rem, 2.4vw, 2rem);
		letter-spacing: 0.2em;
		text-shadow: 0 2px 2px rgb(33 21 13 / 0.55);
		writing-mode: vertical-rl;
		text-orientation: upright;
	}

	.cover-mark {
		position: absolute;
		bottom: 28px;
		left: 50%;
		width: 22px;
		height: 8px;
		transform: translateX(-50%);
		border-top: 1px solid rgb(235 207 153 / 0.68);
		border-bottom: 1px solid rgb(235 207 153 / 0.48);
	}

	@keyframes page-turn {
		0% {
			transform: rotateY(0deg) translateZ(calc(var(--page-index) * 0.4px));
			opacity: 1;
		}
		48% {
			transform: rotateY(-82deg) translateZ(10px) skewY(-1.5deg);
			opacity: 1;
		}
		100% {
			transform: rotateY(-178deg) translateZ(calc(var(--page-index) * 0.3px));
			opacity: 1;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.transition-book,
		.transition-shade,
		.front-cover,
		.page-leaf {
			transition-duration: 0.01ms;
			animation-duration: 0.01ms;
			animation-delay: 0ms;
		}
	}
</style>
