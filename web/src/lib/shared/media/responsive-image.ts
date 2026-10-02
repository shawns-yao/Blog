const imageWidths = [320, 640, 1280];

/** Only local uploads use our image endpoint; external images retain their URLs. */
export function responsiveImage(src: string, sourceWidth?: number, fallbackWidth = 640) {
	if (!/^\/uploads\/pictures\/[^?#]+\.(?:jpe?g|png|webp)$/i.test(src)) {
		return { src, srcset: undefined as string | undefined };
	}
	const intrinsicWidth = sourceWidth && sourceWidth > 0 ? sourceWidth : undefined;
	const candidates = imageWidths.map((preset) => ({
		preset,
		width: intrinsicWidth ? Math.min(preset, intrinsicWidth) : preset
	}));
	const unique = candidates.filter(
		(candidate, index) => candidates.findIndex((item) => item.width === candidate.width) === index
	);
	return {
		src: `${src}?width=${fallbackWidth}`,
		srcset: unique.map(({ preset, width }) => `${src}?width=${preset} ${width}w`).join(', ')
	};
}
