<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { resolveFooterThemeConfig } from '$lib/features/footer/theme';
	import { websiteInfoCtx } from '$lib/features/website-info/context';

	type Props = {
		imageBackground?: boolean;
		onlineCount?: number;
		presenceConnected?: boolean;
		onOpenPresence?: () => void;
	};

	let {
		imageBackground = false,
		onlineCount = 0,
		presenceConnected = false,
		onOpenPresence = () => {}
	}: Props = $props();

	const footerThemeStore = websiteInfoCtx.selectModelData((data) => resolveFooterThemeConfig(data));
	const desktopColumnCount = $derived(Math.min($footerThemeStore.sections.length, 6));
	const desktopGridStyle = $derived.by(() => {
		const sectionCols = Math.max(desktopColumnCount, 1);
		return `repeat(${sectionCols}, minmax(0, 1fr)) minmax(220px, 1.2fr)`;
	});
	const brandFullRow = $derived($footerThemeStore.sections.length + 1 > 6);

	const formatPresenceText = (template: string, count: number): string =>
		template.replaceAll('{count}', String(count));
</script>

<footer
	class:image-background={imageBackground}
	class:presence-connected={presenceConnected}
	class="relative z-1 isolate mt-32 bg-transparent"
>
	<div class="max-w-[1200px] mx-auto px-6 py-12 md:py-16">
		<!-- Mobile Compact Layout (Hidden on Desktop) -->
		<div class="flex flex-col gap-4 md:hidden">
			{#each $footerThemeStore.sections as section (section.title)}
				<div class="flex flex-col gap-2">
					<div
						class="text-sm font-serif font-bold text-ink-900 dark:text-ink-100 flex items-center justify-between"
					>
						{section.title}
						<span class="text-ink-300 dark:text-ink-700 font-mono font-normal">></span>
					</div>
					<div class="flex flex-wrap gap-x-4 gap-y-2">
						{#each section.links as link (link.name)}
							<a
								href={/^(https?:|mailto:)/i.test(link.href) ? link.href : resolvePath(link.href)}
								class="text-sm text-ink-500 hover:text-jade-600 dark:hover:text-jade-400 transition-colors"
							>
								{link.name}
							</a>
						{/each}
					</div>
				</div>
			{/each}

			<!-- Brand Info below mobile footer links -->
			<div class="flex flex-col gap-4 pt-4 border-t border-ink-100 dark:border-ink-800/50">
				<div class="flex flex-col">
					<div class="text-lg font-mono font-bold text-ink-900 dark:text-ink-100">
						{$footerThemeStore.brandName}
					</div>
					<p class="text-[11px] font-mono text-ink-400 mt-1 uppercase tracking-wider">
						{$footerThemeStore.brandTagline}
					</p>
				</div>
				<button
					onclick={onOpenPresence}
					class="flex items-center gap-2 w-fit transition-colors underline-offset-2 hover:underline focus-visible:underline focus-visible:outline-none {presenceConnected
						? 'text-jade-700/80 dark:text-jade-400/80'
						: 'text-red-600 dark:text-red-400'}"
				>
					<span class="relative flex h-1.5 w-1.5">
						<span
							class="absolute inline-flex h-full w-full rounded-full opacity-75 {presenceConnected
								? 'bg-jade-400'
								: 'bg-red-400'}"
						></span>
						<span
							class="relative inline-flex rounded-full h-1.5 w-1.5 {presenceConnected
								? 'bg-jade-500'
								: 'bg-red-500'}"
						></span>
					</span>
					<span class="text-[10px] font-mono">
						{#if presenceConnected}
							{formatPresenceText($footerThemeStore.presenceConnectedText, onlineCount)}
						{:else}
							{$footerThemeStore.presenceLoadingText}
						{/if}
					</span>
				</button>
			</div>
		</div>

		<!-- Desktop Multi-column Layout (Hidden on Mobile) -->
		<div class="hidden md:grid gap-8 lg:gap-12" style:grid-template-columns={desktopGridStyle}>
			{#each $footerThemeStore.sections as section (section.title)}
				<div class="flex flex-col gap-6">
					<h3
						class="text-sm font-serif font-bold text-ink-900 dark:text-ink-100 flex items-center gap-2"
					>
						<span class="w-1 h-3 bg-jade-500 rounded-full"></span>
						{section.title}
					</h3>
					<ul class="flex flex-col gap-3">
						{#each section.links as link (link.name)}
							<li>
								<a
									href={/^(https?:|mailto:)/i.test(link.href) ? link.href : resolvePath(link.href)}
									class="text-sm text-ink-500 hover:text-jade-600 dark:hover:text-jade-400 transition-colors"
								>
									{link.name}
								</a>
							</li>
						{/each}
					</ul>
				</div>
			{/each}

			<!-- Brand Info inside Desktop Grid -->
			<div
				class="flex flex-col gap-6 items-end text-right"
				style:grid-column={brandFullRow ? '1 / -1' : undefined}
			>
				<div class="flex flex-col items-end">
					<div class="text-xl font-mono font-bold text-ink-900 dark:text-ink-100">
						{$footerThemeStore.brandName}
					</div>
					<p class="text-[11px] font-mono text-ink-400 mt-1 uppercase tracking-wider">
						{$footerThemeStore.brandTagline}
					</p>
				</div>
				<button
					onclick={onOpenPresence}
					class="flex items-center gap-2 w-fit transition-colors underline-offset-2 hover:underline focus-visible:underline focus-visible:outline-none {presenceConnected
						? 'text-jade-700/80 dark:text-jade-400/80'
						: 'text-red-600 dark:text-red-400'}"
				>
					<span class="relative flex h-1.5 w-1.5">
						<span
							class="absolute inline-flex h-full w-full rounded-full opacity-75 {presenceConnected
								? 'bg-jade-400'
								: 'bg-red-400'}"
						></span>
						<span
							class="relative inline-flex rounded-full h-1.5 w-1.5 {presenceConnected
								? 'bg-jade-500'
								: 'bg-red-500'}"
						></span>
					</span>
					<span class="text-[10px] font-mono">
						{#if presenceConnected}
							{formatPresenceText($footerThemeStore.presenceConnectedText, onlineCount)}
						{:else}
							{$footerThemeStore.presenceLoadingText}
						{/if}
					</span>
				</button>
			</div>
		</div>
	</div>
</footer>

<style lang="postcss">
	@reference "$routes/layout.css";

	.image-background {
		--footer-text-shadow:
			-1px 0 1px rgb(15 20 17 / 0.9), 1px 0 1px rgb(15 20 17 / 0.9), 0 -1px 1px rgb(15 20 17 / 0.9),
			0 1px 1px rgb(15 20 17 / 0.9), 0 2px 4px rgb(15 20 17 / 0.65);
		text-shadow: var(--footer-text-shadow);
	}

	.image-background .font-bold {
		color: #fffaf0;
	}

	.image-background a,
	.image-background p,
	.image-background .font-normal {
		color: #f1eee6;
	}

	.image-background a {
		font-weight: 500;
	}

	.image-background p {
		font-size: 12px;
		line-height: 1.75;
		letter-spacing: 0;
	}

	.image-background button {
		color: #ffc6c6;
	}

	.image-background.presence-connected button {
		color: #b9f1df;
	}

	.image-background button > span:last-child {
		font-size: 12px;
		line-height: 1.5;
	}

	.image-background a:hover,
	.image-background a:focus-visible,
	.image-background button:hover {
		color: #b9f1df;
		text-decoration: underline;
		text-underline-offset: 4px;
	}

	.image-background a:focus-visible,
	.image-background button:focus-visible {
		outline: 2px solid #fffaf0;
		outline-offset: 4px;
	}
</style>
