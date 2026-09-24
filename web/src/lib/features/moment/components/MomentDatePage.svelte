<script lang="ts">
	import type { MomentBookVisit, MomentDateLeaf } from '$lib/features/moment/book-pages';
	import { formatLeafPageLabel } from '$lib/features/moment/book-pages';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { ArrowLeft, ArrowRight, Search, X } from 'lucide-svelte';
	import MomentLeafEntry from './MomentLeafEntry.svelte';
	import SeasonalPageMark from './SeasonalPageMark.svelte';

	interface Props {
		leaf: MomentDateLeaf | null;
		side: 'left' | 'right';
		search?: string;
		basePath?: string;
		preview?: boolean;
		showBlankNote?: boolean;
		openContext?: Omit<MomentBookVisit, 'momentId' | 'at'>;
		kicker?: string;
		canTurn?: boolean;
		turnLabel?: string;
		turnPageLabel?: string;
		onTurn?: () => void;
	}

	let {
		leaf,
		side,
		search = '',
		basePath = '/moments',
		preview = false,
		showBlankNote = true,
		openContext,
		kicker,
		canTurn = false,
		turnLabel = '',
		turnPageLabel,
		onTurn = () => {}
	}: Props = $props();
	const pageLabel = $derived(leaf ? formatLeafPageLabel(leaf) : '留白');
</script>

<section
	class="date-page"
	class:left-page={side === 'left'}
	class:right-page={side === 'right'}
	aria-label={leaf ? `${pageLabel}的手记` : '留白书页'}
>
	{#if leaf}<SeasonalPageMark dateKey={leaf.dateKey} />{/if}
	<header class="page-head">
		<div>
			<p>{kicker ?? (side === 'left' ? 'YESTERDAY / NOTES' : 'TODAY / NOTES')}</p>
			<h2>
				{leaf
					? `${Number(leaf.month)}月${Number(leaf.day)}日`
					: showBlankNote
						? '留给明天'
						: '手记'}
			</h2>
			{#if leaf}<span
					>{leaf.year} · {leaf.totalItems} 条片刻{leaf.totalParts > 1
						? ` · 第 ${leaf.part} 页`
						: ''}</span
				>{/if}
		</div>
		{#if side === 'right'}
			<form action={resolvePath(`${basePath}/`)} method="GET" role="search">
				<label for="moment-query" class="sr-only">搜索手记</label>
				<input id="moment-query" name="q" value={search} type="search" placeholder="寻找一句话……" />
				<button type="submit" aria-label="搜索手记"><Search size={15} strokeWidth={1.5} /></button>
			</form>
		{/if}
	</header>

	{#if search && side === 'right'}
		<div class="search-note">
			<span>“{search}” 的结果</span><a href={resolvePath(`${basePath}/`)}><X size={11} />清除</a>
		</div>
	{/if}

	<div class="leaf-entries">
		{#if leaf}
			{#each leaf.items as moment (moment.id)}<MomentLeafEntry
					{moment}
					{preview}
					{openContext}
				/>{/each}
		{:else if showBlankNote}
			<div class="blank-note">
				<span aria-hidden="true">✦</span>
				<p>今天写到这里，<br />下一页留给明天。</p>
			</div>
		{/if}
	</div>

	<footer class="page-folio">
		{#if canTurn}
			<button type="button" onclick={onTurn} aria-label={turnLabel} title={turnLabel}>
				{#if side === 'left'}<ArrowLeft size={14} />{/if}
				<span>{turnPageLabel ?? pageLabel}</span>
				{#if side === 'right'}<ArrowRight size={14} />{/if}
			</button>
		{:else}<span>{pageLabel}</span>{/if}
	</footer>
</section>

<style>
	.date-page {
		position: relative;
		display: flex;
		min-height: 100%;
		flex-direction: column;
		font-family: var(--font-serif);
	}
	.page-head {
		position: relative;
		z-index: 1;
		display: flex;
		min-height: 6.25rem;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--book-rule);
	}
	.left-page .page-head {
		padding-left: 3.5rem;
	}
	.page-head p {
		font-family: var(--font-mono);
		font-size: 0.52rem;
		letter-spacing: 0.24em;
		color: var(--book-faint);
	}
	.page-head h2 {
		margin-top: 0.52rem;
		font-size: clamp(1.55rem, 2.35vw, 2.25rem);
		font-weight: 500;
		letter-spacing: 0.08em;
	}
	.page-head span {
		display: block;
		margin-top: 0.34rem;
		font-size: 0.62rem;
		letter-spacing: 0.08em;
		color: var(--book-muted);
	}
	form {
		display: flex;
		width: min(46%, 12rem);
		align-items: center;
		margin-top: 0.25rem;
		border-bottom: 1px solid rgba(56, 47, 40, 0.3);
	}
	input {
		min-width: 0;
		flex: 1;
		padding: 0.42rem 0;
		font-family: var(--font-serif);
		font-size: 0.68rem;
		color: var(--book-ink);
		background: transparent;
	}
	input::placeholder {
		color: var(--book-faint);
	}
	form button {
		display: grid;
		width: 2rem;
		height: 2rem;
		place-items: center;
		color: var(--book-muted);
	}
	form :is(input, button):focus-visible,
	.page-folio button:focus-visible {
		outline: 2px solid rgba(141, 56, 45, 0.5);
		outline-offset: 2px;
	}
	.search-note {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.48rem 0;
		font-size: 0.58rem;
		color: var(--book-faint);
		border-bottom: 1px dashed var(--book-rule);
	}
	.search-note a {
		display: inline-flex;
		align-items: center;
		gap: 0.2rem;
		color: var(--book-accent);
	}
	.leaf-entries {
		display: grid;
		min-height: 0;
		flex: 1;
		grid-template-rows: repeat(4, minmax(0, 1fr));
		padding-top: 0.35rem;
	}
	.blank-note {
		grid-row: 2 / span 2;
		align-self: center;
		justify-self: center;
		text-align: center;
		color: var(--book-faint);
	}
	.blank-note span {
		display: block;
		margin-bottom: 0.75rem;
		font-size: 1rem;
		color: var(--book-accent);
		opacity: 0.45;
	}
	.blank-note p {
		font-size: 0.72rem;
		line-height: 2;
		letter-spacing: 0.09em;
	}
	.page-folio {
		display: flex;
		min-height: 2.6rem;
		align-items: flex-end;
		padding-top: 0.65rem;
		font-family: var(--font-mono);
		font-size: 0.58rem;
		letter-spacing: 0.12em;
		color: var(--book-faint);
	}
	.left-page .page-folio {
		justify-content: flex-start;
	}
	.right-page .page-folio {
		justify-content: flex-end;
	}
	.page-folio button {
		position: relative;
		display: inline-flex;
		min-height: 2.5rem;
		align-items: center;
		gap: 0.45rem;
		padding-inline: 0.3rem;
		color: var(--book-muted);
		transition:
			color 180ms ease,
			transform 180ms ease;
	}
	.page-folio button::after {
		content: '';
		position: absolute;
		bottom: -0.25rem;
		width: 0.75rem;
		height: 0.75rem;
		pointer-events: none;
		background: linear-gradient(135deg, #f8ebcd, #c7ac80);
		clip-path: polygon(0 100%, 100% 0, 100% 100%);
		opacity: 0;
		transform: translateY(0.2rem);
		transition:
			opacity 180ms ease,
			transform 180ms ease;
	}
	.left-page .page-folio button::after {
		left: -0.7rem;
		transform: translateY(0.2rem) scaleX(-1);
	}
	.right-page .page-folio button::after {
		right: -0.7rem;
	}
	.page-folio button:hover::after,
	.page-folio button:focus-visible::after {
		opacity: 0.85;
		transform: translateY(-0.12rem);
	}
	.left-page .page-folio button:hover::after,
	.left-page .page-folio button:focus-visible::after {
		transform: translateY(-0.12rem) scaleX(-1);
	}
	.page-folio button:hover {
		color: var(--book-accent);
		transform: translateY(-1px);
	}
	@media (max-width: 1100px) {
		.page-head {
			min-height: 5.4rem;
		}
		.left-page .page-head {
			padding-left: 2.6rem;
		}
		form {
			width: 8.5rem;
		}
	}
	@media (max-width: 767px) {
		.date-page {
			min-height: calc(100dvh - 6rem);
		}
		.page-head {
			min-height: 6.2rem;
		}
		.left-page .page-head {
			padding-left: 3rem;
		}
		form {
			width: 44%;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.page-folio button {
			transition: none;
		}
	}
</style>
