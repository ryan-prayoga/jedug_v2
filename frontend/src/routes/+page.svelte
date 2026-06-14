<script lang="ts">
	import { ArrowRightIcon, ChartIcon, DocumentIcon, MapIcon } from '$lib/icons';

	const quickLinks = [
		{
			href: '/lapor',
			label: 'Laporkan Jalan Rusak',
			copy: 'Form cepat, mobile-friendly, langsung terhubung ke titik issue.',
			icon: DocumentIcon,
			primary: true
		},
		{
			href: '/issues',
			label: 'Lihat Peta Laporan',
			copy: 'Amati marker dan hotspot area rawan dari peta publik.',
			icon: MapIcon,
			primary: false
		},
		{
			href: '/stats',
			label: 'Statistik Jalan Rusak',
			copy: 'Baca ringkasan issue, status, dan wilayah dengan UI yang ringan.',
			icon: ChartIcon,
			primary: false
		}
	];

	const flow = [
		{
			title: 'Ambil foto dan kirim laporan',
			copy: 'Alur lapor dirancang cepat dan jelas di mobile.'
		},
		{
			title: 'Pantau titik issue di peta',
			copy: 'Marker, list, dan heatmap membantu membaca area rawan.'
		},
		{
			title: 'Moderasi menjaga kualitas publik',
			copy: 'Admin menilai issue tanpa mengubah flow inti warga.'
		}
	];

	const valueCards = [
		{
			title: 'Pelaporan cepat dari HP',
			copy: 'Warga langsung ambil foto, tentukan tingkat keparahan, dan kirim tanpa alur rumit.'
		},
		{
			title: 'Map-first dan mudah dipindai',
			copy: 'Titik kerusakan terlihat rapi di peta publik, lengkap panel detail dan status issue.'
		},
		{
			title: 'Moderasi menjaga kualitas data',
			copy: 'Area admin siap menilai, menyembunyikan, menolak, atau menandai issue selesai.'
		}
	];

	function pad(n: number): string {
		return String(n + 1).padStart(2, '0');
	}
</script>

<div class="public-stack">
	<!-- Hero -->
	<section class="grid gap-10 lg:grid-cols-[minmax(0,1.15fr)_minmax(260px,0.85fr)] lg:items-start">
		<div class="flex flex-col gap-6">
			<p class="kicker">Pelaporan jalan rusak partisipatif</p>
			<h1 class="section-title text-balance max-w-[20ch]">
				Laporan jalan rusak yang jelas, rapi, dan sulit diabaikan.
			</h1>
			<p class="section-copy max-w-[54ch]">
				JEDUG mengubah keluhan lapangan menjadi bukti visual yang mudah dipantau di peta,
				aman untuk pelapor anonim, dan cukup rapi untuk kebutuhan moderasi operasional.
			</p>
			<div class="flex flex-wrap gap-2 pt-1">
				<span class="badge-tint">Mobile-first</span>
				<span class="badge-muted">Anonim-friendly</span>
				<span class="badge-muted">Map-first</span>
				<span class="badge-muted">Moderation-ready</span>
			</div>
		</div>

		<!-- Alur utama — editorial numbered list -->
		<div class="flex flex-col">
			<p class="surface-label mb-4">Alur utama</p>
			<ol class="flex flex-col">
				{#each flow as step, i}
					<li class="flex gap-4 border-t border-hairline py-4">
						<span class="editorial-index pt-0.5">{pad(i)}</span>
						<div class="min-w-0">
							<p class="text-sm font-semibold text-ink">{step.title}</p>
							<p class="mt-1 text-xs leading-5 text-muted">{step.copy}</p>
						</div>
					</li>
				{/each}
			</ol>
		</div>
	</section>

	<!-- Quick links -->
	<section class="flex flex-col gap-3">
		{#each quickLinks as item}
			{@const Icon = item.icon}
			<a
				href={item.href}
				class={item.primary
					? 'flex items-center justify-between gap-4 rounded-[12px] bg-ink px-5 py-5 text-paper transition-colors hover:bg-brand'
					: 'flex items-center justify-between gap-4 rounded-[12px] border border-hairline bg-surface px-5 py-5 transition-colors hover:border-ink'}
			>
				<div class="flex min-w-0 items-center gap-4">
					<Icon class={item.primary ? 'size-5 shrink-0 text-paper' : 'size-5 shrink-0 text-muted'} />
					<div class="min-w-0">
						<p class={item.primary ? 'text-base font-semibold text-paper' : 'text-base font-semibold text-ink'}>
							{item.label}
						</p>
						<p class={item.primary ? 'mt-1 text-sm text-paper/75' : 'mt-1 text-sm leading-6 text-muted'}>
							{item.copy}
						</p>
					</div>
				</div>
				<ArrowRightIcon class={item.primary ? 'size-5 shrink-0 text-paper' : 'size-5 shrink-0 text-subtle'} />
			</a>
		{/each}
	</section>

	<!-- Value props — editorial numbered -->
	<section class="flex flex-col">
		<p class="surface-label mb-4">Kenapa JEDUG</p>
		{#each valueCards as item, i}
			<article class="flex gap-4 border-t border-hairline py-6">
				<span class="editorial-index pt-1">{pad(i)}</span>
				<div class="min-w-0">
					<h2 class="font-serif text-lg font-semibold text-ink">{item.title}</h2>
					<p class="mt-2 text-sm leading-6 text-muted">{item.copy}</p>
				</div>
			</article>
		{/each}
	</section>
</div>
