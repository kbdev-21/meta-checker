<script lang="ts">
	import Info from "@lucide/svelte/icons/info";
	import { page } from "$app/state";
	import brandIcon from "$lib/assets/brand-icon.svg";
	import Tooltip from "$lib/components/shared/Tooltip.svelte";
	import GameSelect from "./GameSelect.svelte";
	import SearchBar from "./SearchBar.svelte";

	// Trang chủ "/" chỉ vào qua logo. Các mục đều có route trong src/routes; prerender
	// crawl link nên route phải tồn tại, không được để href chết.
	// Mục disabled = trang chưa làm: render thành span, không phải link.
	// TODO: prefix /lol đang hardcode vì mới có 1 game; có thêm game thì suy từ GameSelect.
	const NAV = [
		{ label: "Champions", href: "/lol/champions", disabled: false },
		{ label: "What's to pick", href: "/lol/whats-to-pick", disabled: true },
		{ label: "Randomizer", href: "/lol/randomizer", disabled: true },
		{ label: "Game Data", href: "/lol/game-data", disabled: true },
	];
</script>

<header
	class="sticky top-0 z-50 border-b border-line bg-surface"
>
	<div class="mx-auto flex h-15 max-w-[1100px] items-center gap-7 px-5">
		<a
			href="/"
			class="flex shrink-0 items-center gap-2.5 rounded outline-none focus-visible:ring-2 focus-visible:ring-accent/60"
		>
			<!-- logo = icon + wordmark. Wordmark dựng bằng chữ HTML thay vì file svg: <text> trong
			     svg không khai báo font, nhúng qua <img> sẽ ra font serif mặc định. -->
			<img src={brandIcon} alt="" class="size-7" />
			<span class="text-2xl font-bold tracking-tight">
				Meta<span class="text-accent">Checker</span>
			</span>
		</a>

		<GameSelect />

		<!-- search chiếm hết phần còn dư giữa game select và nút info -->
		<div class="min-w-0 flex-1">
			<SearchBar />
		</div>

		<div class="flex shrink-0 items-center gap-4">
			<button
				type="button"
				aria-label="Giới thiệu"
				class="grid size-5.5 shrink-0 place-items-center rounded-full text-muted outline-none transition-colors hover:text-ink focus-visible:ring-2 focus-visible:ring-accent/60"
			>
				<Info class="size-full" strokeWidth={1.8} />
			</button>
		</div>
	</div>

	<nav>
		<div
			class="mx-auto flex h-10 max-w-[1100px] items-stretch gap-8 px-5 text-[15px]"
		>
			{#each NAV as item (item.href)}
				{#if item.disabled}
					<Tooltip content="Coming soon" side="bottom" class="flex">
						<span
							aria-disabled="true"
							class="flex cursor-not-allowed items-center border-b-2 border-transparent font-medium text-muted opacity-50"
						>
							{item.label}
						</span>
					</Tooltip>
				{:else}
					<a
						href={item.href}
						class="flex font-medium items-center border-b-2 outline-none transition-colors focus-visible:ring-2 focus-visible:ring-accent/60 {page
							.url.pathname === item.href
							? 'border-accent text-ink'
							: 'border-transparent text-muted hover:text-ink'}"
					>
						{item.label}
					</a>
				{/if}
			{/each}
		</div>
	</nav>
</header>
