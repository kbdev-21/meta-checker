// GET /api/lol/analytics/* — meta của patch hiện tại.
import { requestOrNull, type RequestOptions } from './client';
import { META_SERVER_GLOBAL, type ChampionStat, type Meta, type MetaServer, type RankBucket } from './types';

// Mặc định của backend khi không truyền query.
export const DEFAULT_META_SERVER: MetaServer = META_SERVER_GLOBAL;
export const DEFAULT_RANK_BUCKET: RankBucket = 'MASTER_PLUS';

export type AnalyticsParams = {
	server?: MetaServer;
	rankBucket?: RankBucket;
};

// Tier list: meta kèm ChampionStatSummary của mọi (champion, position).
// Chưa tổng hợp cho patch hiện tại => null.
export function getAnalytics(
	params: AnalyticsParams = {},
	opts?: RequestOptions
): Promise<Meta | null> {
	return requestOrNull<Meta>('GET', '/api/lol/analytics', {
		...opts,
		query: { ...opts?.query, ...params }
	});
}

// Mọi position của 1 champion, kèm build (spell combo, rune, item).
// Meta chưa tổng hợp => null; champion không có dòng nào => list rỗng.
export function getChampionAnalytics(
	championId: number,
	params: AnalyticsParams = {},
	opts?: RequestOptions
): Promise<ChampionStat[] | null> {
	return requestOrNull<ChampionStat[]>('GET', `/api/lol/analytics/champions/${championId}`, {
		...opts,
		query: { ...opts?.query, ...params }
	});
}
