<script lang="ts">
	import type { HomeActivityPulsePoint } from './types';

	let { points, label }: { points: HomeActivityPulsePoint[]; label: string } = $props();
	let activeDate = $state<string | null>(null);
	let focusIndex = $state<number | null>(null);
	const total = $derived(points.reduce((sum, point) => sum + point.moments, 0));
	const activeDays = $derived(points.filter((point) => point.moments > 0).length);
	const leadingDays = $derived(
		points.length ? (new Date(`${points[0].date}T00:00:00Z`).getUTCDay() + 6) % 7 : 0
	);
	const weeks = $derived(Math.ceil((leadingDays + points.length) / 7));
	const active = $derived(points.find((point) => point.date === activeDate));
	const level = (count: number) =>
		count === 0 ? 0 : count === 1 ? 1 : count <= 3 ? 2 : count <= 6 ? 3 : 4;

	function moveFocus(event: KeyboardEvent, index: number) {
		const offsets: Record<string, number> = {
			ArrowUp: -1,
			ArrowDown: 1,
			ArrowLeft: -7,
			ArrowRight: 7
		};
		let next: number;
		if (event.key === 'Home') next = 0;
		else if (event.key === 'End') next = points.length - 1;
		else if (event.key in offsets) next = index + offsets[event.key];
		else return;
		event.preventDefault();
		next = Math.max(0, Math.min(points.length - 1, next));
		focusIndex = next;
		const button = event.currentTarget as HTMLButtonElement;
		button.parentElement?.querySelectorAll<HTMLButtonElement>('button')[next]?.focus();
	}
</script>

<div aria-label={`${label}创作记录`}>
	<div
		class="flex flex-wrap items-baseline justify-between gap-3 mb-6 text-sm text-ink-700 dark:text-ink-200"
	>
		<span>{label}</span>
		<div class="flex flex-wrap gap-x-6 gap-y-2">
			<span
				>发布 <strong class="text-lg font-medium text-ink-900 dark:text-ink-100">{total}</strong> 条</span
			>
			<span
				><strong class="text-lg font-medium text-ink-900 dark:text-ink-100">{activeDays}</strong> 天有更新</span
			>
		</div>
	</div>
	{#if total === 0}
		<p class="py-4 text-sm text-ink-700 dark:text-ink-200">这段时间还没有公开的创作记录。</p>
	{:else}
		<div class="calendar" style:--weeks={weeks} role="group" aria-label={`${label}每日发布数量`}>
			{#each Array(leadingDays) as _, index (index)}
				<span aria-hidden="true"></span>
			{/each}
			{#each points as point, index (point.date)}
				<button
					type="button"
					class="day"
					data-level={level(point.moments)}
					aria-label={`${point.date}，发布 ${point.moments} 条`}
					title={`${point.date} · 发布 ${point.moments} 条`}
					tabindex={index === (focusIndex ?? points.length - 1) ? 0 : -1}
					onmouseenter={() => (activeDate = point.date)}
					onfocus={() => {
						activeDate = point.date;
						focusIndex = index;
					}}
					onclick={() => {
						activeDate = point.date;
						focusIndex = index;
					}}
					onkeydown={(event) => moveFocus(event, index)}
				></button>
			{/each}
		</div>
		<div
			class="flex flex-wrap items-center justify-between gap-3 mt-3 text-xs text-ink-600 dark:text-ink-300"
		>
			<span>{points[0]?.date} — {points.at(-1)?.date}</span>
			<div class="flex items-center gap-1.5" aria-label="颜色由浅至深表示发布数量从少到多">
				<span>少</span>
				{#each [0, 1, 2, 3, 4] as value (value)}
					<span class="legend-cell" data-level={value} aria-hidden="true"></span>
				{/each}
				<span>多</span>
			</div>
		</div>
		<p
			class="mt-3 min-h-5 text-xs text-ink-700 dark:text-ink-200"
			aria-live="polite"
			aria-atomic="true"
		>
			{active ? `${active.date} · 发布 ${active.moments} 条` : '\u00a0'}
		</p>
	{/if}
</div>

<style>
	.calendar {
		display: grid;
		grid-template-rows: repeat(7, auto);
		grid-template-columns: repeat(var(--weeks), minmax(0, 1fr));
		grid-auto-flow: column;
		gap: 3px;
	}
	.day,
	.legend-cell {
		display: block;
		border-radius: 2px;
		background: rgb(87 100 93 / 0.18);
		border: 1px solid rgb(87 100 93 / 0.2);
	}
	.day {
		width: 100%;
		aspect-ratio: 1;
		max-height: 18px;
		padding: 0;
		cursor: pointer;
	}
	.legend-cell {
		width: 10px;
		height: 10px;
	}
	[data-level='1'] {
		background: #aad4bc;
	}
	[data-level='2'] {
		background: #6ab492;
	}
	[data-level='3'] {
		background: #32856a;
	}
	[data-level='4'] {
		background: #14533f;
	}
	:global(.dark) [data-level='0'] {
		background: rgb(200 215 205 / 0.15);
		border-color: rgb(200 215 205 / 0.2);
	}
	:global(.dark) [data-level='1'] {
		background: #2e6550;
	}
	:global(.dark) [data-level='2'] {
		background: #428b68;
	}
	:global(.dark) [data-level='3'] {
		background: #6cbd8b;
	}
	:global(.dark) [data-level='4'] {
		background: #ace4b8;
	}
	.day:hover,
	.day:focus-visible {
		outline: 2px solid var(--color-ink-900);
		outline-offset: 1px;
	}
	:global(.dark) .day:hover,
	:global(.dark) .day:focus-visible {
		outline-color: var(--color-ink-100);
	}
</style>
