// Fetch wrapper dùng chung cho mọi API call. Các file lol-*.ts chỉ mô tả endpoint,
// còn URL / query / lỗi / parse JSON nằm hết ở đây.
import { PUBLIC_API_URL } from '$env/static/public';

export const API_BASE_URL = PUBLIC_API_URL || 'http://localhost:3000';

// Giá trị null / undefined bị bỏ khỏi query string.
export type QueryParams = Record<string, string | number | boolean | null | undefined>;

export type RequestOptions = {
	query?: QueryParams;
	// fetch của SvelteKit load(); bỏ trống thì dùng fetch global.
	fetch?: typeof globalThis.fetch;
	signal?: AbortSignal;
};

// Response có status không OK. 404 của các endpoint "có thể không tồn tại" được
// requestOrNull nuốt thành null nên sẽ không throw ra đây.
export type ApiError = Error & {
	status: number;
	url: string;
	body: string;
};

export function isApiError(err: unknown): err is ApiError {
	return err instanceof Error && 'status' in err;
}

// Status không OK => ApiError.
export async function request<T>(
	method: string,
	path: string,
	opts: RequestOptions = {}
): Promise<T> {
	const res = await send(method, path, opts);
	if (!res.ok) {
		throw await apiError(res);
	}
	return parseJson<T>(res);
}

// Như request nhưng 404 => null, dùng cho endpoint mà "không tìm thấy" là kết quả hợp lệ
// (player chưa từng tồn tại, meta chưa tổng hợp cho patch hiện tại...).
export async function requestOrNull<T>(
	method: string,
	path: string,
	opts: RequestOptions = {}
): Promise<T | null> {
	const res = await send(method, path, opts);
	if (res.status === 404) {
		return null;
	}
	if (!res.ok) {
		throw await apiError(res);
	}
	return parseJson<T>(res);
}

// ---------- private ----------

async function apiError(res: Response): Promise<ApiError> {
	const body = await res.text().catch(() => '');
	return Object.assign(new Error(`${res.status} ${res.url}${body ? `: ${body}` : ''}`), {
		status: res.status,
		url: res.url,
		body
	});
}

function send(method: string, path: string, opts: RequestOptions): Promise<Response> {
	const doFetch = opts.fetch ?? globalThis.fetch;
	return doFetch(buildUrl(path, opts.query), {
		method,
		headers: { Accept: 'application/json' },
		signal: opts.signal
	});
}

function buildUrl(path: string, query?: QueryParams): string {
	const url = new URL(path, API_BASE_URL);
	for (const [key, value] of Object.entries(query ?? {})) {
		if (value !== null && value !== undefined) {
			url.searchParams.set(key, String(value));
		}
	}
	return url.toString();
}

// 204 (endpoint update player) không có body; các endpoint còn lại luôn trả JSON.
async function parseJson<T>(res: Response): Promise<T> {
	if (res.status === 204) {
		return undefined as T;
	}
	return (await res.json()) as T;
}
