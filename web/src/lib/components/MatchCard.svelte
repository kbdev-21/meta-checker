<script lang="ts">
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import type { GameMode, Match, Server } from '$lib/api';
	import ChampionIcon from '$lib/components/ChampionIcon.svelte';
	import MatchDetail from '$lib/components/MatchDetail.svelte';
	import PerfScoreBadge from '$lib/components/PerfScoreBadge.svelte';
	import PerfScoreLabel from '$lib/components/PerfScoreLabel.svelte';
	import { formatDuration, timeAgo } from '$lib/format';
	import { POSITION_ICONS } from '$lib/positions';
	import { playerHref } from '$lib/riot-id';
	import { lolData } from '$lib/stores/lol-data.svelte';

	type Props = {
		match: Match;
		// Player đang xem trang; dùng để lấy dòng participant của họ.
		playerId: string;
		server: Server;
	};

	let { match, playerId, server }: Props = $props();

	// Đang mở MatchDetail bên dưới card hay không.
	let expanded = $state(false);

	// Click bất kỳ đâu trên card để mở/đóng detail, trừ link (tên player) và nút mũi tên
	// (nút tự toggle, không bỏ qua thì bị toggle 2 lần).
	function onCardClick(e: MouseEvent) {
		if ((e.target as HTMLElement).closest('a, button')) return;
		expanded = !expanded;
	}

	const MODE_LABEL: Record<GameMode, string> = {
		SOLO: 'Ranked Solo',
		FLEX: 'Ranked Flex',
		ARAM: 'ARAM',
		NORMAL: 'Normal'
	};

	// Thắng sky-400, thua red-500. Nền = 1 màu đặc: màu kết quả kéo gần về surface (tối).
	// bar: thanh màu mép trái (không dùng border-l vì border ăn theo rounded nên bị cong ở 2 đầu).
	const RESULT_TONE = {
		WIN: {
			card: 'bg-[color-mix(in_oklab,var(--color-surface)_92%,var(--color-sky-400))]',
			bar: 'bg-sky-400',
			text: 'text-sky-400'
		},
		LOSS: {
			card: 'bg-[color-mix(in_oklab,var(--color-surface)_92%,var(--color-red-500))]',
			bar: 'bg-red-500',
			text: 'text-red-500'
		},
		REMAKE: { card: 'bg-surface', bar: 'bg-muted', text: 'text-muted' }
	} as const;

	// Slot cuối (index 6) là trinket.
	const ITEM_SLOTS = 6;

	// Số người chơi tối đa hiển thị mỗi cột team trong card.
	const TEAM_DISPLAY_COUNT = 5;

	const me = $derived(match.participants.find((p) => p.playerId === playerId));
	const result = $derived<keyof typeof RESULT_TONE>(
		match.isRemake ? 'REMAKE' : me?.isWin ? 'WIN' : 'LOSS'
	);
	const teams = $derived([1, 2].map((t) => match.participants.filter((p) => p.team === t)));

	const champ = $derived(me ? lolData.championById.get(me.championId) : undefined);
	const spells = $derived(me ? [me.spell1Id, me.spell2Id].map((id) => lolData.spellById.get(id)) : []);
	const runes = $derived(me ? [me.keyRune, me.runeSubStyle].map((id) => lolData.runeById.get(id)) : []);
	const items = $derived(me ? me.items.slice(0, ITEM_SLOTS) : []);
	const trinket = $derived(me ? me.items[ITEM_SLOTS] : 0);

	// AI-Score = perfScore. Điểm cao nhất team thắng là MVP, team thua là ACE;
	// còn lại hiện thứ hạng điểm trong cả trận (1st..10th).
	const scoreBadge = $derived.by(() => {
		if (!me || match.isRemake) return null;
		const myTeamBest = Math.max(...match.participants.filter((p) => p.team === me.team).map((p) => p.perfScore));
		if (me.perfScore === myTeamBest) return me.isWin ? 'MVP' : 'ACE';
		const rank = match.participants.filter((p) => p.perfScore > me.perfScore).length + 1;
		return ordinal(rank);
	});

	function ordinal(n: number): string {
		const suffix = n === 1 ? 'st' : n === 2 ? 'nd' : n === 3 ? 'rd' : 'th';
		return `${n}${suffix}`;
	}
</script>

{#snippet itemSlot(itemId: number)}
	{@const item = itemId ? lolData.itemById.get(itemId) : undefined}
	{#if item}
		<img src={item.imgUrl} alt={item.name} title={item.name} class="size-[22px] rounded-md" loading="lazy" />
	{:else}
		<div class="size-[22px] rounded-md bg-black/30"></div>
	{/if}
{/snippet}

{#if me}
	<!-- Bàn phím đã có nút mũi tên (focus + Enter) nên click trên card chỉ là tiện ích cho chuột. -->
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
	<article
		onclick={onCardClick}
		class="relative flex cursor-pointer items-center gap-5 overflow-hidden rounded-lg py-2.5 pl-5 pr-2 {RESULT_TONE[
			result
		].card}"
	>
		<!-- thanh thẳng, bị overflow-hidden của card cắt theo góc bo nên 2 đầu gọn -->
		<span class="absolute inset-y-0 left-0 w-1 {RESULT_TONE[result].bar}" aria-hidden="true"></span>
		<!-- mode, thời gian (trên) và kết quả + thời lượng (dưới), đẩy về 2 mép theo chiều cao card -->
		<div class="flex w-24 shrink-0 flex-col justify-between self-stretch py-0.5 text-sm">
			<div>
				<div class="font-semibold {RESULT_TONE[result].text}">
					{MODE_LABEL[match.mode]}
				</div>
				<div class="text-xs text-muted" title={new Date(match.gameStartAt).toLocaleString()}>
					{timeAgo(match.gameStartAt)}
				</div>
			</div>
			<div class="text-xs">
				<span class={RESULT_TONE[result].text}>
					{result === 'WIN' ? 'Win' : result === 'LOSS' ? 'Lose' : 'Remake'}
				</span>
				<span class="ml-1 text-muted">{formatDuration(match.durationSec)}</span>
			</div>
		</div>

		<!-- champion + spell + rune + KDA, hàng item bên dưới. Width của khối = width dàn item:
		     hàng trên w-0 min-w-full nên không góp vào width khối cha mà chỉ giãn bằng nó. -->
		<div class="flex shrink-0 flex-col gap-2">
			<div class="flex w-0 min-w-full items-center justify-between gap-2">
				<div class="flex gap-1">
					<div class="relative">
						{#if champ}
							<ChampionIcon src={champ.imgUrl} alt={champ.name} title="{champ.name} · Lv {me.champLevel}" class="size-12 rounded-lg" />
						{:else}
							<div class="size-12 rounded-lg bg-elevated"></div>
						{/if}
						{#if POSITION_ICONS[me.position]}
							<img
								src={POSITION_ICONS[me.position]}
								alt={me.position}
								title={me.position}
								class="absolute bottom-0 left-0 size-4 rounded-tr-md rounded-bl-lg bg-base/85 object-contain p-0.5"
							/>
						{/if}
					</div>
					<div class="flex flex-col gap-1">
						{#each spells as spell, i (i)}
							{#if spell}
								<img src={spell.imgUrl} alt={spell.name} title={spell.name} class="size-[22px] rounded-md" loading="lazy" />
							{:else}
								<div class="size-[22px] rounded-md bg-black/30"></div>
							{/if}
						{/each}
					</div>
					<div class="flex flex-col gap-1">
						{#each runes as rune, i (i)}
							{#if rune}
								<img
									src={rune.imgUrl}
									alt={rune.name}
									title={rune.name}
									class="size-[22px] rounded-full bg-black/40 {i === 1 ? 'p-1' : ''}"
									loading="lazy"
								/>
							{:else}
								<div class="size-[22px] rounded-full bg-black/30"></div>
							{/if}
						{/each}
					</div>
				</div>

				<div class="whitespace-nowrap text-center">
					<div class="text-[15px] font-medium">
						{me.kills}<span class="mx-0.5 text-muted">/</span><span class="text-red-400"
							>{me.deaths}</span
						><span class="mx-0.5 text-muted">/</span>{me.assists}
					</div>
					<div class="text-xs text-muted">{me.kda.toFixed(2)} KDA</div>
				</div>
			</div>

			<div class="flex gap-1">
				{#each items as itemId, i (i)}
					{@render itemSlot(itemId)}
				{/each}
				<div class="ml-1">{@render itemSlot(trinket)}</div>
			</div>
		</div>

		<!-- Perf. Score -->
		<div class="flex w-20 shrink-0 flex-col items-center gap-1">
			<div class="text-xs text-muted">Score</div>
			<PerfScoreBadge score={me.perfScore} />
			{#if scoreBadge}
				<PerfScoreLabel label={scoreBadge} />
			{/if}
		</div>

		<!-- 10 người chơi, 2 team -->
		<div class="ml-auto hidden min-w-0 gap-4 md:flex">
			{#each teams as team, t (t)}
				<ul class="flex w-[130px] min-w-0 flex-col gap-0.5">
					<!-- Mode sự kiện có thể > 5 người/team: chỉ hiện 5 để card không cao lên -->
					{#each team.slice(0, TEAM_DISPLAY_COUNT) as p (p.playerId)}
						{@const pc = lolData.championById.get(p.championId)}
						<li class="flex items-center gap-1.5 text-xs">
							{#if pc}
								<ChampionIcon src={pc.imgUrl} alt={pc.name} title={pc.name} class="size-4 rounded" />
							{:else}
								<div class="size-4 rounded bg-elevated"></div>
							{/if}
							<a
								href={playerHref(server, p.name, p.tag)}
								title="{p.name}#{p.tag}"
								class="truncate hover:underline {p.playerId === playerId
									? 'font-normal text-white'
									: 'text-ink/70'}"
							>
								{p.name}
							</a>
						</li>
					{/each}
				</ul>
			{/each}
		</div>

		<!-- mở / đóng MatchDetail bên dưới card -->
		<button
			type="button"
			onclick={() => (expanded = !expanded)}
			aria-expanded={expanded}
			aria-label="Chi tiết trận"
			class="shrink-0 self-end rounded p-1 outline-none transition-colors hover:bg-white/5 focus-visible:ring-2 focus-visible:ring-accent/60 {RESULT_TONE[
				result
			].text}"
		>
			<ChevronDown class="size-5 transition-transform duration-150 {expanded ? 'rotate-180' : ''}" />
		</button>
	</article>

	{#if expanded}
		<div class="mt-1">
			<MatchDetail {match} {playerId} {server} />
		</div>
	{/if}
{/if}
