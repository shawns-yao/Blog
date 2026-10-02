import type { MomentSummary } from './types';

export const MOMENTS_PER_BOOK_PAGE = 4;

export type MomentDateLeaf = {
	dateKey: string;
	year: string;
	month: string;
	day: string;
	items: MomentSummary[];
	totalItems: number;
	part: number;
	totalParts: number;
};

export type MomentBookSpread = {
	left: MomentDateLeaf | null;
	right: MomentDateLeaf | null;
};

export type MomentBookVisit = {
	momentId: number;
	spreadIndex: number;
	returnPath: string;
	at: number;
};

const compareMoments = (left: MomentSummary, right: MomentSummary) =>
	new Date(left.createdAt).getTime() - new Date(right.createdAt).getTime();

export const buildMomentBookSpreads = (moments: MomentSummary[]): MomentBookSpread[] => {
	const days = new Map<string, MomentSummary[]>();

	for (const moment of [...moments].sort(compareMoments)) {
		const dateKey = moment.createdAt.slice(0, 10);
		const current = days.get(dateKey) ?? [];
		current.push(moment);
		days.set(dateKey, current);
	}

	const leaves: MomentDateLeaf[] = [];
	for (const [dateKey, items] of [...days.entries()].sort(([left], [right]) =>
		left.localeCompare(right)
	)) {
		const [year = '', month = '', day = ''] = dateKey.split('-');
		const totalParts = Math.ceil(items.length / MOMENTS_PER_BOOK_PAGE);
		for (let part = 0; part < totalParts; part += 1) {
			leaves.push({
				dateKey,
				year,
				month,
				day,
				items: items.slice(part * MOMENTS_PER_BOOK_PAGE, (part + 1) * MOMENTS_PER_BOOK_PAGE),
				totalItems: items.length,
				part: part + 1,
				totalParts
			});
		}
	}

	const spreads: MomentBookSpread[] = [];
	let pendingLeaf: MomentDateLeaf | null = null;
	let singleLeaves: MomentDateLeaf[] = [];
	const flushSingleLeaves = () => {
		if (pendingLeaf && singleLeaves.length) {
			spreads.push({ left: pendingLeaf, right: singleLeaves.shift()! });
			pendingLeaf = null;
		}
		if (singleLeaves.length % 2 === 1) {
			spreads.push({ left: null, right: singleLeaves.shift()! });
		}
		for (let index = 0; index < singleLeaves.length; index += 2) {
			spreads.push({ left: singleLeaves[index], right: singleLeaves[index + 1] });
		}
		singleLeaves = [];
	};
	for (const dayLeaves of [...days.keys()]
		.sort((left, right) => left.localeCompare(right))
		.map((dateKey) => leaves.filter((leaf) => leaf.dateKey === dateKey))) {
		if (dayLeaves.length > 1) {
			flushSingleLeaves();
			if (pendingLeaf) {
				spreads.push({ left: pendingLeaf, right: null });
				pendingLeaf = null;
			}
			for (let index = 0; index < dayLeaves.length; index += 2) {
				if (dayLeaves[index + 1]) {
					spreads.push({ left: dayLeaves[index], right: dayLeaves[index + 1] });
				} else {
					pendingLeaf = dayLeaves[index];
				}
			}
			continue;
		}

		const leaf = dayLeaves[0];
		if (!leaf) continue;
		singleLeaves.push(leaf);
	}
	flushSingleLeaves();

	if (pendingLeaf) {
		spreads.push({ left: pendingLeaf, right: null });
	}
	return spreads.length ? spreads : [{ left: null, right: null }];
};

export const formatLeafPageLabel = (leaf: MomentDateLeaf): string => {
	const date = `${Number(leaf.month)}月${Number(leaf.day)}日`;
	return `${date} · ${leaf.part}/${leaf.totalParts}`;
};
