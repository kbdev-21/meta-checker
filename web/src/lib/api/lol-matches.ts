// GET /api/lol/matches/*
import { requestOrNull, type RequestOptions } from './client';
import type { Match, MatchListMode, Server } from './types';

// Trùng với matchListDefaultCount / matchListMaxCount của backend; vượt max => 400.
export const MATCH_LIST_DEFAULT_COUNT = 10;
export const MATCH_LIST_MAX_COUNT = 20;

export type MatchListParams = {
	mode?: MatchListMode; // bỏ trống = mọi mode
	start?: number; // offset, mặc định 0
	count?: number; // 1..MATCH_LIST_MAX_COUNT, mặc định MATCH_LIST_DEFAULT_COUNT
};

// Backend tự crawl match còn thiếu từ Riot nên call này có thể chậm.
// Không tìm thấy player => null. Kết quả mới nhất trước.
export function getMatchesByPlayerInfo(
	server: Server,
	name: string,
	tag: string,
	params: MatchListParams = {},
	opts?: RequestOptions
): Promise<Match[] | null> {
	const path = `/api/lol/matches/by-player-info/${server}/${encodeURIComponent(name)}/${encodeURIComponent(tag)}`;
	return requestOrNull<Match[]>('GET', path, {
		...opts,
		query: { ...opts?.query, mode: params.mode, start: params.start, count: params.count }
	});
}
