<script lang="ts">
	import brandIcon from '$lib/assets/brand-icon.svg';

	// Trang được prerender nên đây là năm lúc build.
	const YEAR = new Date().getFullYear();
</script>

<!-- cùng nền với body, ngăn cách bằng đường kẻ ngang full màn hình;
     nội dung bên dưới vẫn giới hạn 1100px như Header -->
<footer class="mt-20 border-t border-line">
	<div class="mx-auto flex max-w-[1100px] flex-col gap-6 px-5 py-12 text-xs text-muted">
		<p class="text-center leading-relaxed">
			MetaChecker isn’t endorsed by Riot Games and doesn’t reflect the views or opinions of Riot Games or
			anyone officially involved in producing or managing League of Legends. League of Legends and Riot
			Games are trademarks or registered trademarks of Riot Games, Inc. League of Legends © Riot Games,
			Inc.
		</p>

		<div class="flex flex-wrap items-center justify-center gap-3">
			<!-- logo giống Header, thu nhỏ -->
			<a href="/" class="flex items-center gap-2 text-white">
				<img src={brandIcon} alt="" class="size-5" />
				<!-- không dùng text-base: --color-base trong theme biến nó thành class màu nền (chữ đen) -->
				<span class="text-[16px] font-bold tracking-tight">
					Meta<span class="text-accent">Checker</span>
				</span>
			</a>
			<div>
				© {YEAR} MetaChecker. Made by
				<a
					href="https://github.com/kbdev-21"
					target="_blank"
					rel="noopener noreferrer"
					class="font-semibold text-ink hover:underline"
				>
					kbdev-21
				</a>
			</div>
		</div>
	</div>
</footer>
