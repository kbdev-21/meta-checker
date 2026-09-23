// GET/POST /api/lol/players/*
import { request, requestOrNull, type RequestOptions } from './client';
import type { Player, Server } from './types';

// Backend luôn giới hạn 20 kết quả. Tìm trong name, tag và cả bản bỏ dấu;
// sắp xếp theo solo rank power giảm dần.
export function searchPlayers(keyword: string, opts?: RequestOptions): Promise<Player[]> {
	return request<Player[]>('GET', '/api/lol/players', {
		...opts,
		query: { ...opts?.query, q: keyword }
	});
}

// Backend tự gọi Riot để làm mới nếu player đã cũ. Không tìm thấy => null.
export function findPlayerByInfo(
	server: Server,
	name: string,
	tag: string,
	opts?: RequestOptions
): Promise<Player | null> {
	return requestOrNull<Player>('GET', playerInfoPath(server, name, tag), opts);
}

// Ép làm mới từ Riot (luôn gọi Riot, không dùng cache). Riot không có player này => false.
export async function updatePlayerByInfo(
	server: Server,
	name: string,
	tag: string,
	opts?: RequestOptions
): Promise<boolean> {
	const res = await requestOrNull<void>('POST', `${playerInfoPath(server, name, tag)}/update`, opts);
	return res !== null;
}

// ---------- private ----------

// name/tag có thể chứa khoảng trắng hoặc ký tự lạ nên phải encode từng segment.
function playerInfoPath(server: Server, name: string, tag: string): string {
	return `/api/lol/players/by-info/${server}/${encodeURIComponent(name)}/${encodeURIComponent(tag)}`;
}
