<script lang="ts">
	import './layout.css';
	import brandIcon from '$lib/assets/brand-icon.svg';
	import Footer from '$lib/components/layout/Footer.svelte';
	import Header from '$lib/components/layout/Header.svelte';
	import TooltipHost from '$lib/components/layout/TooltipHost.svelte';
	import { loadLolData } from '$lib/stores/lol-data.svelte';

	let { children } = $props();

	// lol-data (tên + icon champion, item, rune) dùng ở mọi trang. Đặt trong $effect để
	// chỉ chạy ở browser, không bắn fetch lúc prerender.
	$effect(() => {
		loadLolData();
	});
</script>

<!-- title / description do từng page đặt qua PageMeta: đặt ở đây thì layout render trước page,
     <title> của layout sẽ thắng. -->
<svelte:head><link rel="icon" type="image/svg+xml" href={brandIcon} /></svelte:head>

<!-- flex-col + flex-1: trang ngắn thì footer vẫn nằm sát đáy màn hình -->
<div class="flex min-h-dvh flex-col">
	<Header />
	<div class="flex-1">{@render children()}</div>
	<Footer />
</div>
<!-- 1 tooltip dùng chung cho mọi <Tooltip> trong app -->
<TooltipHost />
