import type { RagHistoryPolicy, RagMessage, RagTurn } from './types';

export function conversationHistory(
	turns: RagTurn[],
	policy: RagHistoryPolicy = { maxRounds: 5, maxCharacters: 20000 }
): RagMessage[] {
	const history: RagMessage[] = [];
	let characters = 0;
	for (let index = turns.length - 1; index >= 0 && history.length < policy.maxRounds * 2; index--) {
		const turn = turns[index];
		if (turn.answer?.status !== 'answered') continue;
		const content = Array.from(turn.answer.answer).slice(0, 6000).join('');
		const length = Array.from(turn.question).length + Array.from(content).length;
		if (characters + length > policy.maxCharacters) break;
		history.unshift({ role: 'user', content: turn.question }, { role: 'assistant', content });
		characters += length;
	}
	return history;
}
