<script lang="ts">
	import { base } from '$app/paths';
	let { class: className = 'size-9' } = $props<{
		class?: string;
	}>();
</script>

<span class={`${className} inline-block`} aria-hidden="true">
	<img
		src={`${base}/images/rag-shuling-avatar_20260930.png`}
		class="size-full object-contain"
		width="256"
		height="256"
		alt=""
		decoding="async"
		draggable="false"
	/>
</span>
