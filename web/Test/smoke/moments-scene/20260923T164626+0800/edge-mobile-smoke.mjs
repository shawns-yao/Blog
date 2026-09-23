import { chromium } from 'playwright';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const runDir = path.dirname(fileURLToPath(import.meta.url));
const browser = await chromium.launch({ channel: 'msedge', headless: true });
const context = await browser.newContext({
	viewport: { width: 390, height: 844 },
	colorScheme: 'dark',
	locale: 'zh-CN'
});
const page = await context.newPage();
const runtimeErrors = [];

await page.route('**/api/v2/**', (route) =>
	route.fulfill({
		status: 200,
		contentType: 'application/json',
		body: JSON.stringify({ code: 0, data: {} })
	})
);

page.on('pageerror', (error) => runtimeErrors.push(`pageerror: ${error.message}`));
page.on('console', (message) => {
	if (message.type() === 'error') runtimeErrors.push(`console: ${message.text()}`);
});

const response = await page.goto('http://127.0.0.1:5173/moments', {
	waitUntil: 'networkidle',
	timeout: 60_000
});
await page.waitForSelector('.moment-book-room', { timeout: 30_000 });
await page.waitForTimeout(500);
await page.screenshot({
	path: path.join(runDir, 'edge-mobile-390x844.png'),
	fullPage: false
});

const directoryButton = page.getByRole('button', { name: '目录', exact: true });
await directoryButton.click();
await page.locator('#moment-mobile-directory').waitFor({ state: 'visible' });
await page.waitForTimeout(400);

const checks = {
	httpOk: response?.ok() ?? false,
	roomVisible: await page.locator('.moment-book-room').isVisible(),
	bookFrameHiddenAtMobile: !(await page.locator('.open-book-frame').isVisible()),
	contentVisible: await page.locator('.moment-book-page').isVisible(),
	mobileDirectoryVisible: await page.locator('#moment-mobile-directory').isVisible()
};

await page.screenshot({
	path: path.join(runDir, 'edge-mobile-directory-390x844.png'),
	fullPage: false
});

console.log(JSON.stringify({ checks, runtimeErrors }, null, 2));
await browser.close();
