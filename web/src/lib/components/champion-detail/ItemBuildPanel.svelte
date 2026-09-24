<script lang="ts">
	import ChevronRight from '@lucide/svelte/icons/chevron-right';

	import type { ItemSetStat, ItemStat } from '$lib/api';
	import Tooltip from '$lib/components/shared/Tooltip.svelte';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { pct2, winRateColor } from '$lib/utils/format';

	type Props = {
		// Các list đều đã được backend sort games giảm dần.
		starterSets: ItemSetStat[]; // itemIds đã sort, giữ trùng (2 bình máu = 2 phần tử)
		bootItems: ItemStat[];
		coreBuilds: ItemSetStat[]; // 3 đồ legendary đầu, đúng thứ tự mua
		// Tổng trận của (champion, position): mẫu số pick rate của boots (tính từ items cuối trận,
		// có ở mọi trận). Starter / core chỉ có ở trận có timeline nên lấy tổng của chính list đó.
		games: number;
		loading: boolean;
	};

	let { starterSets, bootItems, coreBuilds, games, loading }: Props = $props();

	// Số dòng mỗi nhóm (giống op.gg).
	const STARTER_COUNT = 2;
	const BOOT_COUNT = 2;
	const CORE_COUNT = 5;

	const starterTotal = $derived(starterSets.reduce((sum, s) => sum + s.games, 0));
	const coreTotal = $derived(coreBuilds.reduce((sum, s) => sum + s.games, 0));

	// [1055, 2003, 2003] => [{1055, 1}, {2003, 2}]: món trùng hiện 1 icon kèm số lượng.
	function groupItems(itemIds: number[]): { itemId: number; count: number }[] {
		const out: { itemId: number; count: number }[] = [];
		for (const id of itemIds) {
			const last = out[out.length - 1];
			if (last?.itemId === id) last.count++;
			else out.push({ itemId: id, count: 1 });
		}
		return out;
	}
</script>

{#snippet itemIcon(itemId: number, count = 1)}
	{@const item = lolData.itemById.get(itemId)}
	<Tooltip content={item?.name ?? String(itemId)} class="inline-flex shrink-0">
		<div class="relative shrink-0">
			{#if item}
				<img src={item.imgUrl} alt={item.name} class="size-8 rounded" loading="lazy" />
			{:else}
				<div class="size-8 rounded bg-elevated"></div>
			{/if}
			{#if count > 1}
				<span class="absolute -bottom-0.5 -right-0.5 rounded-tl bg-base px-0.5 text-[10px] font-bold leading-none">
					{count}
				</span>
			{/if}
		</div>
	</Tooltip>
{/snippet}

<!-- pick rate + số trận ở giữa, win rate bên phải (giống Spells / Skill Order) -->
{#snippet rates(g: number, total: number, wins: number)}
	<div class="flex-1 text-center text-xs text-muted">
		<div class="text-xs font-bold text-ink">{pct2(g / total)}</div>
		<div>{g.toLocaleString('en-US')} Games</div>
	</div>
	<!-- rộng cố định để "100.00%" không đẩy lệch cột so với các dòng khác -->
	<div class="w-14 text-center text-xs font-bold {winRateColor(wins / g)}">{pct2(wins / g)}</div>
{/snippet}

{#snippet subHeader(label: string)}
	<div class="px-4 py-2 text-xs text-muted">{label}</div>
{/snippet}

{#snippet empty()}
	<p class="px-4 py-3 text-center text-xs text-muted">No data</p>
{/snippet}

<section class="overflow-hidden rounded-lg border border-line bg-surface">
	<h2 class="border-b border-line px-4 py-2.5 text-sm font-semibold">Builds</h2>

	{#if loading}
		<div class="animate-pulse p-4">
			<div class="h-[380px] rounded-lg bg-elevated"></div>
		</div>
	{:else}
		<!-- starter | boots -->
		<div class="grid grid-cols-2 divide-x divide-line border-b border-line bg-base/40">
			{@render subHeader('Starter Items')}
			{@render subHeader('Boots')}
		</div>
		<div class="grid grid-cols-2 divide-x divide-line">
			<div class="divide-y divide-line/50">
				{#each starterSets.slice(0, STARTER_COUNT) as s (s.itemIds.join('-'))}
					<div class="flex items-center gap-4 px-4 py-3">
						<div class="flex gap-1">
							{#each groupItems(s.itemIds) as x (x.itemId)}
								{@render itemIcon(x.itemId, x.count)}
							{/each}
						</div>
						{@render rates(s.games, starterTotal, s.wins)}
					</div>
				{:else}
					{@render empty()}
				{/each}
			</div>
			<div class="divide-y divide-line/50">
				{#each bootItems.slice(0, BOOT_COUNT) as b (b.itemId)}
					<div class="flex items-center gap-4 px-4 py-3">
						{@render itemIcon(b.itemId)}
						{@render rates(b.games, games, b.wins)}
					</div>
				{:else}
					{@render empty()}
				{/each}
			</div>
		</div>

		<!-- core builds: 3 đồ legendary đầu theo thứ tự mua -->
		<div class="border-y border-line bg-base/40">
			{@render subHeader('Core Items')}
		</div>
		<div class="divide-y divide-line/50">
			{#each coreBuilds.slice(0, CORE_COUNT) as c (c.itemIds.join('-'))}
				<div class="grid grid-cols-2">
					<div class="flex items-center gap-1.5 px-4 py-3">
						{#each c.itemIds as itemId, i (i)}
							{#if i > 0}
								<ChevronRight class="size-4 shrink-0 text-muted" />
							{/if}
							{@render itemIcon(itemId)}
						{/each}
					</div>
					<!-- nửa phải dựng giống 1 dòng Boots (ô trống rộng bằng icon) để số thẳng cột với Boots -->
					<div class="flex items-center gap-4 px-4 py-3">
						<div class="size-8 shrink-0"></div>
						{@render rates(c.games, coreTotal, c.wins)}
					</div>
				</div>
			{:else}
				{@render empty()}
			{/each}
		</div>
	{/if}
</section>
