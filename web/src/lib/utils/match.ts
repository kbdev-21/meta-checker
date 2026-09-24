import type { MatchParticipant } from '$lib/api';
import { lolData } from '$lib/stores/lol-data.svelte';

// items[6] là trinket.
const TRINKET_SLOT = 6;

// Item của ô cuối (sau 6 ô đồ). ADC xong role quest thì giày chuyển sang ô riêng (roleBoundItem),
// không còn trong items, nên hiện giày thay cho trinket. Chưa xong quest thì roleBoundItem là item
// theo dõi quest (vd 1202, không có trong lol_items) và giày vẫn nằm trong items => hiện trinket.
export function lastSlotItemOf(p: MatchParticipant): number {
	if (p.position === 'ADC' && lolData.itemById.has(p.roleBoundItem)) return p.roleBoundItem;
	return p.items[TRINKET_SLOT];
}
