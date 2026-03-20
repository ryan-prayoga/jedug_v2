<script lang="ts">
	import { DocumentIcon } from '$lib/icons';

	let {
		message = 'Belum ada data',
		icon = '📭',
		ctaHref = '',
		ctaLabel = ''
	}: {
		message?: string;
		icon?: string;
		ctaHref?: string;
		ctaLabel?: string;
	} = $props();

	const useEmoji = $derived(icon && icon !== '📭');
</script>

<div class="state-panel">
	<div class="mx-auto flex size-14 items-center justify-center rounded-[20px] bg-brand-50 text-brand-600">
		{#if useEmoji}
			<span class="text-2xl leading-none">{icon}</span>
		{:else}
			<DocumentIcon class="size-7" />
		{/if}
	</div>
	<p class="mt-4 text-sm font-semibold text-slate-800">{message}</p>
	<p class="mt-1 text-xs leading-5 text-slate-500">Begitu ada data baru, area ini akan terisi otomatis.</p>
	{#if ctaHref && ctaLabel}
		<a href={ctaHref} class="btn-primary mx-auto mt-4">{ctaLabel}</a>
	{/if}
</div>
