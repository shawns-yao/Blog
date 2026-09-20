<script lang="ts">
	import { onMount } from 'svelte';

	let now = $state<Date | null>(null);
	const second = $derived(now?.getSeconds() ?? 0);
	const minute = $derived((now?.getMinutes() ?? 0) + second / 60);
	const hour = $derived(((now?.getHours() ?? 0) % 12) + minute / 60);
	const timeLabel = $derived(
		now
			? [now.getHours(), now.getMinutes(), now.getSeconds()]
					.map((part) => String(part).padStart(2, '0'))
					.join(':')
			: ''
	);

	onMount(() => {
		const update = () => {
			now = new Date();
		};
		update();
		const timer = window.setInterval(update, 1000);
		return () => window.clearInterval(timer);
	});
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	viewBox="0 0 160 180"
	role="img"
	aria-label={timeLabel ? `本地时间 ${timeLabel}` : '闹钟'}
	focusable="false"
>
	<title>{timeLabel ? `本地时间 ${timeLabel}` : '闹钟'}</title>
	<image href="/alarm-clock.svg" width="160" height="180" />
	{#if now}
		<g transform="rotate({hour * 30} 78 100)">
			<path
				d="M79 102V78"
				stroke="#6b5c43"
				stroke-opacity=".24"
				stroke-width="4"
				stroke-linecap="round"
			/>
			<path d="M78 100V77" stroke="#30372f" stroke-width="3.8" stroke-linecap="round" />
		</g>
		<g transform="rotate({minute * 6} 78 100)">
			<path
				d="M79 102V70"
				stroke="#6b5c43"
				stroke-opacity=".24"
				stroke-width="3"
				stroke-linecap="round"
			/>
			<path d="M78 100V68" stroke="#30372f" stroke-width="2.6" stroke-linecap="round" />
		</g>
		<path
			d="M78 112V63"
			transform="rotate({second * 6} 78 100)"
			stroke="#9c4d36"
			stroke-width="1.1"
		/>
		<circle cx="78" cy="100" r="4.3" fill="#b49a69" stroke="#55462e" stroke-width="1" />
		<circle cx="77" cy="99" r="1.3" fill="#e8d3a3" />
	{/if}
</svg>

<style>
	svg {
		display: block;
		width: 100%;
		height: 100%;
		transform-origin: bottom center;
		transform: rotate(2deg);
	}
</style>
