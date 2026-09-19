export type FooterThemeLink = {
	name: string;
	href: string;
};

export type FooterThemeSection = {
	title: string;
	links: FooterThemeLink[];
};

export type FooterThemeConfig = {
	sections: FooterThemeSection[];
	brandName: string;
	brandTagline: string;
	presenceConnectedText: string;
	presenceLoadingText: string;
};
