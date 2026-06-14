<script lang="ts">
	let { message = 'Memuat...' }: { message?: string } = $props();
</script>

<div class="state-panel">
	<div class="mx-auto size-11 animate-spin rounded-full border-[3px] border-hairline border-t-brand"></div>
	<p class="mt-4 text-sm font-semibold text-ink">{message}</p>
	<p class="mt-1 text-xs leading-5 text-muted">Tunggu sebentar, data sedang disiapkan.</p>
</div>
