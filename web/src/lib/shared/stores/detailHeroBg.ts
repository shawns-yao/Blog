import { writable } from 'svelte/store';

/** Stores the hero background image URL for detail pages (moment cover). */
export const detailHeroBgSrc = writable<string>('');
