<script lang="ts">
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import RagAvatar from '$lib/features/rag/components/RagAvatar.svelte';
</script>

<svelte:head><title>留言板</title></svelte:head>

<div class="mx-auto w-full max-w-3xl px-6 pt-20 pb-24 sm:px-8 sm:pt-28">
	<header class="flex items-center gap-4 border-b border-ink-200 pb-8 dark:border-ink-700">
		<RagAvatar class="size-20 shrink-0" animated />
		<div>
			<h1 class="font-serif text-2xl text-ink-900 dark:text-ink-100">留言板</h1>
			<p class="mt-3 text-sm leading-7 text-ink-600 dark:text-ink-300">
				我是书灵。读后感、问题，或者一句招呼，都可以留在这里。
			</p>
		</div>
	</header>
	<QueryRoot loader={() => import('$lib/features/comment/components/GuestbookClient.svelte')}>
		{#snippet fallback()}<p class="py-12 text-sm text-ink-500" role="status">
				正在打开留言板…
			</p>{/snippet}
	</QueryRoot>
</div>
