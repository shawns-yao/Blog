import { chromium } from 'playwright';
import { promises as fs } from 'node:fs';
import path from 'node:path';

const sourceDir = path.resolve('static/moments/scenes');
const inputs = (await fs.readdir(sourceDir)).filter(
	(name) => name.endsWith('.png') && name !== 'cat-paw.png'
);
const browser = await chromium.launch({ channel: 'msedge', headless: true });
const page = await browser.newPage();

for (const input of inputs) {
	const sourcePath = path.join(sourceDir, input);
	const dataUrl = `data:image/png;base64,${(await fs.readFile(sourcePath)).toString('base64')}`;
	const webp = await page.evaluate(async (src) => {
		const image = new Image();
		image.src = src;
		await image.decode();
		const canvas = document.createElement('canvas');
		canvas.width = image.naturalWidth;
		canvas.height = image.naturalHeight;
		canvas.getContext('2d').drawImage(image, 0, 0);
		return canvas.toDataURL('image/webp', 0.88);
	}, dataUrl);
	const outputPath = path.join(sourceDir, input.replace(/\.png$/, '.webp'));
	await fs.writeFile(outputPath, Buffer.from(webp.split(',')[1], 'base64'));
}

await browser.close();
