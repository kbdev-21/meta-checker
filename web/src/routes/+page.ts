import { redirect } from '@sveltejs/kit';

// Tạm thời: trang chủ chưa có nội dung nên chuyển thẳng sang tier list.
// Prerender vẫn chạy được: build ra trang HTML tự chuyển hướng (meta refresh).
export function load() {
	redirect(307, '/lol/champions');
}
