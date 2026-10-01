<script lang="ts">
	import type { MomentSummary } from '$lib/features/moment/types';
	import type { MomentBookVisit } from '$lib/features/moment/book-pages';
	import { base } from '$app/paths';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { ArrowUpRight } from 'lucide-svelte';

	let {
		moment,
		preview = false,
		openContext,
		onOpen
	}: {
		moment: MomentSummary;
		preview?: boolean;
		openContext?: Omit<MomentBookVisit, 'momentId' | 'at'>;
		onOpen?: (moment: MomentSummary, href: string, link: HTMLAnchorElement) => void;
	} = $props();
	const href = $derived(
		preview
			? `${base}/moments/preview/${encodeURIComponent(moment.shortUrl)}/`
			: `${base}${buildMomentPath(moment.shortUrl, moment.createdAt)}`
	);
	const time = $derived(moment.createdAt.slice(11, 16));

	function handleOpen(event: MouseEvent) {
		if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey)
			return;
		const link = event.currentTarget as HTMLAnchorElement;
		const card = link.closest<HTMLElement>('.leaf-entry');
		if (!card) return;
		const rect = card.getBoundingClientRect();
		try {
			sessionStorage.setItem(
				'moment:open-origin',
				JSON.stringify({
					x: rect.left,
					y: rect.top,
					width: rect.width,
					height: rect.height,
					at: Date.now(),
					...(openContext ? { ...openContext, momentId: moment.id } : {})
				})
			);
		} catch {
			// Continue opening the note when browser storage is unavailable.
		}
		if (onOpen) {
			event.preventDefault();
			onOpen(moment, href, link);
		}
	}
</script>

<article class="leaf-entry">
	<span class="margin-dot" aria-hidden="true"></span>
	<a class="entry-link" {href} onclick={handleOpen}>
		<div class="entry-copy">
			<div class="entry-heading">
				<time datetime={moment.createdAt}>{time || '片刻'}</time>
				<h3>{moment.title || '无题手记'}</h3>
			</div>
			{#if moment.summary}<p>{moment.summary}</p>{/if}
			<div class="entry-meta">
				<span
					>{(moment.topics ?? [])
						.slice(0, 2)
						.map((topic) => `#${topic}`)
						.join(' · ') ||
						moment.columnName ||
						'日常'}</span
				>
				<em>翻开此页 <ArrowUpRight size={12} strokeWidth={1.6} /></em>
			</div>
		</div>
		{#if moment.cover}
			<figure><img src={moment.cover} alt="" loading="lazy" /></figure>
		{/if}
	</a>
</article>

<style>
	.leaf-entry {
		position: relative;
		min-height: 0;
		border-bottom: 1px solid var(--book-rule);
		transition:
			transform 190ms ease,
			background-color 190ms ease,
			box-shadow 190ms ease;
	}
	.margin-dot {
		position: absolute;
		top: 50%;
		left: -0.8rem;
		width: 0.3rem;
		height: 0.3rem;
		border-radius: 50%;
		background: rgba(113, 73, 54, 0.2);
		transform: translateY(-50%);
		transition:
			background-color 180ms ease,
			box-shadow 180ms ease;
	}
	.entry-link {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		gap: 0.8rem;
		min-height: 100%;
		padding: 0.9rem 0.35rem 0.8rem 0.25rem;
	}
	.entry-copy {
		min-width: 0;
	}
	.entry-heading {
		display: grid;
		grid-template-columns: 2.55rem minmax(0, 1fr);
		gap: 0.7rem;
		align-items: baseline;
	}
	time {
		font-family: var(--font-mono);
		font-size: 0.53rem;
		letter-spacing: 0.1em;
		color: var(--book-faint);
	}
	h3 {
		overflow: hidden;
		width: fit-content;
		max-width: 100%;
		padding-bottom: 0.12rem;
		font-family: var(--font-serif);
		font-size: clamp(0.92rem, 1.25vw, 1.12rem);
		font-weight: 600;
		line-height: 1.45;
		letter-spacing: 0.035em;
		text-overflow: ellipsis;
		white-space: nowrap;
		background: linear-gradient(rgba(104, 78, 58, 0.45), rgba(104, 78, 58, 0.45)) left bottom / 0
			1px no-repeat;
		transition: background-size 240ms cubic-bezier(0.16, 1, 0.3, 1);
	}
	p {
		display: -webkit-box;
		overflow: hidden;
		margin: 0.38rem 0 0 3.25rem;
		font-family: var(--font-serif);
		font-size: clamp(0.68rem, 0.82vw, 0.78rem);
		line-height: 1.65;
		color: var(--book-muted);
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 2;
		line-clamp: 2;
	}
	.entry-meta {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.8rem;
		margin: 0.45rem 0 0 3.25rem;
		font-family: var(--font-serif);
		font-size: 0.58rem;
		letter-spacing: 0.05em;
		color: var(--book-faint);
	}
	.entry-meta em {
		display: inline-flex;
		align-items: center;
		gap: 0.2rem;
		flex: none;
		font-style: normal;
		color: var(--book-accent);
		opacity: 0;
		transform: translateX(-0.3rem);
		transition:
			opacity 180ms ease,
			transform 180ms ease;
	}
	figure {
		width: clamp(3.6rem, 5.2vw, 4.8rem);
		align-self: center;
		aspect-ratio: 4 / 3;
		padding: 0.18rem;
		background: rgba(255, 250, 235, 0.55);
		box-shadow: 0 0.35rem 0.75rem rgba(62, 40, 25, 0.12);
		transform: rotate(1.5deg);
	}
	figure img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		filter: saturate(0.78) sepia(0.08);
	}
	.leaf-entry:hover,
	.leaf-entry:focus-within {
		z-index: 2;
		background: rgba(255, 250, 232, 0.2);
		box-shadow: 0 0.45rem 1rem rgba(69, 45, 27, 0.08);
		transform: translateY(-2px);
	}
	.leaf-entry:hover h3,
	.leaf-entry:focus-within h3 {
		background-size: 100% 1px;
	}
	.leaf-entry:hover .margin-dot,
	.leaf-entry:focus-within .margin-dot {
		background: var(--book-accent);
		box-shadow: 0 0 0 3px rgba(141, 56, 45, 0.1);
	}
	.leaf-entry:hover .entry-meta em,
	.leaf-entry:focus-within .entry-meta em {
		opacity: 0.72;
		transform: none;
	}
	.entry-link:focus-visible {
		outline: 2px solid rgba(141, 56, 45, 0.5);
		outline-offset: -2px;
	}
	@media (max-width: 1100px) {
		.entry-meta em {
			display: none;
		}
		p {
			-webkit-line-clamp: 1;
			line-clamp: 1;
		}
	}
	@media (max-width: 767px) {
		.entry-link {
			padding-block: 1.1rem;
		}
		h3 {
			font-size: 1rem;
		}
		p {
			-webkit-line-clamp: 2;
			line-clamp: 2;
		}
		.entry-meta em {
			display: inline-flex;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.leaf-entry,
		h3,
		.margin-dot,
		.entry-meta em {
			transition: none;
		}
	}
</style>
