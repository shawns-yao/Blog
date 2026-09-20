import assert from 'node:assert/strict';

const origin = process.env.CONTENT_TEST_ORIGIN || 'http://127.0.0.1:5173';
async function read(path) {
  const response = await fetch(`${origin}${path}`, { signal: AbortSignal.timeout(60000) });
  assert.equal(response.status, 200, path);
  return response.text();
}

const html = await read('/moments/');
assert.match(html, /id="moments-heading"/);
assert.match(html, /aria-label="主导航"/);
const activeLink = html.match(/<a[^>]*href="([^"]*)"[^>]*aria-current="page"[^>]*>手记<\/a>/);
assert.ok(activeLink, '手记导航应处于选中状态');
assert.equal(new URL(activeLink[1], `${origin}/moments/`).pathname.replace(/\/$/, ''), '/moments');
assert.match(html, /name="q"/);
assert.ok(!html.includes('desktop-shelf-header hidden md:flex'));

const query = 'framework-no-match-20260920';
const searched = await read(`/moments/?q=${query}`);
assert.ok(searched.includes(query));
assert.ok(searched.includes('清除搜索'));
assert.ok(searched.includes('没有找到相关手记'));

const redirect = await fetch(`${origin}/moments/?page=2&q=${query}`, { redirect: 'manual' });
assert.equal(redirect.status, 308);
assert.equal(redirect.headers.get('location'), `/moments/page/2/?q=${query}`);

console.log('定向测试通过：手记框架、导航选中、搜索空状态、分页保留搜索条件。');
console.log('未验证：非空文字图片列表、多页导航、浏览器视觉及交互；未创建或修改内容。');
