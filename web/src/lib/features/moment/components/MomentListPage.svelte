<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { navigating } from '$app/state';
	import { momentListCtx } from '$lib/features/moment/context';
	import type { MomentListResponse } from '$lib/features/moment/types';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import { Search, NotebookPen, X } from 'lucide-svelte';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import MomentFeedItem from './MomentFeedItem.svelte';

	interface Props {
		moments: MomentListResponse;
		search?: string;
		basePath?: string;
		staggerKey?: string;
	}

	let { moments, search = '', basePath = '/moments', staggerKey = 'moments' }: Props = $props();
	momentListCtx.mountModelData(() => moments);
	const totalPages = $derived(
		moments.size > 0 ? Math.max(1, Math.ceil(moments.total / moments.size)) : 1
	);

	function onPageChange(nextPage: number) {
		const safePage = Number.isFinite(nextPage) && nextPage > 1 ? nextPage : 1;
		const path = resolvePath(safePage === 1 ? `${basePath}/` : `${basePath}/page/${safePage}/`);
		goto(`${path}${search ? `?${new URLSearchParams({ q: search })}` : ''}`);
	}
</script>

<section class="mx-auto w-full max-w-4xl pb-16" aria-labelledby="moments-heading">
	<header class="border-b border-ink-200 pb-8 dark:border-ink-800">
		<h1 id="moments-heading" class="font-serif text-3xl font-medium text-ink-900 dark:text-ink-100">
			手记
		</h1>
		<p class="mt-3 text-sm text-ink-600 dark:text-ink-400">日常、片刻与随想</p>
	</header>
	<div
		class="flex flex-wrap items-center justify-between gap-4 border-b border-ink-200 py-5 dark:border-ink-800"
	>
		<p class="text-sm text-ink-600 dark:text-ink-400">
			{search ? '搜索结果' : '全部手记'} <span class="ml-2 font-mono">{moments.total}</span>
		</p>
		<form
			action={resolvePath(`${basePath}/`)}
			method="GET"
			role="search"
			class="flex w-full items-center gap-2 border-b border-ink-400 sm:w-72 dark:border-ink-600"
		>
			<label for="moment-query" class="sr-only">搜索手记</label>
			<input
				id="moment-query"
				name="q"
				value={search}
				type="search"
				placeholder="搜索手记"
				class="min-w-0 flex-1 bg-transparent py-2 text-sm text-ink-900 outline-none focus-visible:ring-2 focus-visible:ring-jade-500 dark:text-ink-100"
			/>
			<button
				type="submit"
				aria-label="搜索手记"
				title="搜索手记"
				class="flex size-9 shrink-0 items-center justify-center text-ink-600 hover:text-jade-600 focus-visible:outline-2 focus-visible:outline-jade-500 dark:text-ink-300"
				><Search size={18} /></button
			>
		</form>
	</div>
	{#if search}
		<div class="flex items-center gap-3 pt-5 text-sm text-ink-600 dark:text-ink-300">
			<span class="min-w-0 break-words">“{search}”</span>
			<a
				href={resolvePath(`${basePath}/`)}
				class="inline-flex shrink-0 items-center gap-1 text-jade-700 underline-offset-4 hover:underline dark:text-jade-400"
				><X size={14} />清除搜索</a
			>
		</div>
	{/if}

	<div aria-busy={!!navigating.to}>
		{#if moments.items.length > 0}
			<StaggerList
				class="flex flex-col"
				staggerDelay={40}
				duration={250}
				y={8}
				key={`${staggerKey}-${search}`}
			>
				{#each moments.items as moment (moment.id)}
					<MomentFeedItem {moment} />
				{/each}
			</StaggerList>

			{#if totalPages > 1}
				<div class="flex justify-center pt-8 pb-4 sm:pt-10 sm:pb-8">
					<Pagination current={moments.page} total={totalPages} {onPageChange} />
				</div>
			{/if}
		{:else}
			<div
				class="flex flex-col items-center gap-4 py-20 text-ink-500 dark:text-ink-400"
				role="status"
			>
				<NotebookPen size={28} strokeWidth={1.25} aria-hidden="true" />
				<p class="text-sm">{search ? '没有找到相关手记' : '还没有公开的手记'}</p>
				{#if search}
					<a
						href={resolvePath(`${basePath}/`)}
						class="text-sm text-jade-700 underline underline-offset-4 dark:text-jade-400"
						>查看全部手记</a
					>
				{/if}
			</div>
		{/if}
	</div>
</section>
