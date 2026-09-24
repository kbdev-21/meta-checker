<script lang="ts">
	import Star from '@lucide/svelte/icons/star';
	import { untrack } from 'svelte';

	import {
		findPlayerByInfo,
		getMatchesByPlayerInfo,
		isApiError,
		MATCH_LIST_MAX_COUNT,
		type Match,
		type MatchListMode,
		type Player,
		type Position
	} from '$lib/api';
	import MatchCard from '$lib/components/player/MatchCard.svelte';
	import PerfScoreBadge from '$lib/components/player/PerfScoreBadge.svelte';
	import RankCard from '$lib/components/player/RankCard.svelte';
	import ChampionIcon from '$lib/components/shared/ChampionIcon.svelte';
	import PageMeta from '$lib/components/shared/PageMeta.svelte';
	import ServerBadge from '$lib/components/shared/ServerBadge.svelte';
	import { isFollowed, toggleFollow } from '$lib/stores/followed-players.svelte';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { championHref } from '$lib/utils/champion';
	import { ddragonVersionOf, profileIconUrl } from '$lib/utils/ddragon';
	import { pct0, rankLabel } from '$lib/utils/format';
	import { POSITION_ICONS } from '$lib/utils/positions';

	let { data } = $props();

	// Tab chưa làm, tạm hiển thị cho đủ layout.
	const TABS = ['Summary', 'Champions', 'Live Game'] as const;

	const MODE_FILTERS: { value: MatchListMode | 'ALL'; label: string }[] = [
		{ value: 'ALL', label: 'All' },
		{ value: 'SOLO', label: 'Ranked Solo' },
		{ value: 'FLEX', label: 'Ranked Flex' }
	];

	// Lần đầu lấy 20 trận (= MATCH_LIST_MAX_COUNT), mỗi lần "Show more" lấy thêm 10.
	const INITIAL_MATCH_COUNT = 10;
	const MORE_MATCH_COUNT = 10;

	const SUMMARY_POSITIONS: Position[] = ['TOP', 'JGL', 'MID', 'ADC', 'SPT'];
	const SIDEBAR_CHAMPION_COUNT = 7;
	const SUMMARY_CHAMPION_COUNT = 3;

	let player = $state<Player | null>(null);
	let playerLoading = $state(true);
	let playerError = $state<unknown>(null);

	let mode = $state<MatchListMode | 'ALL'>('ALL');
	let matches = $state<Match[]>([]);
	let matchesLoading = $state(true);
	let matchesError = $state<unknown>(null);
	let hasMore = $state(false);

	// Response cũ (player trước / mode trước) có thể về sau response mới: chỉ nhận request mới nhất.
	let playerReq = 0;
	let matchReq = 0;

	// Đổi player (navigate sang trang player khác) => load lại. untrack để effect chỉ phụ
	// thuộc đúng các giá trị liệt kê, không dính state đọc bên trong hàm load.
	$effect(() => {
		void [data.server, data.name, data.tag];
		untrack(loadPlayer);
	});

	$effect(() => {
		void [data.server, data.name, data.tag, mode];
		untrack(() => loadMatches(0));
	});

	async function loadPlayer() {
		const id = ++playerReq;
		playerLoading = true;
		playerError = null;
		try {
			const p = await findPlayerByInfo(data.server, data.name, data.tag);
			if (id === playerReq) player = p;
		} catch (e) {
			if (id === playerReq) playerError = e;
		} finally {
			if (id === playerReq) playerLoading = false;
		}
	}

	// start = 0: load lại từ đầu (đổi player / mode); start > 0: "Show more", nối vào cuối.
	async function loadMatches(start: number) {
		const id = ++matchReq;
		const count = start === 0 ? INITIAL_MATCH_COUNT : MORE_MATCH_COUNT;
		matchesLoading = true;
		matchesError = null;
		if (start === 0) matches = [];
		try {
			const list =
				(await getMatchesByPlayerInfo(data.server, data.name, data.tag, {
					mode: mode === 'ALL' ? undefined : mode,
					start,
					count
				})) ?? [];
			if (id !== matchReq) return;
			matches = start === 0 ? list : [...matches, ...list];
			// Trả đủ số đã xin thì có thể còn trận cũ hơn.
			hasMore = list.length === count;
		} catch (e) {
			if (id === matchReq) matchesError = e;
		} finally {
			if (id === matchReq) matchesLoading = false;
		}
	}

	const followed = $derived(player !== null && isFollowed(player.id));

	const ddVersion = $derived(
		ddragonVersionOf(lolData.championById.values().next().value?.imgUrl ?? '')
	);

	// Thống kê từ các trận đã load (bỏ remake), dùng cho card Summary + list champion bên trái.
	const summary = $derived.by(() => {
		const mine = matches
			.filter((m) => !m.isRemake)
			.map((m) => m.participants.find((p) => p.playerId === player?.id))
			.filter((p) => p !== undefined);

		const games = mine.length;
		const wins = mine.filter((p) => p.isWin).length;
		const totalScore = mine.reduce((a, p) => a + p.perfScore, 0);

		const byChampion = new Map<number, { championId: number; games: number; wins: number; kills: number; deaths: number; assists: number }>();
		for (const p of mine) {
			const c = byChampion.get(p.championId) ?? { championId: p.championId, games: 0, wins: 0, kills: 0, deaths: 0, assists: 0 };
			c.games++;
			if (p.isWin) c.wins++;
			c.kills += p.kills;
			c.deaths += p.deaths;
			c.assists += p.assists;
			byChampion.set(p.championId, c);
		}

		return {
			games,
			wins,
			losses: games - wins,
			// Trung bình perfScore, làm tròn để hiển thị bằng PerfScoreBadge như từng trận.
			avgScore: games ? Math.round(totalScore / games) : null,
			champions: [...byChampion.values()].sort((a, b) => b.games - a.games),
			positions: SUMMARY_POSITIONS.map((pos) => ({
				position: pos,
				games: mine.filter((p) => p.position === pos).length
			}))
		};
	});

	// Title / description: lấy name/tag từ player khi đã load (đúng hoa thường), chưa có thì từ URL.
	const riotId = $derived(player ? `${player.name}#${player.tag}` : `${data.name}#${data.tag}`);
	const metaDescription = $derived(
		`${riotId} League of Legends profile on ${data.server}${player ? ` (Solo/Duo: ${rankLabel(player.soloRank, player.soloTier)})` : ''}: rank, recent matches, champion stats and performance score.`
	);

	const kdaOf = (c: { kills: number; deaths: number; assists: number }) =>
		((c.kills + c.assists) / Math.max(c.deaths, 1)).toFixed(2);

	// Vòng tròn win rate: chu vi của r = 28.
	const RING_CIRCUMFERENCE = 2 * Math.PI * 28;
</script>

<PageMeta title="{riotId} ({data.server})" description={metaDescription} />

{#if playerLoading && !player}
	<!-- skeleton phần header trong lúc chờ player -->
	<div class="mx-auto max-w-[1100px] px-5 pt-4">
		<div
			class="flex animate-pulse items-center gap-5 rounded-lg border border-line bg-surface px-6 py-6"
		>
			<div class="size-24 rounded-2xl bg-elevated"></div>
			<div class="flex flex-col gap-2">
				<div class="h-6 w-56 rounded bg-elevated"></div>
				<div class="h-4 w-32 rounded bg-elevated"></div>
			</div>
		</div>
	</div>
{:else if playerError}
	<p class="py-16 text-center text-muted">
		Không tải được player{isApiError(playerError) ? ` (HTTP ${playerError.status})` : ''}.
	</p>
{:else if !player}
	<p class="py-16 text-center text-muted">
		Không tìm thấy player <span class="text-ink">{data.name}#{data.tag}</span> ở server {data.server}.
	</p>
{:else}
	<!-- ===== header player: card riêng, cùng max width với nội dung trang ===== -->
	<div class="mx-auto max-w-[1100px] px-5 pt-4">
		<div class="rounded-lg border border-line bg-surface px-6">
			<div class="flex items-center gap-5 py-6">
				<div class="relative shrink-0">
					{#if ddVersion && player.profileIconId !== null}
						<img
							src={profileIconUrl(ddVersion, player.profileIconId)}
							alt=""
							class="size-24 rounded-2xl"
						/>
					{:else}
						<div class="size-24 rounded-2xl bg-elevated"></div>
					{/if}
					{#if player.level !== null}
						<span
							class="absolute -bottom-2 left-1/2 -translate-x-1/2 rounded-full border border-line bg-base px-2 text-xs font-semibold"
						>
							{player.level}
						</span>
					{/if}
				</div>

				<div class="min-w-0">
					<h1 class="truncate text-2xl font-bold">
						{player.name} <span class="font-normal text-muted">#{player.tag}</span>
					</h1>
					<div class="mt-1 flex">
						<ServerBadge server={player.server} size="MD" />
					</div>
					<div class="mt-3 flex items-center gap-3">
						<!-- chưa follow: nút vàng nổi bật; đã follow: nút tối, sao tô vàng -->
						<button
							type="button"
							onclick={() => toggleFollow(player!)}
							aria-pressed={followed}
							class="flex h-9 items-center gap-2 rounded-lg px-4 text-sm font-semibold outline-none transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-accent/60 {followed
								? 'bg-elevated text-ink ring-1 ring-line'
								: 'bg-accent text-black'}"
						>
							<Star class="size-4 {followed ? 'fill-accent text-accent' : ''}" />
							{followed ? 'Following' : 'Follow'}
						</button>					</div>
				</div>
			</div>

			<!-- tab nằm ở đáy card, gạch chân active trùng mép dưới card -->
			<nav class="flex gap-6 text-sm">
				{#each TABS as tab, i (tab)}
					<button
						type="button"
						disabled={i !== 0}
						title={i !== 0 ? 'Coming soon' : undefined}
						class="border-b-2 pb-2.5 font-medium outline-none {i === 0
							? 'border-accent text-ink'
							: 'border-transparent text-muted disabled:cursor-not-allowed'}"
					>
						{tab}
					</button>
				{/each}
			</nav>
		</div>
	</div>

	<!-- ===== body ===== -->
	<div class="mx-auto grid max-w-[1100px] gap-3 px-5 py-4 lg:grid-cols-[300px_1fr]">
		<!-- cột trái -->
		<aside class="flex flex-col gap-3">
			<RankCard
				title="Ranked Solo/Duo"
				rank={player.soloRank}
				tier={player.soloTier}
				lp={player.soloLp}
				wins={player.soloWins}
				losses={player.soloLosses}
			/>
			<RankCard
				title="Ranked Flex"
				rank={player.flexRank}
				tier={player.flexTier}
				lp={player.flexLp}
				wins={player.flexWins}
				losses={player.flexLosses}
			/>

			<section class="rounded-lg border border-line bg-surface">
				<h2 class="border-b border-line px-4 py-2.5 text-sm font-semibold">Recent champions</h2>
				{#if summary.champions.length}
					<ul class="divide-y divide-line/50">
						{#each summary.champions.slice(0, SIDEBAR_CHAMPION_COUNT) as c (c.championId)}
							{@const champ = lolData.championById.get(c.championId)}
							<li class="flex items-center gap-3 px-4 py-2 text-xs">
								{#if champ}
									<a href={championHref(champ.slug)}>
										<ChampionIcon src={champ.imgUrl} alt={champ.name} class="size-8 rounded-full" />
									</a>
								{:else}
									<div class="size-8 rounded-full bg-elevated"></div>
								{/if}
								<div class="min-w-0 flex-1">
									{#if champ}
										<a href={championHref(champ.slug)} class="block truncate text-sm font-semibold hover:underline">
											{champ.name}
										</a>
									{:else}
										<div class="truncate text-sm font-semibold">{c.championId}</div>
									{/if}
									<div class="text-muted">{kdaOf(c)}:1 KDA</div>
								</div>
								<div class="text-right">
									<div class="font-semibold {c.wins / c.games >= 0.6 ? 'text-red-400' : ''}">
										{pct0(c.wins / c.games)}
									</div>
									<div class="text-muted">{c.games} games</div>
								</div>
							</li>
						{/each}
					</ul>
				{:else}
					<p class="px-4 py-6 text-center text-xs text-muted">No games</p>
				{/if}
			</section>
		</aside>

		<!-- cột phải -->
		<main class="flex min-w-0 flex-col gap-3">
			<!-- lọc mode -->
			<div class="flex gap-1 rounded-lg border border-line bg-surface p-1 text-sm">
				{#each MODE_FILTERS as f (f.value)}
					<button
						type="button"
						onclick={() => (mode = f.value)}
						aria-pressed={mode === f.value}
						class="rounded-md px-3 py-1.5 outline-none transition-colors focus-visible:ring-2 focus-visible:ring-accent/60 {mode ===
						f.value
							? 'bg-elevated font-semibold text-ink'
							: 'text-muted hover:text-ink'}"
					>
						{f.label}
					</button>
				{/each}
			</div>

			<!-- tóm tắt các trận đã load -->
			<section class="flex flex-wrap items-center gap-8 rounded-lg border border-line bg-surface px-5 py-4">
				<div class="flex items-center gap-4">
					<!-- win rate nằm giữa vòng tròn -->
					<div class="relative size-16 shrink-0">
						<svg viewBox="0 0 64 64" class="size-16 -rotate-90">
							<circle cx="32" cy="32" r="28" fill="none" stroke-width="7" class="stroke-red-500/70" />
							<circle
								cx="32"
								cy="32"
								r="28"
								fill="none"
								stroke-width="7"
								class="stroke-sky-500"
								stroke-dasharray={RING_CIRCUMFERENCE}
								stroke-dashoffset={RING_CIRCUMFERENCE * (1 - (summary.games ? summary.wins / summary.games : 0))}
							/>
						</svg>
						<span class="absolute inset-0 grid place-items-center text-sm font-bold">
							{summary.games ? pct0(summary.wins / summary.games) : '-'}
						</span>
					</div>
					<div class="text-xs text-muted">{summary.games}G {summary.wins}W {summary.losses}L</div>
				</div>

				<div class="flex flex-col items-center gap-1">
					<div class="text-xs text-muted">Avg Score</div>
					{#if summary.avgScore !== null}
						<PerfScoreBadge score={summary.avgScore} />
					{:else}
						<span class="text-lg font-bold text-muted">-</span>
					{/if}
				</div>

				<ul class="flex flex-col gap-1.5">
					{#each summary.champions.slice(0, SUMMARY_CHAMPION_COUNT) as c (c.championId)}
						{@const champ = lolData.championById.get(c.championId)}
						<li class="flex items-center gap-2 text-xs">
							{#if champ}
								<a href={championHref(champ.slug)} title={champ.name}>
									<ChampionIcon src={champ.imgUrl} alt={champ.name} class="size-6 rounded-full" />
								</a>
							{/if}
							<span class="font-semibold">{pct0(c.wins / c.games)}</span>
							<span class="text-muted">({c.wins}W {c.games - c.wins}L)</span>
							<span class="text-muted">{kdaOf(c)} KDA</span>
						</li>
					{/each}
				</ul>

				<!-- tỉ lệ vị trí -->
				<div class="ml-auto flex items-end gap-2">
					{#each summary.positions as p (p.position)}
						<div class="flex flex-col items-center gap-1">
							<div class="relative h-14 w-3 overflow-hidden rounded-sm bg-elevated">
								<div
									class="absolute inset-x-0 bottom-0 bg-accent"
									style:height="{summary.games ? (p.games / summary.games) * 100 : 0}%"
								></div>
							</div>
							<img src={POSITION_ICONS[p.position]} alt={p.position} title={p.position} class="size-4 object-contain" />
						</div>
					{/each}
				</div>
			</section>

			<!-- danh sách trận -->
			{#if matchesError}
				<p class="py-8 text-center text-sm text-muted">
					Không tải được lịch sử đấu{isApiError(matchesError) ? ` (HTTP ${matchesError.status})` : ''}.
				</p>
			{:else}
				<div class="flex flex-col gap-2">
					{#each matches as m (m.id)}
						<MatchCard match={m} playerId={player.id} server={player.server} />
					{/each}

					{#if matchesLoading}
						{#each { length: matches.length ? 2 : 5 }, i (i)}
							<div class="h-[108px] animate-pulse rounded-lg bg-surface"></div>
						{/each}
					{:else if !matches.length}
						<p class="py-8 text-center text-sm text-muted">No matches</p>
					{/if}
				</div>

				{#if hasMore && !matchesLoading}
					<button
						type="button"
						onclick={() => loadMatches(matches.length)}
						class="h-10 rounded-lg border border-line bg-surface text-sm text-muted outline-none transition-colors hover:text-ink focus-visible:ring-2 focus-visible:ring-accent/60"
					>
						Show more
					</button>
				{/if}
			{/if}
		</main>
	</div>
{/if}
