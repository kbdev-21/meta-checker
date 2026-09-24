<script lang="ts">
	import type { Rune, RuneStat } from '$lib/api';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { STAT_SHARDS } from '$lib/utils/cdragon';
	import { pct2, winRateColor } from '$lib/utils/format';

	type Props = {
		// Mọi bộ rune của (champion, position), backend đã sort games giảm dần.
		runes: RuneStat[];
		loading: boolean;
	};

	let { runes, loading }: Props = $props();

	// Số tab bộ rune (cột bên trái).
	const PAGE_COUNT = 3;

	// 3 hàng stat shard cố định của game (offense, flex, defense), cùng thứ tự với statRunes.
	const STAT_SHARD_ROWS = [
		[5008, 5005, 5007],
		[5008, 5010, 5001],
		[5011, 5013, 5001]
	];

	type Count = { games: number; wins: number };

	const total = $derived(runes.reduce((sum, r) => sum + r.games, 0));

	// Tab = 1 cặp (cây chính, keystone, cây phụ), cộng dồn mọi bộ rune cùng cặp. Rune được
	// highlight của tab = bộ nhiều trận nhất trong nhóm (gặp đầu tiên vì runes đã sort).
	const pages = $derived.by(() => {
		const byKey = new Map<string, Count & { top: RuneStat }>();
		for (const r of runes) {
			const key = `${r.runePrimaryStyle}-${r.keyRune}-${r.runeSubStyle}`;
			const g = byKey.get(key);
			if (g) {
				g.games += r.games;
				g.wins += r.wins;
			} else {
				byKey.set(key, { games: r.games, wins: r.wins, top: r });
			}
		}
		return [...byKey.values()].sort((a, b) => b.games - a.games).slice(0, PAGE_COUNT);
	});

	// runes đổi (đổi position / server) thì quay về tab đầu; click tab thì ghi đè.
	let selected = $derived.by(() => {
		void runes;
		return 0;
	});
	const page = $derived(pages[selected]?.top);

	// Màu viền rune đang chọn theo cây (giống client game); shard không thuộc cây nào.
	const TREE_COLORS: Record<number, string> = {
		8000: '#c8aa6e', // Precision
		8100: '#d44242', // Domination
		8200: '#9faafc', // Sorcery
		8300: '#49aab9', // Inspiration
		8400: '#a1d586' // Resolve
	};
	const SHARD_COLOR = '#2dd4bf';

	// 5 cây theo thứ tự trong game (backend sort theo sort_order).
	const trees = $derived([...lolData.runeById.values()].filter((r) => r.styleId === null));

	const picked = $derived(new Set(page?.runes ?? []));
	const primaryRows = $derived(page ? treeRows(page.runePrimaryStyle) : []);
	// Cây phụ không có hàng keystone.
	const subRows = $derived(page ? treeRows(page.runeSubStyle).slice(1) : []);

	// Rune của 1 cây, chia theo hàng. lolData giữ đúng thứ tự trong game (backend sort theo sort_order).
	function treeRows(styleId: number): Rune[][] {
		const rows: Rune[][] = [];
		for (const r of lolData.runeById.values()) {
			if (r.styleId !== styleId || r.slot === null) continue;
			(rows[r.slot] ??= []).push(r);
		}
		return rows;
	}
</script>

<!-- icon rune / shard / cây. Đang chọn: sáng, viền màu (nếu có color); còn lại mờ. -->
{#snippet runeIcon(imgUrl: string | undefined, name: string, active: boolean, sizeClass: string, color?: string)}
	{#if imgUrl}
		<img
			src={imgUrl}
			alt={name}
			title={name}
			class="{sizeClass} shrink-0 rounded-full bg-black/40 {active ? '' : 'opacity-40 grayscale'}"
			style:box-shadow={active && color ? `0 0 0 2px ${color}` : undefined}
			loading="lazy"
		/>
	{:else}
		<div class="{sizeClass} shrink-0 rounded-full bg-elevated"></div>
	{/if}
{/snippet}

<!-- hàng chọn cây trên cùng của mỗi cột; cây phụ không hiện cây chính -->
{#snippet treeRow(selectedId: number, excludeId?: number)}
	<div class="flex justify-center gap-2">
		{#each trees.filter((t) => t.id !== excludeId) as t (t.id)}
			{@render runeIcon(t.imgUrl, t.name, t.id === selectedId, 'size-6 p-1')}
		{/each}
	</div>
{/snippet}

<section class="rounded-lg border border-line bg-surface">
	<h2 class="border-b border-line px-4 py-2.5 text-sm font-semibold">Runes</h2>

	{#if loading}
		<div class="flex animate-pulse gap-4 p-4">
			<div class="flex w-40 shrink-0 flex-col gap-2">
				{#each { length: PAGE_COUNT }, i (i)}
					<div class="h-[84px] rounded-lg bg-elevated"></div>
				{/each}
			</div>
			<div class="h-[330px] flex-1 rounded-lg bg-elevated/60"></div>
		</div>
	{:else if !page}
		<p class="px-4 py-10 text-center text-sm text-muted">No data</p>
	{:else}
		<div class="flex gap-4 p-4">
			<!-- tab bộ rune (bên trái): cây chính + keystone + cây phụ, dưới là pick rate / số trận | win rate -->
			<div class="flex w-40 shrink-0 flex-col gap-2">
				{#each pages as p, i (i)}
					{@const active = i === selected}
					{@const primary = lolData.runeById.get(p.top.runePrimaryStyle)}
					{@const keystone = lolData.runeById.get(p.top.keyRune)}
					{@const sub = lolData.runeById.get(p.top.runeSubStyle)}
					<button
						type="button"
						onclick={() => (selected = i)}
						aria-pressed={active}
						class="rounded-lg px-2 py-2 outline-none ring-1 ring-line transition-colors focus-visible:ring-2 focus-visible:ring-accent/60 {active
							? 'bg-white/8'
							: 'hover:bg-white/[0.03]'}"
					>
						<div class="flex items-center justify-center gap-1.5 {active ? '' : 'opacity-60'}">
							<img src={primary?.imgUrl} alt={primary?.name} title={primary?.name} class="size-5" />
							<img
								src={keystone?.imgUrl}
								alt={keystone?.name}
								title={keystone?.name}
								class="size-9 rounded-full bg-black/40"
							/>
							<img src={sub?.imgUrl} alt={sub?.name} title={sub?.name} class="size-5" />
						</div>
						<div class="mt-2 grid grid-cols-2 divide-x divide-line text-center text-xs text-muted">
							<div>
								<div class="text-xs font-bold text-ink">{pct2(p.games / total)}</div>
								<div>{p.games.toLocaleString('en-US')} Games</div>
							</div>
							<!-- ô cao bằng ô pick rate (2 dòng) => căn giữa dọc để win rate nằm giữa 2 dòng đó -->
							<div class="flex items-center justify-center">
								<div class="text-xs font-bold {winRateColor(p.wins / p.games)}">{pct2(p.wins / p.games)}</div>
							</div>
						</div>
					</button>
				{/each}
			</div>

			<!-- rune của tab đang chọn: cây chính | cây phụ + shard -->
			<div class="grid flex-1 grid-cols-2 gap-4">
				<div class="flex flex-col gap-4">
					{@render treeRow(page.runePrimaryStyle)}
					{#each primaryRows as row, i (i)}
						<div class="flex items-center justify-around">
							{#each row as r (r.id)}
								{@render runeIcon(
									r.imgUrl,
									r.name,
									picked.has(r.id),
									i === 0 ? 'size-10' : 'size-8',
									TREE_COLORS[page.runePrimaryStyle]
								)}
							{/each}
						</div>
					{/each}
				</div>

				<div class="flex flex-col gap-4">
					{@render treeRow(page.runeSubStyle, page.runePrimaryStyle)}
					{#each subRows as row, i (i)}
						<div class="flex items-center justify-around">
							{#each row as r (r.id)}
								{@render runeIcon(r.imgUrl, r.name, picked.has(r.id), 'size-8', TREE_COLORS[page.runeSubStyle])}
							{/each}
						</div>
					{/each}
					<div class="flex flex-col gap-2.5">
						{#each STAT_SHARD_ROWS as row, i (i)}
							<div class="flex items-center justify-around">
								{#each row as id (id)}
									{@const shard = STAT_SHARDS[id]}
									{@render runeIcon(
										shard?.imgUrl,
										shard?.name ?? String(id),
										page.statRunes[i] === id,
										'size-6 p-1',
										SHARD_COLOR
									)}
								{/each}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{/if}
</section>
