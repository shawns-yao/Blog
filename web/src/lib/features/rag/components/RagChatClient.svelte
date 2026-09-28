<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { base } from '$app/paths';
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import { ArrowUp, BookOpen, LoaderCircle, RotateCw, Search } from 'lucide-svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import Textarea from '$lib/ui/primitives/textarea/Textarea.svelte';
	import { askRag, getRagAvailability } from '../api';
	import type { RagAnswer } from '../types';

	let { question, sessionId, onQuestionChange, onSearch } = $props<{
		question: string;
		sessionId: string;
		onQuestionChange: (value: string) => void;
		onSearch: () => void;
	}>();
	let input: HTMLTextAreaElement | undefined = $state();
	let answer = $state<RagAnswer | null>(null);
	let askedQuestion = $state('');
	let controller: AbortController | undefined;

	const availability = createQuery(() => ({
		queryKey: ['rag-availability'],
		queryFn: ({ signal }) => getRagAvailability(signal),
		staleTime: 0,
		gcTime: 0,
		retry: false,
		refetchOnWindowFocus: false
	}));
	const mutation = createMutation(() => ({
		mutationFn: (value: string) => {
			controller = new AbortController();
			return askRag(value, sessionId, controller.signal);
		},
		retry: false,
		gcTime: 0,
		onSuccess: (result: RagAnswer) => {
			answer = result;
		},
		onError: () => {
			if (controller?.signal.aborted) return;
			answer = {
				status: 'temporarily_unavailable',
				answer: '',
				reason: '请求暂时未能完成，请稍后重试或使用站内搜索。',
				citations: []
			};
		}
	}));
	let ready = $derived(availability.data?.available === true);
	let availabilityText = $derived(
		availability.isPending
			? '正在检查问答服务…'
			: availability.isError
				? '暂时无法连接问答服务，可以先用搜索查找文章与手记。'
				: ready
					? '仅依据本站已发布的文章与手记回答。'
					: availability.data?.reason === 'index_not_ready'
						? '公开内容正在准备中，请稍后重试。'
						: '问答服务尚未开放，可以先用搜索查找文章与手记。'
	);

	onMount(() => input?.focus());
	onDestroy(() => controller?.abort());

	function submit(event: SubmitEvent) {
		event.preventDefault();
		const value = question.trim();
		if (!ready || !value || mutation.isPending) return;
		askedQuestion = value;
		answer = null;
		mutation.mutate(value);
	}
</script>

<div
	class="min-h-0 flex-1 overflow-y-auto px-6 py-6"
	aria-live="polite"
	aria-busy={mutation.isPending}
>
	{#if mutation.isPending}
		<p class="mb-5 whitespace-pre-wrap text-sm leading-7 text-ink-700 dark:text-ink-200">
			{askedQuestion}
		</p>
		<p class="flex items-center gap-2 text-sm text-ink-600 dark:text-ink-300">
			<LoaderCircle class="size-4 animate-spin motion-reduce:animate-none" aria-hidden="true" />
			正在检索原文并整理回答…
		</p>
	{:else if answer}
		<p class="mb-5 whitespace-pre-wrap text-sm leading-7 text-ink-600 dark:text-ink-300">
			{askedQuestion}
		</p>
		{#if answer.status === 'answered'}
			<p class="whitespace-pre-wrap break-words font-serif text-base leading-8">{answer.answer}</p>
			<h2 class="mb-2 mt-8 font-serif text-sm font-medium">原文依据</h2>
			<ol class="divide-y divide-ink-200 dark:divide-ink-700">
				{#each answer.citations as citation (citation.chunkId)}
					<li class="py-4">
						<a
							href={`${base}${citation.url}`}
							class="block rounded-sm font-serif text-sm leading-6 text-jade-800 underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jade-700 dark:text-jade-200 dark:focus-visible:outline-jade-400"
						>
							[{citation.number}] {citation.title}
						</a>
						{#if citation.contextHeader}
							<p class="mt-2 break-words text-xs leading-5 text-ink-600 dark:text-ink-300">
								{citation.contextHeader}
							</p>
						{/if}
						<details class="mt-3 text-xs leading-6 text-ink-600 dark:text-ink-300">
							<summary
								class="cursor-pointer rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
								>查看原文片段</summary
							>
							<p class="mt-2 whitespace-pre-wrap break-words">{citation.content}</p>
						</details>
					</li>
				{/each}
			</ol>
		{:else}
			<p
				role={answer.status === 'temporarily_unavailable' ? 'alert' : 'status'}
				class="text-sm leading-7 text-ink-700 dark:text-ink-200"
			>
				{answer.reason}
			</p>
		{/if}
	{:else}
		<div class="py-6">
			<BookOpen class="mb-6 size-8 text-jade-800 dark:text-jade-300" aria-hidden="true" />
			<h2 class="font-serif text-2xl leading-relaxed">{ready ? '向书房提问' : '站内问答'}</h2>
			<p class="mt-3 text-sm leading-7 text-ink-600 dark:text-ink-300">{availabilityText}</p>
		</div>
	{/if}
	{#if !ready && !availability.isPending}
		<Button
			variant="ghost"
			type="button"
			onclick={() => availability.refetch()}
			class="mt-4 min-h-10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
		>
			<RotateCw class="size-4" aria-hidden="true" />重新检查
		</Button>
	{/if}
	{#if !mutation.isPending}
		<Button
			variant="secondary"
			type="button"
			onclick={onSearch}
			class="mt-4 min-h-10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
		>
			<Search class="size-4" aria-hidden="true" />使用站内搜索
		</Button>
	{/if}
</div>

<form
	onsubmit={submit}
	class="shrink-0 border-t border-ink-200 bg-ink-100/50 p-5 dark:border-ink-700 dark:bg-ink-800/30"
	style:padding-bottom="calc(var(--spacing) * 5 + env(safe-area-inset-bottom))"
>
	<div class="mb-3 flex items-center justify-between gap-3">
		<label for="rag-question" class="font-serif text-sm">你的问题</label>
		<Button
			variant="ghost"
			size="sm"
			type="button"
			disabled={!question || mutation.isPending}
			onclick={() => {
				onQuestionChange('');
				input?.focus();
			}}
			class="min-h-10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
			>清空草稿</Button
		>
	</div>
	<Textarea
		id="rag-question"
		bind:ref={input}
		value={question}
		oninput={() => onQuestionChange(input?.value ?? '')}
		rows={3}
		maxLength={1000}
		resize="none"
		disabled={mutation.isPending}
		aria-describedby="rag-availability"
		placeholder="输入你的问题…"
		textareaClass="block text-sm leading-6 disabled:opacity-60"
	/>
	<p id="rag-availability" class="sr-only">{availabilityText}</p>
	<div class="mt-3 flex justify-end">
		<Button
			type="submit"
			loading={mutation.isPending}
			disabled={!ready || !question.trim() || mutation.isPending}
			aria-describedby="rag-availability"
			class="min-h-10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
		>
			{mutation.isPending ? '正在回答' : '发送问题'}<ArrowUp class="size-4" aria-hidden="true" />
		</Button>
	</div>
</form>
