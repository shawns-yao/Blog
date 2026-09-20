export function libraryPath(
	options: { column?: number; read?: string; page?: number; q?: string } = {}
): `/gallery/${string}` {
	const query = new URLSearchParams();
	if (options.column) query.set('column', String(options.column));
	if (options.read) query.set('read', options.read);
	if (options.page && options.page > 1) query.set('page', String(options.page));
	if (options.q) query.set('q', options.q);
	return `/gallery/${query.size ? `?${query}` : ''}`;
}
