<script lang="ts">
	import { page } from '$app/state';
	import { resolveHomeThemeConfig } from '$lib/features/home/theme';

	const techStack = $derived(resolveHomeThemeConfig(page.data.websiteInfo).inspiration?.techStack);
</script>

<div class="mx-auto max-w-2xl py-24">
	<h1 class="font-serif text-2xl font-medium text-ink-900 dark:text-ink-100">关于</h1>
	<p class="mt-4 text-sm leading-relaxed text-ink-500 dark:text-ink-400">
		此页面尚未实现，内容将在后续补充。
	</p>
	{#if techStack?.items?.length}
		<section class="mt-10">
			<h2 class="font-serif text-lg text-ink-900 dark:text-ink-100">技术栈</h2>
			<ul class="mt-4 flex flex-wrap gap-x-6 gap-y-3 text-sm text-ink-700 dark:text-ink-200">
				{#each techStack.items as item, index (`${item}-${index}`)}
					<li>{item}</li>
				{/each}
			</ul>
		</section>
	{/if}
</div>
