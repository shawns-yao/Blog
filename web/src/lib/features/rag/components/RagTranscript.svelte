<script lang="ts">
	import { tick } from 'svelte';
	import { LoaderCircle } from 'lucide-svelte';
	import { scrollToElementById } from '$lib/shared/dom/scroll-to-element';
	import type { RagTurn } from '../types';
	import RagAvatar from './RagAvatar.svelte';
	import RagAnswerReading from './RagAnswerReading.svelte';

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
	<ol class="mx-auto max-w-[720px] space-y-8">
		{#each turns as turn (turn.id)}
			<li class="space-y-4">
				<div class="flex justify-end" data-message-role="user">
					<div class="max-w-[85%] min-w-0">
						<p class="mb-1.5 text-right text-xs text-ink-600 dark:text-ink-400">你</p>
						<div
							class="rounded-2xl rounded-tr-sm bg-ink-200/80 px-4 py-3 text-ink-900 dark:bg-ink-800 dark:text-ink-100"
						>
							<p class="whitespace-pre-wrap break-words text-sm leading-7">{turn.question}</p>
						</div>
					</div>
				</div>
				{#if turn.answer?.status === 'answered'}
					<div data-message-role="assistant">
						<RagAnswerReading answer={turn.answer} turnId={turn.id} />
					</div>
				{:else}
					<div class="flex items-start gap-3" data-message-role="assistant">
						<RagAvatar class="size-8 shrink-0" />
						<div class="min-w-0 flex-1">
							<p class="mb-1.5 text-xs text-ink-600 dark:text-ink-400">书灵</p>
							<div class="border-l border-jade-600/40 pl-3 dark:border-jade-400/40">
								{#if !turn.answer}
									<p
										class="flex items-center gap-2 text-sm leading-7 text-ink-600 dark:text-ink-300"
										role="status"
									>
										<LoaderCircle
											class="size-4 shrink-0 animate-spin motion-reduce:animate-none"
											aria-hidden="true"
										/>正在知识的海洋中遨游
									</p>
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
				{/if}
			</li>
		{/each}
	</ol>
	<div id="rag-conversation-end" class="h-px" aria-hidden="true"></div>
</div>
