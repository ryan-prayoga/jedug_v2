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
	const severityColor = ['', '#38a169', '#d69e2e', '#dd6b20', '#e53e3e', '#9b2c2c'];
	const statusLabel: Record<string, string> = {
		open: 'Terbuka',
		closed: 'Selesai',
		in_progress: 'Diproses'
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
				<span class="status-badge">
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
			<a href="/lapor" class="btn btn-primary">📸 Laporkan Lagi</a>
			<a href="/issues" class="btn btn-secondary">← Kembali ke Daftar</a>
		</div>
	{/if}
</div>

<style>
	.detail-page {
		padding-top: 0.5rem;
		padding-bottom: 2rem;
	}
	.primary-photo {
		margin: 0 -16px;
		margin-bottom: 1rem;
	}
	.primary-photo img {
		width: 100%;
		max-height: 350px;
		object-fit: cover;
	}
	.detail-header {
		margin-bottom: 1rem;
	}
	.badges {
		display: flex;
		gap: 8px;
		margin-bottom: 8px;
	}
	.severity-badge {
		font-size: 0.8rem;
		font-weight: 600;
		color: #fff;
		padding: 3px 12px;
		border-radius: 999px;
	}
	.status-badge {
		font-size: 0.8rem;
		color: #718096;
		background: #edf2f7;
		padding: 3px 12px;
		border-radius: 999px;
	}
	.road-name {
		font-size: 1.3rem;
		font-weight: 700;
		margin: 0;
	}
	.road-type {
		font-size: 0.85rem;
		color: #a0aec0;
		margin: 2px 0 0;
		text-transform: capitalize;
	}

	/* Meta grid */
	.meta-grid {
		display: grid;
		grid-template-columns: 1fr 1fr 1fr;
		gap: 8px;
		margin-bottom: 1rem;
	}
	.meta-item {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 10px;
		padding: 10px;
		text-align: center;
	}
	.meta-label {
		display: block;
		font-size: 0.7rem;
		color: #a0aec0;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin-bottom: 2px;
	}
	.meta-value {
		font-size: 0.9rem;
		font-weight: 600;
		color: #2d3748;
	}

	/* Location */
	.location-box {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 10px;
		padding: 12px;
		margin-bottom: 1.5rem;
		font-size: 0.85rem;
		color: #4a5568;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.region {
		font-size: 0.75rem;
		color: #a0aec0;
	}

	/* Gallery */
	.gallery-section {
		margin-bottom: 1.5rem;
	}
	.gallery-section h2 {
		font-size: 1rem;
		margin-bottom: 8px;
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
		border-radius: 8px;
	}

	/* Submissions */
	.submissions-section {
		margin-bottom: 1.5rem;
	}
	.submissions-section h2 {
		font-size: 1rem;
		margin-bottom: 8px;
	}
	.submission-item {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 10px;
		padding: 10px 12px;
		margin-bottom: 6px;
	}
	.sub-header {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 0.85rem;
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
		color: #a0aec0;
		margin-left: auto;
		font-size: 0.75rem;
	}
	.sub-note {
		font-size: 0.85rem;
		color: #4a5568;
		margin-top: 4px;
		line-height: 1.4;
	}

	/* Actions */
	.actions {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-top: 1.5rem;
	}
	.btn {
		display: block;
		text-align: center;
		text-decoration: none;
		padding: 14px 20px;
		font-size: 1rem;
		font-weight: 600;
		border-radius: 12px;
	}
	.btn-primary {
		background: #e53e3e;
		color: #fff;
	}
	.btn-primary:hover {
		opacity: 0.9;
	}
	.btn-secondary {
		background: #fff;
		color: #4a5568;
		border: 1px solid #e2e8f0;
	}
	.btn-secondary:hover {
		background: #f7fafc;
	}
</style>
