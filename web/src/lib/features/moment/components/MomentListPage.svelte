<script lang="ts">
	import { dev } from '$app/environment';
	import { goto } from '$app/navigation';
	import { navigating, page } from '$app/state';
	import { buildMomentBookSpreads, formatLeafPageLabel } from '$lib/features/moment/book-pages';
	import { momentListCtx } from '$lib/features/moment/context';
	import type { MomentListResponse, MomentSummary } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import MomentBookShell from './MomentBookShell.svelte';
	import MomentDatePage from './MomentDatePage.svelte';
	import { onMount } from 'svelte';

	interface Props {
		moments: MomentListResponse;
		search?: string;
		basePath?: string;
		staggerKey?: string;
	}

	const layoutPreviewMoments: MomentSummary[] = [
		{
			id: -1,
			title: '九点半，窗边的光刚刚好',
			shortUrl: 'layout-preview-morning',
			summary: '把咖啡放到手边，读完昨晚折角的那一页。窗外很亮，房间里安静得只剩翻书声。',
			cover: '/home-scenes/11.webp',
			views: 128,
			topics: ['日常', '阅读'],
			likes: 18,
			comments: 3,
			isTop: true,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-22T09:30:00+08:00',
			createdAt: '2026-09-22T09:30:00+08:00',
			updatedAt: '2026-09-22T09:30:00+08:00'
		},
		{
			id: -2,
			title: '把读到一半的书留在桌上',
			shortUrl: 'layout-preview-afternoon',
			summary: '有些内容不必急着读完。停在恰好的地方，等下一次回到桌前，故事仍会从书签旁边继续。',
			views: 96,
			topics: ['片刻', '书房'],
			likes: 12,
			comments: 1,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-22T15:20:00+08:00',
			createdAt: '2026-09-22T15:20:00+08:00',
			updatedAt: '2026-09-22T15:20:00+08:00'
		},
		{
			id: -3,
			title: '台灯亮起以后',
			shortUrl: 'layout-preview-evening',
			summary: '天色暗得很慢，笔尖停在纸面上，刚好听见窗外第一声晚风。',
			views: 73,
			topics: ['夜晚', '随想'],
			likes: 9,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-22T20:10:00+08:00',
			createdAt: '2026-09-22T20:10:00+08:00',
			updatedAt: '2026-09-22T20:10:00+08:00'
		},
		{
			id: -4,
			title: '月亮越过窗框',
			shortUrl: 'layout-preview-moon',
			summary: '写下今天最后一句话，把书页轻轻压平。',
			views: 51,
			topics: ['夜晚'],
			likes: 7,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-22T23:06:00+08:00',
			createdAt: '2026-09-22T23:06:00+08:00',
			updatedAt: '2026-09-22T23:06:00+08:00'
		},
		{
			id: -5,
			title: '秋天好像真的来了',
			shortUrl: 'layout-preview-autumn',
			summary: '傍晚的风开始有了凉意，楼下的梧桐叶也变黄了。',
			views: 81,
			topics: ['日常'],
			likes: 11,
			comments: 2,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-23T08:45:00+08:00',
			createdAt: '2026-09-23T08:45:00+08:00',
			updatedAt: '2026-09-23T08:45:00+08:00'
		},
		{
			id: -6,
			title: '午后的一小段空白',
			shortUrl: 'layout-preview-blank',
			summary: '没有安排的十分钟，也值得被单独留下。',
			views: 42,
			topics: ['片刻'],
			likes: 6,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-23T13:20:00+08:00',
			createdAt: '2026-09-23T13:20:00+08:00',
			updatedAt: '2026-09-23T13:20:00+08:00'
		},
		{
			id: -7,
			title: '钢笔应该放在右手边',
			shortUrl: 'layout-preview-pen',
			summary: '整理桌面时突然发现，熟悉的位置也有自己的秩序。',
			views: 39,
			topics: ['日常', '书房'],
			likes: 5,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-23T16:10:00+08:00',
			createdAt: '2026-09-23T16:10:00+08:00',
			updatedAt: '2026-09-23T16:10:00+08:00'
		},
		{
			id: -8,
			title: '晚饭后的短散步',
			shortUrl: 'layout-preview-walk',
			summary: '绕着街角走了一圈，风里已经有桂花的味道。',
			views: 31,
			topics: ['日常'],
			likes: 4,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-23T18:35:00+08:00',
			createdAt: '2026-09-23T18:35:00+08:00',
			updatedAt: '2026-09-23T18:35:00+08:00'
		},
		{
			id: -9,
			title: '给明天留一张便签',
			shortUrl: 'layout-preview-note',
			summary: '先写下最重要的一件事，其他的等太阳升起来再说。',
			views: 28,
			topics: ['计划'],
			likes: 3,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-23T20:40:00+08:00',
			createdAt: '2026-09-23T20:40:00+08:00',
			updatedAt: '2026-09-23T20:40:00+08:00'
		},
		{
			id: -10,
			title: '听完一首旧歌',
			shortUrl: 'layout-preview-song',
			summary: '熟悉的旋律经过很多年，还是会把人带回同一扇窗前。',
			views: 24,
			topics: ['片刻', '音乐'],
			likes: 3,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-23T22:05:00+08:00',
			createdAt: '2026-09-23T22:05:00+08:00',
			updatedAt: '2026-09-23T22:05:00+08:00'
		},
		{
			id: -11,
			title: '今天写到这里',
			shortUrl: 'layout-preview-goodnight',
			summary: '合上电脑之前，再看一眼窗外安静的月亮。',
			views: 19,
			topics: ['夜晚'],
			likes: 2,
			comments: 0,
			isTop: false,
			isHot: false,
			isOriginal: true,
			contentUpdatedAt: '2026-09-23T23:18:00+08:00',
			createdAt: '2026-09-23T23:18:00+08:00',
			updatedAt: '2026-09-23T23:18:00+08:00'
		}
	];

	let { moments, search = '', basePath = '/moments', staggerKey = 'moments' }: Props = $props();
	momentListCtx.mountModelData(() => moments);
	const isLayoutPreview = $derived(dev && !search && moments.items.length === 0);
	const visibleMoments = $derived(isLayoutPreview ? layoutPreviewMoments : moments.items);
	const spreads = $derived(buildMomentBookSpreads(visibleMoments));
	const totalServerPages = $derived(
		moments.size > 0 ? Math.max(1, Math.ceil(moments.total / moments.size)) : 1
	);
	let spreadIndex = $state(0);
	let turnDirection = $state<'older' | 'newer' | null>(null);
	let currentDatasetKey = $state('');
	let touchStartX: number | null = null;
	let touchStartY: number | null = null;

	$effect(() => {
		const nextKey = `${staggerKey}-${search}-${visibleMoments.map((item) => item.id).join(',')}`;
		if (nextKey !== currentDatasetKey) {
			currentDatasetKey = nextKey;
			spreadIndex = Math.max(0, spreads.length - 1);
		}
	});

	const currentSpread = $derived(
		spreads[Math.min(spreadIndex, spreads.length - 1)] ?? { left: null, right: null }
	);
	const openContext = $derived({
		spread: currentSpread,
		spreadIndex,
		returnPath: page.url.pathname + page.url.search
	});

	onMount(() => {
		try {
			const raw = sessionStorage.getItem('moment:return-spread');
			if (!raw) return;
			sessionStorage.removeItem('moment:return-spread');
			const saved = JSON.parse(raw) as { returnPath: string; spreadIndex: number; at: number };
			if (
				saved.returnPath === page.url.pathname + page.url.search &&
				Date.now() - saved.at < 60_000 &&
				Number.isInteger(saved.spreadIndex) &&
				saved.spreadIndex >= 0 &&
				saved.spreadIndex < spreads.length
			) {
				spreadIndex = saved.spreadIndex;
			}
		} catch {
			sessionStorage.removeItem('moment:return-spread');
		}
	});
	const canTurnOlder = $derived(spreadIndex > 0 || moments.page < totalServerPages);
	const canTurnNewer = $derived(spreadIndex < spreads.length - 1 || moments.page > 1);
	const olderLeaf = $derived(spreads[spreadIndex - 1]?.right ?? spreads[spreadIndex - 1]?.left);
	const newerLeaf = $derived(spreads[spreadIndex + 1]?.left ?? spreads[spreadIndex + 1]?.right);

	function pageHref(page: number) {
		const safePage = Math.max(1, page);
		const path = resolvePath(safePage === 1 ? `${basePath}/` : `${basePath}/page/${safePage}/`);
		return `${path}${search ? `?${new URLSearchParams({ q: search })}` : ''}`;
	}

	function animateTurn(direction: 'older' | 'newer') {
		turnDirection = direction;
		window.setTimeout(() => (turnDirection = null), 520);
	}

	function turnOlder() {
		if (spreadIndex > 0) {
			animateTurn('older');
			spreadIndex -= 1;
			return;
		}
		if (moments.page < totalServerPages) void goto(pageHref(moments.page + 1));
	}

	function turnNewer() {
		if (spreadIndex < spreads.length - 1) {
			animateTurn('newer');
			spreadIndex += 1;
			return;
		}
		if (moments.page > 1) void goto(pageHref(moments.page - 1));
	}

	function rememberTouch(event: TouchEvent) {
		touchStartX = event.changedTouches[0]?.clientX ?? null;
		touchStartY = event.changedTouches[0]?.clientY ?? null;
	}

	function turnFromSwipe(event: TouchEvent) {
		if (touchStartX === null || touchStartY === null) return;
		const deltaX = (event.changedTouches[0]?.clientX ?? touchStartX) - touchStartX;
		const deltaY = (event.changedTouches[0]?.clientY ?? touchStartY) - touchStartY;
		touchStartX = null;
		touchStartY = null;
		if (Math.abs(deltaX) < 60 || Math.abs(deltaX) < Math.abs(deltaY) * 1.3) return;
		if (deltaX < 0) turnOlder();
		else turnNewer();
	}
</script>

<MomentBookShell pageLabel="手记列表">
	{#snippet directory()}
		<div
			class="leaf-motion"
			class:turn-older={turnDirection === 'older'}
			class:turn-newer={turnDirection === 'newer'}
			role="region"
			aria-label="左侧日期书页"
			aria-busy={!!navigating.to}
			ontouchstart={rememberTouch}
			ontouchend={turnFromSwipe}
		>
			{#key `${currentSpread.left?.dateKey ?? 'blank'}-${currentSpread.left?.part ?? 0}`}
				<MomentDatePage
					leaf={currentSpread.left}
					side="left"
					{search}
					{basePath}
					preview={isLayoutPreview}
					{openContext}
					kicker={currentSpread.left?.dateKey &&
					currentSpread.left.dateKey === currentSpread.right?.dateKey
						? `SAME DAY / ${String(currentSpread.left.part).padStart(2, '0')}`
						: 'YESTERDAY / NOTES'}
					canTurn={canTurnOlder}
					turnLabel="翻到更早的手记"
					turnPageLabel={olderLeaf ? formatLeafPageLabel(olderLeaf) : '更早的手记'}
					onTurn={turnOlder}
				/>
			{/key}
		</div>
	{/snippet}

	<div
		class="leaf-motion"
		class:turn-older={turnDirection === 'older'}
		class:turn-newer={turnDirection === 'newer'}
		role="region"
		aria-label="右侧日期书页"
		aria-busy={!!navigating.to}
		ontouchstart={rememberTouch}
		ontouchend={turnFromSwipe}
	>
		{#key `${currentSpread.right?.dateKey ?? 'blank'}-${currentSpread.right?.part ?? 0}`}
			<MomentDatePage
				leaf={currentSpread.right}
				side="right"
				{search}
				{basePath}
				preview={isLayoutPreview}
				{openContext}
				kicker={currentSpread.right?.dateKey &&
				currentSpread.right.dateKey === currentSpread.left?.dateKey
					? `SAME DAY / ${String(currentSpread.right.part).padStart(2, '0')}`
					: 'TODAY / NOTES'}
				canTurn={canTurnNewer}
				turnLabel="翻到更新的手记"
				turnPageLabel={newerLeaf ? formatLeafPageLabel(newerLeaf) : '更新的手记'}
				onTurn={turnNewer}
			/>
		{/key}
	</div>
</MomentBookShell>

<style>
	.leaf-motion {
		height: 100%;
		transform-origin: center;
	}
	.leaf-motion.turn-older {
		animation: turn-to-older 500ms cubic-bezier(0.2, 0.72, 0.2, 1);
	}
	.leaf-motion.turn-newer {
		animation: turn-to-newer 500ms cubic-bezier(0.2, 0.72, 0.2, 1);
	}
	@keyframes turn-to-older {
		0% {
			opacity: 0.15;
			transform: perspective(1000px) rotateY(-8deg) translateX(0.8rem);
			filter: blur(2px);
		}
		100% {
			opacity: 1;
			transform: none;
			filter: none;
		}
	}
	@keyframes turn-to-newer {
		0% {
			opacity: 0.15;
			transform: perspective(1000px) rotateY(8deg) translateX(-0.8rem);
			filter: blur(2px);
		}
		100% {
			opacity: 1;
			transform: none;
			filter: none;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.leaf-motion.turn-older,
		.leaf-motion.turn-newer {
			animation: none;
		}
	}
</style>
