// Global state cho lol-data: public, read-only, không đổi trong 1 patch nên fetch 1 lần
// rồi dùng cả phiên. adapter-static không có server runtime nên state ở module scope
// không bị share giữa các user.
// Chỉ giữ Map vì chỗ nào cũng tra theo id; cần list thì [...lolData.championById.values()]
// (Map giữ nguyên thứ tự backend trả về).
import { getChampions, getItems, getRunes, getSpells } from '$lib/api';
import type { Champion, Item, Rune, Spell } from '$lib/api';

export const lolData = $state({
	championById: new Map<number, Champion>(),
	itemById: new Map<number, Item>(),
	spellById: new Map<number, Spell>(),
	runeById: new Map<number, Rune>(),
	isLoaded: false,
	error: null as unknown
});

let pending: Promise<void> | undefined;

// Gọi bao nhiêu lần cũng chỉ fetch 1 lượt. Không throw vì được gọi trong $effect;
// lỗi nằm ở lolData.error, gọi lại loadLolData() là thử lại.
export function loadLolData(): Promise<void> {
	return (pending ??= fetchAll());
}

async function fetchAll(): Promise<void> {
	try {
		const [champions, items, spells, runes] = await Promise.all([
			getChampions(),
			getItems(),
			getSpells(),
			getRunes()
		]);
		// Gán nguyên Map mới (không mutate) vì $state không theo dõi thay đổi bên trong Map.
		lolData.championById = new Map(champions.map((c) => [c.id, c]));
		lolData.itemById = new Map(items.map((i) => [i.id, i]));
		lolData.spellById = new Map(spells.map((s) => [s.id, s]));
		lolData.runeById = new Map(runes.map((r) => [r.id, r]));
		lolData.isLoaded = true;
	} catch (e) {
		lolData.error = e;
		pending = undefined;
	}
}
