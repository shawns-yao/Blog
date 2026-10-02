<script lang="ts">
	import { tick } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { base } from '$app/paths';
	import { Tooltip } from 'bits-ui';
	import { ArrowUpRight, ChevronDown, Feather, Quote } from 'lucide-svelte';
	import { scrollToElementById } from '$lib/shared/dom/scroll-to-element';
	import { answerParagraphs, readableExcerpt } from '../answer-presentation';
	import type { RagAnswer, RagCitation } from '../types';
	import RagCitationMarker from './RagCitationMarker.svelte';

	let { answer, turnId } = $props<{ answer: RagAnswer; turnId: string }>();
	const paragraphs = $derived(answerParagraphs(answer.answer, answer.citations));
	const openSources = new SvelteSet<number>();
	const sourceId = (citation: RagCitation) => `rag-source-${turnId}-${citation.chunkId}`;

	async function revealSource(citation: RagCitation) {
		openSources.add(citation.number);
		await tick();
		scrollToElementById(sourceId(citation), { block: 'start', behavior: 'auto' })?.focus({
			preventScroll: true
		});
	}
</script>

<article
	aria-label="书灵回答"
	class="mx-auto w-full max-w-[720px] rounded-xl border border-ink-200/80 bg-white/60 px-4 py-5 shadow-[0_3px_18px_rgba(70,55,40,0.035)] sm:px-7 sm:py-7 dark:border-ink-700/80 dark:bg-ink-900/40"
>
	<header
		class="mb-6 flex items-center gap-2.5 border-b border-ink-200/80 pb-4 dark:border-ink-700"
	>
		<Feather class="size-5 shrink-0 text-jade-800 dark:text-jade-300" aria-hidden="true" />
		<h2 class="font-serif text-base font-medium text-jade-800 dark:text-jade-200">书灵回答</h2>
	</header>

	<Tooltip.Provider delayDuration={180}>
		<div class="space-y-5 font-serif text-base leading-[1.9] text-ink-800 dark:text-ink-200">
			{#each paragraphs as paragraph, index (index)}
				<p class="whitespace-pre-wrap text-left indent-[2em] break-words">
					{#each paragraph as part, partIndex (partIndex)}
						{#if part.citation}
							<RagCitationMarker
								citation={part.citation}
								sourceId={sourceId(part.citation)}
								onSelect={() => revealSource(part.citation!)}
							/>
						{:else}{part.text}{/if}
					{/each}
				</p>
			{/each}
		</div>
	</Tooltip.Provider>

	{#if answer.citations.length > 0}
		<section aria-labelledby={`rag-sources-${turnId}`} class="mt-8">
			<div class="mb-5 flex items-center gap-3 text-ink-600 dark:text-ink-400">
				<span class="h-px flex-1 bg-ink-200 dark:bg-ink-700" aria-hidden="true"></span>
				<h3 id={`rag-sources-${turnId}`} class="shrink-0 font-serif text-sm">
					原文依据 <span class="ml-1 font-sans text-xs">· {answer.citations.length} 条</span>
				</h3>
				<span class="h-px flex-1 bg-ink-200 dark:bg-ink-700" aria-hidden="true"></span>
			</div>
			<ol class="space-y-5">
				{#each answer.citations as citation (citation.chunkId)}
					<li
						id={sourceId(citation)}
						tabindex="-1"
						class="scroll-mt-6 rounded-sm border-b border-ink-200/80 pb-5 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jade-700 last:border-b-0 last:pb-0 dark:border-ink-700 dark:focus-visible:outline-jade-400"
					>
						<div class="flex items-start gap-3 sm:gap-5">
							<span
								class="shrink-0 pt-0.5 font-serif text-2xl leading-none text-jade-800/65 sm:text-3xl dark:text-jade-200/70"
								aria-hidden="true"
							>
								{String(citation.number).padStart(2, '0')}
							</span>
							<div class="min-w-0 flex-1">
								<a
									href={`${base}${citation.url}`}
									target="_blank"
									rel="noopener noreferrer"
									aria-label={`在新标签页查看原文：${citation.title}`}
									class="group flex items-start justify-between gap-3 rounded-sm font-serif text-sm leading-7 text-ink-800 hover:text-jade-800 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jade-700 sm:text-base dark:text-ink-200 dark:hover:text-jade-200 dark:focus-visible:outline-jade-400"
								>
									<span class="min-w-0 break-words">{citation.title.replace(/^\d{2}\s+/, '')}</span>
									<ArrowUpRight
										class="mt-1 size-4 shrink-0 text-jade-800 dark:text-jade-300"
										aria-hidden="true"
									/>
								</a>
								<p class="mt-1 text-xs leading-6 text-ink-500 dark:text-ink-400">
									{citation.contentKind === 'article' ? '文章' : '手记'}
								</p>
							</div>
						</div>
						<details
							open={openSources.has(citation.number)}
							ontoggle={(event) => {
								if (event.currentTarget.open) openSources.add(citation.number);
								else openSources.delete(citation.number);
							}}
							class="group mt-3"
						>
							<summary
								class="flex min-h-10 cursor-pointer list-none items-center gap-2 rounded-sm text-sm text-ink-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 [&::-webkit-details-marker]:hidden dark:text-ink-300 dark:focus-visible:outline-jade-400"
							>
								<ChevronDown
									class="size-4 shrink-0 transition-transform group-open:rotate-180 motion-reduce:transition-none"
									aria-hidden="true"
								/>
								<span>查看原文片段</span>
								<span
									class="ml-auto hidden text-xs text-ink-500 group-open:inline dark:text-ink-400"
									>收起</span
								>
							</summary>
							<!-- svelte-ignore a11y_no_noninteractive_tabindex (Long excerpts need keyboard focus for scrolling.) -->
							<blockquote
								tabindex="0"
								aria-label={`引用 ${citation.number} 的原文片段`}
								class="relative mt-2 max-h-80 overflow-y-auto overscroll-contain rounded-r-md border-l-2 border-jade-600/60 bg-[#f8f6f0] py-4 pr-4 pl-5 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:border-jade-400/50 dark:bg-ink-800/50 dark:focus-visible:outline-jade-400"
							>
								<Quote
									class="mb-2 size-4 text-jade-800/35 dark:text-jade-200/40"
									aria-hidden="true"
								/>
								<p
									class="whitespace-pre-wrap font-serif text-sm leading-[1.9] text-ink-700 break-words dark:text-ink-300"
								>
									{readableExcerpt(citation.content)}
								</p>
								{#if citation.contextHeader}
									<footer
										class="mt-4 whitespace-pre-wrap text-xs leading-6 text-ink-500 break-words dark:text-ink-400"
									>
										{citation.contextHeader}
									</footer>
								{/if}
							</blockquote>
						</details>
					</li>
				{/each}
			</ol>
		</section>
	{/if}
</article>
