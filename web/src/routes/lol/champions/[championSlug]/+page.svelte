<script lang="ts">
	import { untrack, type Snippet } from 'svelte';

	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import {
		DEFAULT_META_SERVER,
		getChampionAnalytics,
		isApiError,
		META_SERVERS,
		type ChampionStat,
		type MetaServer,
		type Position,
		type RankBucket
	} from '$lib/api';
	import ItemBuildPanel from '$lib/components/champion-detail/ItemBuildPanel.svelte';
	import RunePanel from '$lib/components/champion-detail/RunePanel.svelte';
	import SkillOrderPanel from '$lib/components/champion-detail/SkillOrderPanel.svelte';
	import SpellComboPanel from '$lib/components/champion-detail/SpellComboPanel.svelte';
	import StatTable from '$lib/components/champion-detail/StatTable.svelte';
	import ChampionIcon from '$lib/components/shared/ChampionIcon.svelte';
	import PageMeta from '$lib/components/shared/PageMeta.svelte';
	import MetaServerSelect from '$lib/components/shared/MetaServerSelect.svelte';
	import TierBadge from '$lib/components/shared/TierBadge.svelte';
	import Tooltip from '$lib/components/shared/Tooltip.svelte';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { championHref } from '$lib/utils/champion';
	import { POSITION_ICONS } from '$lib/utils/positions';

	let { data } = $props();

	const POSITION_LABEL: Record<Position, string> = {
		TOP: 'Top',
		JGL: 'Jungle',
		MID: 'Mid',
		ADC: 'ADC',
		SPT: 'Support',
		UNK: 'Other'
	};

	// Các lane có nút lọc; UNK không phải lane nên không có nút.
	const FILTER_POSITIONS: Position[] = ['TOP', 'JGL', 'MID', 'ADC', 'SPT'];

	// Position chiếm dưới 1% số trận của champion thì không hiện nút (mẫu quá nhỏ).
	const MIN_POSITION_SHARE = 0.01;

	const RANK_BUCKET_LABEL: Record<RankBucket, string> = {
		MASTER_PLUS: 'Master+'
	};

	// slug trên URL có thể gõ tay khác hoa thường => so khớp lowercase.
	const champion = $derived.by(() => {
		const slug = data.slug.toLowerCase();
		for (const c of lolData.championById.values()) {
			if (c.slug.toLowerCase() === slug) return c;
		}
		return undefined;
	});

	let stats = $state<ChampionStat[] | null>(null);
	let loading = $state(true);
	let error = $state<unknown>(null);

	// Trang không prerender nên đọc query ngay lúc khởi tạo được. Giá trị rác => default.
	// position null = chưa chọn => tự lấy position nhiều trận nhất (không ghi lên URL).
	const initialParams = page.url.searchParams;
	let server = $state<MetaServer>(
		oneOf(initialParams.get('server')?.toUpperCase() ?? null, META_SERVERS, DEFAULT_META_SERVER)
	);
	let position = $state<Position | null>(
		oneOf(initialParams.get('position')?.toUpperCase() ?? null, FILTER_POSITIONS, null)
	);

	// replaceState (không push) để đổi filter không làm đầy history.
	$effect(() => {
		const params = new URLSearchParams();
		if (server !== DEFAULT_META_SERVER) params.set('server', server);
		if (position) params.set('position', position);

		const search = params.size ? `?${params}` : '';
		if (search === location.search) return;
		replaceState(`${location.pathname}${search}`, page.state);
	});

	// Link matchup dẫn sang champion khác cùng route nên SvelteKit giữ nguyên component: đọc lại
	// position từ URL mới và bỏ stats của champion cũ (không thì hiện nhầm trong lúc tải).
	let currentSlug = untrack(() => data.slug);
	$effect.pre(() => {
		if (data.slug === currentSlug) return;
		currentSlug = data.slug;
		position = oneOf(page.url.searchParams.get('position')?.toUpperCase() ?? null, FILTER_POSITIONS, null);
		stats = null;
	});

	// Đổi champion / server liên tục thì response cũ có thể về sau: chỉ nhận request mới nhất.
	let reqId = 0;

	$effect(() => {
		const c = champion;
		const s = server;
		if (!c) return;
		untrack(() => load(c.id, s));
	});

	async function load(championId: number, s: MetaServer) {
		const id = ++reqId;
		loading = true;
		error = null;

		try {
			const st = await getChampionAnalytics(championId, { server: s });
			if (id === reqId) stats = st;
		} catch (e) {
			if (id === reqId) error = e;
		} finally {
			if (id === reqId) loading = false;
		}
	}

	// Tỉ lệ trận của từng lane trên tổng trận champion (mọi position), bỏ lane dưới 1%,
	// lane chơi nhiều nhất đứng đầu.
	const positionShares = $derived.by(() => {
		if (!stats?.length) return [];
		const total = stats.reduce((sum, s) => sum + s.games, 0);
		return FILTER_POSITIONS.map((p) => ({
			position: p,
			share: (stats!.find((s) => s.position === p)?.games ?? 0) / total
		}))
			.filter((x) => x.share >= MIN_POSITION_SHARE)
			.sort((a, b) => b.share - a.share);
	});

	// Position đang chọn; chưa chọn, hoặc lane đó không có nút (dưới 1% / đổi server mà
	// champion không chơi lane đó) thì lấy position nhiều trận nhất.
	const stat = $derived.by(() => {
		if (!stats?.length) return null;
		const isShown = positionShares.some((x) => x.position === position);
		return (
			(isShown && stats.find((s) => s.position === position)) ||
			stats.reduce((a, b) => (b.games > a.games ? b : a))
		);
	});

	// Title / description: có stat thì kèm lane + patch, chưa có thì chỉ tên champion.
	const championName = $derived(champion?.name ?? data.slug);
	const metaTitle = $derived(
		`${championName}${stat ? ` ${POSITION_LABEL[stat.position]}` : ''} Build, Runes & Counters`
	);
	const metaDescription = $derived(
		stat
			? `${championName} ${POSITION_LABEL[stat.position]} build for patch ${stat.patch} (${RANK_BUCKET_LABEL[stat.rankBucket]}): best runes, summoner spells, skill order, items and matchups.`
			: `${championName} build: best runes, summoner spells, skill order, items and matchups.`
	);

	function oneOf<T extends string, F>(value: string | null, allowed: readonly T[], fallback: F): T | F {
		return allowed.includes(value as T) ? (value as T) : fallback;
	}
</script>

{#snippet statCell(label: string, value: Snippet)}
	<div class="px-5 py-2.5 text-center">
		<div class="text-xs text-muted">{label}</div>
		<div class="mt-1.5 flex h-6 items-center justify-center text-[15px] font-bold">
			{#if stat}
				{@render value()}
			{:else}
				<div class="h-4 w-10 animate-pulse rounded bg-elevated"></div>
			{/if}
		</div>
	</div>
{/snippet}

{#snippet percent(v: number)}
	{(v * 100).toFixed(1)}<span class="ml-0.5 text-xs font-normal text-muted">%</span>
{/snippet}

<PageMeta title={metaTitle} description={metaDescription} />

<div class="mx-auto max-w-[1100px] px-5 pt-4">
	{#if lolData.isLoaded && !champion}
		<p class="py-16 text-center text-muted">Không tìm thấy champion "{data.slug}".</p>
	{:else}
		<section class="rounded-lg border border-line bg-surface p-4">
			<div class="flex gap-4">
				{#if champion}
					<Tooltip content={champion.name} class="flex shrink-0">
						<ChampionIcon
							src={champion.imgUrl}
							alt={champion.name}
							class="size-24 rounded-lg border border-line"
						/>
					</Tooltip>
				{:else}
					<div class="size-24 shrink-0 animate-pulse rounded-lg bg-elevated"></div>
				{/if}

				<div class="min-w-0">
					<div class="flex items-baseline gap-2.5">
						<h1 class="text-2xl font-bold">{champion?.name ?? ''}</h1>
						{#if stat}
							<span class="text-sm text-muted">
								{POSITION_LABEL[stat.position]} Build, {RANK_BUCKET_LABEL[stat.rankBucket]}, Patch {stat.patch}
							</span>
						{/if}
					</div>

					<!-- position (kèm % số trận của lane) + server, giống toolbar trang tier list -->
					<div class="mt-2 flex flex-wrap items-center gap-2">
						{#if positionShares.length}
							<div class="flex h-10 items-center divide-x divide-line rounded-lg bg-elevated ring-1 ring-line">
								{#each positionShares as x (x.position)}
									{@const active = stat?.position === x.position}
									<button
										type="button"
										onclick={() => (position = x.position)}
										aria-pressed={active}
										title={POSITION_LABEL[x.position]}
										class="flex h-full items-center gap-1.5 px-3 text-sm outline-none transition-colors focus-visible:ring-2 focus-visible:ring-accent/60 {active
											? 'bg-white/12 text-ink'
											: 'text-muted hover:bg-white/5'}"
									>
										<img
											src={POSITION_ICONS[x.position]}
											alt={x.position}
											class="size-4 object-contain {active ? '' : 'opacity-70'}"
										/>
										{(x.share * 100).toFixed(1)}%
									</button>
								{/each}
							</div>
						{:else if loading}
							<div class="h-10 w-48 animate-pulse rounded-lg bg-elevated"></div>
						{/if}

						<MetaServerSelect bind:server />
					</div>
				</div>
			</div>

			{#if error && !loading}
				<p class="mt-4 text-sm text-muted">
					Không tải được dữ liệu{isApiError(error) ? ` (HTTP ${error.status})` : ''}.
				</p>
			{:else if !loading && !stat}
				<p class="mt-4 text-sm text-muted">Chưa có dữ liệu meta cho champion này ở patch hiện tại.</p>
			{:else}
				<!-- chỉ số của position đang chọn, nằm dưới cùng của island header -->
				<div class="mt-4 inline-flex divide-x divide-line rounded-lg border border-line">
					<!-- stat null (đang tải) thì statCell hiện ô giả, không gọi các snippet dưới -->
					{#snippet tier()}<TierBadge tier={stat!.tier} />{/snippet}
					{#snippet winRate()}{@render percent(stat!.winRate)}{/snippet}
					{#snippet pickRate()}{@render percent(stat!.pickRate)}{/snippet}
					{#snippet banRate()}{@render percent(stat!.banRate)}{/snippet}
					{#snippet games()}{stat!.games.toLocaleString('en-US')}{/snippet}
					{@render statCell('Tier', tier)}
					{@render statCell('Win Rate', winRate)}
					{@render statCell('Pick Rate', pickRate)}
					{@render statCell('Ban Rate', banRate)}
					{@render statCell('Games', games)}
				</div>
			{/if}
		</section>

		<!-- cột trái chiếm 2/3 (runes, spells, skill order); cột phải xếp dọc Matchups rồi Legendary Items -->
		{#if !error && (loading || stat)}
			<div class="mt-3 grid items-start gap-3 lg:grid-cols-3">
				<div class="flex min-w-0 flex-col gap-3 lg:col-span-2">
					<RunePanel runes={stat?.bestRunes ?? []} loading={!stat} />
					<SpellComboPanel combos={stat?.bestSpellCombos ?? []} loading={!stat} />
					<SkillOrderPanel orders={stat?.bestSkillsLeveled ?? []} {champion} loading={!stat} />
					<ItemBuildPanel
						starterSets={stat?.bestStarterSets ?? []}
						bootItems={stat?.bestBootItems ?? []}
						coreBuilds={stat?.bestFirstThreeItems ?? []}
						games={stat?.games ?? 0}
						loading={!stat}
					/>
				</div>

				<div class="flex min-w-0 flex-col gap-3">
					<StatTable
						title="Matchups"
						label="Opponent"
						rows={stat?.bestMatchUps ?? []}
						total={stat?.games ?? 0}
						rowKey={(r) => r.opponentChampionId}
						loading={!stat}
					>
						{#snippet entity(r)}
							{@const opp = lolData.championById.get(r.opponentChampionId)}
							{#if opp}
								<a
									href={championHref(opp.slug, { position: stat?.position, server })}
									class="group flex items-center gap-2"
								>
									<Tooltip content={opp.name} class="inline-flex shrink-0">
										<ChampionIcon src={opp.imgUrl} alt={opp.name} class="size-7 rounded-full" />
									</Tooltip>
									<span class="truncate font-medium group-hover:underline">{opp.name}</span>
								</a>
							{:else}
								<div class="flex items-center gap-2">
									<div class="size-7 shrink-0 rounded-full bg-elevated"></div>
									<span class="truncate font-medium">{r.opponentChampionId}</span>
								</div>
							{/if}
						{/snippet}
					</StatTable>

					<StatTable
						title="Items"
						label="Item"
						rows={stat?.bestLegendaryItems ?? []}
						total={stat?.games ?? 0}
						rowKey={(r) => r.itemId}
						loading={!stat}
					>
						{#snippet entity(r)}
							{@const item = lolData.itemById.get(r.itemId)}
							<div class="flex items-center gap-2">
								{#if item}
									<Tooltip content={item.name} class="inline-flex shrink-0">
										<img src={item.imgUrl} alt={item.name} class="size-7 shrink-0 rounded" loading="lazy" />
									</Tooltip>
								{:else}
									<div class="size-7 shrink-0 rounded bg-elevated"></div>
								{/if}
								<span class="truncate font-medium">{item?.name ?? r.itemId}</span>
							</div>
						{/snippet}
					</StatTable>
				</div>
			</div>
		{/if}
	{/if}
</div>
