import { chromium } from 'playwright';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const runDir = path.dirname(fileURLToPath(import.meta.url));
const viewportWidth = Number(process.argv[2] ?? 1440);
const viewportHeight = Number(process.argv[3] ?? 900);
const mockedHour = Number(process.argv[4] ?? Number.NaN);
const mockedMinute = Number(process.argv[5] ?? 0);
const browser = await chromium.launch({ channel: 'msedge', headless: true });
const context = await browser.newContext({
	viewport: { width: viewportWidth, height: viewportHeight },
	colorScheme: 'dark',
	locale: 'zh-CN'
});

if (Number.isFinite(mockedHour)) {
	await context.addInitScript(
		({ hour, minute }) => {
			const RealDate = Date;
			const fixedNow = new RealDate();
			fixedNow.setHours(hour, minute, 0, 0);
			class MockDate extends RealDate {
				constructor(...args) {
					super(...(args.length > 0 ? args : [fixedNow.getTime()]));
				}

				static now() {
					return fixedNow.getTime();
				}
			}
			window.Date = MockDate;
		},
		{ hour: mockedHour, minute: mockedMinute }
	);
}
const page = await context.newPage();
const runtimeErrors = [];
let pageStatus = null;

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

const momentsUrl = 'http://127.0.0.1:5173/moments';
page.on('response', (response) => {
	if (response.url().replace(/\/$/, '') === momentsUrl) pageStatus = response.status();
});

const response = await page.goto(momentsUrl, { waitUntil: 'networkidle', timeout: 60_000 });
await page.waitForSelector('.moment-book-room', { timeout: 30_000 });
await page.waitForTimeout(900);

const expectedScene = Number.isFinite(mockedHour)
	? [
			[0, 'midnight'],
			[330, 'dawn'],
			[480, 'morning'],
			[690, 'noon'],
			[900, 'afternoon'],
			[1080, 'sunset'],
			[1230, 'night'],
			[1380, 'late-night']
		]
			.toReversed()
			.find(([minute]) => mockedHour * 60 + mockedMinute >= minute)?.[1]
	: null;

const checks = {
	httpOk: response?.ok() ?? (pageStatus !== null && pageStatus >= 200 && pageStatus < 400),
	roomVisible: await page.locator('.moment-book-room').isVisible(),
	bookFrameVisible: await page.locator('.open-book-frame').isVisible(),
	directoryVisible: await page.locator('.moment-book-directory').isVisible(),
	contentVisible: await page.locator('.moment-book-page').isVisible(),
	sceneMatches:
		expectedScene === null ||
		(await page.locator('.moment-book-room').getAttribute('data-scene')) === expectedScene,
	lateNightPawMatches:
		expectedScene === 'late-night'
			? await page.locator('.cat-paw').isVisible()
			: (await page.locator('.cat-paw').count()) === 0,
	nightShootingStarMatches:
		expectedScene === 'night'
			? (await page.locator('.shooting-star').count()) === 1
			: (await page.locator('.shooting-star').count()) === 0
};

await page.screenshot({
	path: path.join(
		runDir,
		`edge-desktop-${viewportWidth}x${viewportHeight}${Number.isFinite(mockedHour) ? `-${String(mockedHour).padStart(2, '0')}${String(mockedMinute).padStart(2, '0')}` : ''}.png`
	),
	fullPage: false
});

console.log(JSON.stringify({ checks, runtimeErrors }, null, 2));
await browser.close();
