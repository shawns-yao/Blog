<script lang="ts">
	import { onMount } from 'svelte';
	import { Dialog } from 'bits-ui';
	import { MessageSquare, X } from 'lucide-svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import type { RagTurn } from '../types';

	let open = $state(false);
	let question = $state('');
	let sessionId = $state('');
	let turns = $state<RagTurn[]>([]);
	let searchAfterClose = false;
	onMount(() => {
		sessionId = crypto.randomUUID();
	});

	$effect(() => {
		if (uiState.isSearchOpen) open = false;
	});

	function useSearch() {
		searchAfterClose = true;
		open = false;
	}

	function handleOpenChangeComplete(isOpen: boolean) {
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
		title="站内问答"
		class="fixed right-0 top-1/2 z-50 flex -translate-y-1/2 flex-col items-center gap-2 rounded-l-default border border-r-0 border-ink-300 bg-ink-50 px-3 py-4 font-serif text-sm text-ink-700 shadow-float transition-colors hover:bg-ink-100 hover:text-jade-800 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jade-700 dark:border-ink-700 dark:bg-ink-900 dark:text-ink-200 dark:hover:bg-ink-800 dark:hover:text-jade-200 dark:focus-visible:outline-jade-400"
	>
		<MessageSquare class="size-5" aria-hidden="true" />
		<span>问答</span>
	</Dialog.Trigger>
	<Dialog.Portal>
		<Dialog.Overlay
			class="fixed inset-0 z-(--z-index-rag-overlay) bg-ink-950/20 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 duration-150 motion-reduce:animate-none dark:bg-ink-950/50"
		/>
		<Dialog.Content
			class="fixed inset-y-0 right-0 z-(--z-index-rag-panel) flex h-dvh w-full max-w-rag flex-col border-l border-ink-200 bg-ink-50 text-ink-900 shadow-deep outline-none data-[state=open]:animate-in data-[state=open]:slide-in-from-right data-[state=closed]:animate-out data-[state=closed]:slide-out-to-right duration-150 motion-reduce:animate-none dark:border-ink-700 dark:bg-ink-900 dark:text-ink-100"
			onCloseAutoFocus={(event) => {
				if (searchAfterClose || uiState.isSearchOpen) event.preventDefault();
			}}
		>
			<header
				class="flex shrink-0 items-start justify-between gap-4 border-b border-ink-200 p-5 dark:border-ink-700"
				style:padding-top="calc(var(--spacing) * 5 + env(safe-area-inset-top))"
			>
				<div>
					<Dialog.Title class="font-serif text-xl font-medium">站内问答</Dialog.Title>
					<Dialog.Description class="mt-2 text-sm text-ink-600 dark:text-ink-300">
						与你交流，查找本站文章与手记
					</Dialog.Description>
				</div>
				<Dialog.Close
					aria-label="关闭站内问答"
					title="关闭站内问答"
					class="flex size-10 shrink-0 items-center justify-center rounded-default text-ink-600 hover:bg-ink-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-jade-700 dark:text-ink-300 dark:hover:bg-ink-800 dark:focus-visible:outline-jade-400"
				>
					<X class="size-5" aria-hidden="true" />
				</Dialog.Close>
			</header>

			<QueryRoot
				loader={() => import('./RagChatClient.svelte')}
				loaderProps={{
					question,
					sessionId,
					turns,
					onTurnsChange: (value: RagTurn[]) => {
						turns = value;
					},
					onQuestionChange: (value: string) => {
						question = value;
					},
					onSearch: useSearch
				}}
			>
				{#snippet fallback()}
					<p class="flex-1 p-6 text-sm text-ink-600 dark:text-ink-300" role="status">
						正在加载问答…
					</p>
				{/snippet}
			</QueryRoot>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
