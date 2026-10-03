export interface AudioState {
	paused: boolean;
	loading: boolean;
	time: number;
	duration: number;
	volume: number;
	error: string;
}

interface AudioOptions {
	src: string;
	key?: number;
	autoplay?: boolean;
	onstate: (state: AudioState) => void;
	onend: () => void;
}

export function musicAudio(node: HTMLAudioElement, initial: AudioOptions) {
	let options = initial;
	let source = '';
	let sourceKey: number | undefined;
	let loading = false;
	let error = '';
	let disposed = false;
	let generation = 0;
	function publish() {
		if (!disposed)
			options.onstate({
				paused: node.paused,
				loading,
				time: node.currentTime,
				duration: Number.isFinite(node.duration) ? node.duration : 0,
				volume: node.volume,
				error
			});
	}
	const events = ['timeupdate', 'durationchange', 'pause', 'volumechange', 'seeking', 'seeked'];
	const played = () => {
		error = '';
		publish();
	};
	const waiting = () => {
		loading = true;
		publish();
	};
	const ready = () => {
		loading = false;
		publish();
	};
	const failed = () => {
		loading = false;
		error = '播放失败，请确认登录状态和音乐服务后重试';
		publish();
	};
	const ended = () => options.onend();
	events.forEach((event) => node.addEventListener(event, publish));
	node.addEventListener('waiting', waiting);
	node.addEventListener('play', played);
	node.addEventListener('playing', ready);
	node.addEventListener('canplay', ready);
	node.addEventListener('error', failed);
	node.addEventListener('ended', ended);
	function update(next: AudioOptions) {
		options = next;
		if (source === next.src && sourceKey === next.key) return;
		source = next.src;
		sourceKey = next.key;
		const current = ++generation;
		node.pause();
		error = '';
		loading = !!source;
		if (source) node.src = source;
		else node.removeAttribute('src');
		node.load();
		publish();
		if (source && next.autoplay !== false)
			void node.play().catch(() => {
				if (!disposed && current === generation) {
					loading = false;
					if (!node.error) error = '请点击播放按钮继续播放';
					publish();
				}
			});
	}
	update(initial);
	return {
		update,
		destroy() {
			disposed = true;
			generation++;
			events.forEach((event) => node.removeEventListener(event, publish));
			node.removeEventListener('waiting', waiting);
			node.removeEventListener('play', played);
			node.removeEventListener('playing', ready);
			node.removeEventListener('canplay', ready);
			node.removeEventListener('error', failed);
			node.removeEventListener('ended', ended);
			node.pause();
			node.removeAttribute('src');
			node.load();
		}
	};
}

export async function toggleMusicAudio(node: HTMLAudioElement | undefined) {
	if (!node?.src) return;
	if (node.paused) await node.play();
	else node.pause();
}

export function seekMusicAudio(node: HTMLAudioElement | undefined, seconds: number) {
	if (node && Number.isFinite(node.duration))
		node.currentTime = Math.min(Math.max(seconds, 0), node.duration);
}

export function setMusicVolume(node: HTMLAudioElement | undefined, volume: number) {
	if (node) node.volume = Math.min(Math.max(volume, 0), 1);
}
