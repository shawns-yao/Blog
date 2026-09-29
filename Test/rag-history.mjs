import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const require = createRequire(new URL('../web/package.json', import.meta.url));
const { chromium } = require('playwright');
const entry = process.env.RAG_HISTORY_UI_URL || 'http://127.0.0.1:5179/';
const samples = [];
const checks = [];
const screenshots = [];
let failure;
let lastAsk = 0;
const browser = await chromium.launch({ channel: 'msedge', headless: true });
function isAsk(request) { return request.method() === 'POST' && new URL(request.url()).pathname.endsWith('/public/ask'); }

try {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1000 }, colorScheme: 'dark', reducedMotion: 'reduce' });
  const page = await context.newPage();
  const errors = [];
  page.on('pageerror', (error) => errors.push(error.message));
  await page.goto(entry, { waitUntil: 'domcontentloaded' });
  const statusPromise = page.waitForResponse((response) => new URL(response.url()).pathname.endsWith('/public/rag/status'));
  await page.getByRole('button', { name: '打开站内问答', exact: true }).click();
  const status = (await (await statusPromise).json()).data;
  assert.deepEqual(status.history, { maxRounds: 2, maxCharacters: 3000 });
  const dialog = page.getByRole('dialog', { name: '站内问答', exact: true });
  const textarea = dialog.getByRole('textbox', { name: '你的问题', exact: true });
  const questions = ['你好', '谢谢', '再见', '你好'];
  for (let index = 0; index < questions.length; index++) {
    const waitMs = 15500 - (Date.now() - lastAsk);
    if (waitMs > 0) await new Promise((resolve) => setTimeout(resolve, waitMs));
    const started = Date.now();
    lastAsk = started;
    const requestPromise = page.waitForRequest(isAsk);
    const responsePromise = page.waitForResponse((response) => isAsk(response.request()), { timeout: 100000 });
    await textarea.fill(questions[index]);
    await textarea.press('Enter');
    const body = (await requestPromise).postDataJSON();
    const response = await responsePromise;
    const answer = (await response.json()).data;
    const sample = { question: questions[index], historyMessages: body.history.length, status: answer?.status,
      mode: answer?.mode, passed: false, externalFailure: response.status() !== 200 ? 'network_or_environment' :
        answer?.status === 'temporarily_unavailable' ? 'model_or_environment' : null,
      durationMs: Date.now() - started };
    samples.push(sample);
    assert.equal(response.status(), 200);
    assert.equal(body.history.length, Math.min(index, 2) * 2);
    assert.equal(answer.status, 'answered', answer.reason);
    assert.equal(answer.mode, 'conversation');
    if (index === 3) assert.deepEqual(body.history.filter((item) => item.role === 'user').map((item) => item.content), ['谢谢', '再见']);
    await page.waitForFunction((count) => {
      const nodes = document.querySelectorAll('[data-message-role="assistant"]');
      return nodes.length === count && !nodes[count - 1].querySelector('[role="status"]');
    }, index + 1);
    sample.passed = true;
  }
  assert.equal(await dialog.locator('[data-message-role="user"]').count(), 4);
  assert.equal(await dialog.locator('[data-message-role="assistant"]').count(), 4);
  checks.push({ name: '模型只接收两轮，页面完整展示四轮', passed: true });
  await dialog.getByRole('button', { name: '关闭站内问答', exact: true }).click();
  await dialog.waitFor({ state: 'hidden' });
  await page.getByRole('button', { name: '打开站内问答', exact: true }).click();
  await dialog.getByRole('textbox', { name: '你的问题', exact: true }).waitFor();
  assert.equal(await dialog.locator('[data-message-role="assistant"]').count(), 4);
  checks.push({ name: '关闭重开保留页面历史', passed: true });
  for (const [name, viewport] of [['desktop', { width: 1440, height: 1000 }], ['mobile', { width: 390, height: 844 }]]) {
    await page.setViewportSize(viewport);
    await page.waitForFunction(() => {
      const node = document.querySelector('[role="dialog"]');
      if (!node) return false;
      const bounds = node.getBoundingClientRect();
      return document.documentElement.scrollWidth <= innerWidth && bounds.x >= -1 &&
        bounds.right <= innerWidth + 1 && bounds.height <= innerHeight + 1;
    });
    const box = await dialog.boundingBox();
    assert(box && box.x >= -1 && box.width <= viewport.width + 1 && box.height <= viewport.height + 1);
    const path = fileURLToPath(new URL(`../Image/figures/rag-history-${name}_20260929.png`, import.meta.url));
    await dialog.screenshot({ path });
    screenshots.push(path);
  }
  await page.reload({ waitUntil: 'domcontentloaded' });
  await page.getByRole('button', { name: '打开站内问答', exact: true }).click();
  await page.getByRole('textbox', { name: '你的问题', exact: true }).waitFor();
  assert.equal(await page.locator('[data-message-role="assistant"]').count(), 0);
  checks.push({ name: '刷新按页面内存生命周期清空记录', passed: true });
  assert.deepEqual(errors, []);
  await context.close();
} catch (error) { failure = error.message; }
finally {
  await browser.close();
  const external = samples.filter((sample) => sample.externalFailure).length;
  const report = { testType: '定向测试', entry, browser: '本地 Microsoft Edge',
    sampleFunnel: { original: 4, excluded: 4 - samples.length, valid: samples.length },
    externalFailures: { total: external, rate: samples.length ? external / samples.length : 0,
      testExecutionFailures: failure && !samples.some((sample) => !sample.passed) ? 1 : 0 },
    coreFunction: { denominator: samples.length - external, passed: samples.filter((sample) => sample.passed).length },
    checks, samples, screenshots, failure: failure ?? null,
    limitation: '验证实际前端遵循服务端历史窗口和页面记录生命周期，不作为检索质量评测' };
  await writeFile(fileURLToPath(new URL('./rag-history-results_20260929.json', import.meta.url)), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
  process.stdout.write(JSON.stringify(report, null, 2));
  if (failure) process.exitCode = 1;
}
