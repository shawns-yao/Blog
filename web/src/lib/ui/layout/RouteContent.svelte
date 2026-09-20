<script lang="ts">
	import { onNavigate } from '$app/navigation';
	import { onDestroy, type Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
	let container = $state<HTMLDivElement>();
	let animations: Animation[] = [];
	let generation = 0;

	function clearAnimations() {
		for (const animation of animations) animation.cancel();
		animations = [];
	}

	function targets(): HTMLElement[] {
		if (!container) return [];
		// Keep the homepage's fixed background out of the content animation.
		const home = container.querySelector('.homepage-container');
		return home
			? Array.from(home.querySelectorAll<HTMLElement>('.home-hero, .home-content'))
			: [container];
	}

	onNavigate(async (navigation) => {
		const current = ++generation;
		clearAnimations();
		const from = navigation.from?.url.pathname;
		const to = navigation.to?.url.pathname;
		const albumPhotoNavigation =
			(from && to && /^\/albums\/[^/]+$/.test(from) && to.startsWith(`${from}/photo/`)) ||
			(from && to && /^\/albums\/[^/]+$/.test(to) && from.startsWith(`${to}/photo/`));
		if (
			from === to ||
			navigation.type === 'popstate' ||
			albumPhotoNavigation ||
			window.matchMedia('(prefers-reduced-motion: reduce)').matches
		)
			return;

		void navigation.complete.catch(() => {
			if (current === generation) clearAnimations();
		});
		animations = targets().map((node) =>
			node.animate([{ opacity: getComputedStyle(node).opacity }, { opacity: 0 }], {
				duration: 120,
				fill: 'forwards',
				easing: 'ease-out'
			})
		);
		await Promise.all(animations.map((animation) => animation.finished.catch(() => {})));
		return () => {
			if (current !== generation) return;
			clearAnimations();
			animations = targets().map((node) =>
				node.animate(
					[
						{ opacity: 0, translate: '0 6px' },
						{ opacity: getComputedStyle(node).opacity, translate: '0 0' }
					],
					{ duration: 200, easing: 'ease-out' }
				)
			);
		};
	});

	onDestroy(() => {
		generation++;
		clearAnimations();
	});
</script>

<div bind:this={container} class="content-container min-h-[60vh]">
	{@render children()}
</div>
