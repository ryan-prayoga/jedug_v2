<script lang="ts">
	import { onMount } from 'svelte';
	import { getPublicStats } from '$lib/api/stats';
	import { resolveLocationLabel } from '$lib/api/location';
	import type {
		LocationLabelData,
		PublicStats,
		PublicStatsRegionOption,
		PublicTopIssue
	} from '$lib/api/types';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
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
	let selectedProvinceID = $state('');
	let selectedRegencyID = $state('');
	let locationHint = $state('Wilayah awal akan dicoba menyesuaikan lokasi kamu, lalu tetap bisa diganti manual.');
	let triedLocationDefault = false;

	const isEmpty = $derived.by(() => {
		if (!stats) return false;
		return stats.global.total_issues === 0;
	});

	const statusTotal = $derived.by(() => {
		if (!stats) return 0;
		return stats.status.open + stats.status.fixed + stats.status.archived;
	});

	const activeScopeLabel = $derived(stats?.filters.scope_label || 'Pilih wilayah');

	onMount(() => {
		void initPage();
	});

	async function initPage() {
		await fetchStats({}, { preserveData: false });
		await applyLocationDefault();
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

	async function applyLocationDefault() {
		if (triedLocationDefault || !stats) return;
		triedLocationDefault = true;

		try {
			const point = await getLocation();
			const labelResult = await resolveLocationLabel(point.latitude, point.longitude);
			const label = labelResult.data;
			if (!label) {
				locationHint = 'Lokasi belum bisa dipetakan, jadi statistik memakai wilayah default yang masih bisa kamu ganti manual.';
				return;
			}

			const province = findProvinceOption(label, stats.filters.province_options);
			if (!province) {
				locationHint =
					'Lokasi ditemukan, tetapi provinsinya belum cocok dengan data statistik saat ini. Kamu masih bisa pilih manual.';
				return;
			}

			if (stats.filters.active_province_id !== province.id) {
				await fetchStats({ provinceID: province.id });
			}

			const currentStats = stats;
			if (!currentStats) return;

			const regency = findRegencyOption(label, currentStats.filters.regency_options);
			if (regency) {
				await fetchStats({ provinceID: province.id, regencyID: regency.id });
			}

			locationHint = 'Wilayah awal disesuaikan dengan lokasi kamu saat ini.';
		} catch {
			locationHint =
				'Lokasi tidak tersedia, jadi statistik memakai wilayah default. Kamu tetap bisa ganti provinsi dan kabupaten/kota secara manual.';
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
			.trim()
			.toLocaleLowerCase('id-ID')
			.replace(/\s+/g, ' ')
			.replace(/^(provinsi|kabupaten|kota|kecamatan)\s+/u, '');
	}

	function findOptionByCandidates(
		options: PublicStatsRegionOption[],
		candidates: Array<string | null | undefined>
	): PublicStatsRegionOption | null {
		const normalizedCandidates = candidates
			.map((candidate) => normalizeRegionToken(candidate))
			.filter((candidate) => candidate !== '');

		if (normalizedCandidates.length === 0) return null;

		return (
			options.find((option) => normalizedCandidates.includes(normalizeRegionToken(option.name))) || null
		);
	}

	function findProvinceOption(
		label: LocationLabelData,
		options: PublicStatsRegionOption[]
	): PublicStatsRegionOption | null {
		const level = normalizeRegionToken(label.region_level);

		if (level === 'province') {
			return findOptionByCandidates(options, [label.region_name]);
		}

		if (level === 'city' || level === 'regency') {
			return findOptionByCandidates(options, [label.parent_name, label.grandparent_name]);
		}

		return findOptionByCandidates(options, [label.grandparent_name, label.parent_name]);
	}

	function findRegencyOption(
		label: LocationLabelData,
		options: PublicStatsRegionOption[]
	): PublicStatsRegionOption | null {
		const level = normalizeRegionToken(label.region_level);

		if (level === 'city' || level === 'regency') {
			return findOptionByCandidates(options, [label.region_name]);
		}

		return findOptionByCandidates(options, [label.parent_name]);
	}

	function getIssueName(item: PublicTopIssue): string {
		if (item.road_name && item.road_name.trim() !== '') return item.road_name;
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

	function getGeneratedLabel(statsData: PublicStats | null): string | null {
		if (!statsData?.generated_at) return null;
		return formatDate(statsData.generated_at);
	}

	async function handleProvinceChange(event: Event) {
		const value = Number((event.currentTarget as HTMLSelectElement).value || 0);
		if (!value) return;
		await fetchStats({ provinceID: value });
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

	<div class="stats-page">
		<header class="stats-hero">
		<p class="hero-kicker">Statistik Publik</p>
		<h1>Dashboard Jalan Rusak</h1>
		<p class="hero-description">
			Ringkasan agregasi issue publik untuk civic storytelling yang cepat dipahami.
		</p>
			<div class="hero-meta">
				{#if getGeneratedLabel(stats)}
					<span>Update terakhir: {getGeneratedLabel(stats)}</span>
				{/if}
				<button
					class="refresh-btn"
					onclick={() => fetchStats(getCurrentScope())}
					disabled={loading || refreshing}
				>
					{loading || refreshing ? 'Memuat...' : 'Muat Ulang'}
				</button>
			</div>
		</header>

		{#if loading}
			<LoadingState message="Memuat statistik publik..." />
	{:else if pageErrorMessage}
		<ErrorState message={pageErrorMessage} onretry={() => fetchStats(getCurrentScope())} />
	{:else if stats && isEmpty}
		<EmptyState
			icon="📊"
			message="Belum ada statistik publik yang bisa ditampilkan."
			ctaHref="/lapor"
			ctaLabel="Kirim Laporan Pertama"
		/>
	{:else if stats}
		<section class="section">
			<div class="section-head">
				<div>
					<h2>Filter Wilayah</h2>
					<p class="section-copy">Leaderboard dan top issue mengikuti provinsi + kabupaten/kota yang aktif.</p>
				</div>
				<span class="scope-pill">{activeScopeLabel}</span>
			</div>

			<div class="region-filter-grid">
				<label class="filter-field">
					<span>Provinsi</span>
					<select
						value={selectedProvinceID}
						disabled={refreshing || stats.filters.province_options.length === 0}
						onchange={handleProvinceChange}
					>
						{#each stats.filters.province_options as province (province.id)}
							<option value={province.id}>{province.name}</option>
						{/each}
					</select>
				</label>

				<label class="filter-field">
					<span>Kabupaten/Kota</span>
					<select
						value={selectedRegencyID}
						disabled={refreshing || stats.filters.regency_options.length === 0}
						onchange={handleRegencyChange}
					>
						{#each stats.filters.regency_options as regency (regency.id)}
							<option value={regency.id}>{regency.name}</option>
						{/each}
					</select>
				</label>
			</div>

			<p class="filter-hint">{locationHint}</p>
			{#if inlineErrorMessage}
				<p class="filter-error">{inlineErrorMessage}</p>
			{/if}
		</section>

		<section class="section">
			<div class="section-head">
				<h2>Global Stats</h2>
			</div>
			<div class="stats-grid">
				<article class="stat-card">
					<span class="stat-label">Total Issue</span>
					<strong class="stat-value">{formatNumber(stats.global.total_issues)}</strong>
				</article>
				<article class="stat-card">
					<span class="stat-label">Issue Minggu Ini</span>
					<strong class="stat-value">{formatNumber(stats.global.total_issues_this_week)}</strong>
				</article>
				<article class="stat-card">
					<span class="stat-label">Total Korban</span>
					<strong class="stat-value">{formatNumber(stats.global.total_casualties)}</strong>
				</article>
				<article class="stat-card">
					<span class="stat-label">Total Foto Laporan</span>
					<strong class="stat-value">{formatNumber(stats.global.total_photos)}</strong>
				</article>
				<article class="stat-card">
					<span class="stat-label">Total Laporan</span>
					<strong class="stat-value">{formatNumber(stats.global.total_reports)}</strong>
				</article>
			</div>
		</section>

		<section class="section">
			<div class="section-head">
				<h2>Status Breakdown</h2>
			</div>
			<div class="status-list">
				<article class="status-card">
					<div class="status-row">
						<div>
							<p class="status-name">Issue Open</p>
							<strong>{formatNumber(stats.status.open)}</strong>
						</div>
						<span>{getStatusPercent(stats.status.open)}%</span>
					</div>
					<div class="status-bar">
						<div class="status-fill open" style={`width:${getStatusPercent(stats.status.open)}%`}></div>
					</div>
				</article>
				<article class="status-card">
					<div class="status-row">
						<div>
							<p class="status-name">Issue Fixed</p>
							<strong>{formatNumber(stats.status.fixed)}</strong>
						</div>
						<span>{getStatusPercent(stats.status.fixed)}%</span>
					</div>
					<div class="status-bar">
						<div class="status-fill fixed" style={`width:${getStatusPercent(stats.status.fixed)}%`}></div>
					</div>
				</article>
				<article class="status-card">
					<div class="status-row">
						<div>
							<p class="status-name">Issue Archived</p>
							<strong>{formatNumber(stats.status.archived)}</strong>
						</div>
						<span>{getStatusPercent(stats.status.archived)}%</span>
					</div>
					<div class="status-bar">
						<div
							class="status-fill archived"
							style={`width:${getStatusPercent(stats.status.archived)}%`}
						></div>
					</div>
				</article>
			</div>
		</section>

		<section class="section">
			<div class="section-head">
				<h2>Time Stats</h2>
			</div>
			<div class="time-grid">
				<article class="stat-card">
					<span class="stat-label">Rata-rata Umur Issue</span>
					<strong class="stat-value">{formatDecimal(stats.time.average_issue_age_days)} hari</strong>
				</article>
				<article class="stat-card">
					<span class="stat-label">Issue Tertua yang Masih Open</span>
					<strong class="stat-value">{formatNumber(stats.time.oldest_open_issue_age_days)} hari</strong>
					{#if stats.time.oldest_open_first_seen_at}
						<p class="meta-text">Pertama tercatat {formatDate(stats.time.oldest_open_first_seen_at)}</p>
					{/if}
					{#if stats.time.oldest_open_issue_id}
						<a class="detail-link" href={`/issues/${stats.time.oldest_open_issue_id}`}>Lihat issue tertua</a>
					{/if}
				</article>
			</div>
		</section>

		<section class="section">
			<div class="section-head">
				<div>
					<h2>Region Leaderboard</h2>
					<p class="section-copy">Kecamatan dengan laporan terbanyak di wilayah administratif yang sedang dipilih.</p>
				</div>
			</div>
			{#if stats.regions.length === 0}
				<p class="section-empty">Belum ada data wilayah.</p>
			{:else}
				<div class="leaderboard-list">
					{#each stats.regions as region, index (region.district_name)}
						<article class="leaderboard-item">
							<div class="leaderboard-rank">{index + 1}</div>
							<div class="leaderboard-body">
								<h3>{region.district_name}</h3>
								<p>
									{formatNumber(region.issue_count)} issue · {formatNumber(region.report_count)} laporan ·
									{formatNumber(region.casualty_count)} korban
								</p>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</section>

		<section class="section">
			<div class="section-head">
				<div>
					<h2>Top Issue</h2>
					<p class="section-copy">Kartu issue ini otomatis mengikuti provinsi dan kabupaten/kota yang aktif.</p>
				</div>
			</div>
			{#if stats.top_issues.length === 0}
				<p class="section-empty">Belum ada issue unggulan untuk ditampilkan.</p>
			{:else}
				<div class="top-issue-list">
					{#each stats.top_issues as item (item.category)}
						<article class="top-issue-card">
								<div class="top-issue-head">
									<h3>{item.label}</h3>
									<span class="metric-pill">{formatNumber(item.metric_value)} {item.metric_label}</span>
								</div>
								<p class="top-issue-name">{getIssueName(item)}</p>
								<p class="top-issue-location">{getIssueLocation(item)}</p>
								<p class="top-issue-meta">{getIssueContext(item)}</p>
								<a class="detail-link" href={`/issues/${item.issue_id}`}>Lihat detail issue</a>
							</article>
					{/each}
				</div>
			{/if}
		</section>
	{:else}
		<ErrorState
			message="Data statistik publik tidak tersedia saat ini."
			onretry={() => fetchStats(getCurrentScope())}
		/>
	{/if}
</div>

<style>
	.stats-page {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 16px 0 24px;
	}

	.stats-hero {
		background: linear-gradient(180deg, #FFFFFF 0%, #F8FAFC 100%);
		border: 1px solid #E2E8F0;
		border-radius: 16px;
		padding: 16px;
		box-shadow: 0 1px 3px rgba(0,0,0,0.04);
	}

	.hero-kicker {
		display: inline-flex;
		align-items: center;
		padding: 4px 10px;
		border-radius: 999px;
		border: 1px solid #FECACA;
		background: #FEF2F2;
		color: #E5484D;
		font-size: 11px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.6px;
	}

	.stats-hero h1 {
		font-size: 24px;
		line-height: 1.15;
		margin: 12px 0 0;
		color: #0F172A;
	}

	.hero-description {
		font-size: 14px;
		line-height: 1.55;
		color: #64748B;
		margin: 10px 0 0;
	}

	.hero-meta {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-top: 14px;
		font-size: 12px;
		color: #64748B;
	}

	.refresh-btn {
		flex-shrink: 0;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		background: #FFFFFF;
		color: #0F172A;
		font-size: 12px;
		font-weight: 600;
		padding: 8px 12px;
		cursor: pointer;
	}

	.refresh-btn:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}

	.section {
		background: #FFFFFF;
		border: 1px solid #E2E8F0;
		border-radius: 16px;
		padding: 14px;
		box-shadow: 0 1px 3px rgba(0,0,0,0.04);
	}

	.section-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 10px;
	}

	.section-head h2 {
		font-size: 15px;
		font-weight: 700;
		color: #0F172A;
	}

	.section-copy {
		font-size: 12px;
		color: #64748B;
		line-height: 1.5;
		margin-top: 3px;
	}

	.scope-pill {
		flex-shrink: 0;
		display: inline-flex;
		align-items: center;
		padding: 6px 10px;
		border-radius: 999px;
		background: #FEF2F2;
		color: #E5484D;
		font-size: 11px;
		font-weight: 700;
	}

	.region-filter-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 10px;
	}

	.filter-field {
		display: grid;
		gap: 6px;
	}

	.filter-field span {
		font-size: 12px;
		font-weight: 600;
		color: #475569;
	}

	.filter-field select {
		width: 100%;
		min-height: 44px;
		border-radius: 12px;
		border: 1px solid #CBD5E1;
		background: #FFFFFF;
		padding: 0 12px;
		color: #0F172A;
		font-size: 14px;
	}

	.filter-field select:disabled {
		opacity: 0.6;
		background: #F8FAFC;
	}

	.filter-hint {
		margin-top: 10px;
		font-size: 12px;
		color: #64748B;
		line-height: 1.5;
	}

	.filter-error {
		margin-top: 8px;
		font-size: 12px;
		color: #B42318;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}

	.time-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 10px;
	}

	.stat-card {
		border: 1px solid #E2E8F0;
		background: #F8FAFC;
		border-radius: 12px;
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.stat-label {
		font-size: 12px;
		color: #64748B;
	}

	.stat-value {
		font-size: 20px;
		line-height: 1.15;
		color: #0F172A;
		letter-spacing: -0.3px;
	}

	.meta-text {
		font-size: 12px;
		color: #64748B;
		line-height: 1.45;
	}

	.status-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.status-card {
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 10px;
		background: #FFFFFF;
	}

	.status-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.status-name {
		font-size: 12px;
		color: #64748B;
		margin-bottom: 2px;
	}

	.status-row strong {
		font-size: 16px;
		color: #0F172A;
	}

	.status-row span {
		font-size: 12px;
		font-weight: 600;
		color: #64748B;
	}

	.status-bar {
		margin-top: 8px;
		height: 7px;
		background: #F1F5F9;
		border-radius: 999px;
		overflow: hidden;
	}

	.status-fill {
		height: 100%;
		border-radius: 999px;
	}

	.status-fill.open {
		background: #2563EB;
	}

	.status-fill.fixed {
		background: #16A34A;
	}

	.status-fill.archived {
		background: #64748B;
	}

	.leaderboard-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.leaderboard-item {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 10px 12px;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		background: #FFFFFF;
	}

	.leaderboard-rank {
		min-width: 26px;
		height: 26px;
		border-radius: 999px;
		background: #FEF2F2;
		color: #E5484D;
		font-size: 12px;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.leaderboard-body h3 {
		font-size: 14px;
		line-height: 1.4;
		color: #0F172A;
	}

	.leaderboard-body p {
		font-size: 12px;
		color: #64748B;
		margin-top: 3px;
		line-height: 1.5;
	}

	.top-issue-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.top-issue-card {
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 12px;
		background: #FFFFFF;
	}

	.top-issue-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 8px;
	}

	.top-issue-head h3 {
		font-size: 13px;
		line-height: 1.4;
		color: #0F172A;
	}

	.metric-pill {
		flex-shrink: 0;
		display: inline-flex;
		align-items: center;
		padding: 5px 8px;
		border-radius: 999px;
		background: #FEF2F2;
		color: #E5484D;
		font-size: 11px;
		font-weight: 700;
	}

	.top-issue-name {
		font-size: 14px;
		font-weight: 700;
		line-height: 1.45;
		color: #0F172A;
		margin-top: 8px;
	}

	.top-issue-meta {
		font-size: 12px;
		color: #64748B;
		line-height: 1.5;
		margin-top: 3px;
	}

	.top-issue-location {
		font-size: 12px;
		color: #334155;
		line-height: 1.5;
		margin-top: 4px;
	}

	.section-empty {
		font-size: 13px;
		color: #64748B;
	}

	.detail-link {
		display: inline-flex;
		margin-top: 8px;
		font-size: 12px;
		font-weight: 700;
		color: #E5484D;
		text-decoration: none;
	}

	.detail-link:hover {
		text-decoration: underline;
	}

	@media (min-width: 640px) {
		.stats-grid {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}

		.time-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.region-filter-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (min-width: 960px) {
		.stats-page {
			padding-top: 20px;
		}
	}
</style>
