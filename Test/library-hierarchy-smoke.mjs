import assert from 'node:assert/strict';

const origin = process.env.CONTENT_TEST_ORIGIN || 'http://127.0.0.1:5173';
const api = process.env.CONTENT_TEST_API || 'http://127.0.0.1:8080/api/v2';
async function json(path) {
  const response = await fetch(`${api}${path}`, { signal: AbortSignal.timeout(30000) });
  assert.equal(response.status, 200, path);
  const body = await response.json();
  assert.equal(body.code, 0, path);
  return body.data;
}
async function html(path) {
  const response = await fetch(`${origin}${path}`, { signal: AbortSignal.timeout(30000) });
  assert.equal(response.status, 200, path);
  return (await response.text()).replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '');
}
const columns = await json('/columns');
const roots = columns.filter((item) => item.parentId === null);
const children = columns.filter((item) => item.parentId !== null);
const shelf = await html('/gallery/');
assert.ok(shelf.includes('一级分类书架'));
let articles = 0;
for (const root of roots) {
  const page = await html(`/gallery/?column=${root.id}`);
  assert.ok(page.includes(root.name));
  assert.ok(page.includes('二级分类'));
  const list = await json(`/moments?contentKind=article&columnId=${root.id}&includeChildren=true&sort=newest`);
  for (let index = 1; index < list.items.length; index++) {
    assert.ok(Date.parse(list.items[index - 1].createdAt) >= Date.parse(list.items[index].createdAt));
  }
  for (const article of list.items) {
    const reader = await html(`/gallery/?read=${encodeURIComponent(article.shortUrl)}`);
    assert.ok(reader.includes('open-book'));
    assert.ok(reader.includes('返回时间列表'));
    articles++;
  }
}
for (const child of children) {
  assert.ok(roots.some((root) => root.id === child.parentId));
  const reader = await html(`/gallery/?column=${child.id}`);
  assert.ok(reader.includes('open-book'));
  assert.ok(reader.includes(child.name));
}
console.log(`定向测试通过：${roots.length} 个一级分类、${articles} 次文章阅读、时间排序、${children.length} 个二级分类入口。`);
if (!children.length) console.log('未验证：二级分类非空阅读，当前没有二级分类样本。');
console.log('未验证：浏览器视觉、客户端交互、后台分类写入；未修改数据。');
