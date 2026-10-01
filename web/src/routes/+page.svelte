<script lang="ts">
	import { resolve } from '$app/paths';
	import Hero from '$lib/features/home/Hero.svelte';
	import InspirationGrid from '$lib/features/home/InspirationGrid.svelte';
	import ActivityPulse from '$lib/features/home/ActivityPulse.svelte';
	import HomeMomentItem from '$lib/features/moment/components/HomeMomentItem.svelte';
	import { SlideIn, StaggerList } from '$lib/ui/animation';
	import { scrollFade } from '$lib/shared/actions/scroll-fade';
	import { ArrowRight } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	type HomeScene = {
		startHour: number;
		src: string;
		label: string;
	};

	const HOME_SCENES: readonly HomeScene[] = [
		{ startHour: 0, src: '/home-scenes/00.webp', label: '凌晨' },
		{ startHour: 6, src: '/home-scenes/06.webp', label: '清晨' },
		{ startHour: 8, src: '/home-scenes/08.webp', label: '上午' },
		{ startHour: 11, src: '/home-scenes/11.webp', label: '中午' },
		{ startHour: 14, src: '/home-scenes/14.webp', label: '下午' },
		{ startHour: 17, src: '/home-scenes/17.webp', label: '黄昏' },
		{ startHour: 20, src: '/home-scenes/20.webp', label: '夜间' },
		{ startHour: 23, src: '/home-scenes/23.webp', label: '深夜' }
	];

	const portraitSceneSrc = (scene: HomeScene) =>
		scene.startHour < 6 || scene.startHour >= 20
			? '/home-scenes/portrait-night_20261001.jpg'
			: '/home-scenes/portrait-day_20261001.jpg';

	const resolveSceneIndex = (hour: number) => {
		for (let index = HOME_SCENES.length - 1; index >= 0; index -= 1) {
			if (hour >= HOME_SCENES[index].startHour) return index;
		}
		return 0;
	};

	let currentScene = $state(HOME_SCENES[2]);
	let previousScene = $state<HomeScene | null>(null);

	onMount(() => {
		let transitionTimer: ReturnType<typeof setTimeout> | undefined;

		const syncScene = () => {
			const sceneIndex = resolveSceneIndex(new Date().getHours());
			const nextScene = HOME_SCENES[sceneIndex];

			if (nextScene.src !== currentScene.src) {
				previousScene = currentScene;
				currentScene = nextScene;
				clearTimeout(transitionTimer);
				transitionTimer = setTimeout(() => {
					previousScene = null;
				}, 900);
			}

			const preload = new Image();
			preload.src = HOME_SCENES[(sceneIndex + 1) % HOME_SCENES.length].src;
		};

		const handleVisibilityChange = () => {
			if (document.visibilityState === 'visible') syncScene();
		};

		syncScene();
		const interval = window.setInterval(syncScene, 60_000);
		window.addEventListener('focus', syncScene);
		document.addEventListener('visibilitychange', handleVisibilityChange);

		return () => {
			clearTimeout(transitionTimer);
			window.clearInterval(interval);
			window.removeEventListener('focus', syncScene);
			document.removeEventListener('visibilitychange', handleVisibilityChange);
		};
	});
</script>

<div class="homepage-container">
	<div class="home-backdrop" data-scene={currentScene.label} aria-hidden="true">
		{#if previousScene}
			<picture class="home-backdrop-image home-backdrop-previous">
				<source media="(max-aspect-ratio: 3/4)" srcset={portraitSceneSrc(previousScene)} />
				<img src={previousScene.src} alt="" decoding="async" />
			</picture>
		{/if}
		<picture class="home-backdrop-image home-backdrop-current">
			<source media="(max-aspect-ratio: 3/4)" srcset={portraitSceneSrc(currentScene)} />
			<img src={currentScene.src} alt="" decoding="async" />
		</picture>
	</div>

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

	.homepage-container::after {
		content: '';
		position: fixed;
		inset: 0;
		pointer-events: none;
	}

	.home-backdrop {
		position: fixed;
		inset: 0;
		z-index: -2;
		overflow: hidden;
		pointer-events: none;
	}

	.home-backdrop-image {
		position: absolute;
		inset: 0;
		display: block;
	}

	.home-backdrop-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: center top;
	}

	.home-backdrop-previous {
		z-index: 0;
	}

	.home-backdrop-current {
		z-index: 1;
		animation: scene-crossfade 900ms ease-out both;
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

	@keyframes scene-crossfade {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@media (max-width: 767px) {
		.home-backdrop-image img {
			object-position: 57% center;
		}

		.home-visual {
			min-height: calc(100svh - 5rem);
		}

		.home-hero {
			padding: 0 1rem;
		}
	}

	@media (max-width: 1199px), (max-aspect-ratio: 3/2) {
		.home-visual {
			min-height: max(100svh, 500px);
		}

		.home-hero {
			width: calc(100% - 114px);
			margin: 0 114px 0 0;
			padding: 0 0.75rem;
		}
	}

	@media (max-aspect-ratio: 3/4) {
		.home-backdrop-image img {
			object-position: center;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.home-backdrop-current {
			animation: none;
		}
	}
</style>
