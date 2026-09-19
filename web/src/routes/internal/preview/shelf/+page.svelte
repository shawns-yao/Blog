<script lang="ts">
	import ShelfNav from '$lib/ui/layout/shelf/ShelfNav.svelte';
	import { SHELF_BOOKS } from '$lib/shared/nav/nav-items';
</script>

<div class="mx-auto max-w-4xl py-10">
	<h1 class="text-xl font-medium text-ink-900 dark:text-ink-100">书架导航预览</h1>
	<p class="mt-2 text-sm text-ink-500 dark:text-ink-400">
		静止时东倒西歪，悬停任意一本书会扶正并抬起
	</p>

	<div class="mt-8 rounded-default border border-ink-200 bg-[#FBF9F5] dark:border-ink-800 dark:bg-ink-900">
		<ShelfNav books={SHELF_BOOKS} />
	</div>

	<div class="mt-8 rounded-default border border-ink-200 p-6 dark:border-ink-800">
		<p class="text-sm text-ink-500 dark:text-ink-400">
			下方为内容区示意。正式接入后，书架会固定在页面顶部，替代现有左侧栏。
		</p>
	</div>
</div>
