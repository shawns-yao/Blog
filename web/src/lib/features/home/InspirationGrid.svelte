<script lang="ts">
	import { ArrowUpRight } from 'lucide-svelte';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import type { HomeInspirationThemeConfig } from './types';

	let { config }: { config?: HomeInspirationThemeConfig } = $props();
	const items = $derived((config?.now?.items ?? []).slice(0, 3));
	const quote = $derived(config?.quote);
	const work = $derived(config?.work);
	const updatedAt = $derived(
		config?.now?.updatedAt && Number.isFinite(Date.parse(config.now.updatedAt))
			? new Date(config.now.updatedAt).toISOString().slice(0, 10)
			: ''
	);
	const href = (url: string) => (url.startsWith('/') ? resolvePath(url) : url);
</script>

<section class="mt-16 md:mt-24" aria-labelledby="inspiration-title">
	<h2 id="inspiration-title" class="section-heading">
		{config?.sectionTitle || '灵感与实验场'}
	</h2>
	{#if quote?.text || items.length || work}
		<div class="inspiration-layout" class:has-work={!!work && (!!quote?.text || items.length > 0)}>
			{#if quote?.text || items.length}
				<div class="space-y-10 min-w-0">
					{#if quote?.text}
						<div>
							<h3 class="text-sm font-medium text-jade-700 dark:text-jade-300 mb-4">灵感一则</h3>
							<blockquote class="m-0">
								<p
									class="font-serif text-xl leading-loose text-ink-900 dark:text-ink-100 break-words"
								>
									{quote.text}
								</p>
								{#if quote.author}
									<footer class="mt-3 text-sm text-ink-600 dark:text-ink-300">
										{quote.author}
									</footer>
								{/if}
							</blockquote>
							{#if quote.href}
								<a class="content-link mt-4 inline-flex items-center gap-1" href={href(quote.href)}>
									相关手记 <ArrowUpRight size={14} />
								</a>
							{/if}
						</div>
					{/if}
					{#if items.length}
						<div>
							<div class="flex flex-wrap items-baseline justify-between gap-2 mb-4">
								<h3 class="text-sm font-medium text-jade-700 dark:text-jade-300">最近在做</h3>
								{#if updatedAt}
									<time datetime={updatedAt} class="text-xs text-ink-600 dark:text-ink-300"
										>{updatedAt} 更新</time
									>
								{/if}
							</div>
							<dl class="space-y-4">
								{#each items as item, index (`${item.id}-${index}`)}
									<div class="grid grid-cols-[5rem_minmax(0,1fr)] gap-4 text-sm leading-relaxed">
										<dt class="text-ink-600 dark:text-ink-300 break-words">{item.label}</dt>
										<dd class="text-ink-900 dark:text-ink-100 break-words">{item.value}</dd>
									</div>
								{/each}
							</dl>
						</div>
					{/if}
				</div>
			{/if}
			{#if work}
				<article class="min-w-0">
					<h3 class="text-sm font-medium text-jade-700 dark:text-jade-300 mb-4">小作品 / 实验</h3>
					<a class="work-link block" href={href(work.href)}>
						<img
							src={href(work.image)}
							alt={work.title}
							loading="lazy"
							class="aspect-[16/10] w-full object-cover rounded-sm"
						/>
						<div class="flex items-start justify-between gap-4 mt-4">
							<h4 class="font-serif text-lg text-ink-900 dark:text-ink-100 break-words">
								{work.title}
							</h4>
							<ArrowUpRight size={18} class="shrink-0 text-jade-700 dark:text-jade-300" />
						</div>
						<p class="mt-2 text-sm leading-relaxed text-ink-700 dark:text-ink-200 break-words">
							{work.description}
						</p>
					</a>
				</article>
			{/if}
		</div>
	{:else}
		<p class="py-6 text-sm text-ink-700 dark:text-ink-200">最近的灵感与近况还未记录。</p>
	{/if}
</section>

<style lang="postcss">
	@reference "$routes/layout.css";
	.section-heading {
		@apply mb-8 flex items-center gap-3 border-b border-ink-300/50 pb-4 font-serif text-xl font-medium text-ink-900 dark:border-ink-600/50 dark:text-ink-100;
	}
	.section-heading::before {
		content: '';
		@apply h-px w-8 bg-jade-500/60;
	}
	.inspiration-layout {
		display: grid;
		gap: 2.5rem;
	}
	.content-link {
		@apply text-sm text-jade-700 dark:text-jade-300;
	}
	a:hover {
		text-decoration: underline;
		text-underline-offset: 4px;
	}
	a:focus-visible {
		outline: 2px solid var(--color-jade-500);
		outline-offset: 5px;
	}
	@media (min-width: 768px) {
		.has-work {
			grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
			gap: 3.5rem;
		}
	}
</style>
