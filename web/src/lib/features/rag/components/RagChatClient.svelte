<script lang="ts">
	import { onDestroy, onMount, tick } from 'svelte';
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import { ArrowUp, BookOpen, RotateCw, Search } from 'lucide-svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import Textarea from '$lib/ui/primitives/textarea/Textarea.svelte';
	import { askRag, getRagAvailability } from '../api';
	import { conversationHistory } from '../conversation';
	import type { RagAnswer, RagMessage, RagTurn } from '../types';
	import RagTranscript from './RagTranscript.svelte';

	type Submission = { id: string; question: string; history: RagMessage[] };

	let { question, sessionId, turns, onQuestionChange, onTurnsChange, onSearch } = $props<{
		question: string;
		sessionId: string;
		turns: RagTurn[];
		onQuestionChange: (value: string) => void;
		onTurnsChange: (value: RagTurn[]) => void;
		onSearch: () => void;
	}>();
	let input: HTMLTextAreaElement | undefined = $state();
	let controller: AbortController | undefined;
	let pendingTurnId: string | undefined;
	let disposed = false;

	const availability = createQuery(() => ({
		queryKey: ['rag-availability'],
		queryFn: ({ signal }) => getRagAvailability(signal),
		staleTime: 0,
		gcTime: 0,
		retry: false,
		refetchOnWindowFocus: false
	}));
	const mutation = createMutation(() => ({
		mutationFn: (value: Submission) => {
			controller = new AbortController();
			if (disposed) controller.abort();
			return askRag(value.question, sessionId, controller.signal, value.history);
		},
		retry: false,
		gcTime: 0,
		onSuccess: (result: RagAnswer, value: Submission) => {
			if (disposed || controller?.signal.aborted) return;
			completeTurn(value.id, result);
		},
		onError: (_error: unknown, value: Submission) => {
			if (disposed || controller?.signal.aborted) return;
			completeTurn(value.id, {
				status: 'temporarily_unavailable',
				answer: '',
				reason: '请求暂时未能完成，请稍后重试或使用站内搜索。',
				citations: []
			});
		},
		onSettled: () => {
			if (!disposed) void tick().then(() => input?.focus());
		}
	}));
	let ready = $derived(availability.data?.available === true);
	let availabilityText = $derived(
		availability.isPending
			? '正在检查问答服务…'
			: availability.isError
				? '暂时无法连接问答服务，可以先用搜索查找文章与手记。'
				: ready
					? '可以与我交流，也可以提问本站文章与手记。'
					: availability.data?.reason === 'index_not_ready'
						? '公开内容正在准备中，请稍后重试。'
						: '问答服务尚未开放，可以先用搜索查找文章与手记。'
	);
	let offerSearch = $derived(
		!ready || (turns.at(-1)?.answer && turns.at(-1)?.answer?.status !== 'answered')
	);

	onMount(() => input?.focus());
	onDestroy(() => {
		disposed = true;
		controller?.abort();
		if (pendingTurnId) {
			completeTurn(pendingTurnId, {
				status: 'temporarily_unavailable',
				answer: '',
				citations: [],
				reason: '回答已中止，可以重新发送问题。'
			});
		}
	});

	function completeTurn(id: string, answer: RagAnswer) {
		onTurnsChange(turns.map((turn: RagTurn) => (turn.id === id ? { ...turn, answer } : turn)));
		pendingTurnId = undefined;
	}

	function send() {
		const value = question.trim();
		if (!ready || !value || mutation.isPending) return;
		const history = conversationHistory(turns);
		const id = crypto.randomUUID();
		pendingTurnId = id;
		onTurnsChange([...turns, { id, question: value, answer: null }]);
		onQuestionChange('');
		mutation.mutate({ id, question: value, history });
	}

	function submit(event: SubmitEvent) {
		event.preventDefault();
		send();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key !== 'Enter' || event.shiftKey || event.isComposing || event.keyCode === 229)
			return;
		event.preventDefault();
		send();
	}
</script>

<div class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 py-5 sm:px-5">
	{#if turns.length > 0}
		<RagTranscript {turns} />
	{:else}
		<div class="py-6">
			<BookOpen class="mb-6 size-8 text-jade-800 dark:text-jade-300" aria-hidden="true" />
			<h2 class="font-serif text-2xl leading-relaxed">{ready ? '向书房提问' : '站内问答'}</h2>
			<p class="mt-3 text-sm leading-7 text-ink-600 dark:text-ink-300" role="status">
				{availabilityText}
			</p>
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
	{#if offerSearch && !mutation.isPending}
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
		onkeydown={handleKeydown}
		rows={2}
		maxLength={1000}
		resize="none"
		disabled={mutation.isPending}
		aria-describedby="rag-availability rag-keyboard-help"
		placeholder="输入你的问题…"
		textareaClass="block text-sm leading-6 disabled:opacity-60"
	/>
	<p id="rag-availability" class="sr-only">{availabilityText}</p>
	<div class="mt-3 flex items-center justify-between gap-3">
		<p id="rag-keyboard-help" class="text-xs leading-5 text-ink-600 dark:text-ink-400">
			Enter 发送 · Shift+Enter 换行
		</p>
		<Button
			type="submit"
			loading={mutation.isPending}
			disabled={!ready || !question.trim() || mutation.isPending}
			aria-describedby="rag-availability"
			class="min-h-10 shrink-0 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
		>
			{mutation.isPending ? '正在回答' : '发送问题'}<ArrowUp class="size-4" aria-hidden="true" />
		</Button>
	</div>
</form>
