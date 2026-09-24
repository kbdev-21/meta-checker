<script lang="ts">
	import type { Server } from '$lib/api';

	type Size = 'SM' | 'MD';

	type Props = {
		server: Server;
		// SM: dòng kết quả search; MD: header trang player.
		size?: Size;
	};

	let { server, size = 'SM' }: Props = $props();

	// Size là prop thay vì nhận class từ ngoài: 2 class cỡ chữ trùng nhau thì Tailwind không
	// đảm bảo class truyền vào thắng.
	const SIZE_CLASS: Record<Size, string> = {
		SM: 'px-1.5 text-[10px]',
		MD: 'px-2 py-0.5 text-xs'
	};
</script>

<span class="shrink-0 rounded bg-elevated font-bold text-muted {SIZE_CLASS[size]}">{server}</span>
