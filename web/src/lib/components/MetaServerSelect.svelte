<script lang="ts">
	import Globe from '@lucide/svelte/icons/globe';
	import Triangle from '@lucide/svelte/icons/triangle';

	import { META_SERVERS, type MetaServer } from '$lib/api';
	import { dismissOn } from '$lib/dismiss';

	type Props = {
		server: MetaServer;
	};

	let { server = $bindable() }: Props = $props();

	let open = $state(false);
	let picker: HTMLDivElement;

	$effect(() => (open ? dismissOn(picker, () => (open = false)) : undefined));

	// GLOBAL là "gộp mọi server" nên hiển thị là World cho dễ hiểu.
	const serverLabel = (s: MetaServer) => (s === 'GLOBAL' ? 'Global' : s);
</script>

<!-- chọn server có meta -->
<div class="relative" bind:this={picker}>
	<button
		type="button"
		onclick={() => (open = !open)}
		aria-haspopup="menu"
		aria-expanded={open}
		class="flex h-10 items-center gap-2 rounded-lg bg-elevated px-3.5 text-sm text-ink outline-none ring-1 ring-line transition-colors hover:bg-white/5 focus-visible:ring-2 focus-visible:ring-accent/60"
	>
		<Globe class="size-4 text-muted" />
		{serverLabel(server)}
		<Triangle
			class="size-2 fill-current text-muted transition-transform duration-150 {open
				? 'rotate-0'
				: 'rotate-180'}"
			strokeWidth={0}
		/>
	</button>

	{#if open}
		<div
			role="menu"
			class="absolute left-0 top-full z-10 mt-1.5 w-36 rounded-lg border border-line bg-elevated p-1 shadow-xl shadow-black/50"
		>
			{#each META_SERVERS as s (s)}
				<button
					type="button"
					role="menuitem"
					onclick={() => {
						server = s;
						open = false;
					}}
					class="w-full rounded px-2.5 py-1.5 text-left text-sm outline-none transition-colors hover:bg-white/5 focus-visible:bg-white/5 {s ===
					server
						? 'font-semibold text-accent'
						: 'text-muted'}"
				>
					{serverLabel(s)}
				</button>
			{/each}
		</div>
	{/if}
</div>
