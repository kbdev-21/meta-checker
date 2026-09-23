<script lang="ts">
	import type { ChampionTier } from '$lib/api';

	type Props = {
		tier: ChampionTier;
		// Để nơi dùng tự căn vị trí (vd mx-auto trong ô table).
		class?: string;
	};

	let { tier, class: className = '' }: Props = $props();

	const TIER_BG: Record<ChampionTier, string> = {
		S: 'bg-amber-300',
		A: 'bg-sky-400',
		B: 'bg-emerald-400',
		C: 'bg-gray-400',
		D: 'bg-taupe-500'
	};
</script>

<span
	class="grid size-6 place-items-center rounded text-base font-extrabold text-black {TIER_BG[
		tier
	]} {className}"
>
	{tier}
</span>
