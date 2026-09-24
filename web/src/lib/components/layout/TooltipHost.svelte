<script lang="ts">
	import { Tooltip } from 'bits-ui';
	import type { TooltipSide } from '$lib/components/shared/Tooltip.svelte';

	// Hover bao lâu mới hiện. Vừa đóng 1 tooltip (rê từ icon này sang icon kế bên) trong
	// SKIP_DELAY_MS thì hiện ngay, không đợi lại.
	const DELAY_MS = 200;
	const SKIP_DELAY_MS = 300;

	let open = $state(false);
	let anchor = $state<HTMLElement | null>(null);
	let content = $state('');
	let side = $state<TooltipSide>('top');

	let openTimer: ReturnType<typeof setTimeout> | undefined;
	let lastClosedAt = 0;

	function show(el: HTMLElement) {
		anchor = el;
		content = el.dataset.tooltip ?? '';
		side = (el.dataset.tooltipSide as TooltipSide | undefined) ?? 'top';
		open = true;
	}

	function hide() {
		clearTimeout(openTimer);
		if (open) lastClosedAt = performance.now();
		open = false;
	}

	function tooltipTargetOf(e: Event): HTMLElement | null {
		return e.target instanceof Element ? e.target.closest<HTMLElement>('[data-tooltip]') : null;
	}

	function onPointerOver(e: PointerEvent) {
		// Touch không có hover; chạm vào icon thì để click xử lý như bình thường.
		if (e.pointerType !== 'mouse') return;
		const el = tooltipTargetOf(e);
		if (!el?.dataset.tooltip || (open && el === anchor)) return;

		clearTimeout(openTimer);
		if (open || performance.now() - lastClosedAt < SKIP_DELAY_MS) show(el);
		else openTimer = setTimeout(() => show(el), DELAY_MS);
	}

	function onPointerOut(e: PointerEvent) {
		const el = tooltipTargetOf(e);
		// Rê giữa các phần tử con bên trong cùng 1 target thì bỏ qua.
		if (!el || (e.relatedTarget instanceof Node && el.contains(e.relatedTarget))) return;
		hide();
	}
</script>

<!-- pointerdown: click (đổi tab, mở match...) có thể render lại làm anchor biến mất khỏi DOM mà
     không bắn pointerout, nên đóng luôn. -->
<svelte:document onpointerover={onPointerOver} onpointerout={onPointerOut} onpointerdown={hide} />

<!-- Root không có Trigger: Content neo vào phần tử đang hover qua customAnchor.
     disableHoverableContent: tooltip chỉ để đọc, không cần rê chuột vào trong. -->
<Tooltip.Provider>
	<Tooltip.Root bind:open disableHoverableContent>
		{#if anchor}
			<Tooltip.Portal>
				<Tooltip.Content
					customAnchor={anchor}
					{side}
					sideOffset={6}
					class="pointer-events-none z-50 max-w-xs rounded-md bg-elevated px-2.5 py-1.5 text-center text-xs text-ink shadow-lg ring-1 ring-line"
				>
					{content}
				</Tooltip.Content>
			</Tooltip.Portal>
		{/if}
	</Tooltip.Root>
</Tooltip.Provider>
