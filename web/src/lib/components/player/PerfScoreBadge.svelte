<script lang="ts">
	import Tooltip from '$lib/components/shared/Tooltip.svelte';

	type Size = 'SM' | 'MD';

	type Props = {
		// perfScore 0-100 của participant.
		score: number;
		// MD: card match; SM: chỗ chật như bảng MatchDetail.
		size?: Size;
	};

	let { score, size = 'MD' }: Props = $props();

	const TOOLTIP = "Performance score (0-100) calculated by MetaChecker's formula";

	const SIZE_CLASS: Record<Size, string> = {
		SM: 'h-6 w-8 rounded text-sm',
		MD: 'h-8 w-10 rounded-md text-lg'
	};

	// Mốc dưới của từng bậc: < 50 đỏ, 50-74 cyan, 75-89 vàng, 90-100 premium.
	const CYAN_FROM = 50;
	const GOLD_FROM = 75;
	const PREMIUM_FROM = 90;

	const isPremium = $derived(score >= PREMIUM_FROM);
	const tone = $derived(
		isPremium
			? // gradient + glow + viền sáng mảnh bên trong cho cảm giác "kim loại bóng"
				"bg-linear-to-br from-teal-300 via-cyan-400 to-blue-500 shadow-[0_0_14px_rgba(45,212,191,0.55)] ring-1 ring-inset ring-white/45"
			: score >= GOLD_FROM
				? "bg-amber-400"
				: score >= CYAN_FROM
					? "bg-emerald-500"
					: "bg-red-500",
	);
</script>

<Tooltip content={TOOLTIP}>
	<span
		class="relative inline-grid place-items-center overflow-hidden font-medium text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.35)] {SIZE_CLASS[
			size
		]} {tone}"
	>
		{score}
		{#if isPremium}
			<span class="shine" aria-hidden="true"></span>
		{/if}
	</span>
</Tooltip>

<style>
	/* Vệt sáng quét chéo qua badge premium, nghỉ một nhịp giữa các lần quét. */
	.shine {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			110deg,
			transparent 30%,
			rgb(255 255 255 / 0.6) 50%,
			transparent 70%
		);
		transform: translateX(-120%);
		animation: shine 3s ease-in-out infinite;
		pointer-events: none;
	}

	@keyframes shine {
		0%,
		55% {
			transform: translateX(-120%);
		}
		100% {
			transform: translateX(120%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.shine {
			display: none;
		}
	}
</style>
