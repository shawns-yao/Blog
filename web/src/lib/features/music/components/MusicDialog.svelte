<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import X from 'lucide-svelte/icons/x';
	let {
		open = $bindable(false),
		title,
		children
	} = $props<{ open: boolean; title: string; children: Snippet }>();
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 z-[2100] bg-black/30" />
		<Dialog.Content
			class="fixed left-1/2 top-1/2 z-[2101] max-h-[85dvh] w-[calc(100vw_-_2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-xl bg-white p-6 text-stone-800 shadow-xl [color-scheme:light]"
		>
			<div class="mb-5 flex items-center justify-between gap-3">
				<Dialog.Title class="min-w-0 break-words text-base font-medium">{title}</Dialog.Title>
				<Dialog.Close class="shrink-0 rounded p-2 hover:bg-stone-100" aria-label="关闭"
					><X size={18} /></Dialog.Close
				>
			</div>
			<Dialog.Description class="sr-only">{title}</Dialog.Description>
			{@render children()}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
