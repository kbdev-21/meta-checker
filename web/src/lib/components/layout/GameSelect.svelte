<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import { dismissOn } from '$lib/utils/dismiss';

	// Game mà app hỗ trợ. Thêm game mới = thêm 1 dòng ở đây.
	const GAMES = [{ id: 'LOL', name: 'League of Legends', short: 'LoL' }] as const;

	let selected = $state<(typeof GAMES)[number]>(GAMES[0]);
	let open = $state(false);
	let root: HTMLDivElement;

	$effect(() => (open ? dismissOn(root, () => (open = false)) : undefined));
</script>

<div class="relative" bind:this={root}>
	<button
		type="button"
		onclick={() => (open = !open)}
		aria-haspopup="menu"
		aria-expanded={open}
		class="flex h-9 items-center gap-1.5 rounded-md bg-elevated px-2.5 py-1.5 text-sm text-muted outline-none ring-1 ring-line transition-colors hover:bg-white/5 hover:text-ink focus-visible:ring-2 focus-visible:ring-accent/60"
	>
		<span class="hidden lg:inline">{selected.name}</span>
		<span class="lg:hidden">{selected.short}</span>
		<ChevronDown
			class="size-3.5 transition-transform duration-150 {open ? 'rotate-180' : ''}"
		/>
	</button>

	{#if open}
		<div
			role="menu"
			class="absolute left-0 top-full z-10 mt-2 w-56 rounded-lg border border-line bg-elevated p-1 shadow-xl shadow-black/50"
		>
			{#each GAMES as game (game.id)}
				<button
					type="button"
					role="menuitem"
					onclick={() => {
						selected = game;
						open = false;
					}}
					class="flex w-full items-center justify-between gap-2 rounded px-2.5 py-1.5 text-left text-sm outline-none hover:bg-white/5 focus-visible:bg-white/5"
				>
					{game.name}
					{#if game.id === selected.id}
						<Check class="size-3.5 shrink-0 text-accent" strokeWidth={2.5} />
					{/if}
				</button>
			{/each}
		</div>
	{/if}
</div>
