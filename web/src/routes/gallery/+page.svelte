<script lang="ts">
	import LibraryHierarchy from '$lib/features/library/LibraryHierarchy.svelte';
	import LibraryReader from '$lib/features/library/LibraryReader.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<svelte:head
	><title>{data.selected?.title ?? data.column?.name ?? '图书馆'} · 图书馆</title></svelte:head
>

{#key `${data.columnId ?? ''}:${data.selected?.id ?? ''}:${data.moments.page}:${data.query}`}
	<div class="library-content">
		<LibraryHierarchy
			columns={data.columns}
			root={data.root}
			secondary={data.secondary}
			moments={data.moments}
			query={data.query}
			reading={data.reading}
		/>
		{#if data.reading}
			<LibraryReader
				columns={data.columns}
				column={data.column}
				moments={data.moments}
				selected={data.selected}
			/>
		{/if}
	</div>
{/key}

<style>
	.library-content {
		animation: library-enter 200ms ease-out;
	}
	@keyframes library-enter {
		from {
			opacity: 0;
			transform: translateY(6px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.library-content {
			animation: none;
		}
	}
</style>
