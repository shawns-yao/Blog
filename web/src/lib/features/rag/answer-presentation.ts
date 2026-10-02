import type { RagCitation } from './types';
import { parseMarkdown, type SvmdNode } from 'svmarkdown';
import { extractPlainTextFromNodes } from '$lib/shared/markdown/component-body';

export type AnswerPart = { text: string; citation?: RagCitation };

export function readableExcerpt(content: string): string {
	const body = content.replace(/^\uFEFF?---\s*\r?\n[\s\S]*?\r?\n---(?:\r?\n|$)/, '');
	const nodes = parseMarkdown(body).children;
	const formatText = (items: SvmdNode[]) => {
		for (const node of items) {
			if (node.kind === 'text') {
				// Obsidian links and Chinese-adjacent emphasis can remain literal after parsing.
				// Leave code nodes untouched so operators and indentation retain their meaning.
				node.value = node.value
					.replace(
						/\[\[([^\]\n]+)\]\]/g,
						(_match, target: string) => target.split('|').at(-1) ?? target
					)
					.replace(/(?<![\w*])\*\*([^*\n]+)\*\*(?![\w*])/g, '$1')
					.replace(/\*\*(?=[“”"'‘’。，：；！？])/g, '');
			} else if (node.kind === 'element' || node.kind === 'component') {
				formatText(node.children);
			}
		}
	};
	formatText(nodes);
	return extractPlainTextFromNodes(nodes);
}

/** Keep returned text literal; only known source references become footnotes. */
export function answerParagraphs(answer: string, citations: RagCitation[]): AnswerPart[][] {
	const sources = new Map(citations.map((citation) => [citation.number, citation]));
	return answer
		.trim()
		.split(/\r?\n\s*\r?\n/)
		.filter(Boolean)
		.map((paragraph) => {
			const parts: AnswerPart[] = [];
			let start = 0;
			for (const match of paragraph.matchAll(/\[(\d+)\]/g)) {
				const citation = sources.get(Number(match[1]));
				if (!citation) continue;
				if (match.index > start) parts.push({ text: paragraph.slice(start, match.index) });
				parts.push({ text: match[1], citation });
				start = match.index + match[0].length;
			}
			if (start < paragraph.length) parts.push({ text: paragraph.slice(start) });
			return parts;
		});
}
