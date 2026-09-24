// Player người dùng follow, lưu localStorage (chỉ trên trình duyệt này). Snapshot đủ để hiện
// trong dropdown search mà không phải gọi API; name/tag/icon cập nhật lại mỗi lần follow.
import { browser } from '$app/environment';
import type { Player, Server } from '$lib/api';

export type FollowedPlayer = {
	id: string; // puuid
	server: Server;
	name: string;
	tag: string;
	profileIconId: number | null;
};

const STORAGE_KEY = 'followed-players';

// Mới follow đứng đầu.
export const followedPlayers = $state({ list: load() });

export function isFollowed(playerId: string): boolean {
	return followedPlayers.list.some((p) => p.id === playerId);
}

export function toggleFollow(player: Player): void {
	if (isFollowed(player.id)) {
		followedPlayers.list = followedPlayers.list.filter((p) => p.id !== player.id);
	} else {
		const { id, server, name, tag, profileIconId } = player;
		followedPlayers.list = [{ id, server, name, tag, profileIconId }, ...followedPlayers.list];
	}
	save();
}

// ---------- private ----------

// localStorage có thể bị chặn (private mode, cookie bị tắt) hoặc chứa rác => coi như chưa follow ai.
function load(): FollowedPlayer[] {
	if (!browser) return [];
	try {
		const list = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
		return Array.isArray(list) ? list : [];
	} catch {
		return [];
	}
}

function save(): void {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(followedPlayers.list));
	} catch {
		// Không lưu được thì vẫn giữ trong phiên hiện tại.
	}
}
