// Đóng popup khi click ra ngoài `el` hoặc bấm Escape. Trả cleanup để dùng trong $effect:
//   $effect(() => (open ? dismissOn(root, () => (open = false)) : undefined));
export function dismissOn(el: HTMLElement, close: () => void): () => void {
	const onPointerDown = (e: PointerEvent) => {
		if (!el.contains(e.target as Node)) close();
	};
	const onKeyDown = (e: KeyboardEvent) => {
		if (e.key === 'Escape') close();
	};
	window.addEventListener('pointerdown', onPointerDown);
	window.addEventListener('keydown', onKeyDown);
	return () => {
		window.removeEventListener('pointerdown', onPointerDown);
		window.removeEventListener('keydown', onKeyDown);
	};
}
