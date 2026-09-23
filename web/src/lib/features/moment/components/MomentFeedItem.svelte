<script lang="ts">
	import type { MomentSummary } from '$lib/features/moment/types';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { ArrowUpRight, Pin } from 'lucide-svelte';

	let { moment, preview = false }: { moment: MomentSummary; preview?: boolean } = $props();
	const href = $derived(
		preview
			? `#preview-entry-${Math.abs(moment.id)}`
			: resolvePath(buildMomentPath(moment.shortUrl, moment.createdAt))
	);
	const dateParts = $derived(moment.createdAt.slice(0, 10).split('-'));
	const day = $derived(dateParts[2] ?? '');
	const month = $derived(
		(
			['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC'][
				Math.max(0, Math.min(11, Number(dateParts[1] ?? '1') - 1))
			] ?? ''
		).toUpperCase()
	);
</script>

<article class="book-entry" id={preview ? `preview-entry-${Math.abs(moment.id)}` : undefined}>
	<div class="entry-date">
		<time datetime={moment.createdAt}>
			<strong>{day}</strong>
			<span>{month}</span>
		</time>
		{#if moment.isTop}
			<span class="pinned"><Pin size={11} />置顶</span>
		{/if}
	</div>

	<div class="entry-body">
		<div class:entry-copy-with-cover={!!moment.cover} class="entry-copy-grid">
			<div class="entry-copy">
				{#if moment.title}
					<h2>
						<a {href}>{moment.title}</a>
					</h2>
				{/if}
				{#if moment.summary}
					<p class="entry-summary">{moment.summary}</p>
				{/if}
			</div>

			{#if moment.cover}
				<a class="entry-cover" {href} aria-label={moment.title || '查看图片手记'}>
					<img src={moment.cover} alt={moment.title || '手记图片'} loading="lazy" />
				</a>
			{/if}
		</div>

		<footer class="entry-footer">
			<div class="entry-topics">
				{#each moment.topics ?? [] as topic}<span>#{topic}</span>{/each}
			</div>
			<a class="read-entry" {href}>
				{preview ? '版式预览' : '翻开此页'}
				<ArrowUpRight size={14} />
			</a>
		</footer>
	</div>
</article>

<style>
	.book-entry {
		display: grid;
		grid-template-columns: clamp(4.2rem, 6.5vw, 5.6rem) minmax(0, 1fr);
		gap: clamp(1rem, 2.2vw, 2rem);
		padding: clamp(1.15rem, 2.2vh, 1.8rem) 0;
		border-bottom: 1px solid var(--book-rule);
	}

	.entry-date {
		padding-top: 0.35rem;
	}

	.entry-date time {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		font-family: var(--font-mono);
	}

	.entry-date strong {
		font-family: var(--font-serif);
		font-size: clamp(1.55rem, 2.4vw, 2.05rem);
		font-weight: 500;
		line-height: 1;
		letter-spacing: 0.04em;
		color: var(--book-ink);
	}

	.entry-date time span {
		font-size: 0.54rem;
		letter-spacing: 0.2em;
		color: var(--book-faint);
	}

	.pinned {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		margin-top: 0.8rem;
		font-family: var(--font-serif);
		font-size: 0.62rem;
		letter-spacing: 0.08em;
		color: var(--book-accent);
	}

	.entry-copy-grid.entry-copy-with-cover {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(5.5rem, 8vw, 7.2rem);
		gap: clamp(1.1rem, 2vw, 2rem);
		align-items: start;
	}

	.entry-copy {
		min-width: 0;
	}

	h2 {
		font-family: var(--font-serif);
		font-size: clamp(1.08rem, 1.55vw, 1.42rem);
		font-weight: 500;
		line-height: 1.45;
		letter-spacing: 0.035em;
		text-wrap: balance;
	}

	h2 a {
		transition: color 180ms ease;
	}

	h2 a:hover,
	h2 a:focus-visible {
		color: var(--book-accent);
	}

	h2 a:focus-visible,
	.entry-cover:focus-visible,
	.read-entry:focus-visible {
		outline: 2px solid rgba(141, 56, 45, 0.52);
		outline-offset: 4px;
	}

	.entry-summary {
		max-width: 48rem;
		margin-top: 0.58rem;
		font-family: var(--font-serif);
		font-size: clamp(0.78rem, 0.92vw, 0.88rem);
		line-height: 1.72;
		letter-spacing: 0.025em;
		white-space: pre-line;
		color: var(--book-muted);
	}

	.entry-cover {
		display: block;
		overflow: hidden;
		aspect-ratio: 4 / 3;
		border: 0.28rem solid rgba(255, 250, 235, 0.52);
		box-shadow: 0 0.7rem 1.6rem rgba(65, 43, 28, 0.16);
		transform: rotate(0.7deg);
	}

	.entry-cover img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		filter: saturate(0.84) sepia(0.08);
		transition: transform 420ms cubic-bezier(0.16, 1, 0.3, 1);
	}

	.entry-cover:hover img {
		transform: scale(1.035);
	}

	.entry-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-top: 0.7rem;
	}

	.entry-topics {
		display: flex;
		min-width: 0;
		flex-wrap: wrap;
		gap: 0.35rem 0.75rem;
		font-family: var(--font-mono);
		font-size: 0.57rem;
		letter-spacing: 0.06em;
		color: var(--book-faint);
	}

	.read-entry {
		display: inline-flex;
		flex: none;
		align-items: center;
		gap: 0.35rem;
		font-family: var(--font-serif);
		font-size: 0.7rem;
		letter-spacing: 0.12em;
		color: var(--book-accent);
	}

	@media (max-width: 640px) {
		.book-entry {
			grid-template-columns: 3.6rem minmax(0, 1fr);
			gap: 0.8rem;
			padding: 2.2rem 0;
		}

		.entry-copy-grid.entry-copy-with-cover {
			display: flex;
			flex-direction: column;
		}

		.entry-cover {
			width: min(100%, 18rem);
		}

		.entry-footer {
			align-items: flex-start;
			flex-direction: column;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		h2 a,
		.entry-cover img {
			transition: none;
		}
	}
</style>
