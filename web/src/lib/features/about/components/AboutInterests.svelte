<script lang="ts">
	import { ArrowUpRight, Compass, MessageSquare } from 'lucide-svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { resolveHref } from '$lib/shared/utils/resolve-path';

	const interests = [
		{
			name: 'Agent 运行时',
			description: '关注任务执行、工具调用与长期记忆，探索更清晰的 Agent 运行机制。'
		},
		{
			name: 'RAG 与知识组织',
			description: '从文档分块、检索到有据回答，让知识能够积累，也能够被重新使用。'
		},
		{
			name: '工具治理与 MCP',
			description: '围绕工具边界、权限和调用记录，思考 MCP 如何连接不同能力。'
		}
	];
</script>

<section class="about-interests mt-12" aria-labelledby="interests-heading">
	<header class="flex items-start gap-3.5 sm:gap-5">
		<Compass
			class="mt-0.5 shrink-0 text-[var(--accent)]"
			size={30}
			strokeWidth={1.5}
			aria-hidden="true"
		/>
		<h2
			id="interests-heading"
			class="flex-1 border-b border-[var(--paper-line)] pb-2 text-lg font-medium tracking-wider"
		>
			正在关注
		</h2>
	</header>
	<dl class="mt-2 sm:ml-12.5">
		{#each interests as interest (interest.name)}
			<div
				class="interest-row grid gap-2 border-b border-[var(--paper-line)] py-5 last:border-b-0 sm:gap-6"
			>
				<dt class="text-base font-medium leading-7 text-ink-800 dark:text-ink-200">
					{interest.name}
				</dt>
				<dd class="text-sm leading-7 text-ink-600 dark:text-ink-400">{interest.description}</dd>
			</div>
		{/each}
	</dl>
</section>

<footer
	class="about-contact mt-9 flex flex-wrap items-center justify-between gap-6 border-t border-[var(--paper-line)] py-8"
>
	<div class="flex min-w-0 items-start gap-3.5 sm:gap-5">
		<MessageSquare
			class="mt-0.5 shrink-0 text-[var(--accent)]"
			size={30}
			strokeWidth={1.5}
			aria-hidden="true"
		/>
		<div class="min-w-0">
			<h2 class="text-lg font-medium tracking-wider">交流与留言</h2>
			<p class="mt-2 text-sm leading-7 text-ink-600 dark:text-ink-400">
				有技术问题、项目想法，或想打个招呼，欢迎给我留言。
			</p>
		</div>
	</div>
	<Button
		href={resolveHref('/message/')}
		variant="secondary"
		size="lg"
		class="contact-link min-h-11 min-w-28 font-serif"
	>
		留言
		<ArrowUpRight size={16} strokeWidth={1.5} aria-hidden="true" />
	</Button>
</footer>

<style>
	.about-contact :global(.contact-link) {
		border-color: var(--paper-line);
		background: var(--paper);
	}
	.about-contact :global(.contact-link):hover {
		border-color: var(--accent);
	}
	.about-contact :global(.contact-link):focus-visible {
		outline: 2px solid var(--accent);
		outline-offset: 4px;
	}
	@media (min-width: 640px) {
		.interest-row {
			grid-template-columns: 160px minmax(0, 1fr);
		}
	}
	@media (max-width: 639px) {
		.about-contact :global(.contact-link) {
			margin-left: 44px;
		}
	}
</style>
