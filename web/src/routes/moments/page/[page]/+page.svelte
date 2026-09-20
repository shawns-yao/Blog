<script lang="ts">
	import MomentListPage from '$lib/features/moment/components/MomentListPage.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<MomentListPage
	moments={data.moments}
	search={data.search}
	staggerKey={`moments-${data.moments.page}`}
/>
