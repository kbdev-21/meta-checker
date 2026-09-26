// Logic random của trang Randomizer: chọn ngẫu nhiên champion / rune / summoner spell.
import type { Position, Rune, RuneStat, Spell } from '$lib/api';

// 5 lane được random; UNK không phải lane.
export const LANES: Position[] = ['TOP', 'JGL', 'MID', 'ADC', 'SPT'];

// Rune chỉ gồm cây chính, keystone, cây phụ (không random các rune nhỏ / shard).
export type RuneChoice = {
	primaryStyle: number;
	keyRune: number;
	subStyle: number;
};

export type SpellPair = [number, number];

// Summoner spell dùng được ở Summoner's Rift. Backend lưu nguyên summoner.json của ddragon
// (cả spell ARAM / Arena...) nên phải lọc theo slug.
const SR_SPELL_SLUGS = new Set([
	'SummonerFlash',
	'SummonerDot', // Ignite
	'SummonerTeleport',
	'SummonerHeal',
	'SummonerBarrier',
	'SummonerExhaust',
	'SummonerHaste', // Ghost
	'SummonerBoost', // Cleanse
	'SummonerSmite'
]);
const SMITE_SLUG = 'SummonerSmite';

export function randomItem<T>(items: readonly T[]): T {
	return items[Math.floor(Math.random() * items.length)];
}

// Rune bất kỳ: cây chính, keystone thuộc cây đó, cây phụ khác cây chính.
export function randomRune(runeById: Map<number, Rune>): RuneChoice {
	const runes = [...runeById.values()];
	const trees = runes.filter((r) => r.styleId === null);
	const primary = randomItem(trees);
	// slot 0 = hàng keystone.
	const keystones = runes.filter((r) => r.styleId === primary.id && r.slot === 0);
	const sub = randomItem(trees.filter((t) => t.id !== primary.id));
	return { primaryStyle: primary.id, keyRune: randomItem(keystones).id, subStyle: sub.id };
}

// 2 spell khác nhau. JGL luôn có Smite (không có thì không farm rừng được) + 1 spell random;
// lane khác không bao giờ ra Smite.
export function randomSpells(position: Position, spellById: Map<number, Spell>): SpellPair {
	const spells = [...spellById.values()].filter((s) => SR_SPELL_SLUGS.has(s.slug));
	const smite = spells.find((s) => s.slug === SMITE_SLUG);
	const others = spells.filter((s) => s.slug !== SMITE_SLUG);
	const first = position === 'JGL' && smite ? smite : randomItem(others);
	const second = randomItem(others.filter((s) => s.id !== first.id));
	return [first.id, second.id];
}

// Gộp các bộ rune cùng (cây chính, keystone, cây phụ) như tab của RunePanel, lấy n nhóm
// nhiều trận nhất.
export function topRuneChoices(runes: RuneStat[], n: number): RuneChoice[] {
	const byKey = new Map<string, RuneChoice & { games: number }>();
	for (const r of runes) {
		const key = `${r.runePrimaryStyle}-${r.keyRune}-${r.runeSubStyle}`;
		const g = byKey.get(key);
		if (g) {
			g.games += r.games;
		} else {
			byKey.set(key, {
				primaryStyle: r.runePrimaryStyle,
				keyRune: r.keyRune,
				subStyle: r.runeSubStyle,
				games: r.games
			});
		}
	}
	return [...byKey.values()]
		.sort((a, b) => b.games - a.games)
		.slice(0, n)
		.map(({ primaryStyle, keyRune, subStyle }) => ({ primaryStyle, keyRune, subStyle }));
}
