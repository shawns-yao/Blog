export type RagAvailability = {
	available: boolean;
	reason: 'ready' | 'disabled' | 'not_configured' | 'index_not_ready' | 'temporarily_unavailable';
};

export type RagCitation = {
	number: number;
	chunkId: number;
	momentId: number;
	title: string;
	url: string;
	content: string;
	contextHeader: string;
	kind: string;
	contentKind: 'article' | 'note';
	start: number;
	end: number;
	indexVersion: string;
	createdAt: string;
	updatedAt: string;
};

export type RagAnswer = {
	status: 'answered' | 'no_evidence' | 'temporarily_unavailable' | 'invalid_scope';
	answer: string;
	reason?: string;
	citations: RagCitation[];
	indexVersion?: string;
};
