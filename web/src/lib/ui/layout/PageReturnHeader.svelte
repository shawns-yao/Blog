<script lang="ts">
	import { ArrowLeft } from 'lucide-svelte';
	import { resolveHref } from '$lib/shared/utils/resolve-path';

	let { showReadingMotto = false }: { showReadingMotto?: boolean } = $props();
</script>

<header class="flex items-start justify-between gap-6 px-5 py-5 sm:px-8 sm:py-6 lg:px-12">
	<a
		href={resolveHref('/')}
		class="inline-flex min-h-11 items-center gap-3 rounded-default font-serif text-sm tracking-wider text-ink-600 transition-colors hover:text-jade-700 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jade-600 dark:text-ink-300 dark:hover:text-jade-400"
	>
		<ArrowLeft size={18} strokeWidth={1.4} aria-hidden="true" />
		<span>返回首页</span>
	</a>
	{#if showReadingMotto}
		<p
			class="hidden pt-1 font-serif text-xs italic leading-relaxed text-ink-500 sm:block dark:text-ink-400"
		>
			Good Books<br />Build a Softer You.
			<span class="mt-3 block h-px w-5 bg-ink-300 dark:bg-ink-600"></span>
		</p>
	{/if}
</header>
