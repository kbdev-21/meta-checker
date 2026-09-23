<script lang="ts">
	type Props = {
		src: string;
		alt?: string;
		title?: string;
		// Kích thước + bo góc do nơi dùng quyết định (vd size-7 rounded, size-6 rounded-full).
		class?: string;
	};

	let { src, alt = '', title, class: className = '' }: Props = $props();
</script>

<!-- Ảnh ddragon có sẵn 1 viền đen mảnh quanh mép: zoom 110% rồi cắt phần tràn để giấu viền. -->
<span class="block shrink-0 overflow-hidden {className}" {title}>
	<img {src} {alt} class="size-full scale-110 object-cover" loading="lazy" />
</span>
