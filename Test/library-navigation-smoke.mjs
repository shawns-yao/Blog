import assert from 'node:assert/strict';

const origin = process.env.CONTENT_TEST_ORIGIN || 'http://127.0.0.1:5173';
async function page(path, status = 200) {
  const response = await fetch(`${origin}${path}`, { signal: AbortSignal.timeout(30000) });
  assert.equal(response.status, status, path);
  return (await response.text()).replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '');
}
const shelf = await page('/gallery/');
assert.ok(shelf.includes('aria-label="主导航"'));
assert.ok(shelf.includes('library-nav'));
assert.ok(!shelf.includes('aria-label="主导航"><div class="window-light'));
assert.ok(shelf.includes('搜索馆内文章'));
assert.ok(shelf.includes('最近收录'));
assert.ok(shelf.includes('分类书架') || shelf.includes('暂无主题分类'));
const search = await page('/gallery/?q=library-navigation-smoke');
assert.ok(search.includes('搜索结果'));
await page('/gallery/?read=nonexistent-library-navigation-smoke', 404);
await page('/gallery/?column=2147483647', 404);
const home = await page('/');
assert.ok(home.includes('shelf-scene'));
console.log('冒烟测试通过：馆内紧凑导航、书架入口、搜索、无效文章与分类、首页书架保留。');
console.log('未验证：非空分类阅读、点赞评论及浏览器视觉交互。');
