interface FollowOptions {
	index: number;
	key: string;
}

export function followMusicLyric(node: HTMLElement, initial: FollowOptions) {
	let previous: FollowOptions | undefined;
	let manualUntil = 0;
	let frame = 0;
	const pauseFollowing = () => (manualUntil = Date.now() + 5000);
	const pauseWithKey = (event: KeyboardEvent) => {
		if (['ArrowUp', 'ArrowDown', 'PageUp', 'PageDown', 'Home', 'End'].includes(event.key)) {
			pauseFollowing();
		}
	};
	node.addEventListener('wheel', pauseFollowing, { passive: true });
	node.addEventListener('touchmove', pauseFollowing, { passive: true });
	node.addEventListener('keydown', pauseWithKey);
	function update(options: FollowOptions) {
		const changedSong = previous?.key !== options.key;
		if (!changedSong && previous?.index === options.index) return;
		previous = options;
		cancelAnimationFrame(frame);
		frame = requestAnimationFrame(() => {
			if (changedSong) manualUntil = 0;
			if (Date.now() < manualUntil) return;
			const line = node.querySelector<HTMLElement>(`[data-lyric-index="${options.index}"]`);
			if (!line && !changedSong) return;
			node.scrollTo({
				top: line ? line.offsetTop - node.clientHeight / 2 + line.offsetHeight / 2 : 0,
				behavior:
					changedSong || matchMedia('(prefers-reduced-motion: reduce)').matches
						? 'instant'
						: 'smooth'
			});
		});
	}
	update(initial);
	return {
		update,
		destroy() {
			cancelAnimationFrame(frame);
			node.removeEventListener('wheel', pauseFollowing);
			node.removeEventListener('touchmove', pauseFollowing);
			node.removeEventListener('keydown', pauseWithKey);
		}
	};
}
