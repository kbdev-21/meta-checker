// GET /api/lol-data/* — dữ liệu tĩnh từ ddragon, không đổi trong 1 patch.
import { request, type RequestOptions } from './client';
import type { Champion, Item, Rune, Spell } from './types';

export function getChampions(opts?: RequestOptions): Promise<Champion[]> {
	return request<Champion[]>('GET', '/api/lol-data/champions', opts);
}

export function getItems(opts?: RequestOptions): Promise<Item[]> {
	return request<Item[]>('GET', '/api/lol-data/items', opts);
}

export function getSpells(opts?: RequestOptions): Promise<Spell[]> {
	return request<Spell[]>('GET', '/api/lol-data/spells', opts);
}

// Trả cả cây rune lẫn rune: phần tử có styleId = null là cây.
// Thứ tự: cây trước, rồi rune của từng cây theo thứ tự hàng.
export function getRunes(opts?: RequestOptions): Promise<Rune[]> {
	return request<Rune[]>('GET', '/api/lol-data/runes', opts);
}
