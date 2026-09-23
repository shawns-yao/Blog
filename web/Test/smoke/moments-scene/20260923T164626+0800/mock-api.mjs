import http from 'node:http';

const moments = [
	{
		id: 21,
		title: '秋天好像真的来了',
		shortUrl: 'autumn-is-here',
		summary: '傍晚的风开始有了凉意，楼下的梧桐叶也变黄了。',
		views: 128,
		columnName: '日常',
		columnShortUrl: 'daily',
		topics: [],
		likes: 16,
		comments: 3,
		isTop: false,
		isHot: false,
		isOriginal: true,
		contentUpdatedAt: '2026-09-21T20:30:00+08:00',
		createdAt: '2026-09-21T20:30:00+08:00',
		updatedAt: '2026-09-21T20:30:00+08:00'
	},
	{
		id: 18,
		title: '最近的一些想法',
		shortUrl: 'recent-thoughts',
		summary: '关于工作、生活，和一些未来的可能性。',
		views: 96,
		columnName: '随想',
		columnShortUrl: 'thoughts',
		topics: [],
		likes: 11,
		comments: 2,
		isTop: false,
		isHot: false,
		isOriginal: true,
		contentUpdatedAt: '2026-09-18T19:10:00+08:00',
		createdAt: '2026-09-18T19:10:00+08:00',
		updatedAt: '2026-09-18T19:10:00+08:00'
	},
	{
		id: 12,
		title: 'Librarium',
		shortUrl: 'librarium',
		summary: '打造一个属于自己的数字花园，让思考有处安放。',
		views: 83,
		columnName: '记录',
		columnShortUrl: 'notes',
		topics: [],
		likes: 9,
		comments: 1,
		isTop: false,
		isHot: false,
		isOriginal: true,
		contentUpdatedAt: '2026-09-12T21:00:00+08:00',
		createdAt: '2026-09-12T21:00:00+08:00',
		updatedAt: '2026-09-12T21:00:00+08:00'
	}
];

const server = http.createServer((request, response) => {
	response.setHeader('Content-Type', 'application/json; charset=utf-8');
	if (request.url?.startsWith('/api/v2/moments?')) {
		response.end(
			JSON.stringify({ code: 0, data: { items: moments, total: 27, page: 1, size: 20 } })
		);
		return;
	}
	response.end(JSON.stringify({ code: 0, data: {} }));
});

server.listen(18080, '127.0.0.1', () => {
	console.log('Mock API listening on http://127.0.0.1:18080');
});
