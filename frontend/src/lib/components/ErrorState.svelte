<script lang="ts">
	import { DangerIcon, RefreshIcon } from '$lib/icons';

	let { message = 'Terjadi kesalahan', onretry }: { message?: string; onretry?: () => void } = $props();
</script>

<div class="state-panel border-brand/30">
	<DangerIcon class="mx-auto size-7 text-brand" />
	<p class="mt-4 text-sm font-semibold text-ink">{message}</p>
	<p class="mt-1 text-xs leading-5 text-muted">Coba ulangi beberapa saat lagi atau periksa koneksi internet.</p>
	{#if onretry}
		<button class="btn-danger mx-auto mt-5" onclick={onretry}>
			<RefreshIcon class="size-[18px]" />
			Coba Lagi
		</button>
	{/if}
</div>
