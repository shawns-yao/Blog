import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const api = process.env.RAG_PRIORITY_API || 'http://127.0.0.1:8080/api/v2';
const admin = process.env.RAG_PRIORITY_ADMIN || 'http://127.0.0.1:5799';
const token = process.env.RAG_VERIFY_ADMIN_TOKEN;
assert(token, '请在本机环境中提供现有管理员令牌 RAG_VERIFY_ADMIN_TOKEN；脚本不创建账号或令牌');
assert.equal(new URL(api).hostname, '127.0.0.1');
assert.equal(new URL(admin).hostname, '127.0.0.1');
const checks = [];
const screenshots = [];
let original;
let failure;
let browser;

async function request(path, body, expectedSuccess = true) {
  const response = await fetch(`${api}${path}`, {
    method: body === undefined ? 'GET' : 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal: AbortSignal.timeout(10000),
  });
  const payload = await response.json();
  assert.equal(response.status, 200);
  if (expectedSuccess) assert.equal(payload.code, 0, payload.msg);
  else assert.notEqual(payload.code, 0);
  return payload.data;
}

function assertChannels(settings, priority) {
  assert.deepEqual(settings.chatChannels.filter((channel) => !channel.default).map((channel) => channel.name), priority);
  const fallback = settings.chatChannels.at(-1);
  assert.equal(fallback.name, 'deepseek');
  assert.equal(fallback.default, true);
  assert.equal(settings.chatChannels.filter((channel) => channel.default).length, 1);
  assert.deepEqual(settings.chatChannels.map((channel) => channel.priority), [1, 2, 3, 4, 5]);
  assert(!/apiKey|https?:\/\//i.test(JSON.stringify(settings)));
}

try {
  const access = await request('/auth/access-info');
  assert.equal(access.user.isAdmin, true);
  const baseline = await request('/admin/rag/settings');
  original = baseline.chatChannels.filter((channel) => !channel.default).map((channel) => channel.name);
  assertChannels(baseline, original);
  const gpt = baseline.chatChannels.find((channel) => channel.name === 'gpt');
  assert(gpt && gpt.configured);
  assert.equal(gpt.model, 'gpt-6-sol');
  assert.equal(gpt.reasoningEffort, 'medium');
  checks.push({ name: '真实认证与五级非敏感配置', passed: true });

  const reordered = [...original].reverse();
  assertChannels(await request('/admin/rag/chat-priority', { priority: reordered }), reordered);
  assertChannels(await request('/admin/rag/settings'), reordered);
  assert.deepEqual((await request('/admin/rag/settings')).tuning, baseline.tuning);
  checks.push({ name: '排序保存、重新读取及检索参数保持一致', passed: true });
  for (const priority of [null, [], ['gpt', 'gpt', 'gemini', 'opencode_go'], ['gpt', 'grok', 'gemini', 'default'],
    ['gpt', 'grok', 'gemini', 'deepseek'], ['gpt', 'grok', 'gemini', 'unknown']]) {
    await request('/admin/rag/chat-priority', { priority }, false);
    assertChannels(await request('/admin/rag/settings'), reordered);
  }
  checks.push({ name: '拒绝缺失、重复、未知通道及 default 排序', passed: true });
  await request('/admin/rag/chat-priority', { priority: original });

  const require = createRequire(new URL('../web/package.json', import.meta.url));
  const { chromium } = require('playwright');
  browser = await chromium.launch({ channel: 'msedge', headless: true });
  const context = await browser.newContext({ viewport: { width: 1440, height: 1000 }, reducedMotion: 'reduce' });
  await context.addInitScript(({ existingToken, user, roles, permissions }) => {
    localStorage.setItem('token', existingToken);
    localStorage.setItem('user', JSON.stringify({ ...user, roles, permissions }));
  }, { existingToken: token, ...access });
  const page = await context.newPage();
  const errors = [];
  page.on('pageerror', () => errors.push('page_runtime_error'));
  page.on('request', (req) => assert(!/\/public\/ask$/.test(new URL(req.url()).pathname), '优先级验证禁止问答请求'));
  await page.goto(`${admin}/rag`, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.getByText('检索配置', { exact: true }).click();
  const list = page.getByRole('list', { name: '可调整语言模型优先级' });
  await list.locator('li').first().waitFor();
  assert.equal(await list.locator('li').count(), 4);
  assert.equal(await list.locator('[data-channel="default"]').count(), 0);

  const first = list.locator('li').nth(0);
  const second = list.locator('li').nth(1);
  const from = await first.locator('.rag-priority-handle').boundingBox();
  const to = await second.boundingBox();
  assert(from && to);
  await page.mouse.move(from.x + from.width / 2, from.y + from.height / 2);
  await page.mouse.down();
  await page.mouse.move(to.x + 50, to.y + to.height - 4, { steps: 20 });
  await page.mouse.up();
  const swapped = [original[1], original[0], ...original.slice(2)];
  await page.waitForFunction((expected) => JSON.stringify([...document.querySelectorAll('ol[aria-label="可调整语言模型优先级"] li')].map((node) => node.dataset.channel)) === JSON.stringify(expected), swapped);
  await page.getByRole('button', { name: '保存优先级', exact: true }).click();
  await page.getByText('优先级已保存，新请求将按此顺序调用。', { exact: true }).waitFor();
  assertChannels(await request('/admin/rag/settings'), swapped);
  await page.reload({ waitUntil: 'domcontentloaded' });
  await page.getByText('检索配置', { exact: true }).click();
  await page.waitForFunction((expected) => JSON.stringify([...document.querySelectorAll('ol[aria-label="可调整语言模型优先级"] li')].map((node) => node.dataset.channel)) === JSON.stringify(expected), swapped);
  checks.push({ name: '实际后台拖拽、保存、刷新后保留顺序', passed: true });

  const labels = { gpt: 'GPT', grok: 'Grok', gemini: 'Gemini', opencode_go: 'OpenCode Go' };
  await page.getByRole('button', { name: `上移 ${labels[original[0]]}`, exact: true }).focus();
  await page.keyboard.press('Enter');
  await page.getByRole('button', { name: '保存优先级', exact: true }).click();
  await page.getByText('优先级已保存，新请求将按此顺序调用。', { exact: true }).waitFor();
  assertChannels(await request('/admin/rag/settings'), original);
  checks.push({ name: '键盘调整及固定兜底', passed: true });

  const section = page.locator('section[aria-labelledby="rag-chat-priority-heading"]');
  for (const [name, viewport] of [['desktop', { width: 1440, height: 1000 }], ['mobile', { width: 390, height: 844 }]]) {
    await page.setViewportSize(viewport);
    await page.waitForFunction(() => document.documentElement.scrollWidth <= innerWidth);
    await section.scrollIntoViewIfNeeded();
    const bounds = await section.boundingBox();
    assert(bounds && bounds.x >= 0 && bounds.x + bounds.width <= viewport.width + 1);
    const path = fileURLToPath(new URL(`../Image/figures/rag-chat-priority-${name}_20260929.png`, import.meta.url));
    await section.screenshot({ path, animations: 'disabled' });
    screenshots.push(path);
  }
  assert.deepEqual(errors, []);
  checks.push({ name: '桌面和移动端无横向溢出及运行错误', passed: true });
  await context.close();
} catch (error) {
  failure = error.message;
} finally {
  if (browser) await browser.close();
  if (original) {
    try { await request('/admin/rag/chat-priority', { priority: original }); }
    catch { failure = 'priority_restore_failed'; }
  }
  const report = { testType: '定向测试', entry: admin, browser: '本地 Microsoft Edge',
    checks, screenshots, failure: failure ?? null,
    limitation: '只验证模型配置和优先级管理，不调用问答、嵌入或重排序，不写入或发布文章；不作为模型效果验收' };
  await writeFile(fileURLToPath(new URL('./rag-priority-results_20260929.json', import.meta.url)), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
  process.stdout.write(JSON.stringify(report, null, 2));
  if (failure) process.exitCode = 1;
}
