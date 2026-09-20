<script lang="ts">
	import { onDestroy } from 'svelte';
	import { resolveTheme, themeManager, type Theme } from '$lib/shared/theme/theme.svelte';
	import { Moon, Sun } from 'lucide-svelte';
	import AlarmClock from './AlarmClock.svelte';

	let { compact = false }: { compact?: boolean } = $props();
	const theme = themeManager;
	const resolved = $derived.by(() => resolveTheme(theme.current));
	let clockElement: HTMLSpanElement | undefined;
	let ringing: Animation | undefined;

	onDestroy(() => ringing?.cancel());

	function ringClock() {
		ringing?.cancel();
		if (!clockElement || window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

		ringing = clockElement.animate(
			[0, -9, 9, -8, 8, -6, 6, -4, 4, -2, 2, 0].map((angle) => ({
				transform: `rotate(${angle}deg)`
			})),
			{ duration: 850, easing: 'ease-in-out' }
		);
	}

	type ViewTransitionLike = { ready: Promise<void> };
	type DocumentWithViewTransition = Document & {
		startViewTransition?: (callback: () => void) => ViewTransitionLike;
	};

	const isMobile = () => window.innerWidth < 768;

	const nextTheme = (): Theme => (theme.current === 'light' ? 'dark' : 'light');

	const labelMap: Record<Theme, string> = {
		light: '浅色模式',
		dark: '深色模式'
	};

	const toggleTheme = async (event: MouseEvent) => {
		if (compact) ringClock();
		const next = nextTheme();
		const willChange = resolveTheme(next) !== resolved;
		const doc = document as DocumentWithViewTransition;
		const root = document.documentElement;

		// 实际深浅色没变化或不支持 View Transitions：直接切换
		if (!willChange || !doc.startViewTransition || isMobile()) {
			theme.set(next);
			return;
		}

		root.dataset.themeTransitioning = 'true';
		try {
			const x = event.clientX;
			const y = event.clientY;
			const endRadius = Math.hypot(
				Math.max(x, window.innerWidth - x),
				Math.max(y, window.innerHeight - y)
			);

			const transition = doc.startViewTransition.call(doc, () => {
				theme.set(next);
			});

			await transition.ready;

			const reveal = document.documentElement.animate(
				{
					clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`]
				},
				{
					duration: 350,
					easing: 'ease-out',
					pseudoElement: '::view-transition-new(root)'
				}
			);
			await reveal.finished;
		} finally {
			delete root.dataset.themeTransitioning;
		}
	};
</script>

<button
	type="button"
	class:theme-toggle-compact={compact}
	data-theme={theme.current}
	aria-label={labelMap[theme.current]}
	title={labelMap[theme.current]}
	onclick={toggleTheme}
	class="flex h-10 w-10 items-center justify-center rounded-default text-ink-400 hover:bg-ink-100 hover:text-ink-900 dark:hover:bg-ink-800 dark:hover:text-ink-100"
>
	{#if compact}
		<span class="clock-motion" bind:this={clockElement}>
			<AlarmClock />
		</span>
	{:else if theme.current === 'dark'}
		<Moon class="w-5 h-5 relative z-10" />
	{:else}
		<Sun class="w-5 h-5 relative z-10" />
	{/if}
</button>

<style>
	.clock-motion {
		display: block;
		width: 100%;
		height: 100%;
		transform-origin: 50% 90%;
	}

	.theme-toggle-compact {
		height: 76.5px;
		width: 68px;
		padding: 0;
		border-radius: 4px;
		background: transparent;
		transition: filter 160ms ease;
	}

	.theme-toggle-compact:hover,
	.theme-toggle-compact:focus-visible {
		background: transparent;
		filter: brightness(1.08);
	}

	.theme-toggle-compact:focus-visible {
		outline: 2px solid #d7c295;
		outline-offset: 3px;
	}

	@media (prefers-reduced-motion: reduce) {
		.theme-toggle-compact {
			transition: none;
		}
	}
</style>
