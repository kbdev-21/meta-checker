<script lang="ts">
	import type { Match, MatchParticipant, Server } from '$lib/api';
	import PerfScoreBadge from '$lib/components/player/PerfScoreBadge.svelte';
	import PerfScoreLabel from '$lib/components/player/PerfScoreLabel.svelte';
	import ChampionIcon from '$lib/components/shared/ChampionIcon.svelte';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { championHref } from '$lib/utils/champion';
	import { pct0 } from '$lib/utils/format';
	import { lastSlotItemOf } from '$lib/utils/match';
	import { playerHref } from '$lib/utils/riot-id';

	type Props = {
		match: Match;
		// Player đang xem trang; dòng của họ được highlight.
		playerId: string;
		server: Server;
	};

	let { match, playerId, server }: Props = $props();

	// Slot cuối (index 6) là trinket.
	const ITEM_SLOTS = 6;

	// Team 1 = blue side, team 2 = red side (theo Riot).
	const TEAM_SIDE: Record<number, string> = { 1: 'Blue team', 2: 'Red team' };

	// Cột: champion+tên | score | KDA | damage | wards | CS | items
	const GRID = 'grid grid-cols-[minmax(0,1fr)_92px_88px_84px_44px_44px_168px] items-center gap-2';

	const teams = $derived(
		[1, 2].map((t) => {
			const players = match.participants.filter((p) => p.team === t);
			return {
				team: t,
				players,
				isWin: !match.isRemake && players[0]?.isWin === true,
				kills: t === 1 ? match.team1Kills : match.team2Kills,
				gold: players.reduce((a, p) => a + p.gold, 0),
				baron: t === 1 ? match.team1BaronKills : match.team2BaronKills,
				dragon: t === 1 ? match.team1DragonKills : match.team2DragonKills,
				herald: t === 1 ? match.team1HeraldKills : match.team2HeraldKills
			};
		})
	);

	const maxDmg = $derived(Math.max(1, ...match.participants.map((p) => p.dmgDealt)));

	// Giống MatchCard: điểm cao nhất team thắng là MVP, team thua là ACE; còn lại là hạng trong trận.
	function scoreBadge(p: MatchParticipant): string | null {
		if (match.isRemake) return null;
		const teamBest = Math.max(...match.participants.filter((x) => x.team === p.team).map((x) => x.perfScore));
		if (p.perfScore === teamBest) return p.isWin ? 'MVP' : 'ACE';
		const rank = match.participants.filter((x) => x.perfScore > p.perfScore).length + 1;
		return `${rank}${rank === 1 ? 'st' : rank === 2 ? 'nd' : rank === 3 ? 'rd' : 'th'}`;
	}

	// Nền dòng (inline style vì % trộn tính động): surface trộn 1 chút màu kết quả;
	// dòng của player đang xem trộn đậm hơn. Remake không có màu kết quả => trộn màu elevated.
	function rowTone(isWin: boolean, isMe: boolean): string {
		const color = match.isRemake
			? 'var(--color-elevated)'
			: isWin
				? 'var(--color-sky-400)'
				: 'var(--color-red-500)';
		const surface = isMe ? 84 : 94;
		return `background-color: color-mix(in oklab, var(--color-surface) ${surface}%, ${color})`;
	}

	const resultText = (isWin: boolean) => (match.isRemake ? 'Remake' : isWin ? 'Victory' : 'Defeat');
	const resultColor = (isWin: boolean) =>
		match.isRemake ? 'text-muted' : isWin ? 'text-sky-400' : 'text-red-500';
	const barColor = (isWin: boolean) => (match.isRemake ? 'bg-muted' : isWin ? 'bg-sky-400' : 'bg-red-500');
</script>

{#snippet icon(src: string | undefined, name: string | undefined, cls: string)}
	{#if src}
		<img {src} alt={name ?? ''} title={name} class={cls} loading="lazy" />
	{:else}
		<div class="{cls} bg-black/30"></div>
	{/if}
{/snippet}

{#snippet teamTable(t: (typeof teams)[number])}
	<div>
		<div class="{GRID} px-3 py-1.5 text-xs text-muted">
			<div>
				<span class="font-bold {resultColor(t.isWin)}">{resultText(t.isWin)}</span>
				<span>({TEAM_SIDE[t.team]})</span>
			</div>
			<div class="text-center">Score</div>
			<div class="text-center">KDA</div>
			<div class="text-center">Damage</div>
			<div class="text-center">Wards</div>
			<div class="text-center">CS</div>
			<div class="text-center">Items</div>
		</div>

		{#each t.players as p (p.playerId)}
			{@const champ = lolData.championById.get(p.championId)}
			{@const spell1 = lolData.spellById.get(p.spell1Id)}
			{@const spell2 = lolData.spellById.get(p.spell2Id)}
			{@const keyRune = lolData.runeById.get(p.keyRune)}
			{@const subStyle = lolData.runeById.get(p.runeSubStyle)}
			{@const badge = scoreBadge(p)}
			{@const isMe = p.playerId === playerId}
			<div class="{GRID} border-t border-black/20 px-3 py-1.5 text-xs" style={rowTone(t.isWin, isMe)}>
				<!-- champion + spell/rune + tên -->
				<div class="flex min-w-0 items-center gap-1.5">
					<div class="relative shrink-0">
						{#if champ}
							<a href={championHref(champ.slug, { position: p.position })}>
								<ChampionIcon src={champ.imgUrl} alt={champ.name} title={champ.name} class="size-8 rounded-full" />
							</a>
						{:else}
							<div class="size-8 rounded-full bg-elevated"></div>
						{/if}
						<span
							class="absolute -bottom-1 -left-1 grid size-4 place-items-center rounded-full bg-base text-[9px] font-semibold"
						>
							{p.champLevel}
						</span>
					</div>
					<div class="grid shrink-0 grid-cols-2 gap-0.5">
						{@render icon(spell1?.imgUrl, spell1?.name, 'size-4 rounded')}
						{@render icon(keyRune?.imgUrl, keyRune?.name, 'size-4 rounded-full bg-black/40')}
						{@render icon(spell2?.imgUrl, spell2?.name, 'size-4 rounded')}
						{@render icon(subStyle?.imgUrl, subStyle?.name, 'size-4 rounded-full bg-black/40 p-0.5')}
					</div>
					<div class="min-w-0">
						<a
							href={playerHref(server, p.name, p.tag)}
							title="{p.name}#{p.tag}"
							class="block truncate hover:underline {isMe ? 'font-bold text-white' : 'font-semibold'}"
						>
							{p.name}
						</a>
						<div class="truncate text-[11px] text-muted">#{p.tag}</div>
					</div>
				</div>

				<!-- score + nhãn, dùng chung component với MatchCard -->
				<!-- nhãn có width cố định để cả cụm căn giữa cột mà badge các dòng vẫn thẳng hàng -->
				<div class="flex items-center justify-center gap-2">
					<PerfScoreBadge score={p.perfScore} size="SM" />
					<div class="w-8">
						{#if badge}
							<PerfScoreLabel label={badge} showCrown={false} />
						{/if}
					</div>
				</div>

				<!-- KDA -->
				<div class="text-center">
					<div class="whitespace-nowrap">
						{p.kills}/{p.deaths}/{p.assists}
						<span class="text-muted">({pct0(p.killParticipation)})</span>
					</div>
					<div class="font-semibold">{p.kda.toFixed(2)}:1</div>
				</div>

				<!-- damage gây ra (không hiện damage nhận) -->
				<div class="text-center" title="{p.dmgDealt.toLocaleString('en-US')} damage dealt">
					<div>{p.dmgDealt.toLocaleString('en-US')}</div>
					<div class="mx-auto mt-1 h-1 w-12 overflow-hidden rounded-full bg-black/30">
						<div class="h-full bg-red-500" style:width="{(p.dmgDealt / maxDmg) * 100}%"></div>
					</div>
				</div>

				<!-- vision score + placed/killed -->
				<div class="text-center" title="Vision score · Wards placed / killed">
					<div>{p.visionScore}</div>
					<div class="text-muted">{p.wardsPlaced} / {p.wardsKilled}</div>
				</div>

				<!-- CS -->
				<div class="text-center">
					<div>{p.cs}</div>
					<div class="text-muted">{p.csPerMin.toFixed(1)}/m</div>
				</div>

				<!-- items -->
				<!-- id 0 = ô trống, itemById.get(0) trả undefined nên tự thành ô xám.
				     Ô cuối là trinket, riêng ADC là giày (xem lastSlotItemOf). -->
				<div class="flex gap-0.5">
					{#each [...p.items.slice(0, ITEM_SLOTS), lastSlotItemOf(p)] as itemId, i (i)}
						{@const item = lolData.itemById.get(itemId)}
						<div class={i === ITEM_SLOTS ? 'ml-0.5' : ''}>
							{@render icon(item?.imgUrl, item?.name, 'size-[22px] rounded-md')}
						</div>
					{/each}
				</div>
			</div>
		{/each}
	</div>
{/snippet}

{#snippet objectives(t: (typeof teams)[number])}
	<div class="flex gap-3 text-xs text-muted">
		<span title="Baron">Baron <b class="text-ink">{t.baron}</b></span>
		<span title="Dragon">Dragon <b class="text-ink">{t.dragon}</b></span>
		<span title="Rift Herald">Herald <b class="text-ink">{t.herald}</b></span>
	</div>
{/snippet}

{#snippet compareBar(label: string, left: number, right: number)}
	{@const total = Math.max(1, left + right)}
	<div class="relative flex h-5 overflow-hidden rounded text-[11px] font-semibold text-white">
		<div class="{barColor(teams[0].isWin)} flex items-center pl-2" style:width="{(left / total) * 100}%">
			{left.toLocaleString('en-US')}
		</div>
		<div class="{barColor(teams[1].isWin)} flex flex-1 items-center justify-end pr-2">
			{right.toLocaleString('en-US')}
		</div>
		<span class="absolute inset-0 grid place-items-center">{label}</span>
	</div>
{/snippet}

<div class="overflow-hidden rounded-lg border border-line bg-surface">
	{@render teamTable(teams[0])}

	<!-- so sánh 2 team: mục tiêu + tổng kill + tổng vàng -->
	<div class="flex items-center gap-4 border-y border-line bg-base/60 px-3 py-2.5">
		{@render objectives(teams[0])}
		<div class="flex flex-1 flex-col gap-1">
			{@render compareBar('Total kills', teams[0].kills, teams[1].kills)}
			{@render compareBar('Total gold', teams[0].gold, teams[1].gold)}
		</div>
		{@render objectives(teams[1])}
	</div>

	{@render teamTable(teams[1])}
</div>
