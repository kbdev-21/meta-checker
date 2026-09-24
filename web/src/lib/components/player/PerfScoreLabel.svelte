<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';

	type Props = {
		// 'MVP' | 'ACE' | thứ hạng điểm trong trận ('1st'..'10th').
		label: string;
		// Icon vương miện cạnh MVP/ACE; tắt ở chỗ chật như bảng MatchDetail.
		showCrown?: boolean;
	};

	let { label, showCrown = true }: Props = $props();

	const isCrowned = $derived(showCrown && (label === 'MVP' || label === 'ACE'));
</script>

<span
	class="flex items-center gap-1 text-xs font-bold {label === 'MVP'
		? 'text-amber-400'
		: label === 'ACE'
			? 'text-violet-400'
			: 'text-muted'}"
>
	{#if isCrowned}
		<Crown class="size-3.5" />
	{/if}
	{label}
</span>
