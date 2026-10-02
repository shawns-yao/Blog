<script lang="ts">
	import { base } from '$app/paths';
	let { class: className = 'size-9', animated = false } = $props<{
		class?: string;
		animated?: boolean;
	}>();
</script>

<span class={`${className} inline-block`} class:animated aria-hidden="true">
	<img
		src={`${base}/images/rag-shuling-avatar_20260930.webp`}
		class="size-full object-contain"
		width="256"
		height="256"
		alt=""
		decoding="async"
		draggable="false"
	/>
</span>

<style>
	.animated {
		transition: transform 200ms ease;
	}
	.animated img {
		animation: shuling-sway 4.2s ease-in-out infinite;
	}
	.animated:hover {
		transform: scale(1.04) rotate(-3deg);
	}
	.animated:hover img {
		animation-play-state: paused;
	}
	.animated:active {
		transform: scale(0.97);
	}
	@keyframes shuling-sway {
		0%,
		100% {
			transform: translateY(0) rotate(-1deg);
		}
		50% {
			transform: translateY(-3px) rotate(1deg);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.animated {
			transition: none;
			transform: none;
		}
		.animated img {
			animation: none;
		}
	}
</style>
