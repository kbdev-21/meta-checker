<script lang="ts">
	import Triangle from '@lucide/svelte/icons/triangle';
	import Search from '@lucide/svelte/icons/search';
	import { SERVERS, type Server } from '$lib/api';
	import { dismissOn } from '$lib/dismiss';

	// Server đang chọn: findPlayerByInfo(server, name, tag) bắt buộc có server.
	let server = $state<Server>('VN');
	let keyword = $state('');
	let open = $state(false);
	let picker: HTMLDivElement;

	$effect(() => (open ? dismissOn(picker, () => (open = false)) : undefined));
</script>

<div
	class="flex h-9 w-full items-center rounded-lg bg-elevated pl-3.5 pr-1.5 ring-1 ring-line transition-shadow focus-within:ring-2 focus-within:ring-accent/50"
>
	<!-- chọn server -->
	<div class="relative shrink-0" bind:this={picker}>
		<button
			type="button"
			onclick={() => (open = !open)}
			aria-haspopup="menu"
			aria-expanded={open}
			aria-label="Chọn server"
			class="flex items-center gap-1 rounded py-1 text-sm font-bold text-accent outline-none focus-visible:ring-2 focus-visible:ring-accent/60"
		>
			{server}
			<!-- lucide không có caret-down; triangle là path đặc nên bỏ stroke + fill là ra
			     caret. Icon trỏ lên sẵn nên đóng = xoay 180, mở = trả về 0. -->
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
				class="absolute left-0 top-full z-10 mt-3 grid w-48 grid-cols-3 gap-0.5 rounded-lg border border-line bg-elevated p-1.5 shadow-xl shadow-black/50"
			>
				{#each SERVERS as s (s)}
					<button
						type="button"
						role="menuitem"
						onclick={() => {
							server = s;
							open = false;
						}}
						class="rounded px-1 py-1.5 text-center text-xs outline-none transition-colors hover:bg-white/5 focus-visible:bg-white/5 {s ===
						server
							? 'font-bold text-accent'
							: 'text-muted'}"
					>
						{s}
					</button>
				{/each}
			</div>
		{/if}
	</div>

	<div class="mx-3 h-5 w-px shrink-0 bg-line"></div>

	<input
		type="search"
		bind:value={keyword}
		placeholder="Search for champions, players"
		aria-label="Tìm champion hoặc player"
		class="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-muted"
	/>

	<button
		type="button"
		aria-label="Tìm kiếm"
		class="grid size-8 shrink-0 place-items-center rounded-full text-muted outline-none transition-colors hover:bg-white/5 hover:text-ink focus-visible:ring-2 focus-visible:ring-accent/60"
	>
		<Search class="size-[18px]" />
	</button>
</div>
