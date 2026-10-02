<script lang="ts">
	import { onNavigate } from '$app/navigation';
	import { base } from '$app/paths';
	import type { MomentBookVisit } from '$lib/features/moment/book-pages';
	import type { MomentDetail, MomentListResponse } from '$lib/features/moment/types';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { formatDateDotted } from '$lib/shared/utils/date';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { ArrowLeft, ArrowRight, BookOpenText, Paperclip } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import MomentDetailPaper from './moment-detail/MomentDetailPaper.svelte';

	let {
		moment,
		underlayMoments = { items: [], total: 0, page: 1, size: 20 },
		preview = false
	}: {
		moment: MomentDetail;
		underlayMoments?: MomentListResponse;
		preview?: boolean;
	} = $props();
	const dateStr = $derived(formatDateDotted(moment.createdAt));
	const related = $derived(moment.relatedMoments ?? []);
	const previewIndex = $derived(underlayMoments.items.findIndex((item) => item.id === moment.id));
	const previousMoment = $derived(
		preview ? (underlayMoments.items[previewIndex - 1] ?? null) : (related[0] ?? null)
	);
	const nextMoment = $derived(
		preview ? (underlayMoments.items[previewIndex + 1] ?? null) : (related[1] ?? null)
	);
	let visit = $state<MomentBookVisit | null>(null);
	const returnPath = $derived(visit?.returnPath ?? resolvePath('/moments/'));

	function relatedHref(item: NonNullable<typeof previousMoment>) {
		return preview
			? `${base}/moments/preview/${encodeURIComponent(item.shortUrl)}/`
			: `${base}${buildMomentPath(item.shortUrl, item.createdAt)}`;
	}

	onMount(() => {
		try {
			const raw = sessionStorage.getItem('moment:open-origin');
			sessionStorage.removeItem('moment:open-origin');
			if (!raw) return;
			const origin = JSON.parse(raw) as MomentBookVisit;
			if (
				origin.momentId === moment.id &&
				Date.now() - origin.at < 60_000 &&
				origin.returnPath?.startsWith(resolvePath('/moments')) &&
				Number.isInteger(origin.spreadIndex) &&
				origin.spreadIndex >= 0
			) {
				visit = origin;
			}
		} catch {
			// The standard directory link also works without browser storage.
		}
	});

	onNavigate((navigation) => {
		if (!visit || !navigation.to) return;
		const destination = navigation.to.url.pathname + navigation.to.url.search;
		if (destination !== visit.returnPath) return;
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
			// Navigation does not depend on saving the current book spread.
		}
	});
</script>

<section class="moment-reading-page" aria-label="手记阅读">
	<nav class="reading-header" aria-label="手记导航">
		<a href={returnPath}><ArrowLeft size={17} strokeWidth={1.5} /><span>返回手记</span></a>
		<span class="reading-identity"><BookOpenText size={17} strokeWidth={1.4} />手记</span>
	</nav>
	<article class="reading-sheet">
		<div class="paperclip" aria-hidden="true"><Paperclip size={30} strokeWidth={1.3} /></div>
		<div class="reading-content">
			<MomentDetailPaper {moment} {preview} {dateStr} />
			<nav class="sheet-navigation" aria-label="手记前后篇">
				<a href={returnPath}><ArrowLeft size={15} /><span>返回手记</span></a>
				<div>
					{#if previousMoment}<a
							href={relatedHref(previousMoment)}
							aria-label={`上一篇：${previousMoment.title}`}
						>
							<ArrowLeft size={14} /><span>上一篇</span>
						</a>{/if}
					{#if nextMoment}<a
							href={relatedHref(nextMoment)}
							aria-label={`下一篇：${nextMoment.title}`}
						>
							<span>下一篇</span><ArrowRight size={14} />
						</a>{/if}
				</div>
			</nav>
		</div>
	</article>
</section>

<style>
	.moment-reading-page {
		--book-ink: #37342f;
		--book-muted: #70695f;
		--book-faint: #8b8378;
		--book-rule: rgba(91, 80, 65, 0.16);
		min-height: 100dvh;
		padding: 1.25rem 1.5rem 3rem;
		color: var(--book-ink);
		color-scheme: light;
		background: #f0ede6;
	}
	.reading-header {
		display: flex;
		max-width: 54rem;
		align-items: center;
		justify-content: space-between;
		margin: 0 auto 1.5rem;
		font-family: var(--font-serif);
		font-size: 0.85rem;
		color: var(--book-muted);
	}
	.reading-header a,
	.reading-identity {
		display: inline-flex;
		min-height: 2.75rem;
		align-items: center;
		gap: 0.6rem;
	}
	.reading-sheet {
		position: relative;
		max-width: 54rem;
		margin-inline: auto;
		border: 1px solid rgba(98, 82, 60, 0.16);
		background: #fbf8f0;
		box-shadow: 0 0.7rem 2rem rgba(63, 48, 30, 0.09);
	}
	.reading-sheet::before,
	.reading-sheet::after {
		content: '';
		position: absolute;
		right: -0.15rem;
		left: -0.15rem;
		z-index: 1;
		height: 0.6rem;
		pointer-events: none;
	}
	.reading-sheet::before {
		top: -0.25rem;
		border-radius: 48% 44% 24% 28% / 72% 68% 30% 34%;
		background: linear-gradient(180deg, #e3dacb, #fffdf7 48%, #e6ddcd);
		box-shadow: 0 0.12rem 0.2rem rgba(63, 38, 19, 0.08);
	}
	.reading-sheet::after {
		bottom: -0.25rem;
		border-radius: 26% 30% 55% 52% / 28% 30% 76% 72%;
		background: linear-gradient(180deg, #fffdf7, #e3d9c8);
		box-shadow: 0 0.2rem 0.25rem rgba(43, 25, 14, 0.08);
	}
	.reading-content {
		padding: clamp(2.5rem, 5vw, 4rem) clamp(1.5rem, 6vw, 5.2rem) 1.8rem;
	}
	.paperclip {
		position: absolute;
		top: -0.45rem;
		left: 2.4rem;
		z-index: 2;
		color: #837563;
		transform: rotate(9deg);
	}
	.sheet-navigation {
		display: flex;
		max-width: 42rem;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin: 1.5rem auto 0;
		padding-top: 0.75rem;
		border-top: 1px solid var(--book-rule);
		font-family: var(--font-serif);
		font-size: 0.8rem;
		color: var(--book-muted);
	}
	.sheet-navigation a,
	.sheet-navigation div {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}
	.sheet-navigation a {
		min-height: 2.75rem;
	}
	.sheet-navigation div {
		gap: 1.25rem;
	}
	:is(.reading-header, .sheet-navigation) a:hover {
		color: #843c31;
	}
	:is(.reading-header, .sheet-navigation) a:focus-visible {
		outline: 2px solid #9b6552;
		outline-offset: 3px;
	}
	@media (max-width: 767px) {
		.moment-reading-page {
			padding: 0.5rem 0.6rem 2rem;
		}
		.reading-header {
			margin-bottom: 1rem;
			padding-inline: 0.65rem;
		}
		.reading-content {
			padding: 2.4rem 1.2rem 1.25rem;
		}
		.paperclip {
			left: 1.4rem;
		}
		.sheet-navigation {
			gap: 0.5rem;
			font-size: 0.74rem;
		}
		.sheet-navigation div {
			gap: 0.75rem;
		}
	}
</style>
