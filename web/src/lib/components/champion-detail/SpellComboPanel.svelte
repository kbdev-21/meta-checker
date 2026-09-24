<script lang="ts">
	import type { SpellComboStat } from '$lib/api';
	import Tooltip from '$lib/components/shared/Tooltip.svelte';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { pct2, winRateColor } from '$lib/utils/format';

	type Props = {
		// Mọi cặp summoner spell của (champion, position), backend đã sort games giảm dần.
		combos: SpellComboStat[];
		loading: boolean;
	};

	let { combos, loading }: Props = $props();

	// Số cặp spell hiển thị (giống op.gg).
	const COMBO_COUNT = 2;

	const total = $derived(combos.reduce((sum, c) => sum + c.games, 0));
	const shown = $derived(combos.slice(0, COMBO_COUNT));
</script>

<section class="rounded-lg border border-line bg-surface">
	<h2 class="border-b border-line px-4 py-2.5 text-sm font-semibold">Spells</h2>

	{#if loading}
		<div class="grid animate-pulse grid-cols-2 gap-4 p-4">
			{#each { length: COMBO_COUNT }, i (i)}
				<div class="h-10 rounded-lg bg-elevated"></div>
			{/each}
		</div>
	{:else if !shown.length}
		<p class="px-4 py-6 text-center text-sm text-muted">No data</p>
	{:else}
		<div class="grid grid-cols-2 divide-x divide-line">
			{#each shown as c (`${c.spell1Id}-${c.spell2Id}`)}
				<div class="flex items-center gap-4 px-4 py-3">
					<div class="flex gap-1">
						{#each [c.spell1Id, c.spell2Id] as id (id)}
							{@const spell = lolData.spellById.get(id)}
							{#if spell}
								<Tooltip content={spell.name}>
									<img src={spell.imgUrl} alt={spell.name} class="size-8 rounded" loading="lazy" />
								</Tooltip>
							{:else}
								<div class="size-8 rounded bg-elevated"></div>
							{/if}
						{/each}
					</div>
					<div class="flex-1 text-center text-xs text-muted">
						<div class="text-xs font-bold text-ink">{pct2(c.games / total)}</div>
						<div>{c.games.toLocaleString('en-US')} Games</div>
					</div>
					<div class="text-xs font-bold {winRateColor(c.wins / c.games)}">{pct2(c.wins / c.games)}</div>
				</div>
			{/each}
		</div>
	{/if}
</section>
