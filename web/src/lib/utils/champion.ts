import { DEFAULT_META_SERVER, type MetaServer, type Position } from '$lib/api';

// URL trang chi tiết champion. Query ghi giống cách trang chi tiết tự ghi: server default
// thì bỏ, UNK không phải lane nên bỏ (trang tự chọn lane nhiều trận nhất).
export function championHref(slug: string, opts: { position?: Position; server?: MetaServer } = {}): string {
	const params = new URLSearchParams();
	if (opts.server && opts.server !== DEFAULT_META_SERVER) params.set('server', opts.server);
	if (opts.position && opts.position !== 'UNK') params.set('position', opts.position);
	const search = params.size ? `?${params}` : '';
	return `/lol/champions/${encodeURIComponent(slug)}${search}`;
}
