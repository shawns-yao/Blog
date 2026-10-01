<script lang="ts">
	import { dev } from '$app/environment';
	import { goto, preloadData, pushState, replaceState } from '$app/navigation';
	import { navigating, page } from '$app/state';
	import { buildMomentBookSpreads, formatLeafPageLabel } from '$lib/features/moment/book-pages';
	import { momentListCtx } from '$lib/features/moment/context';
	import { layoutPreviewMoments } from '$lib/features/moment/layout-preview';
	import type { MomentDetail, MomentListResponse, MomentSummary } from '$lib/features/moment/types';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import MomentBookShell from './MomentBookShell.svelte';
	import MomentDatePage from './MomentDatePage.svelte';
	import MomentDetailView from './MomentDetail.svelte';
	import { onDestroy, onMount } from 'svelte';

	interface Props {
		moments: MomentListResponse;
		search?: string;
		basePath?: string;
		staggerKey?: string;
	}
	type MomentOverlay = {
		moment: MomentDetail;
		underlayMoments: MomentListResponse;
		preview: boolean;
		returnPath: string;
	};

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
	let visibleOverlay = $state<MomentOverlay | null>(null);
	let closeTimer: ReturnType<typeof setTimeout> | undefined;
	let openRequestId = 0;
	let lastOpenedLink: HTMLAnchorElement | null = null;
	const routeOverlay = $derived(
		(page.state as { momentOverlay?: MomentOverlay }).momentOverlay ?? null
	);

	$effect(() => {
		if (routeOverlay) {
			clearTimeout(closeTimer);
			closeTimer = undefined;
			visibleOverlay = routeOverlay;
		} else if (visibleOverlay && !closeTimer) {
			const delay = window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 430;
			closeTimer = setTimeout(() => {
				visibleOverlay = null;
				closeTimer = undefined;
				lastOpenedLink?.focus();
			}, delay);
		}
	});
	onDestroy(() => clearTimeout(closeTimer));

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

	async function openDetail(href: string, replace = false, link?: HTMLAnchorElement) {
		const requestId = ++openRequestId;
		const returnPath = routeOverlay?.returnPath ?? page.url.pathname + page.url.search;
		if (link) lastOpenedLink = link;
		try {
			const result = await preloadData(href);
			if (requestId !== openRequestId) return;
			if (result.type !== 'loaded' || result.status !== 200 || !result.data.moment) {
				void goto(href);
				return;
			}
			const overlay: MomentOverlay = {
				moment: result.data.moment as MomentDetail,
				underlayMoments: (result.data.underlayMoments as MomentListResponse | undefined) ?? moments,
				preview: (result.data.moment as MomentDetail).id < 0,
				returnPath
			};
			const nextState = { ...page.state, momentOverlay: overlay } as App.PageState;
			if (replace) replaceState(href, nextState);
			else pushState(href, nextState);
		} catch {
			if (requestId === openRequestId) void goto(href);
		}
	}

	function openEntry(_moment: MomentSummary, href: string, link: HTMLAnchorElement) {
		void openDetail(href, false, link);
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

<div inert={!!visibleOverlay} aria-hidden={!!visibleOverlay}>
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
						onOpen={openEntry}
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
					onOpen={openEntry}
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
</div>

{#if visibleOverlay}
	{#key visibleOverlay.moment.id}
		<MomentDetailView
			moment={visibleOverlay.moment}
			underlayMoments={visibleOverlay.underlayMoments}
			preview={visibleOverlay.preview}
			embedded
			shallowOpen={!!routeOverlay}
			returnPathOverride={visibleOverlay.returnPath}
			onRelatedNavigate={(href) => void openDetail(href, true)}
		/>
	{/key}
{/if}

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
