<script lang="ts">
	import { DangerIcon, RefreshIcon } from '$lib/icons';

	let { message = 'Terjadi kesalahan', onretry }: { message?: string; onretry?: () => void } = $props();
</script>

<div class="state-panel border-rose-200 bg-rose-50/80">
	<div class="mx-auto flex size-12 items-center justify-center rounded-2xl bg-white text-rose-600 shadow-[0_12px_28px_rgba(15,23,42,0.06)]">
		<DangerIcon class="size-6" />
	</div>
	<p class="mt-4 text-sm font-semibold text-rose-700">{message}</p>
	<p class="mt-1 text-xs leading-5 text-rose-600/80">Coba ulangi beberapa saat lagi atau periksa koneksi internet.</p>
	{#if onretry}
		<button class="btn-danger mx-auto mt-4" onclick={onretry}>
			<RefreshIcon class="size-[18px]" />
			Coba Lagi
		</button>
	{/if}
</div>
