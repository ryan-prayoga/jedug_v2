<script lang="ts">
	import { onMount } from 'svelte';
	import { getPublicStats } from '$lib/api/stats';
	import type { PublicStats, PublicTopIssue } from '$lib/api/types';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import { formatDate } from '$lib/utils/date';

	let stats = $state<PublicStats | null>(null);
	let loading = $state(true);
	let errorMessage = $state<string | null>(null);

	const isEmpty = $derived.by(() => {
		if (!stats) return false;
		return stats.global.total_issues === 0;
	});

	const statusTotal = $derived.by(() => {
		if (!stats) return 0;
		return stats.status.open + stats.status.fixed + stats.status.archived;
	});

	onMount(() => {
		void fetchStats();
	});

	async function fetchStats() {
		loading = true;
		errorMessage = null;

		try {
			const result = await getPublicStats();
			if (!result.data) {
				stats = null;
				errorMessage = 'Data statistik publik tidak tersedia saat ini.';
				return;
			}

			stats = result.data;
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Gagal memuat statistik publik.';
		} finally {
			loading = false;
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

	function getIssueName(item: PublicTopIssue): string {
		if (item.road_name && item.road_name.trim() !== '') return item.road_name;
		if (item.region_name && item.region_name.trim() !== '') return item.region_name;
		return 'Issue tanpa nama jalan';
	}

	function getIssueContext(item: PublicTopIssue): string {
		const region = item.region_name || 'Wilayah tidak diketahui';
		return `${region} · ${formatNumber(item.submission_count)} laporan · ${formatNumber(item.casualty_count)} korban`;
	}

	function getGeneratedLabel(statsData: PublicStats | null): string | null {
		if (!statsData?.generated_at) return null;
		return formatDate(statsData.generated_at);
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
			<button class="refresh-btn" onclick={fetchStats} disabled={loading}>
				{loading ? 'Memuat...' : 'Muat Ulang'}
			</button>
		</div>
	</header>

	{#if loading}
		<LoadingState message="Memuat statistik publik..." />
	{:else if errorMessage}
		<ErrorState message={errorMessage} onretry={fetchStats} />
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
				<h2>Region Leaderboard</h2>
			</div>
			{#if stats.regions.length === 0}
				<p class="section-empty">Belum ada data wilayah.</p>
			{:else}
				<div class="leaderboard-list">
					{#each stats.regions as region, index (region.region_name)}
						<article class="leaderboard-item">
							<div class="leaderboard-rank">{index + 1}</div>
							<div class="leaderboard-body">
								<h3>{region.region_name}</h3>
								<p>
									{formatNumber(region.issue_count)} issue · {formatNumber(region.casualty_count)} korban ·
									{formatNumber(region.report_count)} laporan
								</p>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</section>

		<section class="section">
			<div class="section-head">
				<h2>Top Issue</h2>
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
							<p class="top-issue-meta">{getIssueContext(item)}</p>
							<a class="detail-link" href={`/issues/${item.issue_id}`}>Lihat detail issue</a>
						</article>
					{/each}
				</div>
			{/if}
		</section>
	{:else}
		<ErrorState message="Data statistik publik tidak tersedia saat ini." onretry={fetchStats} />
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
	}

	@media (min-width: 960px) {
		.stats-page {
			padding-top: 20px;
		}
	}
</style>
