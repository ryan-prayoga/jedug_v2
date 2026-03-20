<script lang="ts">
	let { message = 'Memuat...' }: { message?: string } = $props();
</script>

<div class="state-panel">
	<div class="mx-auto size-11 animate-spin rounded-full border-[3px] border-slate-200 border-t-brand-500"></div>
	<p class="mt-4 text-sm font-semibold text-slate-700">{message}</p>
	<p class="mt-1 text-xs leading-5 text-slate-500">Tunggu sebentar, data sedang disiapkan.</p>
</div>
