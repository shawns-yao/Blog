<script lang="ts">
	import { resolve } from '$app/paths';
	import Hero from '$lib/features/home/Hero.svelte';
	import InspirationGrid from '$lib/features/home/InspirationGrid.svelte';
	import ActivityPulse from '$lib/features/home/ActivityPulse.svelte';
	import HomeMomentItem from '$lib/features/moment/components/HomeMomentItem.svelte';
	import { SlideIn, StaggerList } from '$lib/ui/animation';
	import { scrollFade } from '$lib/shared/actions/scroll-fade';
	import { ArrowRight } from 'lucide-svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<div class="homepage-container">
	<div class="home-visual" use:scrollFade={{ distance: 420 }}>
		<div class="home-hero">
			<Hero config={data.homeTheme?.hero} />
		</div>
	</div>

	<div class="home-content max-w-300 mx-auto px-6 py-12 md:py-20">
		<!-- Recent Moments -->
		<section>
			<SlideIn direction="left">
				<div
					class="flex items-center justify-between mb-6 border-b border-ink-100 dark:border-ink-800 pb-4"
				>
					<div class="flex items-center gap-3">
						<span class="h-px w-8 bg-jade-500/40"></span>
						<h2 class="text-xl font-serif font-medium text-ink-900 dark:text-ink-100">最近手记</h2>
					</div>
					<a
						href={resolve('/moments', {})}
						class="flex items-center gap-1 text-xs font-mono text-ink-400 hover:text-jade-600 dark:hover:text-jade-400 transition-colors group"
					>
						<span>查看全部</span>
						<ArrowRight size={12} class="group-hover:translate-x-1 transition-transform" />
					</a>
				</div>
			</SlideIn>

			<StaggerList staggerDelay={100} y={16} class="flex flex-col">
				{#each data.recentMoments.items as moment (moment.id)}
					<HomeMomentItem {moment} />
				{/each}
			</StaggerList>
			{#if data.recentMomentsFailed}
				<p role="status" class="py-6 text-sm text-ink-700 dark:text-ink-200">
					手记暂时加载失败，请稍后重试。
				</p>
			{:else if data.recentMoments.items.length === 0}
				<p class="py-6 text-sm text-ink-700 dark:text-ink-200">还没有公开的手记。</p>
			{/if}
		</section>

		<!-- New Inspiration Grid -->
		<InspirationGrid config={data.homeTheme?.inspiration} />

		<!-- New Activity Pulse -->
		<ActivityPulse
			pulse={data.activityPulse}
			failed={data.activityPulseFailed}
			config={data.homeTheme?.activityPulse}
		/>
	</div>
</div>

<style lang="postcss">
	@reference "./layout.css";

	.homepage-container {
		position: relative;
		isolation: isolate;
		background: transparent;
	}

	.homepage-container::before,
	.homepage-container::after {
		content: '';
		position: fixed;
		inset: 0;
		pointer-events: none;
	}

	.homepage-container::before {
		z-index: -2;
		background: url('/bg.png') center top / cover no-repeat;
	}

	:global(.dark) .homepage-container::before {
		background-image: url('/bg-night.png');
	}

	.homepage-container::after {
		z-index: -1;
		background: linear-gradient(180deg, rgb(65 44 28 / 0.03) 0%, rgb(65 44 28 / 0.14) 100%);
	}

	.home-visual {
		--scroll-fade-progress: 0;
		position: relative;
		min-height: 100svh;
		background: transparent;
	}

	.home-hero {
		position: relative;
		z-index: 1;
		width: min(1200px, 100%);
		margin: 0 auto;
		padding: 0 1.5rem;
		opacity: calc(1 - var(--scroll-fade-progress));
		transform: translateY(calc(var(--scroll-fade-progress) * -1.5rem));
		transition:
			opacity 120ms linear,
			transform 120ms linear;
	}

	.home-content {
		position: relative;
		isolation: isolate;
	}

	.home-content::before {
		content: '';
		position: absolute;
		inset: 0 calc(50% - 50vw);
		z-index: -1;
		pointer-events: none;
		background: rgb(255 255 255 / 0.42);
		-webkit-backdrop-filter: blur(8px);
		backdrop-filter: blur(8px);
		mask-image: linear-gradient(
			to bottom,
			transparent,
			black 32px,
			black calc(100% - 32px),
			transparent
		);
	}

	:global(.dark) .home-content::before {
		background: rgb(20 23 22 / 0.42);
	}

	:global(.dark) .homepage-container::after {
		background:
			linear-gradient(90deg, rgb(29 25 21 / 0.44) 0%, rgb(29 25 21 / 0.2) 36%, transparent 72%),
			linear-gradient(180deg, transparent 0%, rgb(17 13 10 / 0.14) 72%, rgb(17 13 10 / 0.38) 100%);
	}

	@media (max-width: 767px) {
		.homepage-container::before {
			background-position: 57% center;
		}

		.home-visual {
			min-height: calc(100svh - 5rem);
		}

		.home-hero {
			padding: 0 1rem;
		}
	}
</style>
