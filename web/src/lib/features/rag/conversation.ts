import type { RagMessage, RagTurn } from './types';

export function conversationHistory(turns: RagTurn[]): RagMessage[] {
	const history: RagMessage[] = [];
	let characters = 0;
	for (let index = turns.length - 1; index >= 0 && history.length < 10; index--) {
		const turn = turns[index];
		if (turn.answer?.status !== 'answered') continue;
		const content = Array.from(turn.answer.answer).slice(0, 6000).join('');
		const length = Array.from(turn.question).length + Array.from(content).length;
		if (characters + length > 20000) break;
		history.unshift({ role: 'user', content: turn.question }, { role: 'assistant', content });
		characters += length;
	}
	return history;
}
