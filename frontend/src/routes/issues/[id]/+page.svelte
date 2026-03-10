<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getIssue } from '$lib/api/issues';
	import type { IssueDetail } from '$lib/api/types';
	import { formatDate, relativeTime } from '$lib/utils/date';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';

	let issue = $state<IssueDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const severityLabel = ['', 'Ringan', 'Sedang', 'Berat', 'Parah', 'Kritis'];
	const severityColor = ['', '#F6C453', '#F97316', '#DC2626', '#DC2626', '#991B1B'];
	const statusLabel: Record<string, string> = {
		open: 'Terbuka',
		fixed: 'Selesai',
		closed: 'Selesai',
		in_progress: 'Diproses'
	};
	const statusStyle: Record<string, string> = {
		open: 'background: #EFF6FF; color: #2563EB',
		fixed: 'background: #F1F5F9; color: #64748B',
		closed: 'background: #F1F5F9; color: #64748B',
		in_progress: 'background: #F0FDF4; color: #16A34A'
	};

	async function fetchIssue() {
		loading = true;
		error = null;
		try {
			const id = page.params.id;
			if (!id) throw new Error('ID tidak valid');
			const res = await getIssue(id);
			issue = res.data || null;
			if (!issue) throw new Error('Issue tidak ditemukan');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat data';
		} finally {
			loading = false;
		}
	}

	onMount(fetchIssue);

	let primaryMedia = $derived(issue?.media?.find((m) => m.is_primary) || issue?.media?.[0]);
	let otherMedia = $derived(issue?.media?.filter((m) => m !== primaryMedia) || []);
</script>

<div class="detail-page">
	{#if loading}
		<LoadingState message="Memuat detail..." />
	{:else if error}
		<ErrorState message={error} onretry={fetchIssue} />
	{:else if issue}
		<!-- Primary photo -->
		{#if primaryMedia}
			<div class="primary-photo">
				<img src={primaryMedia.public_url} alt="Foto jalan rusak" />
			</div>
		{/if}

		<!-- Header info -->
		<div class="detail-header">
			<div class="badges">
				<span
					class="severity-badge"
					style="background: {severityColor[issue.severity_current] || '#888'}"
				>
					{severityLabel[issue.severity_current] || `Level ${issue.severity_current}`}
				</span>
			<span class="status-badge" style="{statusStyle[issue.status] || statusStyle['open']}">
					{statusLabel[issue.status] || issue.status}
				</span>
			</div>

			{#if issue.road_name}
				<h1 class="road-name">{issue.road_name}</h1>
			{/if}
			{#if issue.road_type}
				<p class="road-type">{issue.road_type}</p>
			{/if}
		</div>

		<!-- Meta grid -->
		<div class="meta-grid">
			<div class="meta-item">
				<span class="meta-label">Severity Max</span>
				<span class="meta-value">{severityLabel[issue.severity_max] || issue.severity_max}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Laporan</span>
				<span class="meta-value">{issue.submission_count}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Korban</span>
				<span class="meta-value">{issue.casualty_count}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Foto</span>
				<span class="meta-value">{issue.photo_count}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Pertama</span>
				<span class="meta-value">{relativeTime(issue.first_seen_at)}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Terakhir</span>
				<span class="meta-value">{relativeTime(issue.last_seen_at)}</span>
			</div>
		</div>

		<!-- Location -->
		<div class="location-box">
			<span class="meta-label">📍 Koordinat</span>
			<span>{issue.latitude.toFixed(6)}, {issue.longitude.toFixed(6)}</span>
			{#if issue.region_id}
				<span class="region">Region ID: {issue.region_id}</span>
			{/if}
		</div>

		<!-- Gallery -->
		{#if otherMedia.length > 0}
			<div class="gallery-section">
				<h2>Foto Lainnya</h2>
				<div class="gallery">
					{#each otherMedia as media (media.id)}
						<img src={media.public_url} alt="Foto jalan rusak" class="gallery-img" />
					{/each}
				</div>
			</div>
		{/if}

		<!-- Recent submissions -->
		{#if issue.recent_submissions && issue.recent_submissions.length > 0}
			<div class="submissions-section">
				<h2>Laporan Terbaru</h2>
				{#each issue.recent_submissions as sub (sub.id)}
					<div class="submission-item">
						<div class="sub-header">
							<span
								class="severity-dot"
								style="background: {severityColor[sub.severity] || '#888'}"
							></span>
							<span class="sub-severity">{severityLabel[sub.severity] || `Level ${sub.severity}`}</span>
							<span class="sub-time">{relativeTime(sub.reported_at)}</span>
						</div>
						{#if sub.note}
							<p class="sub-note">{sub.note}</p>
						{/if}
					</div>
				{/each}
			</div>
		{/if}

		<!-- Actions -->
		<div class="actions">
			<a href="/lapor" class="btn btn-primary">Laporkan Lagi</a>
			<a href="/issues" class="btn btn-secondary">← Kembali ke Daftar</a>
		</div>
	{/if}
</div>

<style>
	.detail-page {
		padding-top: 8px;
		padding-bottom: 32px;
	}
	.primary-photo {
		margin: 0 -16px;
		margin-bottom: 16px;
	}
	.primary-photo img {
		width: 100%;
		max-height: 350px;
		object-fit: cover;
	}
	.detail-header {
		margin-bottom: 16px;
	}
	.badges {
		display: flex;
		gap: 8px;
		margin-bottom: 8px;
	}
	.severity-badge {
		font-size: 12px;
		font-weight: 600;
		color: #fff;
		padding: 4px 12px;
		border-radius: 999px;
		line-height: 1;
	}
	.status-badge {
		font-size: 12px;
		font-weight: 500;
		padding: 4px 12px;
		border-radius: 999px;
		line-height: 1;
	}
	.road-name {
		font-size: 20px;
		font-weight: 700;
		margin: 0;
		color: #0F172A;
	}
	.road-type {
		font-size: 13px;
		color: #94A3B8;
		margin: 2px 0 0;
		text-transform: capitalize;
	}

	/* Meta grid */
	.meta-grid {
		display: grid;
		grid-template-columns: 1fr 1fr 1fr;
		gap: 8px;
		margin-bottom: 16px;
	}
	.meta-item {
		background: #fff;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 12px;
		text-align: center;
	}
	.meta-label {
		display: block;
		font-size: 11px;
		color: #94A3B8;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin-bottom: 2px;
	}
	.meta-value {
		font-size: 14px;
		font-weight: 600;
		color: #0F172A;
	}

	/* Location */
	.location-box {
		background: #fff;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 12px;
		margin-bottom: 24px;
		font-size: 13px;
		color: #64748B;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.region {
		font-size: 12px;
		color: #94A3B8;
	}

	/* Gallery */
	.gallery-section {
		margin-bottom: 24px;
	}
	.gallery-section h2 {
		font-size: 16px;
		font-weight: 600;
		margin-bottom: 8px;
		color: #0F172A;
	}
	.gallery {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.gallery-img {
		width: 100%;
		height: 120px;
		object-fit: cover;
		border-radius: 12px;
	}

	/* Submissions */
	.submissions-section {
		margin-bottom: 24px;
	}
	.submissions-section h2 {
		font-size: 16px;
		font-weight: 600;
		margin-bottom: 8px;
		color: #0F172A;
	}
	.submission-item {
		background: #fff;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 12px;
		margin-bottom: 8px;
	}
	.sub-header {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
	}
	.severity-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.sub-severity {
		font-weight: 600;
	}
	.sub-time {
		color: #94A3B8;
		margin-left: auto;
		font-size: 12px;
	}
	.sub-note {
		font-size: 13px;
		color: #64748B;
		margin-top: 4px;
		line-height: 1.5;
	}

	/* Actions */
	.actions {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-top: 24px;
	}
	.btn {
		display: flex;
		align-items: center;
		justify-content: center;
		text-align: center;
		text-decoration: none;
		padding: 14px 20px;
		font-size: 16px;
		font-weight: 600;
		border-radius: 12px;
		min-height: 48px;
		transition: opacity 0.15s, transform 0.1s;
	}
	.btn:active {
		transform: scale(0.97);
	}
	.btn-primary {
		background: #E5484D;
		color: #fff;
	}
	.btn-primary:hover {
		opacity: 0.88;
	}
	.btn-secondary {
		background: #fff;
		color: #0F172A;
		border: 1px solid #E2E8F0;
	}
	.btn-secondary:hover {
		background: #F8FAFC;
	}
</style>
