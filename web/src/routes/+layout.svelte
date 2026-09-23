<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import Header from '$lib/components/Header.svelte';
	import { loadLolData } from '$lib/stores/lol-data.svelte';

	let { children } = $props();

	// lol-data (tên + icon champion, item, rune) dùng ở mọi trang. Đặt trong $effect để
	// chỉ chạy ở browser, không bắn fetch lúc prerender.
	$effect(() => {
		loadLolData();
	});
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<Header />
{@render children()}
