<script module lang="ts">
	type Rgb = [number, number, number];

	// Màu thanh từ trái sang phải: đỏ (fun) → vàng → xanh lá (meta). Tailwind red-500 / yellow-400 / green-500.
	const STOPS: Rgb[] = [
		[239, 68, 68],
		[250, 204, 21],
		[34, 197, 94]
	];

	export const SLIDER_GRADIENT = `linear-gradient(to right, ${STOPS.map(rgb).join(', ')})`;

	// Màu tại vị trí t (0-1) trên thanh, nội suy giữa 2 stop kề nhau.
	export function sliderColor(t: number): string {
		const x = clamp(t) * (STOPS.length - 1);
		const i = Math.min(Math.floor(x), STOPS.length - 2);
		const f = x - i;
		const [from, to] = [STOPS[i], STOPS[i + 1]];
		return rgb(from.map((v, k) => Math.round(v + (to[k] - v) * f)) as Rgb);
	}

	function rgb([r, g, b]: Rgb): string {
		return `rgb(${r} ${g} ${b})`;
	}

	function clamp(t: number): number {
		return Math.min(Math.max(t, 0), 1);
	}
</script>

<script lang="ts">
	type Props = {
		// Index mốc đang chọn, 0 .. count - 1. Chỉ đổi khi buông chuột (hoặc bấm phím).
		value: number;
		count: number;
		label: string;
		valueText?: string;
		disabled?: boolean;
	};

	let { value = $bindable(), count, label, valueText, disabled = false }: Props = $props();

	const max = $derived(count - 1);

	let track: HTMLDivElement;
	// Vị trí 0-1 lúc đang kéo; null = không kéo, thumb nằm đúng mốc của value.
	let dragPos = $state<number | null>(null);
	const dragging = $derived(dragPos !== null);
	const pos = $derived(dragPos ?? value / max);

	// Pointer capture: kéo ra ngoài thanh (kể cả ngoài cửa sổ) vẫn nhận move / up.
	function onpointerdown(e: PointerEvent) {
		if (disabled || e.button !== 0) return;
		(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
		dragPos = pointerPos(e);
	}

	function onpointermove(e: PointerEvent) {
		if (dragPos === null) return;
		dragPos = pointerPos(e);
	}

	// Buông chuột mới dính vào mốc gần nhất.
	function onpointerup() {
		if (dragPos === null) return;
		value = Math.round(dragPos * max);
		dragPos = null;
	}

	function onpointercancel() {
		dragPos = null;
	}

	function onkeydown(e: KeyboardEvent) {
		if (disabled) return;
		const next: Record<string, number> = {
			ArrowLeft: value - 1,
			ArrowDown: value - 1,
			ArrowRight: value + 1,
			ArrowUp: value + 1,
			Home: 0,
			End: max
		};
		if (!(e.key in next)) return;
		e.preventDefault();
		value = Math.min(Math.max(next[e.key], 0), max);
	}

	function pointerPos(e: PointerEvent): number {
		const r = track.getBoundingClientRect();
		return clamp((e.clientX - r.left) / r.width);
	}
</script>

<!-- py-2 để vùng bấm cao hơn thanh mảnh -->
<div
	role="slider"
	tabindex={disabled ? -1 : 0}
	aria-label={label}
	aria-valuemin={0}
	aria-valuemax={max}
	aria-valuenow={value}
	aria-valuetext={valueText}
	aria-disabled={disabled}
	class="group relative touch-none select-none py-2 outline-none {disabled
		? 'cursor-not-allowed opacity-60'
		: dragging
			? 'cursor-grabbing'
			: 'cursor-pointer'}"
	{onpointerdown}
	{onpointermove}
	{onpointerup}
	{onpointercancel}
	{onkeydown}
>
	<div bind:this={track} class="relative h-2 rounded-full">
		<!-- nền: cả dải màu nhưng mờ -->
		<div class="absolute inset-0 rounded-full opacity-25" style:background={SLIDER_GRADIENT}></div>
		<!-- phần từ đầu tới thumb: sáng. Lúc kéo thì bám theo chuột, không transition. -->
		<div
			class="absolute inset-0 rounded-full {dragging ? '' : 'transition-[clip-path] duration-200 ease-out'}"
			style:background={SLIDER_GRADIENT}
			style:clip-path="inset(0 {(1 - pos) * 100}% 0 0)"
		></div>

		<!-- thumb: viền theo màu tại vị trí đang đứng -->
		<div
			class="absolute top-1/2 size-5 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 bg-white shadow-md shadow-black/40 group-focus-visible:ring-2 group-focus-visible:ring-accent/60 {dragging
				? 'scale-115'
				: 'transition-[left,border-color] duration-200 ease-out'}"
			style:left="{pos * 100}%"
			style:border-color={sliderColor(pos)}
		></div>
	</div>
</div>
