<script lang="ts">
	import ChevronRight from '@lucide/svelte/icons/chevron-right';

	import type { Champion, SkillOrderStat } from '$lib/api';
	import { pct2, winRateColor } from '$lib/utils/format';

	type Props = {
		// Thứ tự lên skill (13 cấp đầu) của (champion, position), backend đã sort games giảm dần.
		orders: SkillOrderStat[];
		champion: Champion | undefined;
		loading: boolean;
	};

	let { orders, champion, loading }: Props = $props();

	// skillSlot: 1 = Q, 2 = W, 3 = E, 4 = R.
	const SKILL_KEYS: Record<number, string> = { 1: 'Q', 2: 'W', 3: 'E', 4: 'R' };
	const SKILL_COLORS: Record<number, string> = {
		1: 'text-sky-400',
		2: 'text-emerald-400',
		3: 'text-orange-400',
		// Nền vàng accent của app; chữ tối vì chữ trắng trên nền vàng khó đọc.
		4: 'bg-accent text-(--color-base)'
	};

	// Q/W/E max ở 5 điểm.
	const MAX_SKILL_POINTS = 5;

	// Pick rate tính trên các trận có đủ 13 cấp (chỉ những trận này mới có trong orders).
	const total = $derived(orders.reduce((sum, o) => sum + o.games, 0));
	const top = $derived(orders[0]);
	const maxOrder = $derived(top ? maxOrderOf(top.skills) : []);

	// Thứ tự max Q/W/E: skill đủ 5 điểm trước đứng trước; chưa đủ trong 13 cấp thì xếp theo số điểm.
	function maxOrderOf(skills: number[]): number[] {
		const points: Record<number, number> = { 1: 0, 2: 0, 3: 0 };
		const maxedAt: Record<number, number> = {};
		skills.forEach((slot, level) => {
			if (slot === 4) return;
			points[slot]++;
			if (points[slot] === MAX_SKILL_POINTS) maxedAt[slot] = level;
		});
		return [1, 2, 3].sort(
			(a, b) => (maxedAt[a] ?? Infinity) - (maxedAt[b] ?? Infinity) || points[b] - points[a]
		);
	}
</script>

<section class="rounded-lg border border-line bg-surface">
	<h2 class="border-b border-line px-4 py-2.5 text-sm font-semibold">Skill Order</h2>

	{#if loading}
		<div class="animate-pulse p-4">
			<div class="h-[68px] rounded-lg bg-elevated"></div>
		</div>
	{:else if !top}
		<p class="px-4 py-6 text-center text-sm text-muted">No data</p>
	{:else}
		<div class="flex items-center gap-4 px-4 py-3">
			<div class="flex flex-col gap-2">
				<!-- thứ tự max: icon skill kèm phím -->
				<div class="flex items-center gap-1.5">
					{#each maxOrder as slot, i (slot)}
						{@const skill = champion?.skills[slot - 1]}
						{#if i > 0}
							<ChevronRight class="size-4 text-muted" />
						{/if}
						<div class="relative" title={skill?.name}>
							{#if skill}
								<img src={skill.imgUrl} alt={skill.name} class="size-8 rounded" loading="lazy" />
							{:else}
								<div class="size-8 rounded bg-elevated"></div>
							{/if}
							<span
								class="absolute -bottom-0.5 -right-0.5 rounded-tl bg-base px-0.5 text-[10px] font-bold leading-none {SKILL_COLORS[
									slot
								]}"
							>
								{SKILL_KEYS[slot]}
							</span>
						</div>
					{/each}
				</div>

				<!-- skill lên ở từng cấp 1..13 -->
				<div class="flex gap-1">
					{#each top.skills as slot, level (level)}
						<span
							class="grid size-6 place-items-center rounded text-[11px] font-bold {slot === 4
								? SKILL_COLORS[slot]
								: `bg-elevated ${SKILL_COLORS[slot]}`}"
							title="Level {level + 1}"
						>
							{SKILL_KEYS[slot]}
						</span>
					{/each}
				</div>
			</div>

			<div class="flex-1 text-center text-xs text-muted">
				<div class="text-xs font-bold text-ink">{pct2(top.games / total)}</div>
				<div>{top.games.toLocaleString('en-US')} Games</div>
			</div>
			<div class="text-xs font-bold {winRateColor(top.wins / top.games)}">{pct2(top.wins / top.games)}</div>
		</div>
	{/if}
</section>
