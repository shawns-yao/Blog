export type SiteActivityEvent =
	| 'moment.updated'
	| 'moment.hot_marked'
	| 'comment.created'
	| 'comment.approved';

export type SiteActivityPayload = {
	type: 'site.activity';
	event: SiteActivityEvent;
	contentType: 'moment' | 'comment';
	title: string;
	excerpt?: string;
	url: string;
	at: string;
	commentAreaId?: number;
};
