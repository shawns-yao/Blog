import assert from 'node:assert/strict';

const origin = process.env.HOME_TEST_ORIGIN || 'http://127.0.0.1:5173';
const apiOrigin = process.env.HOME_TEST_API_ORIGIN || 'http://127.0.0.1:8080';

async function read(url, json = false) {
	const response = await fetch(url, { signal: AbortSignal.timeout(20000) });
	assert.equal(response.status, 200, url);
	return json ? response.json() : response.text();
}

const [html, recentResponse, activityResponse, about] = await Promise.all([
	read(origin),
	read(`${apiOrigin}/api/v2/public/moments/recent`, true),
	read(`${apiOrigin}/api/v2/public/home/activity-pulse?days=365`, true),
	read(`${origin}/about`)
]);
const visibleHtml = html.replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '');
const recent = recentResponse.data;
assert.equal(recentResponse.code, 0);
assert.equal(recent.size, 3);
assert.ok(recent.items.length <= 3);
for (let i = 1; i < recent.items.length; i++) {
	assert.ok(Date.parse(recent.items[i - 1].createdAt) >= Date.parse(recent.items[i].createdAt));
}
assert.equal((visibleHtml.match(/class="moment-preview\b/g) || []).length, recent.items.length);
const links = [...visibleHtml.matchAll(/href="([^"]+)"/g)];
assert.ok(links.some((match) => new URL(match[1], origin).pathname.replace(/\/$/, '') === '/moments'));
assert.match(visibleHtml, /灵感与实验场/);
assert.match(visibleHtml, /创作律动/);
assert.doesNotMatch(visibleHtml, /High Energy|COFFEE|COMMITS|Tech Stack|Steady/);
assert.match(about, /技术栈/);
assert.equal(activityResponse.code, 0);
assert.equal(activityResponse.data.points.length, 365);
assert.match(visibleHtml, /近一年/);
assert.match(visibleHtml, /近三个月/);
if (recent.items.length === 0) assert.match(visibleHtml, /还没有公开的手记/);
if (activityResponse.data.totalMoments === 0) {
	assert.match(visibleHtml, /这段时间还没有公开的创作记录/);
	assert.doesNotMatch(visibleHtml, /创作记录暂时加载失败/);
}
console.log('冒烟测试通过：首页、最近手记公开接口、365 天创作接口、关于页面。');
console.log(`当前公开手记样本：${recent.items.length}；非空日历和浏览器交互未在本脚本验证。`);
