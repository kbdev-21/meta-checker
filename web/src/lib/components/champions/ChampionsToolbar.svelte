<script lang="ts">
	import Asterisk from '@lucide/svelte/icons/asterisk';
	import Search from '@lucide/svelte/icons/search';

	import type { MetaServer, Position } from '$lib/api';
	import MetaServerSelect from '$lib/components/shared/MetaServerSelect.svelte';
	import { POSITION_ICONS } from '$lib/utils/positions';

	// Meta solo queue không có UNK nên chỉ lọc theo 5 lane.
	const FILTER_POSITIONS: Position[] = ['TOP', 'JGL', 'MID', 'ADC', 'SPT'];

	type Props = {
		keyword: string;
		position: Position | 'ALL';
		server: MetaServer;
	};

	let { keyword = $bindable(), position = $bindable(), server = $bindable() }: Props = $props();
</script>

<div class="mb-3 flex flex-wrap items-center gap-2">
	<MetaServerSelect bind:server />

	<!-- lọc theo position, chọn 1 -->
	<div class="flex h-10 items-center divide-x divide-line rounded-lg bg-elevated ring-1 ring-line">
		{#each FILTER_POSITIONS as p (p)}
			<button
				type="button"
				onclick={() => (position = p)}
				aria-pressed={position === p}
				title={p}
				class="h-full w-12 place-items-center outline-none transition-colors focus-visible:ring-2 focus-visible:ring-accent/60 {position ===
				p
					? 'bg-white/12'
					: 'opacity-70 hover:bg-white/5 hover:opacity-100'}"
			>
				<img src={POSITION_ICONS[p]} alt={p} class="size-4 object-contain" />
			</button>
		{/each}
		<button
			type="button"
			onclick={() => (position = 'ALL')}
			aria-pressed={position === 'ALL'}
			title="Tất cả"
			class="h-full w-12 place-items-center outline-none transition-colors focus-visible:ring-2 focus-visible:ring-accent/60 {position ===
			'ALL'
				? 'bg-white/12 text-ink'
				: 'text-muted hover:bg-white/5'}"
		>
			<Asterisk class="size-6" />
		</button>
	</div>

	<!-- tìm champion -->
	<div
		class="flex h-10 w-80 items-center rounded-lg bg-elevated px-3.5 ring-1 ring-line focus-within:ring-accent/50"
	>
		<Search class="mr-2 size-4 shrink-0 text-muted" />
		<input
			type="search"
			bind:value={keyword}
			placeholder="Search Champions"
			aria-label="Tìm champion"
			class="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-muted"
		/>
	</div>
</div>
