<script lang="ts">
	import { goto } from '$app/navigation';
	import { navigating } from '$app/state';
	import { dev } from '$app/environment';
	import { momentListCtx } from '$lib/features/moment/context';
	import type { MomentListResponse, MomentSummary } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import { ArrowRight, NotebookPen, Search, X } from 'lucide-svelte';
	import MomentBookShell from './MomentBookShell.svelte';
	import MomentFeedItem from './MomentFeedItem.svelte';

	interface Props {
		moments: MomentListResponse;
		search?: string;
		basePath?: string;
		staggerKey?: string;
	}

	const layoutPreviewMoments: MomentSummary[] = [
		{
			id: -1,
			title: '九点半，窗边的光刚刚好',
			shortUrl: 'layout-preview-morning',
			summary: '把咖啡放到手边，读完昨晚折角的那一页。窗外很亮，房间里安静得只剩翻书声。',
			cover: '/home-scenes/11.webp',
			views: 128,
			topics: ['日常', '阅读'],
			likes: 18,
			comments: 3,
			isTop: true,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-21T09:30:00+08:00',
			createdAt: '2026-09-21T09:30:00+08:00',
			updatedAt: '2026-09-21T09:30:00+08:00'
		},
		{
			id: -2,
			title: '把读到一半的书留在桌上',
			shortUrl: 'layout-preview-afternoon',
			summary: '有些内容不必急着读完。停在恰好的地方，等下一次回到桌前，故事仍会从书签旁边继续。',
			views: 96,
			topics: ['片刻', '书房'],
			likes: 12,
			comments: 1,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-18T15:20:00+08:00',
			createdAt: '2026-09-18T15:20:00+08:00',
			updatedAt: '2026-09-18T15:20:00+08:00'
		},
		{
			id: -3,
			title: '夜里十一点，猫先睡着了',
			shortUrl: 'layout-preview-night',
			summary:
				'台灯把桌面照成一小块温暖的岛。计划还剩两行，猫已经在椅子旁蜷成一团，提醒今天可以到这里。',
			views: 73,
			topics: ['夜晚', '随想'],
			likes: 9,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-14T23:10:00+08:00',
			createdAt: '2026-09-14T23:10:00+08:00',
			updatedAt: '2026-09-14T23:10:00+08:00'
		}
	];
	const monthNames = [
		'JANUARY',
		'FEBRUARY',
		'MARCH',
		'APRIL',
		'MAY',
		'JUNE',
		'JULY',
		'AUGUST',
		'SEPTEMBER',
		'OCTOBER',
		'NOVEMBER',
		'DECEMBER'
	] as const;

	let { moments, search = '', basePath = '/moments', staggerKey = 'moments' }: Props = $props();
	momentListCtx.mountModelData(() => moments);
	const isLayoutPreview = $derived(dev && !search && moments.items.length === 0);
	const visibleMoments = $derived(isLayoutPreview ? layoutPreviewMoments : moments.items);
	const visibleTotal = $derived(isLayoutPreview ? visibleMoments.length : moments.total);
	const monthGroups = $derived.by(() => {
		const groups = new Map<
			string,
			{ key: string; year: string; month: string; shortMonth: string; items: MomentSummary[] }
		>();

		for (const moment of visibleMoments) {
			const [year = '', month = '01'] = moment.createdAt.slice(0, 10).split('-');
			const key = `${year}-${month}`;
			const monthIndex = Math.max(0, Math.min(11, Number(month) - 1));
			const current = groups.get(key) ?? {
				key,
				year,
				month: monthNames[monthIndex],
				shortMonth: monthNames[monthIndex].slice(0, 3),
				items: []
			};
			current.items.push(moment);
			groups.set(key, current);
		}

		return [...groups.values()];
	});
	const topicIndex = $derived.by(() => {
		const counts = new Map<string, number>();
		for (const moment of visibleMoments) {
			for (const topic of moment.topics ?? []) counts.set(topic, (counts.get(topic) ?? 0) + 1);
		}
		return [...counts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 5);
	});
	const archiveIndex = $derived.by(() => {
		const counts = new Map<string, number>();
		for (const moment of visibleMoments) {
			const year = moment.createdAt.slice(0, 4);
			counts.set(year, (counts.get(year) ?? 0) + 1);
		}
		return [...counts.entries()].sort((a, b) => b[0].localeCompare(a[0]));
	});
	const totalPages = $derived(
		moments.size > 0 ? Math.max(1, Math.ceil(moments.total / moments.size)) : 1
	);

	function onPageChange(nextPage: number) {
		const safePage = Number.isFinite(nextPage) && nextPage > 1 ? nextPage : 1;
		const path = resolvePath(safePage === 1 ? `${basePath}/` : `${basePath}/page/${safePage}/`);
		goto(`${path}${search ? `?${new URLSearchParams({ q: search })}` : ''}`);
	}
</script>

<MomentBookShell pageLabel="手记列表">
	{#snippet directory()}
		<nav class="book-directory" aria-label="手记目录">
			<p class="directory-kicker">CONTENTS</p>
			<h2>手记</h2>
			<p class="directory-intro">日常、片刻与随想</p>

			<a class="directory-overview active" href={resolvePath(`${basePath}/`)}>
				<span>全部手记</span>
				<strong>{visibleTotal}</strong>
			</a>

			{#if visibleMoments.length > 0}
				<div class="directory-section">
					<div class="directory-section-head">
						<span>分类索引</span>
						<small>本页</small>
					</div>
					<ul class="directory-count-list">
						{#each topicIndex as [topic, count] (topic)}
							<li><span>{topic}</span><strong>{count}</strong></li>
						{/each}
					</ul>
				</div>
				<div class="directory-section archive-section">
					<div class="directory-section-head">
						<span>ARCHIVE</span>
					</div>
					<ul class="directory-count-list">
						{#each archiveIndex as [year, count] (year)}
							<li><span>{year}</span><strong>{count}</strong></li>
						{/each}
					</ul>
				</div>
			{:else}
				<p class="directory-empty">这一页暂时留白，等待下一篇手记。</p>
			{/if}
		</nav>
	{/snippet}

	<section class="moment-index" aria-labelledby="moments-heading">
		<header class="index-header">
			<div>
				<p class="index-kicker">
					{isLayoutPreview
						? 'LAYOUT PREVIEW'
						: `NOTEBOOK / ${String(moments.page).padStart(2, '0')}`}
				</p>
				<h1 id="moments-heading">手记</h1>
				<p class="index-subtitle">把普通日子里值得记住的片刻，夹进这一册书页。</p>
			</div>
			<div class="entry-count" aria-label={`共 ${visibleTotal} 篇手记`}>
				<strong>{String(visibleTotal).padStart(2, '0')}</strong>
				<span>ENTRIES</span>
			</div>
		</header>

		<div class="index-tools">
			<p>
				{isLayoutPreview ? '版式预览' : search ? '搜索结果' : '全部手记'}
				<span>{visibleTotal}</span>
			</p>
			<form action={resolvePath(`${basePath}/`)} method="GET" role="search">
				<label for="moment-query" class="sr-only">搜索手记</label>
				<input
					id="moment-query"
					name="q"
					value={search}
					type="search"
					placeholder="在手记中寻找……"
				/>
				<button type="submit" aria-label="搜索手记" title="搜索手记">
					<Search size={17} strokeWidth={1.6} aria-hidden="true" />
				</button>
			</form>
		</div>

		{#if search}
			<div class="search-state">
				<span>“{search}”</span>
				<a href={resolvePath(`${basePath}/`)}><X size={13} />清除搜索</a>
			</div>
		{/if}

		<div aria-busy={!!navigating.to}>
			{#if visibleMoments.length > 0}
				<div class="month-index">
					{#each monthGroups as group (group.key)}
						<section class="month-group" aria-labelledby={`month-${group.key}`}>
							<header class="month-heading">
								<h2 id={`month-${group.key}`}>{group.month}</h2>
								<span>{group.shortMonth}. {group.year}</span>
							</header>
							<StaggerList
								class="moment-entry-list"
								staggerDelay={40}
								duration={250}
								y={8}
								key={`${staggerKey}-${search}-${group.key}`}
							>
								{#each group.items as moment (moment.id)}
									<MomentFeedItem {moment} preview={isLayoutPreview} />
								{/each}
							</StaggerList>
						</section>
					{/each}
				</div>

				{#if totalPages > 1}
					<div class="index-pagination">
						<Pagination current={moments.page} total={totalPages} {onPageChange} />
					</div>
				{/if}
			{:else}
				<div class="empty-page" role="status">
					<div class="empty-mark"><NotebookPen size={30} strokeWidth={1.15} /></div>
					<p>{search ? '没有找到相关手记' : '还没有公开的手记'}</p>
					<span>{search ? '换一个词，也许会翻到另一页。' : '第一行文字，会从这里开始。'}</span>
					{#if search}
						<a href={resolvePath(`${basePath}/`)}>
							查看全部手记 <ArrowRight size={14} aria-hidden="true" />
						</a>
					{/if}
				</div>
			{/if}
		</div>
	</section>
</MomentBookShell>

<style>
	.book-directory {
		font-family: var(--font-serif);
	}

	.directory-kicker,
	.index-kicker {
		font-family: var(--font-mono);
		font-size: 0.58rem;
		letter-spacing: 0.28em;
		color: var(--book-faint);
	}

	.book-directory h2 {
		margin-top: 0.8rem;
		font-size: clamp(1.8rem, 3vw, 2.6rem);
		font-weight: 500;
		letter-spacing: 0.08em;
	}

	.directory-intro {
		margin-top: 0.7rem;
		font-size: 0.75rem;
		letter-spacing: 0.08em;
		color: var(--book-muted);
	}

	.directory-overview {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-top: 2.5rem;
		padding: 0.8rem 0;
		font-size: 0.82rem;
		letter-spacing: 0.06em;
		border-top: 1px solid var(--book-rule);
		border-bottom: 1px solid var(--book-rule);
	}

	.directory-overview.active {
		color: var(--book-accent);
	}

	.directory-overview strong {
		font-family: var(--font-mono);
		font-size: 0.62rem;
		font-weight: 500;
	}

	.directory-section {
		margin-top: 2.2rem;
	}

	.directory-section-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		margin-bottom: 0.7rem;
		color: var(--book-faint);
	}

	.directory-section-head span {
		font-size: 0.65rem;
		letter-spacing: 0.18em;
	}

	.directory-section-head small {
		font-family: var(--font-mono);
		font-size: 0.55rem;
	}

	.directory-count-list {
		border-top: 1px solid var(--book-rule);
	}

	.directory-count-list li {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.58rem 0;
		font-size: 0.71rem;
		letter-spacing: 0.04em;
		border-bottom: 1px solid rgba(76, 58, 43, 0.1);
	}

	.directory-count-list strong {
		font-family: var(--font-mono);
		font-size: 0.52rem;
		font-weight: 500;
		color: var(--book-faint);
	}

	.archive-section {
		margin-top: 2.7rem;
	}

	.directory-empty {
		margin-top: 2rem;
		font-size: 0.72rem;
		line-height: 1.9;
		color: var(--book-muted);
	}

	.index-header {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 2rem;
		padding-bottom: clamp(2rem, 4vw, 3.6rem);
		border-bottom: 1px solid var(--book-rule);
	}

	.index-header h1 {
		margin-top: 0.9rem;
		font-family: var(--font-serif);
		font-size: clamp(3.1rem, 7vw, 6.6rem);
		font-weight: 500;
		line-height: 0.98;
		letter-spacing: 0.08em;
	}

	.index-subtitle {
		max-width: 34rem;
		margin-top: 1.35rem;
		font-family: var(--font-serif);
		font-size: clamp(0.86rem, 1.2vw, 1rem);
		line-height: 1.9;
		letter-spacing: 0.05em;
		color: var(--book-muted);
	}

	.entry-count {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		padding-bottom: 0.35rem;
	}

	.entry-count strong {
		font-family: var(--font-serif);
		font-size: clamp(1.8rem, 3.2vw, 3rem);
		font-weight: 400;
		line-height: 1;
		color: var(--book-accent);
	}

	.entry-count span {
		margin-top: 0.45rem;
		font-family: var(--font-mono);
		font-size: 0.54rem;
		letter-spacing: 0.2em;
		color: var(--book-faint);
	}

	.index-tools {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 2rem;
		padding: 1.35rem 0;
		border-bottom: 1px solid var(--book-rule);
	}

	.index-tools > p {
		font-family: var(--font-serif);
		font-size: 0.78rem;
		letter-spacing: 0.08em;
		color: var(--book-muted);
	}

	.index-tools > p span {
		margin-left: 0.5rem;
		font-family: var(--font-mono);
		font-size: 0.62rem;
	}

	.index-tools form {
		display: flex;
		width: min(100%, 20rem);
		align-items: center;
		border-bottom: 1px solid rgba(56, 47, 40, 0.38);
	}

	.index-tools input {
		min-width: 0;
		flex: 1;
		padding: 0.55rem 0;
		font-family: var(--font-serif);
		font-size: 0.78rem;
		color: var(--book-ink);
		background: transparent;
	}

	.index-tools input::placeholder {
		color: var(--book-faint);
	}

	.index-tools button {
		display: grid;
		width: 2.35rem;
		height: 2.35rem;
		place-items: center;
		color: var(--book-muted);
	}

	.index-tools input:focus-visible,
	.index-tools button:focus-visible {
		outline: 2px solid rgba(141, 56, 45, 0.55);
		outline-offset: 2px;
	}

	.search-state {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding-top: 1.2rem;
		font-family: var(--font-serif);
		font-size: 0.76rem;
		color: var(--book-muted);
	}

	.search-state a,
	.empty-page a {
		display: inline-flex;
		align-items: center;
		gap: 0.32rem;
		color: var(--book-accent);
	}

	.index-pagination {
		display: flex;
		justify-content: center;
		padding: 3rem 0 0.5rem;
	}

	.month-index {
		padding-top: clamp(2.2rem, 4vw, 3.8rem);
	}

	.month-group + .month-group {
		margin-top: clamp(3rem, 6vw, 5.2rem);
	}

	.month-heading {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		padding-bottom: 0.8rem;
		border-bottom: 1px solid rgba(56, 47, 40, 0.32);
	}

	.month-heading h2 {
		font-family: var(--font-serif);
		font-size: clamp(1.05rem, 1.8vw, 1.45rem);
		font-weight: 500;
		letter-spacing: 0.14em;
	}

	.month-heading span {
		font-family: var(--font-mono);
		font-size: 0.55rem;
		letter-spacing: 0.18em;
		color: var(--book-faint);
	}

	.empty-page {
		display: flex;
		min-height: 22rem;
		flex-direction: column;
		align-items: center;
		justify-content: flex-start;
		padding-top: clamp(5.5rem, 10vw, 8rem);
		text-align: center;
		color: var(--book-muted);
	}

	.empty-mark {
		display: grid;
		width: 4.8rem;
		height: 4.8rem;
		margin-bottom: 1.3rem;
		place-items: center;
		border: 1px solid var(--book-rule);
		border-radius: 50%;
		color: var(--book-accent);
	}

	.empty-page p {
		font-family: var(--font-serif);
		font-size: 1.05rem;
		letter-spacing: 0.08em;
		color: var(--book-ink);
	}

	.empty-page > span {
		margin-top: 0.65rem;
		font-family: var(--font-serif);
		font-size: 0.76rem;
	}

	.empty-page a {
		margin-top: 1.5rem;
		font-family: var(--font-serif);
		font-size: 0.76rem;
	}

	:global(.moment-entry-list) {
		display: flex;
		flex-direction: column;
	}

	@media (max-width: 767px) {
		.index-header {
			align-items: flex-start;
		}

		.index-header h1 {
			font-size: clamp(3.2rem, 19vw, 5rem);
		}

		.index-subtitle {
			max-width: 15rem;
		}

		.entry-count {
			padding-top: 1.7rem;
		}

		.entry-count strong {
			font-size: 2.2rem;
		}

		.index-tools {
			align-items: flex-start;
			flex-direction: column;
			gap: 0.8rem;
		}

		.index-tools form {
			width: 100%;
		}
	}
</style>
