// Asset lấy từ CommunityDragon (extract từ game client) cho những thứ ddragon không có.
import type { Rank } from '$lib/api';

const STATIC_ASSETS =
	'https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images';

const STAT_MOD_ICONS =
	'https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/v1/perk-images/statmods';

// Stat shard (id trong statRunes): ddragon không có nên lol_runes cũng không có.
// Tên + icon lấy theo perks.json của CommunityDragon.
export const STAT_SHARDS: Record<number, { name: string; imgUrl: string }> = {
	5001: { name: 'Health Scaling', imgUrl: `${STAT_MOD_ICONS}/statmodshealthplusicon.png` },
	5005: { name: 'Attack Speed', imgUrl: `${STAT_MOD_ICONS}/statmodsattackspeedicon.png` },
	5007: { name: 'Ability Haste', imgUrl: `${STAT_MOD_ICONS}/statmodscdrscalingicon.png` },
	5008: { name: 'Adaptive Force', imgUrl: `${STAT_MOD_ICONS}/statmodsadaptiveforceicon.png` },
	5010: { name: 'Move Speed', imgUrl: `${STAT_MOD_ICONS}/statmodsmovementspeedicon.png` },
	5011: { name: 'Health', imgUrl: `${STAT_MOD_ICONS}/statmodshealthscalingicon.png` },
	5013: { name: 'Tenacity and Slow Resist', imgUrl: `${STAT_MOD_ICONS}/statmodstenacityicon.png` }
};

// Emblem rank (Iron..Challenger). Ảnh là canvas 16:9 với emblem nằm giữa chiếm ~25% chiều
// ngang, nên chỗ hiển thị phải phóng to + cắt (xem RankCard). UNRANKED không có emblem => null.
export function rankEmblemUrl(rank: Rank): string | null {
	if (rank === 'UNRANKED') {
		return null;
	}
	return `${STATIC_ASSETS}/ranked-emblem/emblem-${rank.toLowerCase()}.png`;
}
