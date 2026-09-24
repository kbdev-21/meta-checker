<script lang="ts" generics="T extends { games: number; wins: number }">
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import type { Snippet } from 'svelte';

	import { winRateColor } from '$lib/utils/format';

	// Bảng thống kê (item, matchup...) của 1 champion: số trận, % xuất hiện, win rate.
	type Props = {
		title: string;
		// Header cột đầu (vd "Item", "Opponent").
		label: string;
		rows: T[];
		// Tổng trận của champion ở position đang xem, để tính % xuất hiện.
		total: number;
		rowKey: (row: T) => number;
		// Nội dung cột đầu (icon + tên).
		entity: Snippet<[T]>;
		loading?: boolean;
	};

	let { title, label, rows, total, rowKey, entity, loading = false }: Props = $props();

	const SORT_KEYS = ['GAMES', 'WIN_RATE'] as const;
	type SortKey = (typeof SORT_KEYS)[number];
	type SortDir = 'DESC' | 'ASC';

	// Mặc định nhiều trận nhất trước.
	let sortKey = $state<SortKey>('GAMES');
	let sortDir = $state<SortDir>('DESC');

	const SORT_VALUE: Record<SortKey, (r: T) => number> = {
		GAMES: (r) => r.games,
		WIN_RATE: (r) => r.wins / r.games
	};

	// Giống tier list: click cột đang sort thì đảo chiều; cột khác thì bắt đầu từ giảm dần.
	const toggleSort = (key: SortKey) => {
		if (sortKey === key) {
			sortDir = sortDir === 'DESC' ? 'ASC' : 'DESC';
		} else {
			sortKey = key;
			sortDir = 'DESC';
		}
	};

	const sorted = $derived.by(() => {
		const value = SORT_VALUE[sortKey];
		const sign = sortDir === 'DESC' ? -1 : 1;
		return [...rows].sort((a, b) => sign * (value(a) - value(b)));
	});

	const SKELETON_ROWS = 8;

	const pct = (v: number) => `${(v * 100).toFixed(1)}%`;
</script>

{#snippet sortableHeader(key: SortKey, text: string)}
	{@const active = sortKey === key}
	<!-- underline bằng inset shadow để không đổi chiều cao header -->
	<th
		class="w-20 whitespace-nowrap px-2 py-2 text-center font-semibold {active ? 'shadow-[inset_0_-2px_0_var(--color-accent)]' : ''}"
		aria-sort={active ? (sortDir === 'DESC' ? 'descending' : 'ascending') : 'none'}
	>
		<button
			type="button"
			onclick={() => toggleSort(key)}
			class="relative tracking-wide outline-none transition-colors hover:text-ink focus-visible:text-ink {active
				? 'text-ink'
				: ''}"
		>
			{text}
			<!-- mũi tên absolute nên không chiếm chỗ, label luôn căn giữa cột -->
			{#if active}
				{@const Arrow = sortDir === 'ASC' ? ArrowUp : ArrowDown}
				<Arrow class="absolute left-full top-1/2 ml-0.5 size-3 -translate-y-1/2" />
			{/if}
		</button>
	</th>
{/snippet}

<section class="overflow-hidden rounded-lg border border-line bg-surface">
	<h2 class="border-b border-line px-4 py-2.5 text-sm font-semibold">{title}</h2>

	<!-- danh sách dài (hàng chục dòng) nên cuộn trong island, header dính trên cùng.
	     556px = header ~31px + 12 dòng x ~43.7px: vừa đủ hiện 12 dòng, còn lại cuộn. -->
	<div class="max-h-[556px] overflow-y-auto">
		<table class="w-full text-xs">
			<!-- z-10: ChampionIcon có transform (scale) nên tạo stacking context, không có z thì bị icon đè khi cuộn -->
			<thead class="sticky top-0 z-10 bg-surface">
				<tr class="border-b border-line text-[11px] text-muted">
					<th class="px-3 py-2 text-left font-semibold tracking-wide">{label}</th>
					{@render sortableHeader('GAMES', 'Games')}
					{@render sortableHeader('WIN_RATE', 'Win rate')}
				</tr>
			</thead>
			<tbody>
				{#if loading}
					{#each { length: SKELETON_ROWS }, i (i)}
						<tr class="animate-pulse border-b border-line/50 last:border-b-0">
							<td class="px-3 py-1.5">
								<div class="flex items-center gap-2">
									<div class="size-7 rounded bg-elevated"></div>
									<div class="h-3 w-24 rounded bg-elevated"></div>
								</div>
							</td>
							{#each { length: 2 }, j (j)}
								<td class="px-2 py-1.5"><div class="mx-auto h-3 w-9 rounded bg-elevated"></div></td>
							{/each}
						</tr>
					{/each}
				{:else if !sorted.length}
					<tr><td colspan="3" class="py-8 text-center text-muted">Chưa có dữ liệu.</td></tr>
				{:else}
					{#each sorted as row (rowKey(row))}
						<tr class="border-b border-line/50 last:border-b-0">
							<td class="max-w-0 px-3 py-1.5">{@render entity(row)}</td>
							<!-- số trận, % xuất hiện bên dưới -->
							<td class="px-2 py-1.5 text-center">
								<div>{row.games.toLocaleString('en-US')}</div>
								<div class="text-[11px] text-muted">{pct(row.games / total)}</div>
							</td>
							<td class="px-2 py-1.5 text-center font-semibold {winRateColor(row.wins / row.games)}">
								{pct(row.wins / row.games)}
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</section>
