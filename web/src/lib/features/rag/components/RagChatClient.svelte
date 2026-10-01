<script lang="ts">
	import { onDestroy, onMount, tick, untrack } from 'svelte';
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import { ArrowUp, Eraser, Link2, MessageCircle, RotateCw, UserRound } from 'lucide-svelte';
	import { resolveHref } from '$lib/shared/utils/resolve-path';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import Textarea from '$lib/ui/primitives/textarea/Textarea.svelte';
	import { askRag, getRagAvailability } from '../api';
	import { conversationHistory } from '../conversation';
	import type { RagAnswer, RagMessage, RagTurn } from '../types';
	import RagTranscript from './RagTranscript.svelte';
	import RagAvatar from './RagAvatar.svelte';

	type Submission = { id: string; question: string; history: RagMessage[] };
	const suggestedQuestions = [
		{
			label: 'RAG 如何工作',
			question: 'RAG 的工作原理是什么？'
		},
		{
			label: 'Java 值传递',
			question: 'Java 是值传递还是引用传递？'
		},
		{
			label: 'Agent 长期记忆',
			question: 'Agent 如何实现长期记忆？'
		},
		{
			label: 'RAG 与微调',
			question: 'RAG 和微调分别适合解决什么问题？'
		}
	];

	const siteShortcuts = [
		{ label: '留言', href: '/message', icon: MessageCircle },
		{ label: '关于', href: '/about', icon: UserRound },
		{ label: '友链', href: '/friends', icon: Link2 }
	];

	let { question, sessionId, greeting, turns, onQuestionChange, onTurnsChange, onNavigate } =
		$props<{
			question: string;
			sessionId: string;
			greeting: string;
			turns: RagTurn[];
			onQuestionChange: (value: string, sourceSessionId: string) => void;
			onTurnsChange: (value: RagTurn[], sourceSessionId: string) => void;
			onNavigate: () => void;
		}>();
	const activeSessionId = untrack(() => sessionId);
	let input: HTMLTextAreaElement | undefined = $state();
	let controller: AbortController | undefined;
	let pendingTurnId: string | undefined;
	let disposed = false;
	let historyCount = $state(0);
	let showHistory = $state(false);
	let visibleTurns = $derived(showHistory ? turns : turns.slice(historyCount));

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
			return askRag(value.question, activeSessionId, controller.signal, value.history);
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
				reason: '请求暂时未能完成，请稍后重试。',
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
				? '暂时无法连接问答服务，请稍后重试。'
				: ready
					? '可以与我交流，也可以提问本站文章与手记。'
					: availability.data?.reason === 'index_not_ready'
						? '公开内容正在准备中，请稍后重试。'
						: '问答服务尚未开放，请稍后重试。'
	);

	onMount(() => {
		historyCount = turns.length;
		input?.focus();
	});
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
		onTurnsChange(
			turns.map((turn: RagTurn) => (turn.id === id ? { ...turn, answer } : turn)),
			activeSessionId
		);
		pendingTurnId = undefined;
	}

	function send(selectedQuestion?: string) {
		const value = (selectedQuestion ?? question).trim();
		if (!ready || !value || mutation.isPending) return;
		const history = conversationHistory(turns, availability.data?.history);
		const id = crypto.randomUUID();
		pendingTurnId = id;
		onTurnsChange([...turns, { id, question: value, answer: null }], activeSessionId);
		onQuestionChange('', activeSessionId);
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

	function handleShortcut(event: MouseEvent) {
		if (event.button === 0 && !event.ctrlKey && !event.metaKey && !event.shiftKey && !event.altKey)
			onNavigate();
	}
</script>

<div class="flex shrink-0 items-center gap-2 px-3 text-xs text-ink-500 sm:px-5 dark:text-ink-400">
	<span class="h-px min-w-3 flex-1 bg-ink-300/70 dark:bg-ink-700" aria-hidden="true"></span>
	{#if historyCount > 0}
		<span>最近对话记录已{showHistory ? '展开' : '收起'}</span>
		<Button
			variant="ghost"
			type="button"
			aria-expanded={showHistory}
			aria-controls="rag-visible-conversation"
			onclick={() => {
				showHistory = !showHistory;
			}}
			class="min-h-8 p-0! text-xs! text-jade-800 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:text-jade-200 dark:focus-visible:outline-jade-400"
			>点击{showHistory ? '收起' : '展开'}</Button
		>
	{:else}
		<span>本次对话</span>
	{/if}
	<span class="h-px min-w-3 flex-1 bg-ink-300/70 dark:bg-ink-700" aria-hidden="true"></span>
</div>

<div class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 pt-5 pb-5 sm:px-5">
	{#if visibleTurns.length === 0}
		<div class="mb-5" aria-label="书灵的问候">
			<div class="mb-1 flex items-center gap-2">
				<RagAvatar class="size-7 shrink-0" />
				<p class="text-xs text-ink-600 dark:text-ink-300">书灵</p>
			</div>
			<p class="ml-3 border-l border-jade-600/40 pl-5 text-sm leading-7 dark:border-jade-400/40">
				{greeting}，我是书灵。有什么问题想问我？
			</p>
		</div>
		<section aria-labelledby="rag-suggestions" class="ml-4">
			<h2 id="rag-suggestions" class="mb-3 text-xs text-ink-600 dark:text-ink-400">猜你喜欢</h2>
			<ul class="grid w-fit max-w-full grid-cols-2 gap-x-3 gap-y-2">
				{#each suggestedQuestions as suggestion (suggestion.question)}
					<li class="max-w-full">
						<Button
							variant="ghost"
							type="button"
							disabled={!ready || mutation.isPending}
							onclick={() => send(suggestion.question)}
							aria-label={suggestion.question}
							title={suggestion.question}
							class="min-h-9 w-full rounded-full! border border-ink-300/80 px-4! text-xs! text-ink-800 hover:border-jade-600 hover:bg-jade-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:border-ink-600 dark:text-ink-200 dark:hover:border-jade-500 dark:hover:bg-ink-800 dark:focus-visible:outline-jade-400"
						>
							{suggestion.label}
						</Button>
					</li>
				{/each}
			</ul>
		</section>
	{/if}
	<div id="rag-visible-conversation">
		{#if visibleTurns.length > 0}<RagTranscript turns={visibleTurns} />{/if}
	</div>
	{#if !ready && !availability.isPending}
		<p class="mt-5 text-xs leading-6 text-ink-600 dark:text-ink-300" role="status">
			{availabilityText}
		</p>
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
</div>

<form
	onsubmit={submit}
	class="shrink-0 px-2 pt-3 pb-3 sm:px-3"
	style:padding-bottom="calc(var(--spacing) * 3 + env(safe-area-inset-bottom))"
>
	<nav aria-label="博客功能" class="mb-2 flex flex-wrap gap-2">
		{#each siteShortcuts as shortcut (shortcut.href)}
			<Button
				variant="ghost"
				size="sm"
				href={resolveHref(shortcut.href)}
				onclick={handleShortcut}
				class="min-h-8 gap-1.5! rounded-full! border border-ink-300/80 bg-white/60 px-3! text-xs! text-ink-600 hover:border-jade-500 hover:bg-jade-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:border-ink-700 dark:bg-ink-800/40 dark:text-ink-300 dark:hover:border-jade-600 dark:hover:bg-ink-800 dark:focus-visible:outline-jade-400"
			>
				<shortcut.icon class="size-3.5 text-jade-700 dark:text-jade-300" aria-hidden="true" />
				{shortcut.label}
			</Button>
		{/each}
	</nav>
	<label for="rag-question" class="sr-only">你的问题</label>
	<div
		class="rounded-xl border border-ink-300 bg-white/90 p-2.5 shadow-sm transition-colors focus-within:border-jade-700 dark:border-ink-600 dark:bg-ink-900/80 dark:focus-within:border-jade-400"
	>
		<div class="relative">
			<Textarea
				id="rag-question"
				bind:ref={input}
				value={question}
				oninput={() => onQuestionChange(input?.value ?? '', activeSessionId)}
				onkeydown={handleKeydown}
				rows={2}
				maxLength={1000}
				resize="none"
				disabled={mutation.isPending}
				aria-describedby="rag-availability rag-keyboard-help"
				placeholder="请问有什么问题？"
				textareaClass="block border-0! bg-transparent! py-1! pr-1! pl-7! text-sm leading-6 placeholder:text-ink-500! focus:ring-0! disabled:opacity-60 dark:placeholder:text-ink-400!"
			/>
			<MessageCircle
				class="pointer-events-none absolute top-2 left-1 size-4 text-ink-500 dark:text-ink-400"
				aria-hidden="true"
			/>
		</div>
		<div class="mt-1 flex items-center justify-between gap-3">
			<Button
				variant="ghost"
				size="sm"
				type="button"
				aria-label="清空草稿"
				title="清空草稿"
				disabled={!question || mutation.isPending}
				onclick={() => {
					onQuestionChange('', activeSessionId);
					input?.focus();
				}}
				class="size-10 rounded-full! p-0! focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
			>
				<Eraser class="size-4" aria-hidden="true" />
			</Button>
			<Button
				type="submit"
				loading={mutation.isPending}
				disabled={!ready || !question.trim() || mutation.isPending}
				aria-label={mutation.isPending ? '书灵正在回答' : '发送问题'}
				title={mutation.isPending ? '书灵正在回答' : '发送问题'}
				aria-describedby="rag-availability"
				class="size-10 shrink-0 rounded-full! bg-jade-600! p-0! hover:bg-jade-500! hover:text-ink-950 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:focus-visible:outline-jade-400"
			>
				{#if !mutation.isPending}<ArrowUp class="size-5" aria-hidden="true" />{/if}
			</Button>
		</div>
	</div>
	<p id="rag-availability" class="sr-only">{availabilityText}</p>
	<p id="rag-keyboard-help" class="sr-only">Enter 发送 · Shift+Enter 换行</p>
</form>
