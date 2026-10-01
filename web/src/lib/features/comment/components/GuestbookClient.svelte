<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { getGuestbookArea } from '../api';
	import CommentAreaClient from './CommentAreaClient.svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';

	const area = createQuery(() => ({
		queryKey: ['guestbook-area'],
		queryFn: getGuestbookArea,
		retry: false
	}));
</script>

{#if area.isPending}
	<p class="py-12 text-sm text-ink-500" role="status">正在打开留言板…</p>
{:else if area.isError}
	<div class="py-12">
		<p class="mb-4 text-sm text-ink-600 dark:text-ink-300" role="alert">
			留言板暂时无法加载，请稍后重试。
		</p>
		<Button variant="secondary" onclick={() => area.refetch()}>重新加载</Button>
	</div>
{:else if area.data}
	<CommentAreaClient areaId={area.data.id} initialClosed={area.data.isClosed} guestbook />
{/if}
