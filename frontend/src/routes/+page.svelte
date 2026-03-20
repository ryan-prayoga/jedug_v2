<script lang="ts">
	import {
		ChartIcon,
		DocumentIcon,
		LocationIcon,
		MapIcon,
		ShieldCheckIcon
	} from '$lib/icons';

	const quickLinks = [
		{
			href: '/lapor',
			label: 'Laporkan Jalan Rusak',
			copy: 'Form cepat, mobile-friendly, dan langsung terhubung ke titik issue.',
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
			copy: 'Baca ringkasan issue, status, dan wilayah dengan beban UI yang ringan.',
			icon: ChartIcon,
			primary: false
		}
	];

	const valueCards = [
		{
			title: 'Pelaporan cepat dari HP',
			copy: 'Warga bisa langsung ambil foto, tentukan tingkat keparahan, dan kirim tanpa alur yang rumit.',
			icon: DocumentIcon
		},
		{
			title: 'Map-first dan mudah dipindai',
			copy: 'Titik kerusakan terlihat rapi di peta publik, lengkap dengan panel detail dan status issue.',
			icon: LocationIcon
		},
		{
			title: 'Moderasi menjaga kualitas data',
			copy: 'Area admin disiapkan untuk menilai, menyembunyikan, menolak, atau menandai issue selesai.',
			icon: ShieldCheckIcon
		}
	];
</script>

<div class="public-stack">
	<section class="jedug-card overflow-hidden">
		<div class="grid gap-6 px-5 py-6 sm:px-6 lg:grid-cols-[minmax(0,1.2fr)_minmax(260px,0.8fr)]">
			<div class="space-y-5">
				<span class="section-kicker">Civic-tech yang serius, tetap ramah dipakai</span>
				<div class="space-y-4">
					<h1 class="section-title text-balance">
						JEDUG membantu laporan jalan rusak terlihat jelas, rapi, dan sulit diabaikan.
					</h1>
					<p class="section-copy max-w-[52ch]">
						Platform publik ini mengubah keluhan lapangan menjadi bukti visual yang mudah
						dipantau di peta, aman untuk pelapor anonim, dan tetap cukup rapi untuk kebutuhan
						moderasi operasional.
					</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<span class="badge-tint">Mobile-first</span>
					<span class="badge-muted">Anonim-friendly</span>
					<span class="badge-muted">Map-first</span>
					<span class="badge-muted">Moderation-ready</span>
				</div>
			</div>

			<div class="jedug-panel grid gap-3 p-4">
				<div class="space-y-1">
					<p class="surface-label">Alur utama</p>
					<p class="text-base font-bold text-slate-950">Lapor, pantau, tindak lanjuti.</p>
				</div>
				<div class="space-y-3">
					<div class="rounded-[22px] border border-white/80 bg-white px-4 py-3 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
						<p class="text-sm font-semibold text-slate-900">Ambil foto dan kirim laporan</p>
						<p class="mt-1 text-xs leading-5 text-slate-500">Alur lapor dirancang untuk cepat dan jelas di mobile.</p>
					</div>
					<div class="rounded-[22px] border border-white/80 bg-white px-4 py-3 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
						<p class="text-sm font-semibold text-slate-900">Pantau titik issue di peta</p>
						<p class="mt-1 text-xs leading-5 text-slate-500">Marker, list, dan heatmap membantu membaca area rawan.</p>
					</div>
					<div class="rounded-[22px] border border-white/80 bg-white px-4 py-3 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
						<p class="text-sm font-semibold text-slate-900">Moderasi menjaga kualitas publik</p>
						<p class="mt-1 text-xs leading-5 text-slate-500">Admin bisa menilai issue tanpa mengubah flow inti warga.</p>
					</div>
				</div>
			</div>
		</div>
	</section>

	<section class="grid gap-3">
		{#each quickLinks as item}
			<a
				href={item.href}
				class={item.primary
					? 'btn-primary min-h-[76px] justify-start rounded-[26px] px-5 text-left'
					: 'jedug-card-soft flex min-h-[76px] items-center gap-4 px-5 py-4 text-left transition hover:-translate-y-0.5 hover:shadow-[0_16px_34px_rgba(15,23,42,0.08)]'}
			>
				<div
					class={item.primary
						? 'flex size-12 shrink-0 items-center justify-center rounded-[18px] bg-white/12 text-white'
						: 'flex size-12 shrink-0 items-center justify-center rounded-[18px] bg-brand-50 text-brand-600'}
				>
					<svelte:component this={item.icon} class="size-6" />
				</div>
				<div class="min-w-0">
					<p class={item.primary ? 'text-base font-bold text-white' : 'text-base font-bold text-slate-950'}>
						{item.label}
					</p>
					<p class={item.primary ? 'mt-1 text-sm text-white/82' : 'mt-1 text-sm leading-6 text-slate-500'}>
						{item.copy}
					</p>
				</div>
			</a>
		{/each}
	</section>

	<section class="grid gap-3">
		{#each valueCards as item}
			<article class="jedug-card-soft flex gap-4 px-5 py-5">
				<div class="flex size-12 shrink-0 items-center justify-center rounded-[18px] bg-brand-50 text-brand-600">
					<svelte:component this={item.icon} class="size-6" />
				</div>
				<div>
					<h2 class="text-base font-bold text-slate-950">{item.title}</h2>
					<p class="mt-2 text-sm leading-6 text-slate-500">{item.copy}</p>
				</div>
			</article>
		{/each}
	</section>
</div>
