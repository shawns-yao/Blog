import { getApi } from '$lib/shared/clients/api';
import type { RagAnswer, RagAvailability, RagMessage } from './types';

export function getRagAvailability(signal?: AbortSignal): Promise<RagAvailability> {
	return getApi()<RagAvailability>('/public/rag/status', { signal, retry: 0 });
}

export function askRag(
	question: string,
	sessionId: string,
	signal: AbortSignal,
	history: RagMessage[] = []
): Promise<RagAnswer> {
	return getApi()<RagAnswer>('/public/ask', {
		method: 'POST',
		body: { question, sessionId, history },
		signal,
		retry: 0,
		timeout: 95000
	});
}
