import { getApi } from '$lib/shared/clients/api';
import type { Column } from './types';

export const getColumns = async (fetcher?: typeof fetch): Promise<Column[]> => {
	const api = getApi(fetcher);
	const result = await api<Column[]>('/columns');
	return result ?? [];
};
