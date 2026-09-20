<script lang="ts">
	import ActivityCalendar from './ActivityCalendar.svelte';
	import type { HomeActivityPulseData, HomeActivityPulseThemeConfig } from './types';

	let {
		pulse,
		failed = false,
		config
	}: {
		pulse: HomeActivityPulseData | null;
		failed?: boolean;
		config?: HomeActivityPulseThemeConfig;
	} = $props();
	const year = $derived((pulse?.points ?? []).slice(-365));
	const quarter = $derived.by(() => {
		const last = year.at(-1);
		if (!last) return [];
		const end = new Date(`${last.date}T00:00:00Z`);
		const day = end.getUTCDate();
		const start = new Date(Date.UTC(end.getUTCFullYear(), end.getUTCMonth() - 3, 1));
		const maxDay = new Date(
			Date.UTC(start.getUTCFullYear(), start.getUTCMonth() + 1, 0)
		).getUTCDate();
		start.setUTCDate(Math.min(day, maxDay) + 1);
		return year.filter((point) => point.date >= start.toISOString().slice(0, 10));
	});
</script>

<section class="mt-16 md:mt-24 pb-8" aria-labelledby="activity-title">
	<h2
		id="activity-title"
		class="flex items-center gap-3 border-b border-ink-300/50 dark:border-ink-600/50 pb-4 mb-6 text-xl font-serif font-medium text-ink-900 dark:text-ink-100"
	>
		<span class="h-px w-8 bg-jade-500/60"></span>
		{config?.title || '创作律动'}
	</h2>
	{#if failed || !pulse}
		<p role="status" class="py-6 text-sm text-ink-700 dark:text-ink-200">
			创作记录暂时加载失败，请稍后重试。
		</p>
	{:else}
		<div class="hidden md:block"><ActivityCalendar points={year} label="近一年" /></div>
		<div class="md:hidden"><ActivityCalendar points={quarter} label="近三个月" /></div>
	{/if}
</section>
