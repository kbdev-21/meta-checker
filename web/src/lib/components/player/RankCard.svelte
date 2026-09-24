<script lang="ts">
	import Trophy from '@lucide/svelte/icons/trophy';
	import type { Rank, Tier } from '$lib/api';
	import { rankEmblemUrl } from '$lib/utils/cdragon';
	import { pct0, rankLabel } from '$lib/utils/format';

	type Props = {
		title: string;
		rank: Rank | null;
		tier: Tier | null;
		lp: number;
		wins: number;
		losses: number;
	};

	let { title, rank, tier, lp, wins, losses }: Props = $props();

	const games = $derived(wins + losses);
	const isRanked = $derived(!!rank && rank !== 'UNRANKED');
	const emblem = $derived(rank ? rankEmblemUrl(rank) : null);
</script>

<section class="rounded-lg border border-line bg-surface">
	<h2 class="border-b border-line px-4 py-2.5 text-sm font-semibold">{title}</h2>

	<div class="flex items-center gap-4 px-4 py-3">
		{#if emblem}
			<!-- Ảnh gốc là canvas 16:9, emblem chỉ ~25% chiều ngang ở giữa: phóng ảnh lên 4 lần
			     width khung rồi căn giữa, overflow-hidden cắt phần trong suốt thừa. -->
			<div class="relative size-16 shrink-0 overflow-hidden">
				<img
					src={emblem}
					alt={rankLabel(rank, tier)}
					class="absolute left-1/2 top-1/2 w-[400%] max-w-none -translate-x-1/2 -translate-y-1/2"
				/>
			</div>
		{:else}
			<div class="grid size-16 shrink-0 place-items-center rounded-full bg-elevated text-muted">
				<Trophy class="size-7" />
			</div>
		{/if}

		{#if isRanked}
			<div class="min-w-0 flex-1">
				<div class="text-lg font-bold">{rankLabel(rank, tier)}</div>
				<div class="text-xs text-muted">{lp} LP</div>
			</div>
			<div class="text-right text-xs text-muted flex flex-col gap-1">
				<div>{wins}W {losses}L</div>
				<div>WR {games ? pct0(wins / games) : '-'}</div>
			</div>
		{:else}
			<div class="text-sm font-semibold text-muted">Unranked</div>
		{/if}
	</div>
</section>
