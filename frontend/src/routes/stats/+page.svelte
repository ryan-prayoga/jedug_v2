<script lang="ts">
	import { onMount } from 'svelte';
	import { resolveLocationLabel } from '$lib/api/location';
	import { getPublicStats, getPublicStatsRegionOptions } from '$lib/api/stats';
	import type {
		LocationLabelData,
		PublicStats,
		PublicStatsProvinceOption,
		PublicStatsRegion,
		PublicStatsRegionOption,
		PublicStatsRegionOptionsData,
		PublicTopIssue
	} from '$lib/api/types';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import {
		CameraIcon,
		ChartIcon,
		CheckCircleIcon,
		ClockIcon,
		DangerIcon,
		DocumentIcon,
		LocationIcon,
		RankingIcon,
		RefreshIcon,
		RouteIcon,
		TargetIcon,
		WidgetIcon
	} from '$lib/icons';
	import { formatDate } from '$lib/utils/date';
	import { getLocation } from '$lib/utils/geolocation';

	type StatsScope = {
		provinceID?: number | null;
		regencyID?: number | null;
	};

	let stats = $state<PublicStats | null>(null);
	let loading = $state(true);
	let refreshing = $state(false);
	let pageErrorMessage = $state<string | null>(null);
	let inlineErrorMessage = $state<string | null>(null);
	let optionsErrorMessage = $state<string | null>(null);
	let regionOptions = $state<PublicStatsRegionOptionsData | null>(null);
	let optionsLoading = $state(false);
	let applyingLocationDefault = $state(false);
	let selectedProvinceID = $state('');
	let selectedRegencyID = $state('');
	let locationHint = $state('Wilayah awal akan dicoba menyesuaikan lokasi kamu, lalu tetap bisa diganti manual.');
	let triedLocationDefault = false;

	const provinceOptions = $derived.by((): PublicStatsProvinceOption[] => regionOptions?.provinces ?? []);
	const selectedProvince = $derived.by(() => {
		return provinceOptions.find((option) => String(option.id) === selectedProvinceID) ?? null;
	});
	const regencyOptions = $derived.by((): PublicStatsRegionOption[] => selectedProvince?.regencies ?? []);
	const selectedRegency = $derived.by(() => {
		return regencyOptions.find((option) => String(option.id) === selectedRegencyID) ?? null;
	});

	const isGlobalEmpty = $derived.by(() => {
		if (!stats) return false;
		return stats.global.total_issues === 0;
	});

	const isScopedEmpty = $derived.by(() => {
		if (!stats) return false;
		return stats.summary.total_issues === 0;
	});

	const statusTotal = $derived.by(() => {
		if (!stats) return 0;
		return stats.status.open + stats.status.fixed + stats.status.archived;
	});

	const activeScopeLabel = $derived.by(() => {
		const selectedScope = joinLocationParts([selectedRegency?.name, selectedProvince?.name]);
		return stats?.active_scope.label || stats?.filters.scope_label || selectedScope || 'Semua wilayah publik';
	});

	const summaryCards = $derived.by(() => {
		if (!stats) return [];

		return [
			{
				label: 'Total Issue',
				value: formatNumber(stats.summary.total_issues),
				copy: 'Issue publik pada scope aktif',
				icon: WidgetIcon
			},
			{
				label: 'Issue Minggu Ini',
				value: formatNumber(stats.summary.total_issues_this_week),
				copy: 'Issue baru pekan berjalan',
				icon: ChartIcon
			},
			{
				label: 'Total Korban',
				value: formatNumber(stats.summary.total_casualties),
				copy: 'Korban yang pernah tercatat',
				icon: DangerIcon
			},
			{
				label: 'Total Foto',
				value: formatNumber(stats.summary.total_photos),
				copy: 'Media publik di scope ini',
				icon: CameraIcon
			},
			{
				label: 'Total Laporan',
				value: formatNumber(stats.summary.total_reports),
				copy: 'Akumulasi submission warga',
				icon: DocumentIcon
			}
		];
	});

	const statusCards = $derived.by(() => {
		if (!stats) return [];

		return [
			{
				label: 'Issue Open',
				value: stats.status.open,
				percent: getStatusPercent(stats.status.open),
				copy: 'Masih perlu perhatian atau tindak lanjut',
				icon: TargetIcon,
				barClass: 'bg-brand-500'
			},
			{
				label: 'Issue Fixed',
				value: stats.status.fixed,
				percent: getStatusPercent(stats.status.fixed),
				copy: 'Sudah ditandai selesai di sistem',
				icon: CheckCircleIcon,
				barClass: 'bg-emerald-500'
			},
			{
				label: 'Issue Archived',
				value: stats.status.archived,
				percent: getStatusPercent(stats.status.archived),
				copy: 'Diarsipkan dari alur aktif publik',
				icon: DocumentIcon,
				barClass: 'bg-slate-500'
			}
		];
	});

	const timeCards = $derived.by(() => {
		if (!stats) return [];

		return [
			{
				label: 'Rata-rata umur issue',
				value: `${formatDecimal(stats.time.average_issue_age_days)} hari`,
				copy: 'Rata-rata umur seluruh issue pada scope aktif',
				icon: ClockIcon,
				href: null
			},
			{
				label: 'Issue open tertua',
				value: `${formatNumber(stats.time.oldest_open_issue_age_days)} hari`,
				copy: stats.time.oldest_open_first_seen_at
					? `Pertama tercatat ${formatDate(stats.time.oldest_open_first_seen_at)}`
					: 'Tanggal pertama terlihat belum tersedia',
				icon: RefreshIcon,
				href: stats.time.oldest_open_issue_id
					? `/issues/${stats.time.oldest_open_issue_id}`
					: null
			}
		];
	});

	onMount(() => {
		void initPage();
	});

	async function initPage() {
		await Promise.all([fetchRegionOptions(), fetchStats({}, { preserveData: false })]);
		await applyLocationDefault();
	}

	async function fetchRegionOptions() {
		optionsLoading = true;
		optionsErrorMessage = null;

		try {
			const result = await getPublicStatsRegionOptions();
			if (!result.data) {
				optionsErrorMessage = 'Daftar wilayah statistik belum tersedia saat ini.';
				regionOptions = buildFallbackRegionOptions(stats);
				return;
			}

			regionOptions = result.data;
			if (stats) {
				syncSelectedFilters(stats);
			}
		} catch (err) {
			optionsErrorMessage =
				err instanceof Error ? err.message : 'Gagal memuat daftar wilayah statistik.';
			regionOptions = buildFallbackRegionOptions(stats);
		} finally {
			optionsLoading = false;
		}
	}

	async function fetchStats(
		scope: StatsScope = {},
		{ preserveData = true }: { preserveData?: boolean } = {}
	) {
		if (!preserveData || !stats) {
			loading = true;
		} else {
			refreshing = true;
		}

		pageErrorMessage = null;
		inlineErrorMessage = null;

		try {
			const result = await getPublicStats(scope);
			if (!result.data) {
				if (!stats) {
					pageErrorMessage = 'Data statistik publik tidak tersedia saat ini.';
				} else {
					inlineErrorMessage = 'Data statistik belum tersedia untuk wilayah ini.';
				}
				return;
			}

			stats = result.data;
			syncRegionOptionsWithStats(result.data);
			syncSelectedFilters(result.data);
		} catch (err) {
			const message = err instanceof Error ? err.message : 'Gagal memuat statistik publik.';
			if (!stats || !preserveData) {
				pageErrorMessage = message;
			} else {
				inlineErrorMessage = message;
			}
		} finally {
			loading = false;
			refreshing = false;
		}
	}

	function syncSelectedFilters(data: PublicStats) {
		selectedProvinceID = data.filters.active_province_id
			? String(data.filters.active_province_id)
			: '';
		selectedRegencyID = data.filters.active_regency_id ? String(data.filters.active_regency_id) : '';
	}

	function buildFallbackRegionOptions(statsData: PublicStats | null): PublicStatsRegionOptionsData | null {
		if (!statsData || statsData.filters.province_options.length === 0) {
			return null;
		}

		return {
			provinces: statsData.filters.province_options.map((province) => ({
				...province,
				regencies:
					statsData.filters.active_province_id === province.id
						? [...statsData.filters.regency_options]
						: []
			}))
		};
	}

	function syncRegionOptionsWithStats(statsData: PublicStats) {
		const fallback = buildFallbackRegionOptions(statsData);
		if (!fallback) {
			return;
		}

		if (!regionOptions || regionOptions.provinces.length === 0) {
			regionOptions = fallback;
			return;
		}

		const provinceMap = new Map<number, PublicStatsProvinceOption>();
		for (const province of regionOptions.provinces) {
			provinceMap.set(province.id, {
				...province,
				regencies: [...province.regencies]
			});
		}

		for (const province of fallback.provinces) {
			provinceMap.set(province.id, {
				...province,
				regencies:
					statsData.filters.active_province_id === province.id
						? [...province.regencies]
						: provinceMap.get(province.id)?.regencies ?? []
			});
		}

		regionOptions = {
			provinces: fallback.provinces.map((province) => provinceMap.get(province.id) ?? province)
		};
	}

	async function applyLocationDefault(options: { forceFresh?: boolean; manual?: boolean } = {}) {
		if ((!options.manual && triedLocationDefault) || provinceOptions.length === 0) {
			if (provinceOptions.length === 0) {
				locationHint =
					'Opsi wilayah statistik belum siap, jadi pilih provinsi dan kabupaten/kota secara manual saat daftar tersedia.';
			}
			return;
		}
		if (!options.manual) {
			triedLocationDefault = true;
		}

		applyingLocationDefault = true;

		try {
			const point = await getLocation({ forceFresh: options.forceFresh });
			const labelResult = await resolveLocationLabel(point.latitude, point.longitude);
			const label = labelResult.data;
			if (!label) {
				locationHint = 'Lokasi belum bisa dipetakan, jadi statistik memakai wilayah default yang masih bisa kamu ganti manual.';
				return;
			}

			const province = findProvinceOption(label, provinceOptions);
			if (!province) {
				locationHint =
					'Lokasi ditemukan, tetapi provinsinya belum cocok dengan data statistik saat ini. Kamu masih bisa pilih manual dari daftar.';
				return;
			}

			const regency = findRegencyOption(label, province.regencies);
			if (regency) {
				await fetchStats({ provinceID: province.id, regencyID: regency.id });
				locationHint = `Wilayah awal disesuaikan ke ${regency.name}, ${province.name}.`;
				return;
			}

			await fetchStats({ provinceID: province.id, regencyID: null });
			locationHint = `Provinsi ${province.name} berhasil dikenali. Pilih kabupaten/kota yang paling sesuai jika belum terisi otomatis.`;
		} catch {
			locationHint =
				'Lokasi tidak tersedia, jadi statistik memakai wilayah default. Kamu tetap bisa ganti provinsi dan kabupaten/kota secara manual.';
		} finally {
			applyingLocationDefault = false;
		}
	}

	function formatNumber(value: number | null | undefined): string {
		const safe = typeof value === 'number' && Number.isFinite(value) ? value : 0;
		return new Intl.NumberFormat('id-ID').format(Math.max(0, Math.trunc(safe)));
	}

	function formatDecimal(value: number | null | undefined): string {
		const safe = typeof value === 'number' && Number.isFinite(value) ? value : 0;
		return new Intl.NumberFormat('id-ID', {
			minimumFractionDigits: 1,
			maximumFractionDigits: 1
		}).format(Math.max(0, safe));
	}

	function getStatusPercent(value: number): number {
		if (statusTotal <= 0) return 0;
		return Math.round((value / statusTotal) * 100);
	}

	function normalizeRegionToken(value: string | null | undefined): string {
		return (value || '')
			.normalize('NFKD')
			.replace(/[\u0300-\u036f]/gu, '')
			.toLocaleLowerCase('id-ID')
			.replace(/[.,()/_-]+/gu, ' ')
			.replace(/\bdaerah khusus ibukota\b/gu, 'dki')
			.replace(/\bdaerah istimewa\b/gu, 'di')
			.replace(/\bkab\.?\b/gu, 'kabupaten')
			.replace(/\bkec\.?\b/gu, 'kecamatan')
			.replace(/\bkota administrasi\b/gu, 'kota')
			.replace(/\bkabupaten administrasi\b/gu, 'kabupaten')
			.replace(/\badministrasi\b/gu, ' ')
			.replace(/\s+/gu, ' ')
			.trim();
	}

	function stripRegionTypePrefix(value: string): string {
		return value
			.replace(/^(provinsi|province)\s+/u, '')
			.replace(/^(kabupaten|regency)\s+/u, '')
			.replace(/^(kota|city)\s+/u, '')
			.replace(/^(kecamatan|district)\s+/u, '')
			.replace(/^(kelurahan|desa|village|subdistrict)\s+/u, '')
			.trim();
	}

	function buildRegionKeys(value: string | null | undefined): string[] {
		const normalized = normalizeRegionToken(value);
		if (!normalized) return [];

		const stripped = stripRegionTypePrefix(normalized);
		const keys = new Set<string>([normalized, stripped].filter(Boolean));

		if (normalized === 'dki jakarta') {
			keys.add('jakarta');
		}
		if (normalized === 'di yogyakarta') {
			keys.add('yogyakarta');
		}

		return Array.from(keys);
	}

	function hasSameRegionMeaning(left: string, right: string): boolean {
		if (left === right) return true;
		if (!left || !right) return false;

		const leftTokens = new Set(left.split(' ').filter(Boolean));
		const rightTokens = new Set(right.split(' ').filter(Boolean));
		const shorter = leftTokens.size <= rightTokens.size ? leftTokens : rightTokens;
		const longer = leftTokens.size <= rightTokens.size ? rightTokens : leftTokens;

		for (const token of shorter) {
			if (!longer.has(token)) {
				return false;
			}
		}

		return shorter.size > 0;
	}

	function findOptionByCandidates<T extends { name: string }>(
		options: T[],
		candidates: Array<string | null | undefined>
	): T | null {
		const candidateKeys = candidates.flatMap((candidate) => buildRegionKeys(candidate));

		if (candidateKeys.length === 0) return null;

		for (const option of options) {
			const optionKeys = buildRegionKeys(option.name);
			if (
				optionKeys.some((optionKey) =>
					candidateKeys.some((candidateKey) => hasSameRegionMeaning(optionKey, candidateKey))
				)
			) {
				return option;
			}
		}

		return null;
	}

	function findProvinceOption(
		label: LocationLabelData,
		options: PublicStatsProvinceOption[]
	): PublicStatsProvinceOption | null {
		return findOptionByCandidates(options, [
			label.province_name,
			label.grandparent_name,
			label.parent_name,
			label.region_name
		]);
	}

	function findRegencyOption(
		label: LocationLabelData,
		options: PublicStatsRegionOption[]
	): PublicStatsRegionOption | null {
		return findOptionByCandidates(options, [
			label.regency_name,
			label.parent_name,
			label.region_name
		]);
	}

	function getIssueName(item: PublicTopIssue): string {
		if (
			item.road_name &&
			item.road_name.trim() !== '' &&
			!/^Kawasan sekitar\s*-?\d+(?:\.\d+)?\s*,\s*-?\d+(?:\.\d+)?$/iu.test(item.road_name.trim())
		) {
			return item.road_name;
		}
		const adminFallback = joinLocationParts([item.district_name, item.regency_name, item.province_name]);
		if (adminFallback !== '') return adminFallback;
		if (item.region_name && item.region_name.trim() !== '') return item.region_name;
		return 'Issue tanpa nama jalan';
	}

	function joinLocationParts(parts: Array<string | null | undefined>): string {
		const unique = new Set<string>();
		for (const part of parts) {
			const trimmed = part?.trim();
			if (trimmed) unique.add(trimmed);
		}
		return Array.from(unique).join(', ');
	}

	function getIssueLocation(item: PublicTopIssue): string {
		const location = joinLocationParts([item.district_name, item.regency_name, item.province_name]);
		if (location !== '') return location;
		if (item.region_name && item.region_name.trim() !== '') return item.region_name;
		return 'Wilayah administratif belum tersedia';
	}

	function getIssueContext(item: PublicTopIssue): string {
		return `${formatNumber(item.submission_count)} laporan · ${formatNumber(item.casualty_count)} korban · ${formatNumber(item.age_days)} hari`;
	}

	function getRegionLevelLabel(level: string | null | undefined): string {
		switch (level) {
			case 'district':
				return 'Kecamatan';
			case 'regency':
				return 'Kabupaten/Kota';
			case 'province':
				return 'Provinsi';
			default:
				return 'Wilayah';
		}
	}

	function getRegionContext(item: PublicStatsRegion): string {
		switch (item.region_level) {
			case 'district':
				return joinLocationParts([item.regency_name, item.province_name]);
			case 'regency':
				return joinLocationParts([item.province_name]);
			default:
				return '';
		}
	}

	function getGeneratedLabel(statsData: PublicStats | null): string | null {
		if (!statsData?.generated_at) return null;
		return formatDate(statsData.generated_at);
	}

	async function handleProvinceChange(event: Event) {
		const value = Number((event.currentTarget as HTMLSelectElement).value || 0);
		selectedProvinceID = value ? String(value) : '';
		selectedRegencyID = '';
		await fetchStats({ provinceID: value || null, regencyID: null });
	}

	async function handleRegencyChange(event: Event) {
		const provinceID = Number(selectedProvinceID || 0);
		const regencyID = Number((event.currentTarget as HTMLSelectElement).value || 0);
		if (!provinceID) return;
		await fetchStats({
			provinceID,
			regencyID: regencyID || null
		});
	}

	function getCurrentScope(): StatsScope {
		const provinceID = Number(selectedProvinceID || 0);
		const regencyID = Number(selectedRegencyID || 0);

		return {
			provinceID: provinceID || null,
			regencyID: regencyID || null
		};
	}
</script>

<svelte:head>
	<title>Statistik Publik Jalan Rusak | JEDUG</title>
	<meta
		name="description"
		content="Dashboard statistik publik JEDUG: total issue, status, umur issue, leaderboard wilayah, dan top issue."
	/>
</svelte:head>

<div class="public-stack pb-10">
	<section class="jedug-card overflow-hidden">
		<div class="grid gap-6 bg-[radial-gradient(circle_at_top_left,rgba(229,72,77,0.14),transparent_32%),linear-gradient(180deg,#fff9f8_0%,#ffffff_100%)] p-5 md:p-6">
			<div class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
				<div class="max-w-[58ch]">
					<span class="section-kicker">
						<ChartIcon class="size-4" />
						Statistik publik
					</span>
					<h1 class="mt-4 text-balance text-[clamp(2rem,4vw,3.2rem)] font-[800] tracking-[-0.05em] text-slate-950">
						Dashboard jalan rusak yang lebih mudah dipindai.
					</h1>
					<p class="mt-3 text-sm leading-6 text-slate-600 sm:text-[15px]">
						Ringkasan agregasi issue publik untuk civic storytelling yang cepat dipahami, dengan filter wilayah yang tetap aman dan ringan untuk dipakai warga.
					</p>
				</div>

				<div class="flex flex-wrap gap-2 xl:justify-end">
					<span class="badge-tint">
						<RouteIcon class="size-4" />
						{activeScopeLabel}
					</span>
					{#if getGeneratedLabel(stats)}
						<span class="badge-muted">
							<ClockIcon class="size-4" />
							Update terakhir {getGeneratedLabel(stats)}
						</span>
					{/if}
					<button
						type="button"
						class="btn-secondary"
						onclick={() => initPage()}
						disabled={loading || refreshing}
					>
						<RefreshIcon class={`size-[18px] ${loading || refreshing ? 'animate-spin' : ''}`} />
						{loading || refreshing ? 'Memuat...' : 'Muat ulang'}
					</button>
				</div>
			</div>

			<div class="grid gap-3 sm:grid-cols-3">
				<article class="rounded-[22px] border border-white/70 bg-white/88 px-4 py-4 shadow-[0_10px_26px_rgba(15,23,42,0.05)]">
					<div class="flex items-center gap-2 text-slate-500">
						<LocationIcon class="size-[18px]" />
						<span class="surface-label">Scope aktif</span>
					</div>
					<strong class="mt-2 block text-sm font-bold text-slate-950">{activeScopeLabel}</strong>
					<p class="mt-1 text-xs leading-5 text-slate-500">
						Leaderboard, top issue, dan seluruh metrik mengikuti wilayah yang sama.
					</p>
				</article>

				<article class="rounded-[22px] border border-white/70 bg-white/88 px-4 py-4 shadow-[0_10px_26px_rgba(15,23,42,0.05)]">
					<div class="flex items-center gap-2 text-slate-500">
						<TargetIcon class="size-[18px]" />
						<span class="surface-label">Kecocokan lokasi</span>
					</div>
					<strong class="mt-2 block text-sm font-bold text-slate-950">
						{applyingLocationDefault ? 'Mencocokkan lokasi...' : 'Bisa otomatis atau manual'}
					</strong>
					<p class="mt-1 text-xs leading-5 text-slate-500">
						Statistik mencoba mengikuti lokasi kamu terlebih dulu, lalu tetap bisa diubah manual.
					</p>
				</article>

				<article class="rounded-[22px] border border-white/70 bg-white/88 px-4 py-4 shadow-[0_10px_26px_rgba(15,23,42,0.05)]">
					<div class="flex items-center gap-2 text-slate-500">
						<WidgetIcon class="size-[18px]" />
						<span class="surface-label">Snapshot global</span>
					</div>
					<strong class="mt-2 block text-sm font-bold text-slate-950">
						{stats ? `${formatNumber(stats.global.total_issues)} issue publik` : 'Menunggu data'}
					</strong>
					<p class="mt-1 text-xs leading-5 text-slate-500">
						Ringkasan global tetap tersedia untuk menjaga konteks nasional saat scope dipersempit.
					</p>
				</article>
			</div>
		</div>
	</section>

	{#if loading}
		<LoadingState message="Memuat statistik publik..." />
	{:else if pageErrorMessage}
		<ErrorState message={pageErrorMessage} onretry={() => initPage()} />
	{:else if stats && isGlobalEmpty}
		<EmptyState
			message="Belum ada statistik publik yang bisa ditampilkan."
			ctaHref="/lapor"
			ctaLabel="Kirim Laporan Pertama"
		/>
	{:else if stats}
		<section class="jedug-card p-5 md:p-6">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div class="max-w-[56ch]">
					<span class="section-kicker">
						<LocationIcon class="size-4" />
						Filter wilayah
					</span>
					<h2 class="mt-4 text-2xl font-[800] tracking-[-0.04em] text-slate-950">Atur scope statistik aktif</h2>
					<p class="mt-3 text-sm leading-6 text-slate-500">
						Ringkasan, status, umur issue, leaderboard, dan top issue mengikuti wilayah aktif yang sama agar pembacaan tetap konsisten.
					</p>
				</div>
				<span class="badge-tint self-start lg:self-auto">
					<RouteIcon class="size-4" />
					{activeScopeLabel}
				</span>
			</div>

			<div class="mt-5 grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] lg:items-end">
				<label class="input-shell">
					<span class="input-label">Provinsi</span>
					<select
						class="select-field w-full"
						bind:value={selectedProvinceID}
						disabled={refreshing || optionsLoading || provinceOptions.length === 0}
						onchange={handleProvinceChange}
					>
						<option value="" disabled>
							{optionsLoading
								? 'Memuat provinsi...'
								: provinceOptions.length === 0
									? 'Provinsi belum tersedia'
									: 'Pilih provinsi'}
						</option>
						{#each provinceOptions as province (province.id)}
							<option value={province.id}>{province.name}</option>
						{/each}
					</select>
				</label>

				<label class="input-shell">
					<span class="input-label">Kabupaten/Kota</span>
					<select
						class="select-field w-full"
						bind:value={selectedRegencyID}
						disabled={refreshing || !selectedProvinceID || regencyOptions.length === 0}
						onchange={handleRegencyChange}
					>
						<option value="">
							{!selectedProvinceID
								? 'Pilih provinsi dulu'
								: regencyOptions.length === 0
									? 'Kabupaten/kota belum tersedia'
									: 'Semua kabupaten/kota di provinsi ini'}
						</option>
						{#each regencyOptions as regency (regency.id)}
							<option value={regency.id}>{regency.name}</option>
						{/each}
					</select>
				</label>

				<button
					type="button"
					class="btn-secondary lg:min-w-[210px]"
					onclick={() => applyLocationDefault({ forceFresh: true, manual: true })}
					disabled={refreshing || optionsLoading || applyingLocationDefault || provinceOptions.length === 0}
				>
					<LocationIcon class={`size-[18px] ${applyingLocationDefault ? 'animate-pulse' : ''}`} />
					{applyingLocationDefault ? 'Mencocokkan lokasi...' : 'Gunakan lokasi saya'}
				</button>
			</div>

			<div class="mt-4 grid gap-3">
				<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-4">
					<div class="flex items-start gap-3">
						<LocationIcon class="mt-0.5 size-5 shrink-0 text-brand-600" />
						<p class="text-sm leading-6 text-slate-600">{locationHint}</p>
					</div>
				</div>

				{#if optionsLoading}
					<div class="notice-panel">Memuat daftar wilayah statistik...</div>
				{:else if optionsErrorMessage}
					<div class="notice-panel">{optionsErrorMessage}</div>
				{:else if provinceOptions.length > 0}
					<div class="rounded-[20px] border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
						Kamu tetap bisa pilih manual jika lokasi browser tidak cocok persis.
					</div>
				{/if}

				{#if inlineErrorMessage}
					<div class="error-panel">{inlineErrorMessage}</div>
				{/if}
				{#if isScopedEmpty}
					<div class="rounded-[20px] border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-600">
						Belum ada issue publik di scope ini. Kamu masih bisa ganti wilayah dari filter di atas.
					</div>
				{/if}
			</div>
		</section>

		<section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
			{#each summaryCards as item}
				{@const ItemIcon = item.icon}
				<article class="metric-card">
					<div class="flex items-center gap-2 text-slate-500">
						<ItemIcon class="size-[18px]" />
						<span class="metric-label">{item.label}</span>
					</div>
					<strong class="metric-value">{item.value}</strong>
					<p class="metric-copy">{item.copy}</p>
				</article>
			{/each}
		</section>

		<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
			<div class="flex flex-col gap-5">
				<section class="jedug-card p-5 md:p-6">
					<div class="flex items-start gap-3">
						<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-slate-100 text-slate-700">
							<TargetIcon class="size-6" />
						</div>
						<div>
							<h2 class="text-xl font-[800] tracking-[-0.03em] text-slate-950">Status breakdown</h2>
							<p class="mt-2 text-sm leading-6 text-slate-500">
								Porsi issue open, fixed, dan archived pada scope aktif saat ini.
							</p>
						</div>
					</div>

					<div class="mt-5 grid gap-3">
						{#each statusCards as item}
							{@const ItemIcon = item.icon}
							<article class="rounded-[24px] border border-slate-200 bg-white px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
								<div class="flex items-start justify-between gap-3">
									<div class="flex items-center gap-3">
										<div class="flex size-10 items-center justify-center rounded-[18px] bg-slate-100 text-slate-700">
											<ItemIcon class="size-5" />
										</div>
										<div>
											<p class="text-sm font-bold text-slate-900">{item.label}</p>
											<p class="mt-1 text-xs leading-5 text-slate-500">{item.copy}</p>
										</div>
									</div>
									<div class="text-right">
										<strong class="text-lg font-[800] text-slate-950">{formatNumber(item.value)}</strong>
										<p class="mt-1 text-xs font-semibold text-slate-500">{item.percent}%</p>
									</div>
								</div>
								<div class="mt-4 h-2.5 rounded-full bg-slate-100">
									<div class={`h-full rounded-full ${item.barClass}`} style={`width:${item.percent}%`}></div>
								</div>
							</article>
						{/each}
					</div>
				</section>

				<section class="jedug-card p-5 md:p-6">
					<div class="flex items-start gap-3">
						<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-brand-50 text-brand-600">
							<RankingIcon class="size-6" />
						</div>
						<div>
							<h2 class="text-xl font-[800] tracking-[-0.03em] text-slate-950">Region leaderboard</h2>
							<p class="mt-2 text-sm leading-6 text-slate-500">
								Wilayah administratif dengan laporan terbanyak di scope yang sedang dipilih.
							</p>
						</div>
					</div>

					{#if stats.regions.length === 0}
						<div class="mt-5 rounded-[24px] border border-dashed border-slate-200 bg-slate-50 px-4 py-5">
							<EmptyState message="Wilayah administratif belum tersedia untuk scope ini." />
						</div>
					{:else}
						<div class="mt-5 grid gap-3">
							{#each stats.regions as region, index (region.region_id)}
								<article class="rounded-[24px] border border-slate-200 bg-white px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
									<div class="flex items-start gap-3">
										<div class="flex size-10 shrink-0 items-center justify-center rounded-full bg-brand-50 text-sm font-bold text-brand-700">
											{index + 1}
										</div>
										<div class="min-w-0 flex-1">
											<h3 class="text-sm font-bold leading-6 text-slate-950">{region.region_name}</h3>
											{#if getRegionContext(region) !== ''}
												<p class="mt-1 text-xs leading-5 text-slate-500">
													{getRegionLevelLabel(region.region_level)} · {getRegionContext(region)}
												</p>
											{/if}
											<p class="mt-2 text-sm leading-6 text-slate-600">
												{formatNumber(region.issue_count)} issue · {formatNumber(region.report_count)} laporan ·
												{formatNumber(region.casualty_count)} korban
											</p>
										</div>
									</div>
								</article>
							{/each}
						</div>
					{/if}
				</section>
			</div>

			<div class="flex flex-col gap-5">
				<section class="jedug-card p-5 md:p-6">
					<div class="flex items-start gap-3">
						<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-sky-50 text-sky-600">
							<ClockIcon class="size-6" />
						</div>
						<div>
							<h2 class="text-xl font-[800] tracking-[-0.03em] text-slate-950">Time stats</h2>
							<p class="mt-2 text-sm leading-6 text-slate-500">
								Umur issue dan titik backlog paling tua pada scope aktif.
							</p>
						</div>
					</div>

					<div class="mt-5 grid gap-3">
						{#each timeCards as item}
							{@const ItemIcon = item.icon}
							<article class="metric-card">
								<div class="flex items-center gap-2 text-slate-500">
									<ItemIcon class="size-[18px]" />
									<span class="metric-label">{item.label}</span>
								</div>
								<strong class="mt-3 block text-xl font-[800] tracking-[-0.03em] text-slate-950">
									{item.value}
								</strong>
								<p class="mt-2 text-sm leading-6 text-slate-500">{item.copy}</p>
								{#if item.href}
									<a
										class="mt-4 inline-flex items-center gap-2 text-sm font-bold text-brand-600 transition hover:text-brand-700"
										href={item.href}
									>
										<RouteIcon class="size-[18px]" />
										Lihat issue tertua
									</a>
								{/if}
							</article>
						{/each}
					</div>
				</section>

				<section class="jedug-card p-5 md:p-6">
					<div class="flex items-start gap-3">
						<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-amber-50 text-amber-700">
							<ChartIcon class="size-6" />
						</div>
						<div>
							<h2 class="text-xl font-[800] tracking-[-0.03em] text-slate-950">Top issue</h2>
							<p class="mt-2 text-sm leading-6 text-slate-500">
								Kartu issue unggulan otomatis mengikuti provinsi dan kabupaten/kota yang aktif.
							</p>
						</div>
					</div>

					{#if stats.top_issues.length === 0}
						<div class="mt-5 rounded-[24px] border border-dashed border-slate-200 bg-slate-50 px-4 py-5">
							<EmptyState message="Belum ada issue unggulan untuk ditampilkan." />
						</div>
					{:else}
						<div class="mt-5 grid gap-3">
							{#each stats.top_issues as item (item.category)}
								<article class="rounded-[24px] border border-slate-200 bg-white px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
									<div class="flex items-start justify-between gap-3">
										<h3 class="text-sm font-bold leading-6 text-slate-950">{item.label}</h3>
										<span class="badge-tint">
											{formatNumber(item.metric_value)} {item.metric_label}
										</span>
									</div>
									<p class="mt-3 text-sm font-bold leading-6 text-slate-900">{getIssueName(item)}</p>
									<p class="mt-1 text-sm leading-6 text-slate-500">{getIssueLocation(item)}</p>
									<p class="mt-2 text-xs leading-5 text-slate-500">{getIssueContext(item)}</p>
									<a
										class="mt-4 inline-flex items-center gap-2 text-sm font-bold text-brand-600 transition hover:text-brand-700"
										href={`/issues/${item.issue_id}`}
									>
										<RouteIcon class="size-[18px]" />
										Lihat detail issue
									</a>
								</article>
							{/each}
						</div>
					{/if}
				</section>
			</div>
		</div>
	{:else}
		<ErrorState
			message="Data statistik publik tidak tersedia saat ini."
			onretry={() => initPage()}
		/>
	{/if}
</div>
