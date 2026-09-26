<script lang="ts">
	import Asterisk from '@lucide/svelte/icons/asterisk';
	import Dices from '@lucide/svelte/icons/dices';
	import Lock from '@lucide/svelte/icons/lock';
	import { onMount, untrack } from 'svelte';

	import {
		getAnalytics,
		getChampionAnalytics,
		isApiError,
		META_SERVER_GLOBAL,
		type ChampionTier,
		type Meta,
		type Position
	} from '$lib/api';
	import LevelSlider, { sliderColor } from '$lib/components/randomizer/LevelSlider.svelte';
	import SlotReel from '$lib/components/randomizer/SlotReel.svelte';
	import ChampionIcon from '$lib/components/shared/ChampionIcon.svelte';
	import PageMeta from '$lib/components/shared/PageMeta.svelte';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { POSITION_ICONS } from '$lib/utils/positions';
	import {
		LANES,
		randomItem,
		randomRune,
		randomSpells,
		topRuneChoices,
		type RuneChoice,
		type SpellPair
	} from '$lib/utils/randomizer';

	const POSITION_LABEL: Record<Position, string> = {
		TOP: 'Top',
		JGL: 'Jungle',
		MID: 'Mid',
		ADC: 'ADC',
		SPT: 'Support',
		UNK: 'Other'
	};

	// Mức của slider, từ trái (fun) sang phải (mạnh).
	const LEVELS = ['FUN', 'BALANCED', 'META'] as const;
	type Level = (typeof LEVELS)[number];

	const LEVEL_INFO: Record<Level, { label: string; description: string }> = {
		FUN: {
			label: 'Totally random',
			description: 'Any champion in any lane, with random runes and spells. You might get reported. Be careful.'
		},
		BALANCED: {
			label: 'Balanced',
			description:
				'Champions from the tier list with playable runes and spells.'
		},
		META: {
			label: 'Only meta picks',
			description: 'Only S and A tier champions with strongest runes and spells.'
		}
	};

	// Giống tier list: pick rate dưới 0.5% thì không có trong tier list.
	const MIN_PICK_RATE = 0.005;
	const META_TIERS: ChampionTier[] = ['S', 'A'];
	// Số lựa chọn phổ biến nhất được random ở mức BALANCED.
	const BALANCED_TOP = 3;

	// Thời gian quay của từng reel (ms): dừng lần lượt từ trái sang phải như máy slot.
	const SPIN_MS = { position: 1200, champion: 1800, rune: 2400, spells: 3000 };
	// Số ô lướt qua mỗi giây quay, đủ nhiều để thấy rõ đang quay.
	const CELLS_PER_SECOND = 16;
	// Số item của strip lướt idle lúc chưa roll.
	const IDLE_CELLS = 12;

	type RollResult = {
		position: Position;
		championId: number;
		rune: RuneChoice;
		spells: SpellPair;
	};

	let position = $state<Position | 'ALL'>('ALL');
	let levelIndex = $state(1);
	const level = $derived(LEVELS[levelIndex]);

	// Meta GLOBAL, chỉ cần cho mức BALANCED / META.
	let meta = $state<Meta | null>(null);
	let metaLoading = $state(true);
	let metaError = $state<unknown>(null);

	onMount(() => {
		getAnalytics({ server: META_SERVER_GLOBAL })
			.then((m) => (meta = m))
			.catch((e) => (metaError = e))
			.finally(() => (metaLoading = false));
	});

	// Rỗng tới khi lolData tải xong (lúc đó mới có item để các reel lướt idle).
	let positionStrip = $state<Position[]>([]);
	let championStrip = $state<number[]>([]);
	let runeStrip = $state<RuneChoice[]>([]);
	let spellStrip = $state<SpellPair[]>([]);

	// Chưa roll lần nào thì các reel lướt chậm (idle) thay vì đứng yên.
	let hasRolled = $state(false);
	// Reel rune / spell tắt được (bỏ tick): mờ đi, lướt idle, roll không quay reel đó.
	let runeEnabled = $state(true);
	let spellEnabled = $state(true);
	// Reel rune / spell chỉ hết idle khi thực sự quay (reel tắt thì roll vẫn không quay).
	let runeSpun = $state(false);
	let spellSpun = $state(false);
	let rolling = $state(false);
	let rollError = $state<string | null>(null);

	const needsMeta = $derived(level !== 'FUN');
	const canRoll = $derived(!rolling && lolData.isLoaded && (!needsMeta || meta !== null));

	// Item ngẫu nhiên lấp vào strip (idle và các ô lướt qua lúc quay), chỉ để nhìn.
	const randomLane = () => randomItem(LANES);
	const randomChampionId = () => randomItem([...lolData.championById.values()]).id;
	const randomRuneFiller = () => randomRune(lolData.runeById);
	const randomSpellFiller = () => randomSpells(randomItem(LANES), lolData.spellById);

	// untrack: chỉ chạy lại khi lolData tải xong, không phải mỗi lần đổi position.
	$effect(() => {
		if (!lolData.isLoaded) return;
		untrack(resetReels);
	});

	// Đã có kết quả mà đổi config (position, build style, bật / tắt reel) thì kết quả cũ không
	// còn khớp config => cả 4 reel về idle. Config bị khóa lúc đang roll nên không reset giữa chừng.
	$effect(() => {
		void [position, level, runeEnabled, spellEnabled];
		untrack(() => {
			if (hasRolled) resetReels();
		});
	});

	// Chọn 1 lane thì reel position khóa luôn vào lane đó; về ALL khi chưa roll thì lướt idle lại.
	function selectPosition(p: Position | 'ALL') {
		position = p;
		if (p !== 'ALL') {
			positionStrip = [p];
		} else if (!hasRolled && lolData.isLoaded) {
			positionStrip = idleStrip(randomLane);
		}
	}

	// Mọi reel về idle (position đang khóa lane thì hiện lane đó).
	function resetReels() {
		hasRolled = false;
		runeSpun = false;
		spellSpun = false;
		positionStrip = position === 'ALL' ? idleStrip(randomLane) : [position];
		championStrip = idleStrip(randomChampionId);
		runeStrip = idleStrip(randomRuneFiller);
		spellStrip = idleStrip(randomSpellFiller);
	}

	async function roll() {
		if (!canRoll) return;
		rolling = true;
		rollError = null;

		let result: RollResult;
		try {
			result = level === 'FUN' ? rollFun() : await rollFromMeta(level);
		} catch (e) {
			rollError = isApiError(e) ? `Không tải được dữ liệu (HTTP ${e.status}).` : String((e as Error).message ?? e);
			rolling = false;
			return;
		}

		hasRolled = true;
		positionStrip =
			position === 'ALL'
				? spinStrip(positionStrip, result.position, randomLane, SPIN_MS.position)
				: [result.position];
		championStrip = spinStrip(championStrip, result.championId, randomChampionId, SPIN_MS.champion);
		if (runeEnabled) {
			runeSpun = true;
			runeStrip = spinStrip(runeStrip, result.rune, randomRuneFiller, SPIN_MS.rune);
		}
		if (spellEnabled) {
			spellSpun = true;
			spellStrip = spinStrip(spellStrip, result.spells, randomSpellFiller, SPIN_MS.spells);
		}

		// Xong khi reel bật cuối cùng (bên phải nhất) dừng.
		const lastStop = spellEnabled ? SPIN_MS.spells : runeEnabled ? SPIN_MS.rune : SPIN_MS.champion;
		setTimeout(() => (rolling = false), lastStop);
	}

	function rollFun(): RollResult {
		const lane = position === 'ALL' ? randomItem(LANES) : position;
		return {
			position: lane,
			championId: randomItem([...lolData.championById.values()]).id,
			rune: randomRune(lolData.runeById),
			spells: randomSpells(lane, lolData.spellById)
		};
	}

	// Random lane trước (mỗi lane xác suất như nhau), rồi random champion trong lane đó.
	async function rollFromMeta(lvl: Exclude<Level, 'FUN'>): Promise<RollResult> {
		const rows = (meta?.championStats ?? []).filter(
			(r) => r.pickRate >= MIN_PICK_RATE && (lvl !== 'META' || META_TIERS.includes(r.tier))
		);
		const lanes = position === 'ALL' ? LANES.filter((l) => rows.some((r) => r.position === l)) : [position];
		const lane = randomItem(lanes);
		const laneRows = rows.filter((r) => r.position === lane);
		if (!laneRows.length) {
			throw new Error('Không có champion nào phù hợp với lựa chọn này.');
		}
		const row = randomItem(laneRows);

		// Summary của tier list không kèm build => lấy build của champion vừa ra.
		const stats = await getChampionAnalytics(row.championId, { server: META_SERVER_GLOBAL });
		const stat = stats?.find((s) => s.position === lane);
		const top = lvl === 'META' ? 1 : BALANCED_TOP;
		const runes = topRuneChoices(stat?.bestRunes ?? [], top);
		// backend đã sort games giảm dần.
		const combos = (stat?.bestSpellCombos ?? []).slice(0, top);
		const combo = combos.length ? randomItem(combos) : null;

		return {
			position: lane,
			championId: row.championId,
			// Build thiếu dữ liệu (hiếm) thì random như mức FUN để vẫn ra kết quả.
			rune: runes.length ? randomItem(runes) : randomRune(lolData.runeById),
			spells: combo ? [combo.spell1Id, combo.spell2Id] : randomSpells(lane, lolData.spellById)
		};
	}

	// Ô đầu = kết quả cũ (để bắt đầu quay không bị giật), giữa là ô random, cuối là kết quả mới.
	function spinStrip<T>(prev: T[], final: T, filler: () => T, ms: number): T[] {
		const count = Math.round((ms / 1000) * CELLS_PER_SECOND);
		return [prev[prev.length - 1], ...Array.from({ length: count }, filler), final];
	}

	function idleStrip<T>(filler: () => T): T[] {
		return Array.from({ length: IDLE_CELLS }, filler);
	}
</script>

{#snippet positionCell(p: Position)}
	<img src={POSITION_ICONS[p]} alt={p} class="size-12 object-contain" />
	<span class="text-sm font-semibold">{POSITION_LABEL[p]}</span>
{/snippet}

{#snippet championCell(id: number)}
	{@const champ = lolData.championById.get(id)}
	{#if champ}
		<ChampionIcon src={champ.imgUrl} alt={champ.name} class="size-14 rounded-lg" />
		<span class="max-w-full truncate px-2 text-sm font-semibold">{champ.name}</span>
	{/if}
{/snippet}

{#snippet runeCell(r: RuneChoice)}
	{@const primary = lolData.runeById.get(r.primaryStyle)}
	{@const keystone = lolData.runeById.get(r.keyRune)}
	{@const sub = lolData.runeById.get(r.subStyle)}
	{#if primary && keystone && sub}
		<div class="flex items-center gap-2">
			<img src={primary.imgUrl} alt={primary.name} class="size-6" />
			<img src={keystone.imgUrl} alt={keystone.name} class="size-14 rounded-full bg-black/40" />
			<img src={sub.imgUrl} alt={sub.name} class="size-6" />
		</div>
		<span class="max-w-full truncate px-2 text-sm font-semibold">{keystone.name}</span>
	{/if}
{/snippet}

{#snippet spellCell(pair: SpellPair)}
	{@const spells = pair.map((id) => lolData.spellById.get(id))}
	{#if spells[0] && spells[1]}
		<div class="flex gap-2">
			{#each spells as s (s!.id)}
				<img src={s!.imgUrl} alt={s!.name} class="size-12 rounded-lg" />
			{/each}
		</div>
		<span class="max-w-full truncate px-2 text-sm font-semibold">{spells[0].name} + {spells[1].name}</span>
	{/if}
{/snippet}

<PageMeta title="LoL Randomizer" description="Randomize a League of Legends champion, build and runes for your next game." />

{#snippet reelTitle(label: string, locked = false)}
	<h2 class="flex items-center justify-center gap-1.5 border-b border-line px-4 py-2.5 text-sm font-semibold">
		{label}
		{#if locked}
			<Lock class="size-3.5 text-muted" aria-label="Locked" />
		{/if}
	</h2>
{/snippet}

<!-- checkbox bật / tắt reel, nằm góc phải trên của header reel -->
{#snippet reelToggle(checked: boolean, onchange: (checked: boolean) => void, label: string)}
	<input
		type="checkbox"
		{checked}
		onchange={(e) => onchange(e.currentTarget.checked)}
		disabled={rolling}
		aria-label="Roll {label}"
		title="Roll {label}"
		class="absolute right-3 top-1/2 size-4 -translate-y-1/2 cursor-pointer accent-accent disabled:cursor-not-allowed"
	/>
{/snippet}

<div class="mx-auto max-w-[1100px] px-5 py-6">
	<div class="mb-4">
		<h1 class="text-xl font-semibold">LoL Randomizer</h1>
		<p class="mt-1 text-sm text-muted">Don't know what champion to pick? Random position, champion, runes and summoner spells for your next game.</p>
	</div>

	<!-- config -->
	<section class="flex flex-wrap items-center gap-x-10 gap-y-4 rounded-lg border border-line bg-surface p-4">
		<div>
			<div class="mb-2 text-xs font-semibold uppercase tracking-wide text-muted">Position</div>
			<!-- giống nút lọc position của tier list -->
			<div class="flex h-10 w-fit items-center divide-x divide-line rounded-lg bg-elevated ring-1 ring-line">
				{#each LANES as p (p)}
					<button
						type="button"
						onclick={() => selectPosition(p)}
						disabled={rolling}
						aria-pressed={position === p}
						title={POSITION_LABEL[p]}
						class="h-full w-12 place-items-center outline-none transition-colors focus-visible:ring-2 focus-visible:ring-accent/60 disabled:cursor-not-allowed {position ===
						p
							? 'bg-white/12'
							: 'opacity-70 hover:bg-white/5 hover:opacity-100'}"
					>
						<img src={POSITION_ICONS[p]} alt={p} class="mx-auto size-4 object-contain" />
					</button>
				{/each}
				<button
					type="button"
					onclick={() => selectPosition('ALL')}
					disabled={rolling}
					aria-pressed={position === 'ALL'}
					title="All"
					class="grid h-full w-12 place-items-center outline-none transition-colors focus-visible:ring-2 focus-visible:ring-accent/60 disabled:cursor-not-allowed {position ===
					'ALL'
						? 'bg-white/12 text-ink'
						: 'text-muted hover:bg-white/5'}"
				>
					<Asterisk class="size-6" />
				</button>
			</div>
		</div>

		<div class="w-72">
			<div
				class="mb-2 text-center text-xs font-semibold uppercase tracking-wide"
				style:color={sliderColor(levelIndex / (LEVELS.length - 1))}
			>
				{LEVEL_INFO[level].label}
			</div>
			<LevelSlider
				bind:value={levelIndex}
				count={LEVELS.length}
				label="Build style"
				valueText={LEVEL_INFO[level].label}
				disabled={rolling}
			/>
			<div class="mt-1 flex justify-between text-xs text-muted">
				<span>More chaos</span>
				<span>More powerful</span>
			</div>
		</div>

		<p class="min-w-60 flex-1 text-sm text-muted">{LEVEL_INFO[level].description}</p>
	</section>

	<!-- reels -->
	<section class="mt-3 grid grid-cols-2 gap-3 md:grid-cols-4">
		<div class="rounded-lg border border-line bg-surface">
			{@render reelTitle('Position', position !== 'ALL')}
			<SlotReel
				strip={positionStrip}
				idle={!hasRolled && position === 'ALL'}
				duration={SPIN_MS.position}
				item={positionCell}
			/>
		</div>
		<div class="rounded-lg border border-line bg-surface">
			{@render reelTitle('Champion')}
			<SlotReel strip={championStrip} idle={!hasRolled} duration={SPIN_MS.champion} item={championCell} />
		</div>
		<div class="rounded-lg border border-line bg-surface">
			<div class="relative">
				{@render reelTitle('Rune')}
				{@render reelToggle(runeEnabled, (v) => (runeEnabled = v), 'rune')}
			</div>
			<SlotReel
				strip={runeStrip}
				idle={!runeSpun}
				disabled={!runeEnabled}
				duration={SPIN_MS.rune}
				item={runeCell}
			/>
		</div>
		<div class="rounded-lg border border-line bg-surface">
			<div class="relative">
				{@render reelTitle('Spells')}
				{@render reelToggle(spellEnabled, (v) => (spellEnabled = v), 'spells')}
			</div>
			<SlotReel
				strip={spellStrip}
				idle={!spellSpun}
				disabled={!spellEnabled}
				duration={SPIN_MS.spells}
				item={spellCell}
			/>
		</div>
	</section>

	<div class="mt-5 flex flex-col items-center gap-3">
		<button
			type="button"
			onclick={roll}
			disabled={!canRoll}
			class="flex h-12 items-center gap-2 rounded-lg bg-accent px-10 text-base font-bold text-black outline-none transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-accent/60 focus-visible:ring-offset-2 focus-visible:ring-offset-base disabled:cursor-not-allowed disabled:opacity-50"
		>
			<Dices class="size-5" />
			{rolling ? 'Rolling...' : 'Roll'}
		</button>

		{#if rollError}
			<p class="text-sm text-red-400">{rollError}</p>
		{:else if needsMeta && !metaLoading && !meta}
			<p class="text-sm text-muted">
				{metaError
					? `Không tải được meta${isApiError(metaError) ? ` (HTTP ${metaError.status})` : ''}.`
					: 'Chưa tổng hợp meta cho patch hiện tại.'}
				Chỉ dùng được mức Fun.
			</p>
		{/if}
	</div>
</div>
