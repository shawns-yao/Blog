interface ScrollFadeOptions {
	distance?: number;
}

const clamp = (value: number) => Math.min(1, Math.max(0, value));

export function scrollFade(node: HTMLElement, options: ScrollFadeOptions = {}) {
	let frame = 0;
	let start = 0;
	const distance = options.distance ?? 420;

	const measure = () => {
		start = node.getBoundingClientRect().top + window.scrollY;
	};

	const update = () => {
		frame = 0;
		const progress = clamp((window.scrollY - start) / distance);
		node.style.setProperty('--scroll-fade-progress', progress.toFixed(3));
		node.style.pointerEvents = progress >= 0.98 ? 'none' : '';
	};

	const scheduleUpdate = () => {
		if (frame) return;
		frame = window.requestAnimationFrame(update);
	};

	const handleResize = () => {
		measure();
		scheduleUpdate();
	};

	measure();
	update();
	window.addEventListener('scroll', scheduleUpdate, { passive: true });
	window.addEventListener('resize', handleResize);

	return {
		destroy() {
			window.removeEventListener('scroll', scheduleUpdate);
			window.removeEventListener('resize', handleResize);
			if (frame) window.cancelAnimationFrame(frame);
		}
	};
}
