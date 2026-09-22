export type Category = {
	id: number;
	name: string;
	shortUrl: string;
	createdAt: string;
	updatedAt: string;
};

export type Column = {
	parentId: number | null;
	id: number;
	name: string;
	shortUrl: string;
	createdAt: string;
	updatedAt: string;
};
