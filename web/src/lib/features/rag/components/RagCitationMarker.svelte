<script lang="ts">
	import { Tooltip } from 'bits-ui';
	import type { RagCitation } from '../types';
	import { readableExcerpt } from '../answer-presentation';

	let { citation, sourceId, onSelect } = $props<{
		citation: RagCitation;
		sourceId: string;
		onSelect: () => void;
	}>();
	const excerpt = $derived(readableExcerpt(citation.content).replace(/\s+/g, ' '));
</script>

<sup class="relative -top-0.5 ml-0.5 inline-block indent-0">
	<Tooltip.Root>
		<Tooltip.Trigger
			type="button"
			aria-label={`查看引用 ${citation.number}：${citation.title}`}
			aria-controls={sourceId}
			onclick={onSelect}
			class="inline-flex size-6 items-center justify-center rounded-sm text-[11px] leading-none font-medium text-jade-800 transition-colors hover:bg-jade-100/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:text-jade-200 dark:hover:bg-jade-900/40 dark:focus-visible:outline-jade-400"
		>
			{citation.number}
		</Tooltip.Trigger>
		<Tooltip.Portal>
			<Tooltip.Content
				role="tooltip"
				side="top"
				sideOffset={6}
				style="z-index: calc(var(--z-index-rag-panel) + 1)"
				class="max-w-[min(320px,calc(100vw-32px))] rounded-md border border-ink-200 bg-ink-50 px-4 py-3 text-left text-xs leading-6 text-ink-700 shadow-float dark:border-ink-700 dark:bg-ink-900 dark:text-ink-200"
			>
				<p class="font-serif font-medium text-ink-900 dark:text-ink-100">
					[{citation.number}] {citation.title}
				</p>
				<p class="mt-1.5">{excerpt.slice(0, 180)}{excerpt.length > 180 ? '…' : ''}</p>
			</Tooltip.Content>
		</Tooltip.Portal>
	</Tooltip.Root>
</sup>
