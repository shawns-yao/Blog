<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { goto, preloadData } from '$app/navigation';
	import { BookOpenText } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { type Snippet } from 'svelte';
	import OpenBookFrame from './OpenBookFrame.svelte';

	interface Props {
		directory: Snippet;
		children: Snippet;
		overlay?: Snippet;
		pageLabel?: string;
	}

	let { directory, children, overlay, pageLabel = '手记' }: Props = $props();
	let isReturningHome = $state(false);
	const homePath = resolvePath('/');

	type MomentScene = {
		key: string;
		label: string;
		startMinute: number;
		src: string;
	};

	const MOMENT_SCENES: readonly MomentScene[] = [
		{ key: 'midnight', label: '凌晨', startMinute: 0, src: '/moments/scenes/00-midnight.webp' },
		{ key: 'dawn', label: '清晨', startMinute: 330, src: '/moments/scenes/05-dawn.webp' },
		{ key: 'morning', label: '上午', startMinute: 480, src: '/moments/scenes/08-morning.webp' },
		{ key: 'noon', label: '午间', startMinute: 690, src: '/moments/scenes/11-noon.webp' },
		{ key: 'afternoon', label: '下午', startMinute: 900, src: '/moments/scenes/15-afternoon.webp' },
		{ key: 'sunset', label: '黄昏', startMinute: 1080, src: '/moments/scenes/18-sunset.webp' },
		{ key: 'night', label: '夜间', startMinute: 1230, src: '/moments/scenes/20-night.webp' },
		{
			key: 'late-night',
			label: '深夜',
			startMinute: 1380,
			src: '/moments/scenes/23-late-night.webp'
		}
	];

	const resolveSceneIndex = (date: Date) => {
		const currentMinute = date.getHours() * 60 + date.getMinutes();
		for (let index = MOMENT_SCENES.length - 1; index >= 0; index -= 1) {
			if (currentMinute >= MOMENT_SCENES[index].startMinute) return index;
		}
		return 0;
	};

	let currentScene = $state(MOMENT_SCENES[resolveSceneIndex(new Date())]);
	let previousScene = $state<MomentScene | null>(null);

	function returnToShelf(event: MouseEvent) {
		if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
			return;
		}
		event.preventDefault();
		if (isReturningHome) return;
		isReturningHome = true;
		void preloadData(homePath).catch(() => undefined);
		const navigationDelay = window.matchMedia('(prefers-reduced-motion: reduce)').matches
			? 0
			: 1150;
		window.setTimeout(() => void goto(homePath), navigationDelay);
	}

	onMount(() => {
		let transitionTimer: ReturnType<typeof setTimeout> | undefined;

		const syncScene = () => {
			const sceneIndex = resolveSceneIndex(new Date());
			const nextScene = MOMENT_SCENES[sceneIndex];
			if (nextScene.key !== currentScene.key) {
				previousScene = currentScene;
				currentScene = nextScene;
				clearTimeout(transitionTimer);
				transitionTimer = setTimeout(() => (previousScene = null), 900);
			}

			const preload = new Image();
			preload.src = MOMENT_SCENES[(sceneIndex + 1) % MOMENT_SCENES.length].src;
		};

		const handleVisibilityChange = () => {
			if (document.visibilityState === 'visible') syncScene();
		};

		syncScene();
		const interval = window.setInterval(syncScene, 30_000);
		window.addEventListener('focus', syncScene);
		document.addEventListener('visibilitychange', handleVisibilityChange);

		return () => {
			clearTimeout(transitionTimer);
			window.clearInterval(interval);
			window.removeEventListener('focus', syncScene);
			document.removeEventListener('visibilitychange', handleVisibilityChange);
		};
	});
</script>

<section
	class="moment-book-room"
	data-scene={currentScene.key}
	aria-label={`${pageLabel}书页，当前为${currentScene.label}场景`}
>
	<div class="moment-room-backdrop" aria-hidden="true">
		{#if previousScene}
			<div
				class="moment-room-image moment-room-image-previous"
				style:background-image={`url('${previousScene.src}')`}
			></div>
		{/if}
		{#key currentScene.key}
			<div
				class="moment-room-image moment-room-image-current"
				style:background-image={`url('${currentScene.src}')`}
			></div>
		{/key}
	</div>

	{#if currentScene.key === 'night'}
		<span class="shooting-star" aria-hidden="true"></span>
	{/if}

	{#if currentScene.key === 'late-night'}
		<img class="cat-paw" src="/moments/scenes/cat-paw.png" alt="" aria-hidden="true" />
	{/if}
	<header class="moment-book-topbar" inert={!!overlay}>
		<a
			class:ladybug-is-flying={isReturningHome}
			class="back-to-shelf"
			href={homePath}
			aria-label="回到书架"
			onclick={returnToShelf}
		>
			<span class="ladybug-flight" aria-hidden="true">
				<svg class="ladybug" viewBox="0 0 48 34" role="presentation">
					<g class="ladybug-legs" fill="none" stroke="currentColor" stroke-linecap="round">
						<path d="M17 18 10 13M16 22 8 22M18 26 11 31M31 18l7-5M32 22h8M30 26l7 5" />
					</g>
					<g class="ladybug-wings">
						<ellipse class="ladybug-wing ladybug-wing-left" cx="21" cy="18" rx="11" ry="6" />
						<ellipse class="ladybug-wing ladybug-wing-right" cx="27" cy="18" rx="11" ry="6" />
					</g>
					<ellipse class="ladybug-body" cx="24" cy="21" rx="13" ry="10" />
					<g class="ladybug-shell-half ladybug-shell-left">
						<path
							class="ladybug-shell"
							d="M23.6 11.2C16.5 11.4 11 15.4 11 21c0 5.5 5.4 9.5 12.6 9.8Z"
						/>
						<circle class="ladybug-spot" cx="18.3" cy="17" r="1.8" />
						<circle class="ladybug-spot" cx="17.2" cy="24.7" r="1.55" />
						<path class="ladybug-shell-shine" d="M15.2 16.2c1.4-1.8 3.1-2.6 5.1-2.9" />
					</g>
					<g class="ladybug-shell-half ladybug-shell-right">
						<path
							class="ladybug-shell"
							d="M24.4 11.2C31.5 11.4 37 15.4 37 21c0 5.5-5.4 9.5-12.6 9.8Z"
						/>
						<circle class="ladybug-spot" cx="29.7" cy="17" r="1.8" />
						<circle class="ladybug-spot" cx="30.8" cy="24.7" r="1.55" />
						<path class="ladybug-shell-shine" d="M27.7 13.3c2 0.3 3.7 1.1 5.1 2.9" />
					</g>
					<circle class="ladybug-head" cx="24" cy="9.5" r="6" />
					<g class="ladybug-eyes">
						<circle cx="21.5" cy="7.8" r="0.85" />
						<circle cx="26.5" cy="7.8" r="0.85" />
					</g>
					<g class="ladybug-antennae" fill="none" stroke="currentColor" stroke-linecap="round">
						<path d="M21 5.1c-2.2-3-4.4-2.7-5.4-1.4M27 5.1c2.2-3 4.4-2.7 5.4-1.4" />
					</g>
				</svg>
			</span>
			<span>回到书架</span>
		</a>
		<div class="book-identity" aria-label="当前书籍：手记">
			<BookOpenText size={15} strokeWidth={1.5} aria-hidden="true" />
			<span>手记</span>
			<i aria-hidden="true"></i>
			<small>NOTEBOOK</small>
		</div>
	</header>

	<div class:under-overlay={!!overlay} class="moment-book-spread" inert={!!overlay}>
		<OpenBookFrame />
		<aside class="moment-book-directory" aria-label="左侧书页" role="region">
			<div class="directory-sticky">{@render directory()}</div>
		</aside>
		<div class="moment-book-page" aria-label="右侧书页" role="region">
			{@render children()}
		</div>
		<div class="book-ribbon" aria-hidden="true">
			<span>生活很长<br />记得慢慢记录。</span>
			<i>Shawn</i>
		</div>
	</div>

	{#if overlay}
		<div class="moment-book-overlay" role="presentation">{@render overlay()}</div>
	{/if}
</section>

<style>
	.moment-book-room {
		--paper: #eee1c5;
		--paper-deep: #dfcfac;
		--paper-edge: #c3a878;
		--book-ink: #382f28;
		--book-muted: #7a6958;
		--book-faint: #9b8970;
		--book-rule: rgba(76, 58, 43, 0.16);
		--book-accent: #8d382d;
		position: relative;
		z-index: 2;
		height: 100dvh;
		min-height: 42rem;
		overflow: hidden;
		color: var(--book-ink);
		color-scheme: light;
		background-color: #17110d;
	}

	.moment-book-room[data-scene='dawn'] {
		--paper: #eee0c7;
		--paper-deep: #dac8a7;
	}

	.moment-book-room[data-scene='morning'],
	.moment-book-room[data-scene='noon'] {
		--paper: #f1e7d0;
		--paper-deep: #dfcfaf;
	}

	.moment-book-room[data-scene='afternoon'] {
		--paper: #f0e1c3;
		--paper-deep: #ddc8a2;
	}

	.moment-book-room[data-scene='sunset'] {
		--paper: #eddbba;
		--paper-deep: #d9bf96;
	}

	.moment-room-backdrop,
	.moment-room-image {
		position: absolute;
		inset: 0;
	}

	.moment-room-backdrop {
		z-index: 0;
		overflow: hidden;
		pointer-events: none;
	}

	.moment-room-backdrop::after {
		content: '';
		position: absolute;
		inset: 0;
		z-index: 2;
		background:
			linear-gradient(90deg, rgba(20, 13, 9, 0.2), transparent 28% 72%, rgba(20, 13, 9, 0.24)),
			linear-gradient(180deg, rgba(19, 13, 10, 0.02), rgba(19, 13, 10, 0.2));
	}

	.moment-room-image {
		background-position: center;
		background-repeat: no-repeat;
		background-size: cover;
	}

	.moment-room-image-previous {
		z-index: 0;
	}

	.moment-room-image-current {
		z-index: 1;
		animation: scene-crossfade 900ms ease-out both;
	}

	.moment-book-room::after {
		content: '';
		position: absolute;
		inset: auto 0 0;
		z-index: 1;
		height: 44%;
		pointer-events: none;
		background: radial-gradient(ellipse at 47% 82%, rgba(255, 205, 131, 0.12), transparent 58%);
		mix-blend-mode: screen;
	}

	.shooting-star {
		position: absolute;
		top: 12%;
		left: 58%;
		z-index: 1;
		width: 5.5rem;
		height: 1px;
		pointer-events: none;
		background: linear-gradient(90deg, transparent, rgba(255, 247, 219, 0.9));
		filter: drop-shadow(0 0 4px rgba(220, 236, 255, 0.75));
		transform: rotate(-21deg) translateX(-10rem);
		animation: shooting-star 13s ease-in infinite;
	}

	.shooting-star::after {
		content: '';
		position: absolute;
		top: -1px;
		right: -1px;
		width: 3px;
		height: 3px;
		border-radius: 50%;
		background: #fff8dc;
	}

	.cat-paw {
		position: absolute;
		right: -1.5rem;
		bottom: 9vh;
		z-index: 3;
		width: clamp(9rem, 14vw, 17rem);
		pointer-events: none;
		filter: drop-shadow(-0.45rem 0.55rem 0.7rem rgba(41, 24, 14, 0.28));
		transform: rotate(-7deg) translate(1.5rem, 1rem);
		transform-origin: right bottom;
		animation: paw-peek 1.15s cubic-bezier(0.2, 0.8, 0.2, 1) 520ms both;
	}

	.moment-book-topbar {
		position: absolute;
		top: 1rem;
		left: 50%;
		z-index: 10;
		display: flex;
		width: min(92vw, 92rem);
		min-height: 2.8rem;
		align-items: center;
		justify-content: space-between;
		color: rgba(246, 233, 209, 0.8);
		text-shadow: 0 1px 10px rgba(24, 14, 8, 0.72);
		transform: translateX(-50%);
	}

	.back-to-shelf,
	.book-identity {
		display: inline-flex;
		align-items: center;
		gap: 0.55rem;
	}

	.back-to-shelf {
		min-height: 2.5rem;
		padding: 0 0.35rem 0 0.1rem;
		font-family: var(--font-serif);
		font-size: 0.78rem;
		letter-spacing: 0.14em;
		cursor: pointer;
		transition: color 180ms ease;
	}

	.ladybug-flight {
		position: relative;
		display: grid;
		width: 2.9rem;
		height: 2.15rem;
		place-items: center;
		transform-origin: 50% 60%;
		filter: drop-shadow(0 0.28rem 0.25rem rgba(26, 15, 8, 0.48));
		will-change: transform, opacity;
	}

	.ladybug {
		display: block;
		width: 2.75rem;
		height: auto;
		overflow: visible;
		color: #211915;
		transform-box: fill-box;
		transform-origin: center;
	}

	.ladybug-legs,
	.ladybug-antennae {
		stroke-width: 1.35;
	}

	.ladybug-body,
	.ladybug-head,
	.ladybug-spot {
		fill: #201814;
	}

	.ladybug-shell {
		fill: #b83d33;
		stroke: #5b211d;
		stroke-width: 0.8;
	}

	.ladybug-shell-half {
		transform-box: fill-box;
		transition: transform 180ms ease;
	}

	.ladybug-shell-shine {
		fill: none;
		stroke: rgba(255, 194, 160, 0.68);
		stroke-width: 1.15;
		stroke-linecap: round;
	}

	.ladybug-eyes {
		fill: #f7dfb2;
	}

	.ladybug-shell-left {
		transform-origin: right center;
	}

	.ladybug-shell-right {
		transform-origin: left center;
	}

	.ladybug-wing {
		fill: rgba(245, 225, 180, 0.82);
		stroke: rgba(82, 55, 36, 0.72);
		stroke-width: 0.65;
		transform-box: fill-box;
		opacity: 0;
	}

	.ladybug-wing-left {
		transform-origin: right center;
	}

	.ladybug-wing-right {
		transform-origin: left center;
	}

	.back-to-shelf:hover .ladybug-shell-left,
	.back-to-shelf:focus-visible .ladybug-shell-left {
		transform: rotate(-7deg);
	}

	.back-to-shelf:hover .ladybug-shell-right,
	.back-to-shelf:focus-visible .ladybug-shell-right {
		transform: rotate(7deg);
	}

	.back-to-shelf.ladybug-is-flying {
		pointer-events: none;
	}

	.back-to-shelf.ladybug-is-flying > span:last-child {
		animation: shelf-label-leave 420ms ease-out both;
	}

	.back-to-shelf.ladybug-is-flying .ladybug-flight {
		animation: ladybug-takeoff 1150ms cubic-bezier(0.42, 0, 0.24, 1) both;
	}

	.back-to-shelf.ladybug-is-flying .ladybug {
		animation: ladybug-heading 1150ms linear both;
	}

	.back-to-shelf.ladybug-is-flying .ladybug-shell-left {
		animation: ladybug-shell-left-open 190ms ease-out both;
	}

	.back-to-shelf.ladybug-is-flying .ladybug-shell-right {
		animation: ladybug-shell-right-open 190ms ease-out both;
	}

	.back-to-shelf.ladybug-is-flying .ladybug-wing-left {
		animation: ladybug-wing-left-flutter 140ms ease-in-out 120ms infinite alternate;
	}

	.back-to-shelf.ladybug-is-flying .ladybug-wing-right {
		animation: ladybug-wing-right-flutter 140ms ease-in-out 120ms infinite alternate;
	}

	.back-to-shelf:hover,
	.back-to-shelf:focus-visible {
		color: #fff7e9;
	}

	.back-to-shelf:focus-visible {
		outline: 2px solid #cba874;
		outline-offset: 3px;
	}

	.book-identity {
		font-family: var(--font-serif);
		font-size: 0.76rem;
		letter-spacing: 0.14em;
	}

	.book-identity i {
		width: 1px;
		height: 0.9rem;
		background: rgba(242, 230, 207, 0.3);
	}

	.book-identity small {
		font-family: var(--font-mono);
		font-size: 0.56rem;
		letter-spacing: 0.24em;
		opacity: 0.72;
	}

	.moment-book-spread {
		position: absolute;
		bottom: 6vh;
		left: 50%;
		z-index: 2;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		width: min(92vw, 106rem);
		height: 90vh;
		min-height: 31rem;
		isolation: isolate;
		transform: translateX(-50%) perspective(1900px) rotateX(3.2deg);
		transform-origin: center bottom;
		animation: book-arrive 720ms cubic-bezier(0.16, 1, 0.3, 1) both;
	}

	.moment-book-directory,
	.moment-book-page {
		position: relative;
		z-index: 1;
		min-width: 0;
		height: 100%;
		overflow-y: auto;
		overscroll-behavior: contain;
		scrollbar-width: none;
	}

	.moment-book-directory::-webkit-scrollbar,
	.moment-book-page::-webkit-scrollbar {
		display: none;
	}

	.moment-book-directory:focus-visible,
	.moment-book-page:focus-visible {
		outline: 2px solid rgba(141, 56, 45, 0.5);
		outline-offset: -1.2rem;
	}

	.moment-book-directory {
		padding: clamp(3.3rem, 5.2vh, 4.4rem) clamp(2rem, 3.2vw, 4rem) 3.6rem;
		background:
			linear-gradient(90deg, rgba(92, 62, 35, 0.08), transparent 9%),
			linear-gradient(90deg, transparent 84%, rgba(74, 49, 29, 0.13));
		box-shadow: inset -1.2rem 0 1.7rem -1.5rem rgba(49, 30, 17, 0.5);
	}

	.directory-sticky {
		position: relative;
		height: 100%;
	}

	.moment-book-page {
		padding: clamp(3.3rem, 5.2vh, 4.4rem) clamp(2rem, 3.2vw, 4rem) 3.6rem;
		background:
			linear-gradient(90deg, rgba(81, 54, 31, 0.11), transparent 2.8rem),
			radial-gradient(circle at 82% 8%, rgba(255, 250, 230, 0.2), transparent 28rem);
		box-shadow: inset 1.25rem 0 1.8rem -1.55rem rgba(45, 28, 17, 0.62);
	}

	.moment-book-page::after {
		content: '';
		position: absolute;
		top: 0;
		right: 0;
		bottom: 0;
		width: clamp(1.2rem, 2vw, 2.2rem);
		pointer-events: none;
		background: linear-gradient(90deg, transparent, rgba(72, 48, 29, 0.12));
		mix-blend-mode: multiply;
	}

	.book-ribbon {
		position: absolute;
		top: 0.7rem;
		right: 1.4rem;
		z-index: 4;
		display: flex;
		width: 3.6rem;
		height: 10.8rem;
		flex-direction: column;
		align-items: center;
		justify-content: space-between;
		padding: 1.35rem 0.55rem 1.55rem;
		color: rgba(245, 221, 184, 0.84);
		border: 1px solid rgba(84, 43, 30, 0.5);
		background:
			linear-gradient(90deg, rgba(78, 37, 27, 0.24), transparent 32%, rgba(255, 218, 167, 0.08)),
			#8e4f3a;
		box-shadow:
			0 0.8rem 1.2rem rgba(50, 27, 19, 0.24),
			inset 0 0 0 1px rgba(239, 200, 151, 0.13);
		clip-path: polygon(0 0, 100% 0, 100% 91%, 50% 100%, 0 91%);
		transform: rotate(-1.4deg);
	}

	.book-ribbon span {
		font-family: var(--font-serif);
		font-size: 0.54rem;
		line-height: 1.9;
		letter-spacing: 0.1em;
		writing-mode: vertical-rl;
	}

	.book-ribbon i {
		font-family: var(--font-serif);
		font-size: 0.62rem;
		font-style: italic;
		letter-spacing: 0.06em;
	}

	.moment-book-room :global(.moment-book-copy) {
		color: var(--book-ink);
	}

	.moment-book-room :global(.moment-book-muted) {
		color: var(--book-muted);
	}

	.moment-book-room :global(.moment-book-rule) {
		border-color: var(--book-rule);
	}

	.moment-book-room :global(.moment-book-accent) {
		color: var(--book-accent);
	}

	.moment-book-spread.under-overlay {
		filter: brightness(0.55) saturate(0.72) blur(2px);
		transform: translateX(-50%) perspective(1900px) rotateX(3.2deg) scale(0.992);
		transition:
			filter 320ms ease,
			transform 420ms cubic-bezier(0.16, 1, 0.3, 1);
	}

	.moment-book-overlay {
		position: fixed;
		inset: 0;
		z-index: 40;
	}

	@keyframes book-arrive {
		from {
			opacity: 0;
			filter: blur(4px);
			transform: translateX(-50%) translateY(2.2rem) perspective(1900px) rotateX(7deg) scale(0.97);
		}
		to {
			opacity: 1;
			filter: blur(0);
			transform: translateX(-50%) perspective(1900px) rotateX(3.2deg) scale(1);
		}
	}

	@keyframes scene-crossfade {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@keyframes shooting-star {
		0%,
		72% {
			opacity: 0;
			transform: rotate(-21deg) translateX(-10rem);
		}
		74% {
			opacity: 0.9;
		}
		79% {
			opacity: 0;
			transform: rotate(-21deg) translateX(14rem);
		}
		100% {
			opacity: 0;
		}
	}

	@keyframes paw-peek {
		from {
			opacity: 0;
			transform: rotate(-2deg) translate(5rem, 3.5rem);
		}
		to {
			opacity: 1;
			transform: rotate(-7deg) translate(1.5rem, 1rem);
		}
	}

	@keyframes ladybug-shell-left-open {
		to {
			transform: rotate(-28deg);
		}
	}

	@keyframes ladybug-shell-right-open {
		to {
			transform: rotate(28deg);
		}
	}

	@keyframes ladybug-wing-left-flutter {
		from {
			opacity: 0.75;
			transform: rotate(-18deg);
		}
		to {
			opacity: 0.9;
			transform: rotate(-42deg);
		}
	}

	@keyframes ladybug-wing-right-flutter {
		from {
			opacity: 0.75;
			transform: rotate(18deg);
		}
		to {
			opacity: 0.9;
			transform: rotate(42deg);
		}
	}

	@keyframes ladybug-takeoff {
		0% {
			opacity: 1;
			transform: translate3d(0, 0, 0) rotate(0deg) scale(1);
		}
		14% {
			opacity: 1;
			transform: translate3d(3vw, -1vh, 0) rotate(8deg) scale(1.02);
		}
		31% {
			opacity: 1;
			transform: translate3d(10vw, 4vh, 0) rotate(22deg) scale(1.03);
		}
		49% {
			opacity: 1;
			transform: translate3d(19vw, 6vh, 0) rotate(46deg) scale(1);
		}
		67% {
			opacity: 0.96;
			transform: translate3d(29vw, 1vh, 0) rotate(78deg) scale(0.94);
		}
		84% {
			opacity: 0.75;
			transform: translate3d(37vw, -6vh, 0) rotate(108deg) scale(0.82);
		}
		100% {
			opacity: 0;
			transform: translate3d(45vw, -13vh, 0) rotate(132deg) scale(0.68);
		}
	}

	@keyframes ladybug-heading {
		0% {
			transform: rotate(0deg);
		}
		20% {
			transform: rotate(52deg);
		}
		50% {
			transform: rotate(112deg);
		}
		75% {
			transform: rotate(65deg);
		}
		100% {
			transform: rotate(43deg);
		}
	}

	@keyframes shelf-label-leave {
		to {
			opacity: 0;
			transform: translateX(-0.25rem);
		}
	}

	@media (min-width: 768px) and (max-width: 1180px) {
		.moment-book-spread {
			grid-template-columns: repeat(2, minmax(0, 1fr));
			width: 95vw;
		}

		.moment-book-directory,
		.moment-book-page {
			padding-right: 2rem;
			padding-left: 2rem;
		}

		.book-ribbon {
			display: none;
		}
	}

	@media (min-width: 768px) and (max-height: 760px) {
		.moment-book-spread {
			bottom: 1vh;
			height: 78vh;
			min-height: 29rem;
		}

		.moment-book-directory,
		.moment-book-page {
			padding-top: 2.8rem;
			padding-bottom: 3.6rem;
		}
	}

	@media (max-width: 767px) {
		.moment-book-room {
			height: auto;
			min-height: 100dvh;
			overflow: visible;
			background-color: var(--paper);
			background-image: none;
		}

		.moment-book-room::after {
			display: none;
		}

		.moment-room-backdrop,
		.shooting-star,
		.cat-paw {
			display: none;
		}

		.moment-book-topbar {
			position: sticky;
			top: 0;
			left: auto;
			z-index: 20;
			width: 100%;
			min-height: 3.6rem;
			padding: 0 1rem 0 1.1rem;
			color: var(--book-ink);
			border-bottom: 1px solid var(--book-rule);
			background: rgba(238, 225, 197, 0.94);
			backdrop-filter: blur(12px);
			transform: none;
			text-shadow: none;
		}

		.back-to-shelf {
			font-size: 0.72rem;
			letter-spacing: 0.1em;
		}

		.ladybug-flight {
			width: 2.55rem;
			height: 1.95rem;
		}

		.ladybug {
			width: 2.45rem;
		}

		.book-identity small,
		.book-identity i,
		.book-identity :global(svg) {
			display: none;
		}

		.book-identity {
			color: var(--book-muted);
		}

		.moment-book-spread {
			position: relative;
			bottom: auto;
			left: auto;
			display: block;
			width: 100%;
			height: auto;
			min-height: calc(100dvh - 3.6rem);
			overflow: clip;
			transform: none;
			animation: none;
		}

		.moment-book-directory,
		.moment-book-page {
			display: block;
			height: auto;
			min-height: calc(100dvh - 3.6rem);
			overflow: visible;
			padding: 2.5rem 1.25rem 3.5rem 2.15rem;
			box-shadow: none;
			background: linear-gradient(90deg, rgba(112, 72, 42, 0.1), transparent 1.2rem), var(--paper);
		}

		.moment-book-directory {
			border-bottom: 1px solid var(--paper-edge);
		}

		.moment-book-page::after,
		.book-ribbon {
			display: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.back-to-shelf {
			transition: none;
		}

		.back-to-shelf.ladybug-is-flying .ladybug-flight,
		.back-to-shelf.ladybug-is-flying .ladybug,
		.back-to-shelf.ladybug-is-flying .ladybug-shell-left,
		.back-to-shelf.ladybug-is-flying .ladybug-shell-right,
		.back-to-shelf.ladybug-is-flying .ladybug-wing-left,
		.back-to-shelf.ladybug-is-flying .ladybug-wing-right,
		.back-to-shelf.ladybug-is-flying > span:last-child {
			animation: none;
		}

		.moment-book-spread {
			animation: none;
		}

		.moment-room-image-current,
		.shooting-star,
		.cat-paw {
			animation: none;
		}
	}
</style>
