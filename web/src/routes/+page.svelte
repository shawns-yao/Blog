<script lang="ts">
	import { resolve } from '$app/paths';
	import Hero from '$lib/features/home/Hero.svelte';
	import InspirationGrid from '$lib/features/home/InspirationGrid.svelte';
	import ActivityPulse from '$lib/features/home/ActivityPulse.svelte';
	import HomeMomentItem from '$lib/features/moment/components/HomeMomentItem.svelte';
	import { SlideIn, StaggerList } from '$lib/ui/animation';
	import { ArrowRight } from 'lucide-svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<div class="homepage-container">
	<div class="home-visual">
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
		</section>

		<!-- New Inspiration Grid -->
		<InspirationGrid config={data.homeTheme?.inspiration} stats={data.inspirationStats} />

		<!-- New Activity Pulse -->
		<ActivityPulse pulse={data.activityPulse} config={data.homeTheme?.activityPulse} />
	</div>
</div>

<style lang="postcss">
	@reference "./layout.css";

	.homepage-container {
		position: relative;
		background: #f4eee3;
	}

	.home-visual {
		position: relative;
		min-height: 100svh;
		background:
			linear-gradient(
				90deg,
				rgb(247 240 228 / 0.8) 0%,
				rgb(247 240 228 / 0.56) 30%,
				transparent 66%
			),
			linear-gradient(0deg, rgb(38 27 18 / 0.2), transparent 35%),
			url('/bg.png') center / cover no-repeat;
		box-shadow: inset 0 -34px 48px rgb(45 30 19 / 0.15);
	}

	.home-hero {
		position: relative;
		z-index: 1;
		width: min(1200px, 100%);
		margin: 0 auto;
		padding: 0 1.5rem;
	}

	.home-content {
		position: relative;
		background: rgb(255 252 246 / 0.94);
		box-shadow: 0 -18px 48px rgb(61 42 26 / 0.13);
	}

	:global(.dark) .homepage-container {
		background: #211d19;
	}

	:global(.dark) .home-visual {
		background:
			linear-gradient(90deg, rgb(29 25 21 / 0.82) 0%, rgb(29 25 21 / 0.54) 30%, transparent 68%),
			linear-gradient(0deg, rgb(17 13 10 / 0.36), transparent 38%),
			url('/bg.png') center / cover no-repeat;
	}

	:global(.dark) .home-content {
		background: rgb(31 27 23 / 0.95);
	}

	@media (max-width: 767px) {
		.home-visual {
			min-height: calc(100svh - 5rem);
			background-position: 57% center;
		}

		.home-hero {
			padding: 0 1rem;
			background: linear-gradient(180deg, rgb(248 242 231 / 0.34), rgb(248 242 231 / 0.62));
		}
	}
</style>
