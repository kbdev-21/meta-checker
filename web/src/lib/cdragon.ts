// Asset lấy từ CommunityDragon (extract từ game client) cho những thứ ddragon không có.
import type { Rank } from '$lib/api';

const STATIC_ASSETS =
	'https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images';

// Emblem rank (Iron..Challenger). Ảnh là canvas 16:9 với emblem nằm giữa chiếm ~25% chiều
// ngang, nên chỗ hiển thị phải phóng to + cắt (xem RankCard). UNRANKED không có emblem => null.
export function rankEmblemUrl(rank: Rank): string | null {
	if (rank === 'UNRANKED') {
		return null;
	}
	return `${STATIC_ASSETS}/ranked-emblem/emblem-${rank.toLowerCase()}.png`;
}
