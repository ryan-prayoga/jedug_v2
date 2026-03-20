<script lang="ts">
	import { CameraIcon, DangerIcon, DocumentIcon, NotificationIcon } from '$lib/icons';

	const numberFormatter = new Intl.NumberFormat('id-ID');

	let {
		submissionCount,
		photoCount,
		casualtyCount,
		reactionCount
	}: {
		submissionCount: number;
		photoCount: number;
		casualtyCount: number;
		reactionCount: number;
	} = $props();

	const items = $derived.by(() => [
		{
			label: 'Laporan',
			value: submissionCount,
			copy: 'Total laporan terkumpul',
			icon: DocumentIcon,
			alert: false
		},
		{
			label: 'Foto',
			value: photoCount,
			copy: 'Media publik tersedia',
			icon: CameraIcon,
			alert: false
		},
		{
			label: 'Korban',
			value: casualtyCount,
			copy: casualtyCount > 0 ? 'Laporan korban tercatat' : 'Belum ada korban tercatat',
			icon: DangerIcon,
			alert: casualtyCount > 0
		},
		{
			label: 'Reaksi',
			value: reactionCount,
			copy: reactionCount > 0 ? 'Dukungan publik tercatat' : 'Belum ada reaksi publik',
			icon: NotificationIcon,
			alert: false
		}
	]);
</script>

<section class="grid grid-cols-2 gap-3 md:grid-cols-4" aria-label="Statistik ringkas issue">
	{#each items as item}
		{@const ItemIcon = item.icon}
		<article class={`metric-card ${item.alert ? 'border-rose-200 bg-rose-50/70' : ''}`}>
			<div class="flex items-center gap-2 text-slate-500">
				<ItemIcon class="size-[18px]" />
				<span class="metric-label">{item.label}</span>
			</div>
			<strong class:text-rose-700={item.alert} class="metric-value">
				{numberFormatter.format(item.value)}
			</strong>
			<p class="metric-copy">{item.copy}</p>
		</article>
	{/each}
</section>
