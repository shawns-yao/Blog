import assert from 'node:assert/strict';
import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const entry = process.env.RAG_TEST_API || 'http://127.0.0.1:18089/api/v2';
const expected = process.env.RAG_TEST_EXPECTED_PROVIDER || 'gpt';
const order = ['gpt', 'grok', 'gemini', 'opencode_go', 'deepseek'];
const selected = process.env.RAG_CHANNEL_CASES?.split(',').map(Number) || [0, 1];
const questions = ['你好', '根据《星舟项目运行手册》，电池巡检周期和补充充电阈值是多少？'];
assert(order.includes(expected));
assert(selected.length > 0 && selected.every((index) => index === 0 || index === 1));
assert(['127.0.0.1', 'localhost'].includes(new URL(entry).hostname));
assert.equal(new URL(entry).port, '18089', '仅用于独立测试服务');
const sessionId = crypto.randomUUID();
const samples = [];
const checks = [];
const screenshots = [];
let executionFailure;

async function bootstrap() {
  const account = `rag-channels-${crypto.randomUUID().slice(0, 8)}`;
  const password = crypto.randomUUID();
  let token = '';
  async function request(path, body) {
    const response = await fetch(`${entry}${path}`, { method: body ? 'POST' : 'GET',
      headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
      body: body ? JSON.stringify(body) : undefined, signal: AbortSignal.timeout(95000) });
    const payload = await response.json();
    assert.equal(response.status, 200);
    assert.equal(payload.code, 0, payload.msg);
    return payload.data;
  }
  await request('/auth/register', { username: account, nickname: '模型链路验证', email: `${account}@example.invalid`, password });
  token = (await request('/auth/login', { credential: account, password })).token;
  const settings = await request('/admin/rag/settings');
  assert.deepEqual(settings.chatChannels.map((channel) => channel.name), order);
  assert.deepEqual(settings.chatChannels.map((channel) => channel.priority), [1, 2, 3, 4, 5]);
  assert.equal(settings.chatChannels.at(-1).default, true);
  assert(settings.chatChannels.every((channel) => channel.configured));
  assert(!/apiKey|https?:\/\//i.test(JSON.stringify(settings)));
  checks.push({ name: '五级非敏感配置通过真实管理员接口读取', passed: true });
  const article = await request('/moments/', { title: '星舟项目运行手册',
    content: '# 电池巡检\n\n电池每 14 天巡检一次；电量低于 30% 时补充充电。日志保留 90 天。\n',
    shortUrl: 'rag-channels-inspection', isPublished: true, extInfo: { contentKind: 'article' } });
  const until = Date.now() + 150000;
  let ready = false;
  while (Date.now() < until) {
    const list = await request('/admin/rag/documents?pageSize=100');
    if (list.items.some((item) => item.momentId === article.id && item.status === 'ready')) { ready = true; break; }
    await new Promise((resolve) => setTimeout(resolve, 2000));
  }
  assert(ready, '通过真实内容入口创建的文档必须完成索引');
  checks.push({ name: '独立文档通过项目内容入口索引', passed: true });

  const require = createRequire(new URL('../web/package.json', import.meta.url));
  const { chromium } = require('playwright');
  const browser = await chromium.launch({ channel: 'msedge', headless: true });
  try {
    const context = await browser.newContext({ viewport: { width: 1440, height: 1000 }, reducedMotion: 'reduce' });
    const page = await context.newPage();
    const admin = process.env.RAG_TEST_ADMIN || 'http://127.0.0.1:5879';
    await page.goto(`${admin}/sign-in?r=/rag`, { waitUntil: 'domcontentloaded', timeout: 60000 });
    await page.getByPlaceholder('请输入账号或邮箱').fill(account);
    await page.getByPlaceholder('请输入密码', { exact: true }).fill(password);
    await page.getByRole('button', { name: /^登\s*录$/ }).click();
    await page.getByRole('heading', { name: /^RAG 知识库/ }).waitFor({ timeout: 60000 });
    await page.getByText('检索配置', { exact: true }).click();
    const priorityList = page.getByRole('list', { name: '可调整语言模型优先级' });
    for (const label of ['GPT', 'Grok', 'Gemini', 'OpenCode Go'])
      await priorityList.getByText(label, { exact: true }).waitFor();
    await page.locator('[data-channel="default"]').getByText('DeepSeek 官方', { exact: true }).waitFor();
    for (const [name, viewport] of [['desktop', { width: 1440, height: 1000 }], ['mobile', { width: 390, height: 844 }]]) {
      await page.setViewportSize(viewport);
      await page.waitForFunction(() => document.documentElement.scrollWidth <= innerWidth);
      if (name === 'mobile') {
        await page.waitForFunction(() => {
          const labels = [...document.querySelectorAll('ol[aria-label="可调整语言模型优先级"] li')];
          const first = labels[0];
          const second = labels[1];
          return first && second && Math.abs(first.getBoundingClientRect().x - second.getBoundingClientRect().x) < 1;
        });
      }
      const path = fileURLToPath(new URL(`../Image/figures/rag-model-priority-${name}_20260929.png`, import.meta.url));
      await page.screenshot({ path, fullPage: true, animations: 'disabled' });
      screenshots.push(path);
    }
    checks.push({ name: '后台真实登录与五级模型桌面、移动端展示', passed: true });
    await context.close();
  } finally { await browser.close(); }
}

if (process.env.RAG_CHANNEL_BOOTSTRAP === 'true') {
  try { await bootstrap(); } catch (error) { executionFailure = error.message; }
}
for (const index of selected) {
  const question = questions[index];
  if (executionFailure) break;
  if (index > 0) await new Promise((resolve) => setTimeout(resolve, 15500));
  const started = Date.now();
  const sample = { question, passed: false, externalFailure: null };
  samples.push(sample);
  try {
    const response = await fetch(`${entry}/public/ask`, { method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question, sessionId }), signal: AbortSignal.timeout(95000) });
    const payload = await response.json();
    const answer = payload.data;
    Object.assign(sample, { httpStatus: response.status, status: answer?.status, mode: answer?.mode,
      trace: answer?.trace, citationCount: answer?.citations?.length ?? 0 });
    if (response.status !== 200) sample.externalFailure = 'network_or_environment';
    else if (answer?.status === 'temporarily_unavailable') sample.externalFailure = 'model_or_environment';
    else if (answer?.trace?.answerProvider !== expected && answer?.trace?.answerFailures?.some((failure) =>
      failure.provider === expected && ['timeout', 'provider_unavailable'].includes(failure.reason)))
      sample.externalFailure = 'preferred_provider_unavailable';
    assert.equal(response.status, 200);
    assert.equal(payload.code, 0);
    assert.equal(answer.status, 'answered', answer.reason);
    assert.equal(answer.trace.answerProvider, expected);
    assert.deepEqual(answer.trace.answerAttempts, order.slice(0, order.indexOf(expected) + 1));
    assert.equal(answer.mode, index === 0 ? 'conversation' : 'grounded');
    if (index === 0) assert.equal(answer.citations.length, 0);
    else {
      assert(answer.answer.includes('14') && answer.answer.includes('30'));
      assert(answer.citations.some((citation) => citation.title === '星舟项目运行手册'));
    }
    sample.passed = true;
  } catch (error) {
    sample.failure = error.message;
    if (!('httpStatus' in sample)) sample.externalFailure = 'network_or_environment';
  }
  sample.durationMs = Date.now() - started;
}
const external = samples.filter((sample) => sample.externalFailure).length;
const report = { testType: '定向测试', entry, expectedProvider: expected,
  sampleFunnel: { original: selected.length, excluded: selected.length - samples.length, valid: samples.length },
  externalFailures: { total: external, rate: samples.length ? external / samples.length : 0, testExecutionFailures: executionFailure ? 1 : 0 },
  coreFunction: { denominator: samples.length - external, passed: samples.filter((sample) => !sample.externalFailure && sample.passed).length,
    qualification: external / samples.length > 0.1 ? '受外部因素影响，仅供参考' : null },
  samples, checks, screenshots, executionFailure: executionFailure ?? null,
  limitation: '通过关闭端口制造先前通道真实连接失败，验证实际项目入口的有序降级和引用；不作为模型效果评测。' };
await writeFile(fileURLToPath(new URL(process.env.RAG_CHANNEL_REPORT_FILE || `./rag-chat-channels-${expected}-results_20260929.json`, import.meta.url)), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
process.stdout.write(JSON.stringify(report, null, 2));
if (executionFailure || samples.some((sample) => !sample.passed)) process.exitCode = 1;
