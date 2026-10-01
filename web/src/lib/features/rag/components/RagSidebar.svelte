<script lang="ts">
	import { onMount } from 'svelte';
	import { Dialog } from 'bits-ui';
	import { ChevronsRight, MessageCirclePlus } from 'lucide-svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import type { RagTurn } from '../types';
	import RagAvatar from './RagAvatar.svelte';

	let open = $state(false);
	let question = $state('');
	let sessionId = $state('');
	let turns = $state<RagTurn[]>([]);
	let greeting = $state('你好');
	let searchAfterClose = false;
	// 销毁旧组件时可能读取旧状态，使用当前标识隔离中止回调。
	let currentSessionId = '';
	onMount(() => {
		currentSessionId = crypto.randomUUID();
		sessionId = currentSessionId;
		updateGreeting();
	});

	function updateGreeting() {
		const hour = new Date().getHours();
		greeting = hour < 6 ? '你好' : hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好';
	}

	function startConversation() {
		currentSessionId = crypto.randomUUID();
		sessionId = currentSessionId;
		question = '';
		turns = [];
		updateGreeting();
	}

	function updateTurns(value: RagTurn[], sourceSessionId: string) {
		if (sourceSessionId === currentSessionId) turns = value;
	}

	function updateQuestion(value: string, sourceSessionId: string) {
		if (sourceSessionId === currentSessionId) question = value;
	}

	$effect(() => {
		if (uiState.isSearchOpen) open = false;
	});

	function useSearch() {
		searchAfterClose = true;
		open = false;
	}

	function handleOpenChangeComplete(isOpen: boolean) {
		if (isOpen) updateGreeting();
		if (!isOpen && searchAfterClose) {
			searchAfterClose = false;
			uiState.openSearch();
		}
	}
</script>

<Dialog.Root bind:open onOpenChangeComplete={handleOpenChangeComplete}>
	<Dialog.Trigger
		disabled={!sessionId}
		aria-label="打开站内问答"
		title="问问书灵"
		class="shuling-entry group fixed right-4 bottom-6 z-50 flex size-20 flex-col items-center justify-center rounded-2xl bg-transparent focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jade-700 sm:right-7 sm:size-24 dark:focus-visible:outline-jade-400"
		style="bottom: calc(var(--spacing) * 6 + env(safe-area-inset-bottom))"
	>
		<span
			aria-hidden="true"
			class="pointer-events-none absolute right-0 -top-9 whitespace-nowrap rounded-full border border-ink-300 bg-ink-50 px-3 py-1.5 text-xs text-ink-700 opacity-0 shadow-sm transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100 dark:border-ink-700 dark:bg-ink-900 dark:text-ink-200"
			>问问书灵</span
		>
		<span
			class="shuling-pose relative z-10 transition-transform duration-200 group-hover:-rotate-6 group-hover:scale-105 group-active:scale-95 motion-reduce:transform-none"
		>
			<span class="shuling-float block"><RagAvatar class="size-20 sm:size-24" /></span>
		</span>
		<span
			class="shuling-shadow absolute bottom-0 h-1.5 w-9 rounded-full bg-ink-800/15 blur-sm dark:bg-ink-400/15"
			aria-hidden="true"
		></span>
	</Dialog.Trigger>
	<Dialog.Portal>
		<Dialog.Overlay
			class="fixed inset-0 z-(--z-index-rag-overlay) bg-ink-950/20 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 duration-150 motion-reduce:animate-none dark:bg-ink-950/50"
		/>
		<Dialog.Content
			class="fixed inset-y-0 right-0 z-(--z-index-rag-panel) flex h-dvh w-full max-w-rag flex-col border-l border-ink-200 bg-linear-to-b from-ink-100 via-ink-50 to-white text-ink-900 shadow-deep outline-none data-[state=open]:animate-in data-[state=open]:slide-in-from-right data-[state=closed]:animate-out data-[state=closed]:slide-out-to-right duration-150 motion-reduce:animate-none dark:border-ink-700 dark:from-ink-900 dark:via-ink-900 dark:to-ink-950 dark:text-ink-100"
			onCloseAutoFocus={(event) => {
				if (searchAfterClose || uiState.isSearchOpen) event.preventDefault();
			}}
		>
			<header
				class="relative flex shrink-0 items-center gap-3 px-3 pt-6 pb-4 sm:px-5"
				style:padding-top="calc(var(--spacing) * 6 + env(safe-area-inset-top))"
			>
				<div class="flex min-w-0 items-center gap-3">
					<RagAvatar class="size-20 shrink-0" animated />
					<div class="min-w-0 pt-4">
						<Dialog.Title class="font-serif text-base font-medium"
							>{greeting}，我是书灵</Dialog.Title
						>
						<Dialog.Description class="mt-1.5 text-xs leading-6 text-ink-600 dark:text-ink-300">
							关于站内知识的问题，都可以问我。
						</Dialog.Description>
					</div>
				</div>
				<div class="absolute top-3 right-2 flex items-center gap-1 sm:right-3">
					<Button
						variant="ghost"
						type="button"
						aria-label="新建对话"
						title="新建对话"
						onclick={startConversation}
						class="size-9 rounded-full! border border-ink-300/80 p-0! text-ink-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:border-ink-700 dark:text-ink-300 dark:focus-visible:outline-jade-400"
					>
						<MessageCirclePlus class="size-4" aria-hidden="true" />
					</Button>
					<Dialog.Close
						aria-label="关闭站内问答"
						title="关闭站内问答"
						class="flex size-9 items-center justify-center rounded-full text-ink-600 hover:bg-ink-200/70 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:text-ink-300 dark:hover:bg-ink-800 dark:focus-visible:outline-jade-400"
					>
						<ChevronsRight class="size-5" aria-hidden="true" />
					</Dialog.Close>
				</div>
			</header>

			{#key sessionId}
				<QueryRoot
					loader={() => import('./RagChatClient.svelte')}
					loaderProps={{
						question,
						sessionId,
						greeting,
						turns,
						onTurnsChange: updateTurns,
						onQuestionChange: updateQuestion,
						onSearch: useSearch
					}}
				>
					{#snippet fallback()}
						<div class="min-h-0 flex-1 overflow-y-auto px-4 py-5 sm:px-5">
							<p class="text-xs text-ink-600 dark:text-ink-300" role="status">正在加载问答…</p>
						</div>
					{/snippet}
				</QueryRoot>
			{/key}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	.shuling-float {
		animation: shuling-float 3.6s ease-in-out infinite;
	}
	.shuling-shadow {
		animation: shuling-shadow 3.6s ease-in-out infinite;
	}
	:global(.shuling-entry):hover .shuling-float,
	:global(.shuling-entry):focus-visible .shuling-float {
		animation-play-state: paused;
	}
	@keyframes shuling-float {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-6px);
		}
	}
	@keyframes shuling-shadow {
		0%,
		100% {
			transform: scaleX(1);
			opacity: 1;
		}
		50% {
			transform: scaleX(0.75);
			opacity: 0.6;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.shuling-float,
		.shuling-shadow {
			animation: none;
		}
		.shuling-pose {
			transition: none;
			transform: none;
		}
	}
</style>
