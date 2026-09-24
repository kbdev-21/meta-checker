<script module lang="ts">
	export type TooltipSide = 'top' | 'right' | 'bottom' | 'left';
</script>

<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		content: string;
		side?: TooltipSide;
		// Class của span bọc ngoài. Mặc định inline-flex để không đổi layout của thứ được bọc.
		class?: string;
		children: Snippet;
	};

	let { content, side = 'top', class: className = 'inline-flex', children }: Props = $props();
</script>

<!-- Chỉ đánh dấu phần tử cần tooltip. Việc hiện tooltip do TooltipHost (1 instance duy nhất ở layout
     gốc) lo qua event delegation, nên trang có hàng trăm icon cũng không tốn thêm state / listener nào.
     align-top luôn có: span nằm trong phần tử inline (vd <a>) thì mặc định căn theo baseline, chừa
     thêm khoảng trống dưới icon cho chân chữ => khối cha cao lên (MatchCard từng phình 108 => 128px). -->
<span data-tooltip={content} data-tooltip-side={side} class="align-top {className}">{@render children()}</span>
