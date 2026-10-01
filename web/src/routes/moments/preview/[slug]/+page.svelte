<script lang="ts">
	import MomentDetail from '$lib/features/moment/components/MomentDetail.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<MomentDetail moment={data.moment} underlayMoments={data.underlayMoments} preview />
