// Riot ID trên URL dạng "name-tag" giống op.gg: '#' là fragment nên không đặt trong path được.
// Tag của Riot không chứa '-' nên tách ở dấu '-' cuối cùng; name thì có thể chứa '-'.
import type { Server } from '$lib/api';

export function riotIdSlug(name: string, tag: string): string {
	return `${name}-${tag}`;
}

// Slug sai format (không có '-', thiếu name hoặc tag) => null.
export function parseRiotIdSlug(slug: string): { name: string; tag: string } | null {
	const i = slug.lastIndexOf('-');
	if (i <= 0 || i === slug.length - 1) {
		return null;
	}
	return { name: slug.slice(0, i), tag: slug.slice(i + 1) };
}

// Tag mặc định Riot gán theo server (người dùng gõ tên không kèm #tag thì dùng tag này).
export const DEFAULT_TAGS: Record<Server, string> = {
	BR: 'br1',
	EUNE: 'eune',
	EUW: 'euw',
	JP: 'jp1',
	KR: 'kr1',
	LAN: 'lan',
	LAS: 'las',
	ME: 'me1',
	NA: 'na1',
	OCE: 'oce',
	RU: 'ru',
	SG: 'sg2',
	TR: 'tr1',
	TW: 'tw2',
	VN: 'vn2'
};

export function playerHref(server: Server, name: string, tag: string): string {
	return `/lol/players/${server.toLowerCase()}/${encodeURIComponent(riotIdSlug(name, tag))}`;
}
