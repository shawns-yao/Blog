<script lang="ts">
	import { BookOpen, FolderOpen } from 'lucide-svelte';
	import { resolveHref } from '$lib/shared/utils/resolve-path';
	import AboutInterests from './AboutInterests.svelte';

	const skillGroups = [
		{ name: '后端工程', skills: ['Java', 'Go', 'TypeScript'] },
		{ name: 'AI 系统', skills: ['Agent', 'RAG', 'MCP'] },
		{ name: '基础设施', skills: ['Docker', 'Kubernetes', 'Linux'] }
	];
	const projects = [
		{ name: 'DevCLI', description: 'Coding Agent Runtime / Tool Governance / Memory' },
		{ name: 'Librarium', description: '个人知识空间' }
	];
</script>

<article class="about-profile mx-auto max-w-3xl pb-12 font-serif text-ink-800 dark:text-ink-200">
	<header class="about-heading text-center">
		<h1 class="text-4xl font-medium tracking-[0.12em] text-ink-900 sm:text-5xl dark:text-ink-100">
			关于
		</h1>
		<p class="mt-5 text-lg tracking-[0.14em] sm:text-xl">Librarium 的馆主档案</p>
		<span class="heading-rule mx-auto my-6 block h-px w-9" aria-hidden="true"></span>
		<p class="text-sm leading-relaxed tracking-wide text-ink-500 sm:text-base dark:text-ink-400">
			在技术与阅读之间，收集更好的自己。
		</p>
	</header>

	<section class="profile-card noise-surface mt-9 grid" aria-label="馆主介绍">
		<aside class="profile-illustration">
			<img
				src={resolveHref('/about/books-and-branch.webp')}
				alt="旧书与一枝干燥的植物"
				width="512"
				height="536"
				class="profile-books"
			/>
			<p class="illustration-caption">A QUIETER<br />BRIGHTER YOU</p>
			<span class="illustration-rule" aria-hidden="true"></span>
		</aside>
		<div class="profile-details">
			<header class="profile-identity flex items-start justify-between gap-5">
				<div>
					<h2 class="text-4xl font-normal tracking-wide text-ink-900 dark:text-ink-100">Shawn</h2>
					<p class="profile-role mt-2 text-sm tracking-[0.16em]">Builder &amp; Explorer</p>
				</div>
				<div class="library-stamp flex items-center gap-3" aria-label="Librarium，始于 2024 年">
					<p class="stamp-copy">
						<span>LIBRARIUM</span>
						<small>SINCE 2024<br />FOR A BETTER YOU</small>
					</p>
					<div class="stamp-seal" aria-hidden="true">
						<BookOpen size={27} strokeWidth={1.3} />
						<svg class="stamp-postmark" width="60" height="34" viewBox="0 0 60 34" fill="none">
							<path
								d="M1 5c10-8 15 8 25 0s15 8 25 0M1 13c10-8 15 8 25 0s15 8 25 0M1 21c10-8 15 8 25 0s15 8 25 0M1 29c10-8 15 8 25 0s15 8 25 0"
								stroke="currentColor"
							/>
						</svg>
					</div>
				</div>
			</header>
			<p
				class="profile-description mt-7 border-t pt-7 text-base leading-loose tracking-wide sm:text-lg"
			>
				这里是 Librarium，一个记录技术探索、<br
					class="hidden sm:block"
				/>项目实践与阅读思考的个人数字图书馆。
			</p>
		</div>
	</section>

	<section class="profile-section mt-12" aria-labelledby="skills-heading">
		<header class="section-header flex items-start gap-5">
			<BookOpen
				class="section-icon mt-0.5 shrink-0"
				size={30}
				strokeWidth={1.5}
				aria-hidden="true"
			/>
			<div
				class="section-title flex min-w-0 flex-1 flex-wrap items-baseline justify-between gap-x-5 gap-y-1 border-b pb-2"
			>
				<h2 id="skills-heading" class="text-lg font-medium tracking-wider">技术书架</h2>
				<p class="section-motto text-xs italic sm:text-sm">Tools for a Brighter Tomorrow.</p>
			</div>
		</header>
		<dl class="section-content mt-5 space-y-4">
			{#each skillGroups as group (group.name)}
				<div class="skill-row grid items-start gap-3">
					<dt class="pt-1 text-base tracking-wider">{group.name}</dt>
					<dd>
						<ul class="flex flex-wrap gap-2.5" aria-label={group.name}>
							{#each group.skills as skill (skill)}
								<li class="skill-tag rounded-full px-4 py-1 text-sm">{skill}</li>
							{/each}
						</ul>
					</dd>
				</div>
			{/each}
		</dl>
	</section>

	<section class="profile-section mt-12" aria-labelledby="projects-heading">
		<header class="section-header flex items-start gap-5">
			<FolderOpen
				class="section-icon mt-0.5 shrink-0"
				size={30}
				strokeWidth={1.5}
				aria-hidden="true"
			/>
			<div
				class="section-title flex min-w-0 flex-1 flex-wrap items-baseline justify-between gap-x-5 gap-y-1 border-b pb-2"
			>
				<h2 id="projects-heading" class="text-lg font-medium tracking-wider">项目收藏</h2>
				<p class="section-motto text-xs italic sm:text-sm">Ideas Live Longer.</p>
			</div>
		</header>
		<ul class="section-content mt-5 space-y-5">
			{#each projects as project (project.name)}
				<li class="project-row flex items-baseline gap-4">
					<span class="project-dot mt-2 size-2 shrink-0 rounded-full" aria-hidden="true"></span>
					<div class="flex flex-wrap items-baseline gap-x-4 gap-y-1">
						<span class="text-lg text-ink-900 dark:text-ink-100">{project.name}</span>
						<span class="project-dash text-ink-400" aria-hidden="true">—</span>
						<span class="text-sm leading-relaxed text-ink-500 dark:text-ink-400"
							>{project.description}</span
						>
					</div>
				</li>
			{/each}
		</ul>
	</section>

	<AboutInterests />
</article>

<style>
	.about-profile {
		--paper: #f5f1e8;
		--paper-line: #ded6c9;
		--accent: #a6845c;
		--stamp: #aa9780;
	}
	:global(.dark) .about-profile {
		--paper: #292622;
		--paper-line: #48413a;
		--accent: #c7aa83;
		--stamp: #b3a08a;
	}
	.heading-rule,
	.project-dot {
		background: var(--accent);
	}
	.profile-card {
		grid-template-columns: 155px minmax(0, 1fr);
		gap: 36px;
		padding: 32px 36px;
		border: 1px solid var(--paper-line);
		border-radius: 12px;
		background-color: var(--paper);
		box-shadow:
			0 14px 30px -18px #43332355,
			inset 0 1px 0 #ffffff88;
	}
	.profile-illustration {
		border-right: 1px solid var(--paper-line);
		padding-right: 25px;
	}
	.profile-books {
		width: 125px;
		height: 145px;
		object-fit: contain;
	}
	.illustration-caption {
		margin-top: 13px;
		color: var(--accent);
		font-size: 10px;
		line-height: 1.7;
		letter-spacing: 0.16em;
	}
	.illustration-rule {
		display: block;
		width: 18px;
		height: 1px;
		margin-top: 10px;
		background: var(--paper-line);
	}
	.profile-identity {
		padding-top: 9px;
	}
	.profile-role,
	.about-profile :global(.section-icon),
	.section-motto {
		color: var(--accent);
	}
	.library-stamp {
		color: var(--stamp);
	}
	.stamp-copy {
		padding-top: 2px;
		font-size: 13px;
		letter-spacing: 0.1em;
	}
	.stamp-copy small {
		display: block;
		margin-top: 7px;
		font-size: 8px;
		line-height: 1.8;
		letter-spacing: 0.15em;
	}
	.stamp-seal {
		position: relative;
		display: grid;
		place-items: center;
		width: 57px;
		height: 57px;
		border: 2px solid currentColor;
		border-radius: 50%;
		opacity: 0.65;
	}
	.stamp-postmark {
		position: absolute;
		left: 42px;
		top: 13px;
		opacity: 0.55;
	}
	.profile-description,
	.section-title {
		border-color: var(--paper-line);
	}
	.section-content {
		margin-left: 50px;
	}
	.skill-row {
		grid-template-columns: 112px minmax(0, 1fr);
	}
	.skill-tag {
		border: 1px solid var(--paper-line);
		background: var(--paper);
		box-shadow: inset 0 1px 2px #43332308;
	}
	@media (max-width: 850px) {
		.stamp-seal {
			display: none;
		}
	}
	@media (max-width: 639px) {
		.profile-card {
			grid-template-columns: minmax(0, 1fr);
			gap: 20px;
			padding: 24px;
		}
		.profile-illustration {
			display: flex;
			align-items: center;
			gap: 18px;
			border-right: 0;
			padding-right: 0;
		}
		.profile-books {
			width: 85px;
			height: 95px;
		}
		.illustration-caption {
			margin-top: 0;
		}
		.illustration-rule,
		.stamp-seal,
		.project-dash {
			display: none;
		}
		.profile-identity {
			padding-top: 0;
		}
		.profile-identity h2 {
			font-size: 30px;
		}
		.profile-role {
			font-size: 11px;
		}
		.stamp-copy {
			font-size: 10px;
		}
		.stamp-copy small {
			font-size: 7px;
		}
		.profile-description {
			margin-top: 20px;
			padding-top: 20px;
		}
		.section-header {
			gap: 14px;
		}
		.section-content {
			margin-left: 0;
		}
		.skill-row {
			grid-template-columns: 80px minmax(0, 1fr);
			gap: 10px;
		}
		.skill-tag {
			padding: 4px 12px;
			font-size: 12px;
		}
		.project-row > div {
			flex-direction: column;
		}
	}
</style>
