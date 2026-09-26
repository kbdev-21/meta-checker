<script lang="ts" generics="T">
	import { untrack, type Snippet } from 'svelte';

	type Props = {
		// idle = false: dải item lướt qua, item cuối là kết quả. Đổi strip => quay từ item đầu tới
		// item cuối. Item đầu nên là kết quả của lần quay trước để lúc bắt đầu quay không bị giật.
		// idle = true: lướt chậm vòng lặp qua strip (lúc chưa roll lần nào).
		strip: T[];
		idle?: boolean;
		// Reel bị tắt: chỉ mờ đi, vẫn lướt idle bình thường.
		disabled?: boolean;
		duration: number;
		item: Snippet<[T]>;
	};

	let { strip, idle = false, disabled = false, duration, item }: Props = $props();

	// Chiều cao 1 ô và của khung nhìn (px). Khung cao hơn 1 ô để thấy lấp ló ô trên / dưới như máy slot.
	const CELL = 96;
	const WINDOW = 160;
	// Tốc độ lướt lúc idle: số ms cho mỗi ô.
	const IDLE_MS_PER_CELL = 1600;

	// translateY để ô thứ i (có thể lẻ) nằm giữa khung.
	const offset = (i: number) => (WINDOW - CELL) / 2 - i * CELL;

	let track: HTMLDivElement;

	// Ô thực sự render. Idle: strip lặp 2 lần để vòng lặp liền mạch.
	let rendered = $state<T[]>([]);
	// Vị trí (theo ô) bắt đầu quay.
	let fromPos = 0;
	let wasIdle = false;

	// Chạy trước khi DOM đổi nên animation cũ vẫn đang áp lên track => đọc được vị trí đang lướt.
	$effect.pre(() => {
		const s = strip;
		const isIdle = idle;
		untrack(() => {
			if (isIdle) {
				rendered = [...s, ...s];
			} else if (wasIdle && s.length > 1) {
				// Từ idle chuyển sang quay: giữ lại các ô idle đang hiện (1 ô trên, ô giữa, 2 ô dưới)
				// rồi nối strip mới (bỏ item đầu), quay từ đúng vị trí đang lướt => không bị giật.
				const pos = currentPos();
				const start = Math.max(0, Math.floor(pos) - 1);
				fromPos = pos - start;
				rendered = [...rendered.slice(start, start + 4), ...s.slice(1)];
			} else {
				rendered = s;
				fromPos = 0;
			}
			wasIdle = isIdle;
		});
	});

	// Web Animations thay vì CSS transition: không cần reset transform rồi đợi 1 frame mới chạy.
	// Hủy animation cũ thì track về offset(0), mà ô 0 của strip mới = kết quả cũ => liền mạch.
	$effect(() => {
		const n = rendered.length;
		let anim: Animation;
		if (n === 0) {
			return;
		} else if (idle) {
			// Lướt từ ô 1 tới ô n/2 + 1 (cùng item với ô 1) rồi lặp. Bắt đầu từ ô 1 để ô phía
			// trên luôn có item, không bị trống lúc vòng lặp quay về đầu.
			const cells = n / 2;
			anim = track.animate(
				[{ transform: `translateY(${offset(1)}px)` }, { transform: `translateY(${offset(cells + 1)}px)` }],
				{ duration: cells * IDLE_MS_PER_CELL, iterations: Infinity }
			);
			// Các reel lệch pha nhau cho tự nhiên.
			anim.currentTime = Math.random() * cells * IDLE_MS_PER_CELL;
		} else if (n >= 2) {
			anim = track.animate(
				[{ transform: `translateY(${offset(fromPos)}px)` }, { transform: `translateY(${offset(n - 1)}px)` }],
				{ duration, easing: 'cubic-bezier(0.15, 0.85, 0.3, 1)', fill: 'forwards' }
			);
		} else {
			return;
		}
		return () => anim.cancel();
	});

	// Vị trí (theo ô, có phần lẻ) của ô đang nằm giữa khung.
	function currentPos(): number {
		const y = new DOMMatrix(getComputedStyle(track).transform).m42;
		return (offset(0) - y) / CELL;
	}
</script>

<div
	class="relative overflow-hidden transition-opacity {disabled ? 'opacity-35 grayscale' : ''} [mask-image:linear-gradient(to_bottom,transparent,black_30%,black_70%,transparent)]"
	style:height="{WINDOW}px"
>
	<!-- vạch đánh dấu ô kết quả ở giữa khung -->
	<div
		class="pointer-events-none absolute inset-x-2 rounded-lg bg-white/[0.04] ring-1 ring-line"
		style:top="{(WINDOW - CELL) / 2}px"
		style:height="{CELL}px"
	></div>

	<div bind:this={track} style:transform="translateY({offset(0)}px)">
		{#each rendered as value, i (i)}
			<div class="flex flex-col items-center justify-center gap-1.5" style:height="{CELL}px">
				{@render item(value)}
			</div>
		{/each}
	</div>
</div>
