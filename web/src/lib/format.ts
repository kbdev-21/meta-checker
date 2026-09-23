// Format dùng chung cho hiển thị.
import type { Rank, Tier } from '$lib/api';

const relativeTime = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

const TIME_UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
	['year', 365 * 24 * 3600],
	['month', 30 * 24 * 3600],
	['day', 24 * 3600],
	['hour', 3600],
	['minute', 60]
];

// "3 hours ago", "yesterday"...; dưới 1 phút => "just now".
export function timeAgo(iso: string): string {
	const diffSec = (new Date(iso).getTime() - Date.now()) / 1000;
	for (const [unit, sec] of TIME_UNITS) {
		if (Math.abs(diffSec) >= sec) {
			return relativeTime.format(Math.round(diffSec / sec), unit);
		}
	}
	return 'just now';
}

// 1710 => "28m 30s"
export function formatDuration(sec: number): string {
	return `${Math.floor(sec / 60)}m ${sec % 60}s`;
}

// "EMERALD" => "Emerald"
export function titleCase(s: string): string {
	return s.charAt(0) + s.slice(1).toLowerCase();
}

// Master trở lên chỉ có 1 bậc nên không hiện tier. Chưa có rank => "Unranked".
export function rankLabel(rank: Rank | null, tier: Tier | null): string {
	if (!rank || rank === 'UNRANKED') {
		return 'Unranked';
	}
	const isApex = rank === 'MASTER' || rank === 'GRANDMASTER' || rank === 'CHALLENGER';
	return isApex || !tier ? titleCase(rank) : `${titleCase(rank)} ${tier}`;
}

// Phân số 0-1 => "53%"
export function pct0(v: number): string {
	return `${Math.round(v * 100)}%`;
}
