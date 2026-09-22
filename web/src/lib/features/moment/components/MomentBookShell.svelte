<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { ArrowLeft, BookOpenText, List, X } from 'lucide-svelte';
	import { cubicOut } from 'svelte/easing';
	import { fade, fly } from 'svelte/transition';
	import { type Snippet } from 'svelte';
	import OpenBookFrame from './OpenBookFrame.svelte';

	interface Props {
		directory: Snippet;
		children: Snippet;
		pageLabel?: string;
	}

	let { directory, children, pageLabel = '手记' }: Props = $props();
	let isDirectoryOpen = $state(false);

	function closeDirectory() {
		isDirectoryOpen = false;
	}
</script>

<section class="moment-book-room" aria-label={`${pageLabel}书页`}>
	<header class="moment-book-topbar">
		<a class="back-to-shelf" href={resolvePath('/')}>
			<ArrowLeft size={16} strokeWidth={1.6} aria-hidden="true" />
			<span>回到书架</span>
		</a>
		<div class="book-identity" aria-label="当前书籍：手记">
			<BookOpenText size={15} strokeWidth={1.5} aria-hidden="true" />
			<span>手记</span>
			<i aria-hidden="true"></i>
			<small>NOTEBOOK</small>
		</div>
	</header>

	<button
		type="button"
		class="directory-tab"
		aria-expanded={isDirectoryOpen}
		aria-controls="moment-mobile-directory"
		onclick={() => (isDirectoryOpen = true)}
	>
		<List size={15} strokeWidth={1.7} aria-hidden="true" />
		<span>目录</span>
	</button>

	<div class="moment-book-spread">
		<OpenBookFrame />
		<aside class="moment-book-directory" aria-label="手记目录">
			<div class="directory-sticky">{@render directory()}</div>
		</aside>
		<div class="moment-book-page">
			{@render children()}
		</div>
		<div class="book-ribbon" aria-hidden="true">
			<span>生活很长<br />记得慢慢记录。</span>
			<i>Shawn</i>
		</div>
	</div>

	{#if isDirectoryOpen}
		<button
			type="button"
			class="directory-overlay"
			aria-label="关闭目录"
			onclick={closeDirectory}
			transition:fade={{ duration: 180 }}
		></button>
		<aside
			id="moment-mobile-directory"
			class="mobile-directory"
			aria-label="手记目录"
			transition:fly={{ x: -24, duration: 260, easing: cubicOut, opacity: 0 }}
		>
			<div class="mobile-directory-head">
				<div>
					<span>CONTENTS</span>
					<strong>手记目录</strong>
				</div>
				<button type="button" aria-label="关闭目录" onclick={closeDirectory}>
					<X size={18} strokeWidth={1.5} aria-hidden="true" />
				</button>
			</div>
			<div class="mobile-directory-body">
				{@render directory()}
			</div>
		</aside>
	{/if}
</section>

<style>
	.moment-book-room {
		--paper: #eadfc6;
		--paper-deep: #ded0b3;
		--paper-edge: #c5ae86;
		--book-ink: #382f28;
		--book-muted: #7d6e5f;
		--book-faint: #a1917c;
		--book-rule: rgba(76, 58, 43, 0.16);
		--book-accent: #8d382d;
		position: relative;
		z-index: 2;
		min-height: 100vh;
		padding: 1.25rem clamp(1rem, 3vw, 3.5rem) clamp(2rem, 5vw, 5rem);
		color: var(--book-ink);
		color-scheme: light;
		background-color: #181411;
		background-image:
			linear-gradient(
				90deg,
				rgba(20, 14, 10, 0.56),
				rgba(28, 19, 13, 0.22) 48%,
				rgba(18, 12, 9, 0.58)
			),
			linear-gradient(180deg, rgba(24, 16, 11, 0.06), rgba(24, 16, 11, 0.46)),
			url('/moments/moments-room-night.png');
		background-position: center top;
		background-size: cover;
		background-attachment: fixed;
	}

	.moment-book-room::before {
		content: '';
		position: fixed;
		inset: 0;
		pointer-events: none;
		opacity: 0.025;
		background-image: var(--texture-noise);
		mix-blend-mode: soft-light;
	}

	.moment-book-topbar {
		position: relative;
		z-index: 3;
		display: flex;
		width: min(100%, 94rem);
		margin: 0 auto 4.8rem;
		align-items: center;
		justify-content: space-between;
		color: rgba(242, 230, 207, 0.74);
		text-shadow: 0 1px 8px rgba(25, 15, 9, 0.56);
	}

	.back-to-shelf,
	.book-identity {
		display: inline-flex;
		align-items: center;
		gap: 0.55rem;
	}

	.back-to-shelf {
		min-height: 2.5rem;
		padding: 0 0.2rem;
		font-family: var(--font-serif);
		font-size: 0.8rem;
		letter-spacing: 0.14em;
		transition: color 180ms ease;
	}

	.back-to-shelf:hover,
	.back-to-shelf:focus-visible {
		color: #fff7e9;
	}

	.back-to-shelf:focus-visible,
	.directory-tab:focus-visible,
	.mobile-directory button:focus-visible {
		outline: 2px solid #cba874;
		outline-offset: 3px;
	}

	.book-identity {
		display: none;
		font-family: var(--font-serif);
		font-size: 0.76rem;
		letter-spacing: 0.14em;
	}

	.book-identity i {
		width: 1px;
		height: 0.9rem;
		background: rgba(242, 230, 207, 0.26);
	}

	.book-identity small {
		font-family: var(--font-mono);
		font-size: 0.58rem;
		letter-spacing: 0.24em;
		opacity: 0.65;
	}

	.moment-book-spread {
		position: relative;
		z-index: 2;
		display: grid;
		grid-template-columns: minmax(14.5rem, 1fr) minmax(0, 4fr);
		width: min(100%, 86rem);
		min-height: clamp(46rem, calc(100vh - 13rem), 54rem);
		margin: 0 auto;
		overflow: visible;
		border: 0;
		border-radius: 0;
		background: transparent;
		box-shadow: none;
		transform: perspective(1800px) rotateX(4deg);
		transform-origin: center bottom;
	}

	.moment-book-spread::after {
		content: '';
		position: absolute;
		inset: 0;
		z-index: 2;
		pointer-events: none;
		opacity: 0.045;
		background-image: var(--texture-noise);
		mix-blend-mode: multiply;
	}

	.moment-book-directory,
	.moment-book-page {
		position: relative;
		z-index: 1;
	}

	.moment-book-directory {
		min-width: 0;
		padding: clamp(2.1rem, 4vw, 4.4rem) clamp(1.35rem, 2.4vw, 2.7rem);
		border-right: 1px solid rgba(78, 58, 39, 0.2);
		background:
			linear-gradient(90deg, rgba(94, 65, 38, 0.08), transparent 9%),
			linear-gradient(90deg, transparent 80%, rgba(76, 51, 29, 0.1));
		box-shadow:
			inset -1.1rem 0 1.8rem -1.5rem rgba(49, 30, 17, 0.6),
			inset 0.28rem 0 rgba(129, 83, 48, 0.2);
		border-radius: 5px 0 0 5px;
	}

	.directory-sticky {
		position: sticky;
		top: 2rem;
	}

	.moment-book-page {
		min-width: 0;
		padding: clamp(2.4rem, 5vw, 5.6rem) clamp(1.8rem, 5.4vw, 6.3rem);
		background:
			linear-gradient(90deg, rgba(81, 54, 31, 0.1), transparent 2.6rem),
			radial-gradient(circle at 90% 8%, rgba(255, 252, 237, 0.32), transparent 30rem);
		box-shadow: inset 1.3rem 0 1.8rem -1.7rem rgba(45, 28, 17, 0.7);
		border-radius: 0 8px 8px 0;
	}

	.moment-book-page::after {
		content: '';
		position: absolute;
		inset: 0 0 0 auto;
		width: clamp(1.2rem, 2vw, 2.2rem);
		pointer-events: none;
		border-radius: 0 8px 8px 0;
		background: linear-gradient(
			90deg,
			transparent,
			rgba(103, 75, 44, 0.035) 35%,
			rgba(72, 48, 29, 0.14) 100%
		);
		mix-blend-mode: multiply;
	}

	.book-ribbon {
		position: absolute;
		top: -0.2rem;
		right: 0.8rem;
		z-index: 4;
		display: flex;
		width: 3.75rem;
		height: 11.5rem;
		flex-direction: column;
		align-items: center;
		justify-content: space-between;
		padding: 1.4rem 0.6rem 1.65rem;
		color: rgba(245, 221, 184, 0.82);
		border: 1px solid rgba(84, 43, 30, 0.5);
		background:
			linear-gradient(
				90deg,
				rgba(78, 37, 27, 0.24),
				transparent 30%,
				rgba(255, 218, 167, 0.08) 65%,
				rgba(64, 29, 22, 0.18)
			),
			#8e4f3a;
		box-shadow:
			0 0.8rem 1.2rem rgba(50, 27, 19, 0.24),
			inset 0 0 0 1px rgba(239, 200, 151, 0.13);
		clip-path: polygon(0 0, 100% 0, 100% 91%, 50% 100%, 0 91%);
		transform: rotate(-1.4deg);
	}

	.book-ribbon span {
		font-family: var(--font-serif);
		font-size: 0.56rem;
		line-height: 1.9;
		letter-spacing: 0.1em;
		writing-mode: vertical-rl;
	}

	.book-ribbon i {
		font-family: var(--font-serif);
		font-size: 0.64rem;
		font-style: italic;
		letter-spacing: 0.06em;
	}

	.directory-tab,
	.directory-overlay,
	.mobile-directory {
		display: none;
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

	@media (max-width: 767px) {
		.moment-book-room {
			padding: 0;
			background-color: var(--paper);
			background-image: none;
			background-attachment: scroll;
		}

		.moment-book-room::before {
			position: absolute;
			opacity: 0.025;
			mix-blend-mode: multiply;
		}

		.moment-book-topbar {
			position: sticky;
			top: 0;
			z-index: 20;
			width: 100%;
			min-height: 3.6rem;
			margin: 0;
			padding: 0 1rem 0 1.1rem;
			color: var(--book-ink);
			border-bottom: 1px solid var(--book-rule);
			background: rgba(234, 223, 198, 0.94);
			backdrop-filter: blur(12px);
		}

		.back-to-shelf {
			font-size: 0.72rem;
			letter-spacing: 0.1em;
		}

		.book-identity small,
		.book-identity i,
		.book-identity :global(svg) {
			display: none;
		}

		.book-identity {
			display: inline-flex;
			color: var(--book-muted);
		}

		.moment-book-spread {
			display: block;
			width: 100%;
			min-height: calc(100vh - 3.6rem);
			border: 0;
			border-radius: 0;
			box-shadow: none;
			overflow: clip;
			transform: none;
		}

		.moment-book-directory {
			display: none;
		}

		.moment-book-page {
			min-height: calc(100vh - 3.6rem);
			padding: 2.5rem 1.25rem 4.5rem 2.15rem;
			box-shadow: none;
			border-radius: 0;
			background: linear-gradient(90deg, rgba(112, 72, 42, 0.1), transparent 1.2rem), var(--paper);
		}

		.moment-book-page::after,
		.book-ribbon {
			display: none;
		}

		.directory-tab {
			position: fixed;
			top: 42%;
			left: 0;
			z-index: 30;
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 0.3rem;
			padding: 0.65rem 0.42rem 0.7rem;
			color: #efe3c9;
			border: 1px solid rgba(219, 195, 157, 0.34);
			border-left: 0;
			border-radius: 0 5px 5px 0;
			background: #5b4030;
			box-shadow: 0 0.45rem 1.1rem rgba(43, 27, 17, 0.26);
		}

		.directory-tab span {
			font-family: var(--font-serif);
			font-size: 0.68rem;
			letter-spacing: 0.12em;
			writing-mode: vertical-rl;
		}

		.directory-overlay {
			position: fixed;
			inset: 0;
			z-index: 70;
			display: block;
			border: 0;
			background: rgba(24, 18, 14, 0.58);
			backdrop-filter: blur(2px);
		}

		.mobile-directory {
			position: fixed;
			inset: 0 auto 0 0;
			z-index: 80;
			display: flex;
			width: min(84vw, 22rem);
			flex-direction: column;
			color: var(--book-ink);
			border-right: 1px solid var(--paper-edge);
			background:
				linear-gradient(90deg, rgba(103, 69, 39, 0.08), transparent 1.2rem), var(--paper-deep);
			box-shadow: 1.4rem 0 3.2rem rgba(0, 0, 0, 0.36);
		}

		.mobile-directory-head {
			display: flex;
			min-height: 4.8rem;
			align-items: center;
			justify-content: space-between;
			padding: 0.8rem 1.1rem 0.75rem 1.35rem;
			border-bottom: 1px solid var(--book-rule);
		}

		.mobile-directory-head div {
			display: flex;
			flex-direction: column;
			gap: 0.22rem;
		}

		.mobile-directory-head span {
			font-family: var(--font-mono);
			font-size: 0.55rem;
			letter-spacing: 0.24em;
			color: var(--book-faint);
		}

		.mobile-directory-head strong {
			font-family: var(--font-serif);
			font-size: 1.05rem;
			font-weight: 600;
		}

		.mobile-directory-head button {
			display: grid;
			width: 2.4rem;
			height: 2.4rem;
			place-items: center;
			border-radius: 50%;
			color: var(--book-muted);
		}

		.mobile-directory-body {
			overflow-y: auto;
			padding: 1.5rem 1.35rem 3rem;
		}
	}

	@media (min-width: 768px) and (max-width: 1100px) {
		.book-ribbon {
			display: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.back-to-shelf {
			transition: none;
		}
	}
</style>
