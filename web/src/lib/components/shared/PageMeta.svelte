<script lang="ts">
	// Title + metadata của 1 trang. Mỗi page tự đặt (không đặt ở layout: layout render trước
	// page nên <title> của layout sẽ thắng).
	type Props = {
		// Không có => title chỉ là tên site.
		title?: string;
		description: string;
	};

	let { title, description }: Props = $props();

	const SITE_NAME = 'MetaChecker';

	const fullTitle = $derived(title ? `${title} | ${SITE_NAME}` : SITE_NAME);
</script>

<svelte:head>
	<title>{fullTitle}</title>
	<meta name="description" content={description} />
	<meta property="og:site_name" content={SITE_NAME} />
	<meta property="og:type" content="website" />
	<meta property="og:title" content={fullTitle} />
	<meta property="og:description" content={description} />
	<meta name="twitter:card" content="summary" />
</svelte:head>
