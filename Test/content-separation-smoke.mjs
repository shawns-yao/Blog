import assert from 'node:assert/strict';

// Exercise public HTTP entries only; this script never creates or edits content.
const api = process.env.CONTENT_TEST_API_ORIGIN || 'http://127.0.0.1:8080';
const web = process.env.CONTENT_TEST_ORIGIN || 'http://127.0.0.1:5173';
for (const kind of ['note', 'article', 'unclassified']) {
  const response = await fetch(`${api}/api/v2/moments?contentKind=${kind}`, {
    signal: AbortSignal.timeout(20000),
  });
  assert.equal(response.status, 200);
  const result = await response.json();
  assert.equal(result.code, 0);
  assert.ok(Array.isArray(result.data.items));
  for (const item of result.data.items) assert.equal(item.contentKind, kind);
  console.log(`${kind}: ${result.data.items.length} public samples`);
}
const invalid = await fetch(`${api}/api/v2/moments?contentKind=invalid`, {
  signal: AbortSignal.timeout(20000),
});
const rejected = await invalid.json();
assert.notEqual(rejected.code, 0);
for (const [path, title] of [['/moments/', '手记'], ['/gallery/', '图书馆'], ['/gallery/?page=2', '图书馆']]) {
  const response = await fetch(`${web}${path}`, { signal: AbortSignal.timeout(30000) });
  assert.equal(response.status, 200);
  const html = await response.text();
  assert.ok(html.includes(title));
  if (path.startsWith('/gallery')) assert.ok(!html.includes('全部内容 · 手记与相册'));
}
console.log('冒烟测试通过：公开类型筛选、非法类型拒绝、手记与图书馆页面及分页入口。');
console.log('未验证：登录后的发布流程、非空数据分流、浏览器视觉和交互。');
