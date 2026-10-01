<script lang="ts">
	import { goto, onNavigate } from '$app/navigation';
	import { buildMomentBookSpreads, type MomentBookVisit } from '$lib/features/moment/book-pages';
	import type { MomentDetail, MomentListResponse } from '$lib/features/moment/types';
	import { detailHeroBgSrc } from '$lib/shared/stores/detailHeroBg';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { formatDateDotted } from '$lib/shared/utils/date';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { ArrowLeft, ArrowRight, Paperclip, X } from 'lucide-svelte';
	import { onDestroy, onMount } from 'svelte';
	import MomentBookShell from './MomentBookShell.svelte';
	import MomentDatePage from './MomentDatePage.svelte';
	import MomentDetailPaper from './moment-detail/MomentDetailPaper.svelte';

	let {
		moment,
		underlayMoments = { items: [], total: 0, page: 1, size: 20 }
	}: { moment: MomentDetail; underlayMoments?: MomentListResponse } = $props();
	const dateStr = $derived(formatDateDotted(moment.createdAt));
	const related = $derived(moment.relatedMoments ?? []);
	const previousMoment = $derived(related[0] ?? null);
	const nextMoment = $derived(related[1] ?? null);
	let contentRoot: HTMLElement | null = $state(null);
	let activeAnchor: string | null = $state(null);
	let sheetElement: HTMLElement | null = $state(null);
	let dialogElement: HTMLElement | null = $state(null);
	let enteringFromCard = $state(false);
	let isClosing = $state(false);
	let visit = $state<MomentBookVisit | null>(null);
	const fallbackSpread = $derived.by(() => {
		const spreads = buildMomentBookSpreads(underlayMoments.items);
		return (
			spreads.find((spread) =>
				[spread.left, spread.right].some((leaf) =>
					leaf?.items.some((item) => item.id === moment.id)
				)
			) ?? spreads[spreads.length - 1]
		);
	});
	const underlaySpread = $derived(visit?.spread ?? fallbackSpread);
	const returnPath = $derived(visit?.returnPath ?? resolvePath('/moments'));

	$effect(() => detailHeroBgSrc.set(moment.cover ?? ''));
	onDestroy(() => detailHeroBgSrc.set(''));

	function closeDetail() {
		if (!isClosing) void goto(returnPath);
	}

	function relatedHref(item: NonNullable<typeof previousMoment>) {
		return resolvePath(buildMomentPath(item.shortUrl, item.createdAt));
	}

	onNavigate(async (navigation) => {
		const destination = navigation.to?.url.pathname;
		if (!destination || /(?:^|\/)moments\/\d{4}\/\d{2}\/\d{2}\/[^/]+\/?$/.test(destination)) return;

		if (visit && destination === new URL(returnPath, window.location.origin).pathname) {
			try {
				sessionStorage.setItem(
					'moment:return-spread',
					JSON.stringify({
						returnPath: visit.returnPath,
						spreadIndex: visit.spreadIndex,
						at: Date.now()
					})
				);
			} catch {
				// Navigation remains available when browser storage is disabled.
			}
		}
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		isClosing = true;
		await new Promise<void>((resolve) => window.setTimeout(resolve, 430));
	});

	onMount(() => {
		dialogElement?.focus();
		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') closeDetail();
		};
		window.addEventListener('keydown', handleKeyDown);

		try {
			const raw = sessionStorage.getItem('moment:open-origin');
			sessionStorage.removeItem('moment:open-origin');
			if (raw && sheetElement) {
				const origin = JSON.parse(raw) as MomentBookVisit & {
					x: number;
					y: number;
					width: number;
					height: number;
					at: number;
				};
				if (
					Date.now() - origin.at < 5000 &&
					(!origin.momentId || origin.momentId === moment.id) &&
					origin.width > 0 &&
					origin.height > 0
				) {
					const sheetRect = sheetElement.getBoundingClientRect();
					const centerX = origin.x + origin.width / 2;
					const centerY = origin.y + origin.height / 2;
					sheetElement.style.setProperty(
						'--open-x',
						`${centerX - sheetRect.left - sheetRect.width / 2}px`
					);
					sheetElement.style.setProperty(
						'--open-y',
						`${centerY - sheetRect.top - sheetRect.height / 2}px`
					);
					sheetElement.style.setProperty('--open-scale-x', `${origin.width / sheetRect.width}`);
					sheetElement.style.setProperty('--open-scale-y', `${origin.height / sheetRect.height}`);
					enteringFromCard = true;
					if (
						origin.momentId === moment.id &&
						origin.returnPath?.startsWith(resolvePath('/moments')) &&
						origin.spread
					) {
						visit = origin;
					}
				}
			}
		} catch {
			sessionStorage.removeItem('moment:open-origin');
		}

		return () => window.removeEventListener('keydown', handleKeyDown);
	});
</script>

<MomentBookShell pageLabel={moment.title || '手记详情'}>
	{#snippet directory()}
		<div class="underlay-date-page" aria-hidden="true">
			<MomentDatePage
				leaf={underlaySpread.left}
				side="left"
				showBlankNote={!!visit || underlayMoments.items.length > 0}
			/>
		</div>
	{/snippet}

	<div class="underlay-date-page" aria-hidden="true">
		<MomentDatePage
			leaf={underlaySpread.right}
			side="right"
			showBlankNote={!!visit || underlayMoments.items.length > 0}
		/>
	</div>

	{#snippet overlay()}
		<button
			class="detail-backdrop"
			class:closing={isClosing}
			type="button"
			aria-label="关闭手记正文"
			onclick={closeDetail}
		></button>
		<div
			class="kraft-stage"
			bind:this={dialogElement}
			role="dialog"
			aria-modal="true"
			aria-labelledby="moment-detail-title"
			tabindex="-1"
		>
			<article
				bind:this={sheetElement}
				class:from-card={enteringFromCard}
				class:closing={isClosing}
				class="kraft-sheet"
			>
				<div class="paperclip" aria-hidden="true"><Paperclip size={44} strokeWidth={1.35} /></div>
				<a class="close-sheet" href={returnPath} aria-label="放回手记" title="放回手记"
					><X size={18} strokeWidth={1.5} /></a
				>
				<div class="kraft-scroll">
					<div id="moment-detail-title" class="sr-only">{moment.title || '无题手记'}</div>
					<MomentDetailPaper
						{moment}
						{dateStr}
						onContentRootChange={(node) => (contentRoot = node)}
						onActiveAnchorChange={(anchor) => (activeAnchor = anchor)}
					/>

					<nav class="sheet-navigation" aria-label="手记前后篇">
						<a class="return-index" href={returnPath}
							><ArrowLeft size={15} /><span>放回手记</span></a
						>
						<div>
							{#if previousMoment}<a
									href={relatedHref(previousMoment)}
									aria-label={`上一篇：${previousMoment.title}`}
									><ArrowLeft size={14} /><span>上一篇</span></a
								>{/if}
							{#if nextMoment}<a
									href={relatedHref(nextMoment)}
									aria-label={`下一篇：${nextMoment.title}`}
									><span>下一篇</span><ArrowRight size={14} /></a
								>{/if}
						</div>
					</nav>
				</div>
			</article>
		</div>
	{/snippet}
</MomentBookShell>

<style>
	.underlay-date-page {
		min-height: 100%;
	}
	.detail-backdrop {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		border: 0;
		background: rgba(24, 15, 10, 0.43);
		backdrop-filter: blur(2.2px);
		animation: backdrop-arrive 280ms ease-out both;
	}
	.detail-backdrop.closing {
		animation: backdrop-leave 430ms ease-in both;
	}
	.kraft-stage {
		position: absolute;
		inset: 0;
		display: grid;
		place-items: center;
		padding: 0.8rem 1.2rem;
		pointer-events: none;
	}
	.kraft-sheet {
		--open-x: 0px;
		--open-y: 0px;
		--open-scale-x: 0.24;
		--open-scale-y: 0.24;
		position: relative;
		width: min(53rem, calc(100vw - 2rem));
		height: min(64rem, calc(100dvh - 3rem));
		min-height: min(36rem, calc(100dvh - 3rem));
		max-height: calc(100dvh - 3rem);
		pointer-events: auto;
		color: #3e3025;
		filter: drop-shadow(0 1.8rem 2.2rem rgba(26, 15, 9, 0.52));
		animation: paper-unfold 620ms cubic-bezier(0.16, 1, 0.3, 1) both;
	}
	.kraft-sheet.from-card {
		animation-name: paper-from-card;
	}
	.kraft-sheet.closing {
		animation: paper-to-spine 430ms cubic-bezier(0.65, 0, 0.84, 0.2) both;
	}
	.kraft-sheet.closing.from-card {
		animation-name: paper-to-card;
	}
	.kraft-sheet::before,
	.kraft-sheet::after {
		content: '';
		position: absolute;
		right: -0.18rem;
		left: -0.18rem;
		z-index: 3;
		height: 1.9rem;
		pointer-events: none;
	}
	.kraft-sheet::before {
		top: -0.72rem;
		border-radius: 48% 44% 24% 28% / 72% 68% 30% 34%;
		background: linear-gradient(
			180deg,
			#956238 0%,
			#d4a776 18%,
			#efd3a5 42%,
			#d5a46e 70%,
			#a06c3e 100%
		);
		box-shadow:
			0 -0.12rem 0.3rem rgba(67, 40, 20, 0.17),
			0 0.35rem 0.48rem rgba(63, 38, 19, 0.23),
			inset 0 0.18rem 0.28rem rgba(255, 233, 192, 0.42);
		transform: rotate(-0.4deg) perspective(260px) rotateX(-14deg);
	}
	.kraft-sheet::after {
		bottom: -0.76rem;
		border-radius: 26% 30% 55% 52% / 28% 30% 76% 72%;
		background: linear-gradient(180deg, #ecd0a0 0%, #d5a46c 38%, #b17c48 74%, #936035 100%);
		box-shadow:
			0 0.55rem 0.6rem rgba(43, 25, 14, 0.34),
			inset 0 0.15rem 0.22rem rgba(255, 231, 183, 0.36);
		transform: rotate(0.25deg) perspective(260px) rotateX(14deg);
	}
	.kraft-scroll {
		position: relative;
		display: flex;
		flex-direction: column;
		height: 100%;
		min-height: 0;
		max-height: none;
		overflow-y: auto;
		overscroll-behavior: contain;
		padding: clamp(3.9rem, 6vh, 5.2rem) clamp(3rem, 6vw, 5.6rem) 2.6rem;
		border: 1px solid rgba(98, 62, 32, 0.36);
		background-color: #e3c79b;
		background-image:
			linear-gradient(
				180deg,
				rgba(88, 51, 24, 0.1),
				transparent 6%,
				transparent 92%,
				rgba(84, 49, 24, 0.13)
			),
			linear-gradient(
				90deg,
				rgba(85, 48, 21, 0.055),
				transparent 8%,
				transparent 92%,
				rgba(85, 48, 21, 0.055)
			),
			radial-gradient(circle at 24% 11%, rgba(255, 235, 196, 0.24), transparent 32%);
		box-shadow: inset 0 0 2.3rem rgba(91, 52, 25, 0.08);
		clip-path: polygon(
			0.3% 0.4%,
			25% 0.2%,
			60% 0.5%,
			99.6% 0.3%,
			99.7% 45%,
			99.5% 99.7%,
			50% 99.5%,
			0.3% 99.7%,
			0.4% 55%
		);
		scrollbar-width: thin;
		scrollbar-color: rgba(103, 64, 34, 0.35) transparent;
	}
	.paperclip {
		position: absolute;
		top: -0.4rem;
		left: 2.4rem;
		z-index: 6;
		color: #6e5945;
		filter: drop-shadow(0 0.18rem 0.12rem rgba(61, 40, 24, 0.3));
		transform: rotate(9deg);
	}
	.close-sheet {
		position: absolute;
		top: 1.35rem;
		right: 1.45rem;
		z-index: 7;
		display: grid;
		width: 2.35rem;
		height: 2.35rem;
		place-items: center;
		border: 1px solid rgba(74, 49, 30, 0.22);
		border-radius: 50%;
		color: rgba(62, 48, 37, 0.64);
		background: rgba(225, 187, 125, 0.62);
		backdrop-filter: blur(4px);
	}
	.close-sheet:hover,
	.close-sheet:focus-visible {
		color: #843c31;
		border-color: rgba(132, 60, 49, 0.4);
	}
	.close-sheet:focus-visible,
	.sheet-navigation a:focus-visible {
		outline: 2px solid rgba(132, 60, 49, 0.55);
		outline-offset: 3px;
	}
	.sheet-navigation {
		display: flex;
		width: 100%;
		max-width: 37rem;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin: auto auto 0;
		padding-top: 1rem;
		border-top: 1px solid rgba(75, 51, 32, 0.22);
		font-family: var(--font-serif);
		font-size: 0.74rem;
		letter-spacing: 0.08em;
		color: rgba(76, 52, 35, 0.74);
	}
	.sheet-navigation a,
	.sheet-navigation div {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}
	.sheet-navigation div {
		gap: 1.25rem;
	}
	.sheet-navigation a:hover {
		color: #843c31;
	}
	@keyframes backdrop-arrive {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
	@keyframes backdrop-leave {
		to {
			opacity: 0;
		}
	}
	@keyframes paper-unfold {
		from {
			opacity: 0;
			transform: translateY(1.2rem) scale(0.94);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
	@keyframes paper-from-card {
		from {
			opacity: 0.25;
			transform: translate(var(--open-x), var(--open-y))
				scale(var(--open-scale-x), var(--open-scale-y)) rotate(-1.5deg);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
	@keyframes paper-to-card {
		to {
			opacity: 0.12;
			transform: translate(var(--open-x), var(--open-y))
				scale(var(--open-scale-x), var(--open-scale-y)) rotate(-1.5deg);
		}
	}
	@keyframes paper-to-spine {
		to {
			opacity: 0;
			transform: translateY(0.8rem) scale(0.76);
		}
	}
	@media (max-width: 767px) {
		.kraft-stage {
			padding: 4.35rem 0.45rem 0.55rem;
		}
		.kraft-sheet {
			width: 100%;
			height: calc(100dvh - 5rem);
			min-height: 0;
			max-height: calc(100dvh - 5rem);
			filter: drop-shadow(0 0.8rem 1.2rem rgba(26, 15, 9, 0.3));
		}
		.kraft-scroll {
			height: 100%;
			min-height: 0;
			max-height: none;
			padding: 4.3rem 1.45rem 2.7rem;
		}
		.paperclip {
			left: 1.35rem;
		}
		.close-sheet {
			top: 1rem;
			right: 1rem;
		}
		.sheet-navigation {
			align-items: center;
			gap: 0.45rem;
		}
		.sheet-navigation div {
			gap: 0.55rem;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.detail-backdrop,
		.kraft-sheet,
		.kraft-sheet.from-card {
			animation: none;
		}
	}
</style>
