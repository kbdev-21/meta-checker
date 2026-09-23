import type { PageLoad } from './$types';

// Data (champion, stat) đều fetch ở browser nên không prerender: dùng fallback 200.html
// (xem vite.config.ts) như trang player.
export const prerender = false;

// URL: /lol/champions/Fiora. slug = id ddragon; so khớp không phân biệt hoa thường ở page
// vì danh sách champion chỉ có sau khi lol-data load xong.
export const load: PageLoad = ({ params }) => {
	return { slug: params.championSlug };
};
