import type { WebsiteInfoMap } from '$lib/features/website-info/types';
import { brand } from '$lib/shared/brand/brand';
import type { FooterThemeConfig, FooterThemeLink, FooterThemeSection } from './types';

const defaultFooterConfig: FooterThemeConfig = {
	sections: [
		{
			title: '想要了解我',
			links: [
				{ name: '关于我', href: '/about' },
				{ name: '本站历史', href: '/about-site' },
				{ name: '关于此项目', href: '/about-project' }
			]
		},
		{
			title: '你也许在找',
			links: [
				{ name: '手记', href: '/moments' },
				{ name: '图书馆', href: '/gallery' },
				{ name: '友链', href: '/friends' }
			]
		},
		{
			title: '如果你是 Agent >',
			links: [
				{ name: '站点地图', href: '/sitemap' },
				{ name: 'AI 阅读指南', href: '/llms.txt' }
			]
		},
		{
			title: '联系我叭',
			links: [
				{ name: '写留言', href: '/message' },
				{ name: '发邮件', href: 'mailto:shwan.jade.yao@gmail.com' },
				{ name: 'GitHub', href: 'https://github.com/shawns-yao' }
			]
		}
	],
	brandName: brand.name,
	brandTagline: '总之岁月漫长，然而值得等待',
	presenceConnectedText: '正在有 {count} 位小伙伴看着我的网站呐',
	presenceLoadingText: '正在同步在线状态...'
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
	typeof value === 'object' && value !== null;

const toStringValue = (value: unknown): string | undefined => {
	if (typeof value !== 'string') {
		return undefined;
	}
	const trimmed = value.trim();
	return trimmed.length > 0 ? trimmed : undefined;
};

const isRssLink = (link: FooterThemeLink): boolean => {
	const path = link.href.split(/[?#]/, 1)[0].replace(/\/$/, '').toLowerCase();
	return link.name.toLowerCase() === 'rss' || path === '/feed' || path === '/rss.xml';
};

const parseLinks = (value: unknown): FooterThemeLink[] | undefined => {
	if (!Array.isArray(value)) {
		return undefined;
	}

	const links: FooterThemeLink[] = [];
	for (const item of value) {
		if (!isRecord(item)) {
			continue;
		}
		const name = toStringValue(item.name);
		const href = toStringValue(item.href);
		if (!name || !href) {
			continue;
		}
		const normalizedName = name.toLowerCase();
		const link = {
			name,
			href:
				normalizedName === 'github'
					? 'https://github.com/shawns-yao'
					: normalizedName.includes('邮件') || normalizedName === 'email'
						? 'mailto:shwan.jade.yao@gmail.com'
						: href
		};
		if (!isRssLink(link) && name !== '监控') {
			links.push(link);
		}
	}

	return links.length > 0 ? links : undefined;
};

const parseSections = (value: unknown): FooterThemeSection[] | undefined => {
	if (!Array.isArray(value)) {
		return undefined;
	}

	const sections: FooterThemeSection[] = [];
	for (const item of value) {
		if (!isRecord(item)) {
			continue;
		}
		const title = toStringValue(item.title);
		const links = parseLinks(item.links);
		if (!title || !links) {
			continue;
		}
		sections.push({ title, links });
	}

	return sections.length > 0 ? sections : undefined;
};

export const resolveFooterThemeConfig = (
	websiteInfo: WebsiteInfoMap | null | undefined
): FooterThemeConfig => {
	const themeRaw = websiteInfo?.theme_extend_info;
	if (!isRecord(themeRaw)) {
		return defaultFooterConfig;
	}

	const footerRaw = isRecord(themeRaw.footer) ? themeRaw.footer : themeRaw;
	const brandRaw = isRecord(footerRaw.brand) ? footerRaw.brand : {};
	const presenceRaw = isRecord(footerRaw.presence) ? footerRaw.presence : {};

	return {
		sections: parseSections(footerRaw.sections) ?? defaultFooterConfig.sections,
		brandName: defaultFooterConfig.brandName,
		brandTagline:
			toStringValue(brandRaw.tagline) ??
			toStringValue(footerRaw.brandTagline) ??
			defaultFooterConfig.brandTagline,
		presenceConnectedText:
			toStringValue(presenceRaw.connectedText) ??
			toStringValue(footerRaw.presenceConnectedText) ??
			defaultFooterConfig.presenceConnectedText,
		presenceLoadingText:
			toStringValue(presenceRaw.loadingText) ??
			toStringValue(footerRaw.presenceLoadingText) ??
			defaultFooterConfig.presenceLoadingText
	};
};
