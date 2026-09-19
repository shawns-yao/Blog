<script lang="ts">
	import { uiState } from '$lib/shared/stores/ui.svelte';
</script>

<div class="mx-auto max-w-2xl py-24 text-center">
	<h1 class="font-serif text-2xl font-medium text-ink-900 dark:text-ink-100">搜索</h1>
	<p class="mt-4 text-sm leading-relaxed text-ink-500 dark:text-ink-400">
		独立搜索页尚未实现。当前可先通过搜索面板检索手记内容。
	</p>
	<button
		type="button"
		onclick={() => uiState.openSearch()}
		class="mt-8 rounded-default border border-ink-300 px-5 py-2 text-sm text-ink-700 transition-colors hover:bg-ink-100 dark:border-ink-700 dark:text-ink-300 dark:hover:bg-ink-800"
	>
		打开搜索面板
	</button>
</div>
