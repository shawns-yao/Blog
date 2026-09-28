<script lang="ts">
	import { tick } from 'svelte';
	import { base } from '$app/paths';
	import { LoaderCircle } from 'lucide-svelte';
	import { scrollToElementById } from '$lib/shared/dom/scroll-to-element';
	import type { RagTurn } from '../types';

	let { turns } = $props<{ turns: RagTurn[] }>();
	$effect(() => {
		const latest = turns.at(-1);
		if (!latest) return;
		// Scroll when a turn is appended or its answer arrives, preserving earlier turns.
		void latest.answer;
		void tick().then(() =>
			scrollToElementById('rag-conversation-end', { block: 'end', behavior: 'instant' })
		);
	});
</script>

<div role="log" aria-label="问答记录" aria-live="polite" aria-relevant="additions text">
	<ol class="space-y-6">
		{#each turns as turn (turn.id)}
			<li class="space-y-4">
				<div class="flex justify-end" data-message-role="user">
					<div class="max-w-[85%] min-w-0">
						<p class="mb-1.5 text-right text-xs text-ink-600 dark:text-ink-400">你</p>
						<div
							class="rounded-2xl rounded-tr-sm bg-jade-700 px-4 py-3 text-ink-50 dark:bg-jade-800"
						>
							<p class="whitespace-pre-wrap break-words text-sm leading-7">{turn.question}</p>
						</div>
					</div>
				</div>
				<div class="flex justify-start" data-message-role="assistant">
					<div class="max-w-[95%] min-w-0">
						<p class="mb-1.5 text-xs text-ink-600 dark:text-ink-400">
							AI 助手{turn.answer?.mode === 'conversation' ? ' · 一般交流' : ''}
						</p>
						<div
							class="rounded-2xl rounded-tl-sm border border-ink-200/70 bg-ink-100/70 px-4 py-3 dark:border-ink-700/70 dark:bg-ink-800/70"
						>
							{#if !turn.answer}
								<p
									class="flex items-center gap-2 text-sm leading-7 text-ink-600 dark:text-ink-300"
									role="status"
								>
									<LoaderCircle
										class="size-4 shrink-0 animate-spin motion-reduce:animate-none"
										aria-hidden="true"
									/>正在整理回答…
								</p>
							{:else if turn.answer.status === 'answered'}
								<p class="whitespace-pre-wrap break-words text-sm leading-7">
									{turn.answer.answer}
								</p>
								{#if turn.answer.citations.length > 0}
									<div class="mt-4 border-t border-ink-200 pt-3 dark:border-ink-700">
										<h2 class="mb-2 font-serif text-xs text-ink-600 dark:text-ink-300">原文依据</h2>
										<ol class="space-y-3">
											{#each turn.answer.citations as citation (citation.chunkId)}
												<li>
													<a
														href={`${base}${citation.url}`}
														class="block break-words text-xs leading-6 text-jade-800 underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:text-jade-200 dark:focus-visible:outline-jade-400"
														>[{citation.number}] {citation.title}</a
													>
													<details class="mt-1 text-xs leading-6 text-ink-600 dark:text-ink-300">
														<summary
															class="cursor-pointer rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
															>查看原文片段</summary
														>
														{#if citation.contextHeader}<p class="mt-2 break-words">
																{citation.contextHeader}
															</p>{/if}
														<p class="mt-2 whitespace-pre-wrap break-words">{citation.content}</p>
													</details>
												</li>
											{/each}
										</ol>
									</div>
								{/if}
							{:else}
								<p
									role={turn.answer.status === 'temporarily_unavailable' ? 'alert' : 'status'}
									class="whitespace-pre-wrap break-words text-sm leading-7 text-ink-700 dark:text-ink-200"
								>
									{turn.answer.reason}
								</p>
							{/if}
						</div>
					</div>
				</div>
			</li>
		{/each}
	</ol>
	<div id="rag-conversation-end" class="h-px" aria-hidden="true"></div>
</div>
