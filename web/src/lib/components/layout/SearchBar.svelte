<script lang="ts">
	import Triangle from '@lucide/svelte/icons/triangle';
	import Search from '@lucide/svelte/icons/search';

	import { goto } from '$app/navigation';
	import { searchPlayers, SERVERS, type Player, type Server } from '$lib/api';
	import ChampionIcon from '$lib/components/shared/ChampionIcon.svelte';
	import ServerBadge from '$lib/components/shared/ServerBadge.svelte';
	import { followedPlayers, type FollowedPlayer } from '$lib/stores/followed-players.svelte';
	import { lolData } from '$lib/stores/lol-data.svelte';
	import { championHref } from '$lib/utils/champion';
	import { ddragonVersionOf, profileIconUrl } from '$lib/utils/ddragon';
	import { dismissOn } from '$lib/utils/dismiss';
	import { DEFAULT_TAGS, playerHref } from '$lib/utils/riot-id';

	// Số kết quả tối đa mỗi nhóm (champion, player).
	const RESULT_LIMIT = 5;
	// Chờ ngừng gõ rồi mới gọi API tìm player.
	const SEARCH_DEBOUNCE_MS = 200;

	// Server đang chọn: findPlayerByInfo(server, name, tag) bắt buộc có server.
	let server = $state<Server>('VN');
	let keyword = $state('');
	let open = $state(false);
	let picker: HTMLDivElement;
	let root: HTMLDivElement;
	let input: HTMLInputElement;

	// Dropdown kết quả: mở khi gõ / focus, đóng khi click ra ngoài, Escape hoặc sau khi điều hướng.
	let resultsOpen = $state(false);
	let players = $state<Player[]>([]);

	$effect(() => (open ? dismissOn(picker, () => (open = false)) : undefined));
	$effect(() => (resultsOpen ? dismissOn(root, () => (resultsOpen = false)) : undefined));

	const query = $derived(keyword.trim());
	// "Faker#KR1" => tìm player theo phần name: search của backend không hiểu '#'.
	const playerQuery = $derived(query.split('#')[0].trim());

	// Lọc ngay trên client. Tên bắt đầu bằng keyword xếp trước ("va" => Varus, Vayne rồi mới tới Nidalee...).
	const champions = $derived.by(() => {
		const kw = query.toLowerCase();
		if (!kw) return [];
		const startsWith = (name: string) => (name.toLowerCase().startsWith(kw) ? 0 : 1);
		return [...lolData.championById.values()]
			.filter((c) => c.name.toLowerCase().includes(kw) || c.slug.toLowerCase().includes(kw))
			.sort((a, b) => startsWith(a.name) - startsWith(b.name) || a.name.localeCompare(b.name))
			.slice(0, RESULT_LIMIT);
	});

	// Gõ liên tục thì response cũ có thể về sau: chỉ nhận request mới nhất.
	let reqId = 0;
	$effect(() => {
		const q = playerQuery;
		const id = ++reqId;
		if (!q) {
			players = [];
			return;
		}
		const timer = setTimeout(() => {
			searchPlayers(q)
				.then((list) => id === reqId && (players = list.slice(0, RESULT_LIMIT)))
				.catch(() => id === reqId && (players = []));
		}, SEARCH_DEBOUNCE_MS);
		return () => clearTimeout(timer);
	});

	// Player đã follow theo name (không phân biệt hoa thường), trùng name thì theo tag.
	const followed = $derived(
		[...followedPlayers.list].sort(
			(a, b) =>
				a.name.localeCompare(b.name, undefined, { sensitivity: 'base' }) ||
				a.tag.localeCompare(b.tag, undefined, { sensitivity: 'base' })
		)
	);

	// Đây cũng là thứ tự chọn bằng phím. Chưa gõ gì: các player đã follow.
	// Có keyword: champion trước, player sau.
	const resultHrefs = $derived(
		query
			? [
					...champions.map((c) => championHref(c.slug)),
					...players.map((p) => playerHref(p.server, p.name, p.tag))
				]
			: followed.map((p) => playerHref(p.server, p.name, p.tag))
	);

	// -1 = chưa chọn gì; bấm mũi tên xuống lần đầu mới chọn kết quả đầu tiên.
	// Kết quả đổi (gõ thêm / player về) thì bỏ chọn.
	let activeIndex = $derived.by(() => {
		void resultHrefs;
		return -1;
	});

	const ddVersion = $derived(ddragonVersionOf(lolData.championById.values().next().value?.imgUrl ?? ''));

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			if (!resultHrefs.length) return;
			e.preventDefault();
			resultsOpen = true;
			activeIndex = Math.min(activeIndex + 1, resultHrefs.length - 1);
		} else if (e.key === 'ArrowUp') {
			if (activeIndex < 0) return;
			e.preventDefault();
			activeIndex--;
		} else if (e.key === 'Enter') {
			e.preventDefault();
			submit();
		}
	}

	// Enter / nút kính lúp: đang chọn kết quả thì tới kết quả đó, không thì tới trang player.
	function submit() {
		const href = activeIndex >= 0 ? resultHrefs[activeIndex] : playerHrefOf(query);
		if (!href) return;
		reset();
		goto(href);
	}

	// "name#tag" (# nằm giữa) => name-tag ở server đang chọn. Không có #, hoặc # ở đầu / cuối
	// => bỏ # và dùng tag mặc định của server đang chọn (VN => vn2).
	function playerHrefOf(q: string): string | null {
		const i = q.indexOf('#');
		if (i > 0 && i < q.length - 1) {
			return playerHref(server, q.slice(0, i).trim(), q.slice(i + 1).trim());
		}
		const name = q.replace(/^#+|#+$/g, '').trim();
		return name ? playerHref(server, name, DEFAULT_TAGS[server]) : null;
	}

	// Sau khi điều hướng: xoá keyword, đóng dropdown, bỏ focus.
	function reset() {
		keyword = '';
		resultsOpen = false;
		input.blur();
	}
</script>

<!-- 1 dòng player (kết quả search / đã follow); index = vị trí trong resultHrefs -->
{#snippet playerRow(p: FollowedPlayer, index: number)}
	<a
		href={resultHrefs[index]}
		onclick={reset}
		class="flex items-center gap-3 px-3 py-2 text-sm transition-colors {index === activeIndex
			? 'bg-white/8'
			: 'hover:bg-white/5'}"
	>
		{#if ddVersion && p.profileIconId !== null}
			<img src={profileIconUrl(ddVersion, p.profileIconId)} alt="" class="size-8 shrink-0 rounded-full" />
		{:else}
			<div class="size-8 shrink-0 rounded-full bg-elevated"></div>
		{/if}
		<ServerBadge server={p.server} />
		<span class="min-w-0 truncate font-semibold">
			{p.name} <span class="font-normal text-muted">#{p.tag}</span>
		</span>
	</a>
{/snippet}

<div class="relative" bind:this={root}>
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
			bind:this={input}
			type="search"
			bind:value={keyword}
			oninput={() => (resultsOpen = true)}
			onfocus={() => (resultsOpen = true)}
			onkeydown={onKeydown}
			placeholder="Search for champions, players"
			aria-label="Tìm champion hoặc player"
			class="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-muted"
		/>

		<button
			type="button"
			onclick={submit}
			aria-label="Tìm kiếm"
			class="grid size-8 shrink-0 place-items-center rounded-full text-muted outline-none transition-colors hover:bg-white/5 hover:text-ink focus-visible:ring-2 focus-visible:ring-accent/60"
		>
			<Search class="size-[18px]" />
		</button>
	</div>

	<!-- chưa gõ gì: player đã follow. Có keyword: tối đa 5 champion ở trên, 5 player ở dưới -->
	{#if resultsOpen && (query ? champions.length || players.length : followed.length)}
		<div class="absolute inset-x-0 top-full z-20 mt-2 flex flex-col gap-2">
			{#if !query}
				<div class="overflow-hidden rounded-lg border border-line bg-surface shadow-xl shadow-black/50">
					<div class="border-b border-line px-3 py-1.5 text-xs font-semibold text-muted">Followed</div>
					{#each followed as p, i (p.id)}
						{@render playerRow(p, i)}
					{/each}
				</div>
			{/if}

			{#if query && champions.length}
				<div class="overflow-hidden rounded-lg border border-line bg-surface shadow-xl shadow-black/50">
					<div class="border-b border-line px-3 py-1.5 text-xs font-semibold text-muted">Champions</div>
					{#each champions as c, i (c.id)}
						<a
							href={resultHrefs[i]}
							onclick={reset}
							class="flex items-center gap-3 px-3 py-2 text-sm font-semibold transition-colors {i === activeIndex
								? 'bg-white/8'
								: 'hover:bg-white/5'}"
						>
							<ChampionIcon src={c.imgUrl} alt={c.name} class="size-8 rounded" />
							{c.name}
						</a>
					{/each}
				</div>
			{/if}

			{#if query && players.length}
				<div class="overflow-hidden rounded-lg border border-line bg-surface shadow-xl shadow-black/50">
					<div class="border-b border-line px-3 py-1.5 text-xs font-semibold text-muted">Players</div>
					{#each players as p, i (p.id)}
						{@render playerRow(p, champions.length + i)}
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>
