<script lang="ts">
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import { onMount } from 'svelte';

	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import {
		getAnalytics,
		isApiError,
		META_SERVERS,
		POSITIONS,
		type ChampionStatSummary,
		type Meta,
		type MetaServer,
		type Position,
		type RankBucket
	} from '$lib/api';
	import ChampionsToolbar from '$lib/components/champions/ChampionsToolbar.svelte';
	import ChampionIcon from '$lib/components/shared/ChampionIcon.svelte';
	import TierBadge from '$lib/components/shared/TierBadge.svelte';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { championHref } from '$lib/utils/champion';
	import { POSITION_ICONS } from '$lib/utils/positions';

	let meta = $state<Meta | null>(null);
	let loading = $state(true);
	let error = $state<unknown>(null);

	// Cột sort được; mặc định TIER giảm dần (tierScore backend tính từ winRate + pickRate).
	const SORT_KEYS = ['TIER', 'WIN_RATE', 'PICK_RATE', 'BAN_RATE', 'DAMAGE'] as const;
	type SortKey = (typeof SORT_KEYS)[number];
	const SORT_DIRS = ['DESC', 'ASC'] as const;
	type SortDir = (typeof SORT_DIRS)[number];

	const POSITION_FILTERS = ['ALL', ...POSITIONS] as const;

	// Giá trị bằng default thì không ghi lên URL cho gọn.
	const DEFAULT_SERVER: MetaServer = 'GLOBAL';
	const DEFAULT_POSITION: Position | 'ALL' = 'TOP';
	const DEFAULT_SORT_KEY: SortKey = 'TIER';
	const DEFAULT_SORT_DIR: SortDir = 'DESC';

	// Bộ lọc: server đổi thì gọi lại API, keyword/position lọc ngay trên client.
	let server = $state<MetaServer>(DEFAULT_SERVER);
	let keyword = $state('');
	let position = $state<Position | 'ALL'>(DEFAULT_POSITION);
	let sortKey = $state<SortKey>(DEFAULT_SORT_KEY);
	let sortDir = $state<SortDir>(DEFAULT_SORT_DIR);

	// Trang được prerender nên không đọc query lúc render được: đợi mount mới đọc URL,
	// ready = true rồi mới fetch và ghi ngược state lên URL (tránh ghi đè query bằng default).
	let ready = $state(false);

	onMount(() => {
		const params = new URL(location.href).searchParams;
		server = oneOf(params.get('server'), META_SERVERS, DEFAULT_SERVER);
		position = oneOf(params.get('position'), POSITION_FILTERS, DEFAULT_POSITION);
		keyword = params.get('q') ?? '';
		sortKey = oneOf(params.get('sort'), SORT_KEYS, DEFAULT_SORT_KEY);
		sortDir = oneOf(params.get('dir'), SORT_DIRS, DEFAULT_SORT_DIR);
		ready = true;
	});

	// replaceState (không push) để gõ keyword / đổi filter không làm đầy history.
	$effect(() => {
		if (!ready) return;

		const params = new URLSearchParams();
		if (server !== DEFAULT_SERVER) params.set('server', server);
		if (position !== DEFAULT_POSITION) params.set('position', position);
		if (keyword) params.set('q', keyword);
		if (sortKey !== DEFAULT_SORT_KEY) params.set('sort', sortKey);
		if (sortDir !== DEFAULT_SORT_DIR) params.set('dir', sortDir);

		const search = params.size ? `?${params}` : '';
		if (search === location.search) return;
		replaceState(`${location.pathname}${search}`, page.state);
	});

	// Đổi server liên tục thì response cũ có thể về sau response mới; chỉ nhận kết quả
	// của request mới nhất. Để let thường vì không cần reactive.
	let reqId = 0;

	$effect(() => {
		if (!ready) return;

		const s = server;
		const id = ++reqId;
		loading = true;
		error = null;
		getAnalytics({ server: s })
			.then((m) => id === reqId && (meta = m))
			.catch((e) => id === reqId && (error = e))
			.finally(() => id === reqId && (loading = false));
	});

	// Pick rate dưới 0.5% thì mẫu quá nhỏ, win rate không đáng tin nên ẩn khỏi tier list.
	const MIN_PICK_RATE = 0.005;

	const SORT_VALUE: Record<SortKey, (r: ChampionStatSummary) => number> = {
		TIER: (r) => r.tierScore,
		WIN_RATE: (r) => r.winRate,
		PICK_RATE: (r) => r.pickRate,
		BAN_RATE: (r) => r.banRate,
		DAMAGE: (r) => r.avgDmgPerMin
	};

	// Click cột đang sort thì đảo chiều; cột khác thì bắt đầu từ giảm dần.
	const toggleSort = (key: SortKey) => {
		if (sortKey === key) {
			sortDir = sortDir === 'DESC' ? 'ASC' : 'DESC';
		} else {
			sortKey = key;
			sortDir = 'DESC';
		}
	};

	const rows = $derived.by(() => {
		const kw = keyword.trim().toLowerCase();
		const value = SORT_VALUE[sortKey];
		const sign = sortDir === 'DESC' ? -1 : 1;
		return (meta?.championStats ?? [])
			.filter((r) => r.pickRate >= MIN_PICK_RATE)
			.filter((r) => position === 'ALL' || r.position === position)
			.filter((r) => {
				if (!kw) return true;
				const name = lolData.championById.get(r.championId)?.name ?? r.championSlug;
				return name.toLowerCase().includes(kw) || r.championSlug.toLowerCase().includes(kw);
			})
			.sort((a, b) => sign * (value(a) - value(b)));
	});

	const RANK_BUCKET_LABEL: Record<RankBucket, string> = {
		MASTER_PLUS: 'Master+'
	};

	// Số dòng giả hiển thị trong lúc chờ data.
	const SKELETON_ROWS = 10;

	// winRate / pickRate / banRate là phân số 0-1.
	const pct = (v: number) => `${(v * 100).toFixed(1)}%`;

	// Matchup gặp ít hơn 3% số trận của champ thì mẫu quá nhỏ, không tính là counter.
	const MIN_MATCHUP_SHARE = 0.03;
	// Champ vẫn thắng quá 48% khi gặp thì đối thủ đó không đủ khắc chế để gọi là counter.
	const MAX_COUNTER_WIN_RATE = 0.45;
	const COUNTER_COUNT = 3;

	// Counter = đối thủ mà champ có win rate thấp nhất khi gặp.
	const topCounters = (r: ChampionStatSummary) =>
		r.bestMatchUps
			.filter((m) => m.games >= r.games * MIN_MATCHUP_SHARE)
			.filter((m) => m.wins / m.games <= MAX_COUNTER_WIN_RATE)
			.sort((a, b) => a.wins / a.games - b.wins / b.games)
			.slice(0, COUNTER_COUNT);

	// Tỉ lệ vật lý / phép / chuẩn trong tổng damage, dùng làm độ rộng từng đoạn của thanh.
	const damageShares = (r: ChampionStatSummary) => {
		const total = r.avgPhysicalDmg + r.avgMagicDmg + r.avgTrueDmg;
		if (total === 0) return { physical: 0, magic: 0, true: 0 };
		return {
			physical: r.avgPhysicalDmg / total,
			magic: r.avgMagicDmg / total,
			true: r.avgTrueDmg / total
		};
	};

	// Query param người dùng tự sửa có thể là giá trị rác => không hợp lệ thì về default.
	function oneOf<T extends string>(value: string | null, allowed: readonly T[], fallback: T): T {
		return allowed.includes(value as T) ? (value as T) : fallback;
	}
</script>

{#snippet sortableHeader(key: SortKey, label: string, width: string)}
	{@const active = sortKey === key}
	<!-- underline bằng inset shadow để không đổi chiều cao header -->
	<th
		class="{width} px-3 py-2.5 text-center font-semibold {active
			? 'shadow-[inset_0_-2px_0_var(--color-accent)]'
			: ''}"
		aria-sort={active ? (sortDir === 'DESC' ? 'descending' : 'ascending') : 'none'}
	>
		<button
			type="button"
			onclick={() => toggleSort(key)}
			class="relative uppercase tracking-wide outline-none transition-colors hover:text-ink focus-visible:text-ink {active
				? 'text-ink'
				: ''}"
		>
			{label}
			<!-- mũi tên absolute bên phải label nên không chiếm chỗ: label luôn căn giữa đúng cột -->
			{#if active}
				{@const Arrow = sortDir === 'ASC' ? ArrowUp : ArrowDown}
				<Arrow class="absolute left-full top-1/2 ml-1 size-3 -translate-y-1/2" />
			{/if}
		</button>
	</th>
{/snippet}

<div class="mx-auto max-w-[1100px] px-5 py-6">
	<div class="mb-4">
		<h1 class="text-xl font-semibold">LOL Champion Tier List</h1>
		{#if meta}
			<p class="mt-1 text-sm text-muted">
				{meta.patch} · {RANK_BUCKET_LABEL[meta.rankBucket]} · {meta.totalMatches.toLocaleString()} matches
			</p>
		{/if}
	</div>

	<ChampionsToolbar bind:keyword bind:position bind:server />

	{#if error && !loading}
		<p class="py-16 text-center text-muted">
			Không tải được dữ liệu{isApiError(error) ? ` (HTTP ${error.status})` : ''}.
		</p>
	{:else if !meta && !loading}
		<p class="py-16 text-center text-muted">Chưa tổng hợp meta cho patch hiện tại.</p>
	{:else}
		<div class="overflow-hidden rounded-lg border border-line bg-surface">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-line font-semibold text-xs uppercase tracking-wide text-muted">
						<th class="w-12 px-3 py-2.5 text-center">#</th>
						<th class="w-24 px-3 py-2.5 text-center">Pos</th>
						<th class="px-3 py-2.5 text-left">Champion</th>
						{@render sortableHeader('TIER', 'Tier', 'w-16')}
						{@render sortableHeader('WIN_RATE', 'Win rate', 'w-28')}
						{@render sortableHeader('PICK_RATE', 'Pick rate', 'w-28')}
						{@render sortableHeader('BAN_RATE', 'Ban rate', 'w-28')}
						{@render sortableHeader('DAMAGE', 'Damage', 'w-28')}
						<th class="w-32 px-3 py-2.5 text-center">Counters</th>
					</tr>
				</thead>
				{#if loading}
					<!-- skeleton: giữ nguyên khung table, ô là các mảng xám cùng kích thước nội dung thật
					     để lúc data về không bị nhảy layout. -->
					<tbody class="animate-pulse">
						{#each { length: SKELETON_ROWS }, i (i)}
							<tr class="border-b border-line/50 last:border-b-0">
								<td class="px-3 py-2"><div class="mx-auto h-3 w-4 rounded bg-elevated"></div></td>
								<td class="px-3 py-2"><div class="mx-auto size-5 rounded bg-elevated"></div></td>
								<td class="px-3 py-2">
									<div class="flex items-center gap-2.5">
										<div class="size-7 rounded bg-elevated"></div>
										<div class="h-3.5 w-24 rounded bg-elevated"></div>
									</div>
								</td>
								<td class="px-3 py-2"><div class="mx-auto size-6 rounded bg-elevated"></div></td>
								<td class="px-3 py-2"><div class="mx-auto h-3.5 w-12 rounded bg-elevated"></div></td>
								<td class="px-3 py-2"><div class="mx-auto h-3.5 w-12 rounded bg-elevated"></div></td>
								<td class="px-3 py-2"><div class="mx-auto h-3.5 w-12 rounded bg-elevated"></div></td>
								<td class="px-3 py-2">
									<div class="mx-auto mb-1.5 h-3 w-12 rounded bg-elevated"></div>
									<div class="mx-auto h-1 w-16 rounded-full bg-elevated"></div>
								</td>
								<td class="px-3 py-2">
									<div class="flex items-center justify-center gap-1">
										<div class="size-6 rounded-full bg-elevated"></div>
										<div class="size-6 rounded-full bg-elevated"></div>
										<div class="size-6 rounded-full bg-elevated"></div>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				{:else}
					<tbody>
						{#each rows as row, i (`${row.championId}-${row.position}`)}
							{@const champ = lolData.championById.get(row.championId)}
							{@const posIcon = POSITION_ICONS[row.position]}
							{@const dmg = damageShares(row)}
							<tr class="border-b border-line/50 transition-colors last:border-b-0 hover:bg-white/[0.03]">
								<td class="px-3 py-2 text-center tabular-nums text-muted">{i + 1}</td>
								<td class="px-3 py-2 text-center">
									{#if posIcon}
										<img src={posIcon} alt={row.position} title={row.position} class="mx-auto size-5 object-contain" />
									{:else}
										<span class="text-muted">{row.position}</span>
									{/if}
								</td>
								<td class="px-3 py-2">
									<a
										href={championHref(row.championSlug, { position: row.position, server })}
										class="group flex w-fit items-center gap-2.5"
									>
										{#if champ}
											<ChampionIcon src={champ.imgUrl} class="size-7 rounded" />
										{:else}
											<div class="size-7 rounded bg-elevated"></div>
										{/if}
										<span class="font-semibold text-white group-hover:underline">{champ?.name ?? row.championSlug}</span>
									</a>
								</td>
								<td class="px-3 py-2">
									<TierBadge tier={row.tier} class="mx-auto" />
								</td>
								<td class="px-3 py-2 text-center tabular-nums">{pct(row.winRate)}</td>
								<td class="px-3 py-2 text-center tabular-nums">{pct(row.pickRate)}</td>
								<td class="px-3 py-2 text-center tabular-nums">{pct(row.banRate)}</td>
								<td class="px-3 py-2">
									<div class="mb-1 text-center text-xs tabular-nums text-white/80">
										{Math.round(row.avgDmgPerMin).toLocaleString()}/min
									</div>
									<div
										class="mx-auto flex h-1 w-16 overflow-hidden rounded-full bg-elevated"
										title="Vật lý {pct(dmg.physical)} · Chuẩn {pct(dmg.true)} · Phép {pct(dmg.magic)}"
									>
										<div class="bg-red-500" style:width="{dmg.physical * 100}%"></div>
										<div class="bg-white" style:width="{dmg.true * 100}%"></div>
										<div class="bg-sky-500" style:width="{dmg.magic * 100}%"></div>
									</div>
								</td>
								<td class="px-3 py-2">
									<div class="flex items-center justify-center gap-1">
										{#each topCounters(row) as m (m.opponentChampionId)}
											{@const opp = lolData.championById.get(m.opponentChampionId)}
											{#if opp}
												<a href={championHref(opp.slug, { position: row.position, server })}>
													<ChampionIcon
														src={opp.imgUrl}
														alt={opp.name}
														title="{opp.name} · {pct(m.wins / m.games)}"
														class="size-6 rounded-full"
													/>
												</a>
											{:else}
												<div class="size-6 rounded-full bg-elevated"></div>
											{/if}
										{/each}
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				{/if}
			</table>
		</div>
	{/if}
</div>
