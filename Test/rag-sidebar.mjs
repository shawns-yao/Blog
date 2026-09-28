import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import assert from 'node:assert/strict';

const require = createRequire(new URL('../web/package.json', import.meta.url));
const { chromium } = require('playwright');
const entry = process.env.RAG_UI_TEST_URL || 'http://127.0.0.1:5173/';
const browser = await chromium.launch({ channel: 'msedge', headless: true });
const results = [];

try {
  for (const scenario of [
    { name: 'desktop', width: 1440, height: 900 },
    { name: 'mobile', width: 390, height: 844 }
  ]) {
    const context = await browser.newContext({
      viewport: { width: scenario.width, height: scenario.height },
      reducedMotion: 'reduce'
    });
    const page = await context.newPage();
    const requests = [];
    const errors = [];
    page.on('request', (request) => {
      if (/\/public\/(rag\/status|ask)(?:\?|$)/.test(request.url())) {
        requests.push(new URL(request.url()).pathname);
      }
    });
    page.on('pageerror', (error) => errors.push(error.message));
    await page.goto(entry, { waitUntil: 'domcontentloaded' });
    const trigger = page.getByRole('button', { name: '打开站内问答', exact: true });
    await trigger.click();
    const dialog = page.getByRole('dialog', { name: '站内问答', exact: true });
    const textarea = page.getByRole('textbox', { name: '你的问题', exact: true });
    await textarea.waitFor();
    await page.waitForFunction(() => document.activeElement?.id === 'rag-question');
    await textarea.fill('怎样查找与 RAG 相关的文章？');
    const unavailable = dialog.getByText(
      /^(暂时无法连接问答服务，可以先用搜索查找文章与手记。|问答服务尚未开放，可以先用搜索查找文章与手记。|公开内容正在准备中，请稍后重试。)$/
    ).first();
    await unavailable.waitFor();
    assert.equal(await dialog.getByRole('button', { name: '发送问题', exact: true }).isDisabled(), true);
    const count = requests.length;
    await textarea.fill('保留这条问题草稿');
    assert.equal(requests.length, count, '输入不得触发问答请求');
    for (let i = 0; i < 12; i++) {
      await page.keyboard.press('Tab');
      assert.equal(await dialog.evaluate((node) => node.contains(document.activeElement)), true);
    }
    const statusCount = requests.filter((path) => path.endsWith('/rag/status')).length;
    await dialog.getByRole('button', { name: '重新检查', exact: true }).click();
    await unavailable.waitFor();
    assert.equal(requests.filter((path) => path.endsWith('/rag/status')).length, statusCount + 1);
    assert.equal(requests.filter((path) => path.endsWith('/ask')).length, 0);
    const bounds = await dialog.boundingBox();
    assert(bounds && bounds.x >= 0 && bounds.y >= 0);
    assert(bounds.x + bounds.width <= scenario.width + 1 && bounds.height <= scenario.height + 1);
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
    const screenshot = fileURLToPath(new URL(
      `../Image/figures/rag-integration-${scenario.name}_20260928.png`, import.meta.url
    ));
    await page.screenshot({ path: screenshot, fullPage: true });
    await page.keyboard.press('Escape');
    await dialog.waitFor({ state: 'hidden' });
    assert.equal(await trigger.evaluate((node) => document.activeElement === node), true);
    await trigger.click();
    await textarea.waitFor();
    assert.equal(await textarea.inputValue(), '保留这条问题草稿');
    await dialog.getByRole('button', { name: '清空草稿', exact: true }).click();
    assert.equal(await textarea.inputValue(), '');
    await dialog.getByRole('button', { name: '使用站内搜索', exact: true }).click();
    await dialog.waitFor({ state: 'hidden' });
    await page.getByRole('button', { name: '关闭搜索弹窗', exact: true }).waitFor();
    assert.deepEqual(errors, []);
    results.push({ viewport: scenario.name, passed: true, askRequests: 0 });
    await context.close();
  }
  process.stdout.write(JSON.stringify({
    testType: '定向测试', entry, results,
    limitation: '仅验证真实前端的服务不可用路径，未执行后端检索或生成'
  }, null, 2));
} finally {
  await browser.close();
}
