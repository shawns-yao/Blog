<script lang="ts">
	let { dateKey = '' }: { dateKey?: string } = $props();
	const [, month = '09', day = '01'] = $derived(dateKey.split('-'));
	const season = $derived.by(() => {
		const value = Number(month);
		if (value >= 3 && value <= 5) return 'spring';
		if (value >= 6 && value <= 8) return 'summer';
		if (value >= 9 && value <= 11) return 'autumn';
		return 'winter';
	});
	const position = $derived(Number(day) % 3);
</script>

<div
	class="seasonal-mark"
	class:middle={position === 1}
	class:near-spine={position === 2}
	aria-hidden="true"
>
	{#if season === 'autumn'}
		<svg viewBox="0 0 84 92">
			<path
				class="leaf-fill"
				d="M42 4 49 22 61 14 59 31 78 29 65 43 79 50 57 56 61 78 45 64 42 89 38 64 21 79 27 56 5 50 19 42 6 29 26 31 23 14 36 22Z"
			/>
			<path class="leaf-vein" d="M42 10v72M42 49 20 34M42 55 66 37M41 63 26 67M43 67l13 8" />
		</svg>
	{:else if season === 'spring'}
		<svg viewBox="0 0 90 94">
			<path class="stem" d="M45 90c-2-26 0-47 4-67" />
			<circle class="seed-core" cx="50" cy="22" r="5" />
			{#each [0, 36, 72, 108, 144, 180, 216, 252, 288, 324] as angle}
				<g transform={`rotate(${angle} 50 22)`}
					><path class="seed" d="M50 17 50 3M50 3l-4 6M50 3l4 6" /></g
				>
			{/each}
			<path class="seed-fly" d="M66 17c6-8 12-10 18-11M78 5l-5 1M78 5l-2 5" />
		</svg>
	{:else if season === 'summer'}
		<svg viewBox="0 0 86 92">
			<path class="leaf-fill summer" d="M43 6C9 19 5 55 40 79c3-25 4-48 3-73Z" />
			<path class="leaf-fill summer pale" d="M46 9c30 17 29 48-3 70 2-24 3-47 3-70Z" />
			<path class="leaf-vein" d="M43 10 42 87M41 34 22 25M42 48 17 43M43 58l20-24M43 69l17-13" />
		</svg>
	{:else}
		<svg viewBox="0 0 86 94">
			<path
				class="stem"
				d="M19 88C39 65 49 39 59 7M38 60 20 43M48 38l20-16M31 70l-11-2M53 25l-6-13"
			/>
			<circle class="berry" cx="19" cy="42" r="4" /><circle
				class="berry"
				cx="68"
				cy="21"
				r="4"
			/><circle class="berry" cx="46" cy="11" r="3.5" />
		</svg>
	{/if}
</div>

<style>
	.seasonal-mark {
		position: absolute;
		top: 1.55rem;
		left: 1.7rem;
		width: 3.65rem;
		opacity: 0.68;
		color: #91483b;
		transform: rotate(-12deg);
		pointer-events: none;
	}
	.seasonal-mark.middle {
		top: 43%;
		left: 1.2rem;
		transform: rotate(9deg);
	}
	.seasonal-mark.near-spine {
		top: 1.35rem;
		left: auto;
		right: 1rem;
		transform: rotate(17deg);
		opacity: 0.48;
	}
	.seasonal-mark svg {
		display: block;
		width: 100%;
		height: auto;
		overflow: visible;
	}
	.leaf-fill {
		fill: rgba(151, 56, 42, 0.18);
		stroke: currentColor;
		stroke-width: 1.1;
	}
	.leaf-fill.summer {
		fill: rgba(87, 109, 67, 0.14);
		stroke: #657653;
	}
	.leaf-fill.pale {
		opacity: 0.58;
	}
	.leaf-vein,
	.stem,
	.seed,
	.seed-fly {
		fill: none;
		stroke: currentColor;
		stroke-width: 1.15;
		stroke-linecap: round;
		stroke-linejoin: round;
	}
	.seed-core,
	.berry {
		fill: currentColor;
		opacity: 0.66;
	}
	@media (max-width: 767px) {
		.seasonal-mark {
			width: 3rem;
			left: 0.9rem;
		}
	}
</style>
