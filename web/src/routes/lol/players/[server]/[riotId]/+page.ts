import { error } from '@sveltejs/kit';
import { SERVERS, type Server } from '$lib/api';
import { parseRiotIdSlug } from '$lib/riot-id';
import type { PageLoad } from './$types';

// Vô số player nên không prerender được: build ra fallback 200.html (xem vite.config.ts),
// router client tự render route này.
export const prerender = false;

// URL: /lol/players/vn/Name-TAG. Server viết thường trên URL, app dùng viết hoa.
export const load: PageLoad = ({ params }) => {
	const server = params.server.toUpperCase() as Server;
	const riotId = parseRiotIdSlug(params.riotId);
	if (!SERVERS.includes(server) || !riotId) {
		error(404, 'Player not found');
	}
	return { server, name: riotId.name, tag: riotId.tag };
};
