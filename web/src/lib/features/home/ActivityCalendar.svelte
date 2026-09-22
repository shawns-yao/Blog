<script lang="ts">
	import { Tooltip as BitsTooltip } from 'bits-ui';
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
		<BitsTooltip.Provider delayDuration={80} skipDelayDuration={120}>
			<div class="calendar" style:--weeks={weeks} role="group" aria-label={`${label}每日发布数量`}>
				{#each Array(leadingDays) as _, index (index)}
					<span aria-hidden="true"></span>
				{/each}
				{#each points as point, index (point.date)}
					<BitsTooltip.Root disableCloseOnTriggerClick={true}>
						<BitsTooltip.Trigger
							type="button"
							class="day"
							data-level={level(point.moments)}
							aria-label={`${point.date}，发布 ${point.moments} 条`}
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
						></BitsTooltip.Trigger>
						<BitsTooltip.Portal>
							<BitsTooltip.Content side="top" sideOffset={7} class="activity-tip">
								<span class="activity-tip-date">{point.date}</span>
								<span class="activity-tip-count">
									<span
										class="activity-tip-dot"
										data-level={level(point.moments)}
										aria-hidden="true"
									></span>
									{point.moments === 0 ? '没有发布' : `发布 ${point.moments} 条`}
								</span>
								<BitsTooltip.Arrow class="activity-tip-arrow" />
							</BitsTooltip.Content>
						</BitsTooltip.Portal>
					</BitsTooltip.Root>
				{/each}
			</div>
		</BitsTooltip.Provider>
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
		<p class="sr-only" aria-live="polite" aria-atomic="true">
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
	.calendar :global(.day),
	.legend-cell {
		display: block;
		border-radius: 2px;
		background: rgb(87 100 93 / 0.18);
		border: 1px solid rgb(87 100 93 / 0.2);
	}
	.calendar :global(.day) {
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
	.calendar :global(.day[data-level='1']),
	.legend-cell[data-level='1'],
	:global(.activity-tip-dot[data-level='1']) {
		background: #aad4bc;
	}
	.calendar :global(.day[data-level='2']),
	.legend-cell[data-level='2'],
	:global(.activity-tip-dot[data-level='2']) {
		background: #6ab492;
	}
	.calendar :global(.day[data-level='3']),
	.legend-cell[data-level='3'],
	:global(.activity-tip-dot[data-level='3']) {
		background: #32856a;
	}
	.calendar :global(.day[data-level='4']),
	.legend-cell[data-level='4'],
	:global(.activity-tip-dot[data-level='4']) {
		background: #14533f;
	}
	:global(.dark) .calendar :global(.day[data-level='0']),
	:global(.dark) .legend-cell[data-level='0'] {
		background: rgb(200 215 205 / 0.15);
		border-color: rgb(200 215 205 / 0.2);
	}
	:global(.dark) .calendar :global(.day[data-level='1']),
	:global(.dark) .legend-cell[data-level='1'],
	:global(.dark .activity-tip-dot[data-level='1']) {
		background: #2e6550;
	}
	:global(.dark) .calendar :global(.day[data-level='2']),
	:global(.dark) .legend-cell[data-level='2'],
	:global(.dark .activity-tip-dot[data-level='2']) {
		background: #428b68;
	}
	:global(.dark) .calendar :global(.day[data-level='3']),
	:global(.dark) .legend-cell[data-level='3'],
	:global(.dark .activity-tip-dot[data-level='3']) {
		background: #6cbd8b;
	}
	:global(.dark) .calendar :global(.day[data-level='4']),
	:global(.dark) .legend-cell[data-level='4'],
	:global(.dark .activity-tip-dot[data-level='4']) {
		background: #ace4b8;
	}
	.calendar :global(.day:hover),
	.calendar :global(.day:focus-visible) {
		outline: 2px solid var(--color-ink-900);
		outline-offset: 1px;
	}
	:global(.dark) .calendar :global(.day:hover),
	:global(.dark) .calendar :global(.day:focus-visible) {
		outline-color: var(--color-ink-100);
	}
	:global(.activity-tip) {
		z-index: 60;
		display: grid;
		gap: 2px;
		min-width: 116px;
		padding: 8px 10px;
		border: 1px solid rgb(255 255 255 / 0.28);
		border-radius: 7px;
		background: rgb(37 34 29 / 0.94);
		box-shadow: 0 8px 24px rgb(19 16 12 / 0.28);
		color: #f5efe3;
		font-family: var(--font-serif);
		backdrop-filter: blur(8px);
		animation: activity-tip-in 120ms ease-out;
	}
	:global(.activity-tip-date) {
		font-size: 11px;
		color: rgb(245 239 227 / 0.72);
		letter-spacing: 0.02em;
	}
	:global(.activity-tip-count) {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		font-weight: 500;
	}
	:global(.activity-tip-dot) {
		width: 7px;
		height: 7px;
		border-radius: 2px;
		border: 1px solid rgb(255 255 255 / 0.24);
	}
	:global(.activity-tip-arrow) {
		fill: rgb(37 34 29 / 0.94);
	}
	@keyframes activity-tip-in {
		from {
			opacity: 0;
			transform: translateY(2px) scale(0.98);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		:global(.activity-tip) {
			animation: none;
		}
	}
</style>
