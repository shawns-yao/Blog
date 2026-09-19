<script lang="ts">
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';
	import { resolveHref } from '$lib/shared/utils/resolve-path';

	const { icon, name, href } = $props<{
		icon: string;
		name: string;
		href: string;
	}>();
</script>

<div class="social-item-container hover:text-jade-600 cursor-pointer">
	<a
		href={href.startsWith('/') ? resolveHref(href) : href}
		class="flex items-center gap-2"
		target="_blank"
		rel="noopener noreferrer"
	>
		<DynamicLucideIcon name={icon} size={14} />
		<span class="font-mono hover:underline">{name}</span>
	</a>
</div>
