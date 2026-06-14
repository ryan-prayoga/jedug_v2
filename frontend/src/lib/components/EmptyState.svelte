<script lang="ts">
	import { DocumentIcon } from '$lib/icons';

	let {
		message = 'Belum ada data',
		ctaHref = '',
		ctaLabel = ''
	}: {
		message?: string;
		icon?: string;
		ctaHref?: string;
		ctaLabel?: string;
	} = $props();
</script>

<div class="state-panel">
	<DocumentIcon class="mx-auto size-7 text-subtle" />
	<p class="mt-4 text-sm font-semibold text-ink">{message}</p>
	<p class="mt-1 text-xs leading-5 text-muted">Begitu ada data baru, area ini akan terisi otomatis.</p>
	{#if ctaHref && ctaLabel}
		<a href={ctaHref} class="btn-primary mx-auto mt-5">{ctaLabel}</a>
	{/if}
</div>
