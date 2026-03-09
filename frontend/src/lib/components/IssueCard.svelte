<script lang="ts">
	import { relativeTime } from '$lib/utils/date';
	import type { Issue } from '$lib/api/types';

	let { issue }: { issue: Issue } = $props();

	const severityLabel = ['', 'Ringan', 'Sedang', 'Berat', 'Parah', 'Kritis'];
	const severityColor = ['', '#38a169', '#d69e2e', '#dd6b20', '#e53e3e', '#9b2c2c'];
	const statusLabel: Record<string, string> = {
		open: 'Terbuka',
		closed: 'Selesai',
		in_progress: 'Diproses'
	};
</script>

<a href="/issues/{issue.id}" class="issue-card">
	<div class="card-body">
		<div class="card-header">
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

		<div class="card-location">
			{#if issue.road_name}
				<strong>{issue.road_name}</strong>
			{:else}
				<span class="coords">{issue.latitude.toFixed(4)}, {issue.longitude.toFixed(4)}</span>
			{/if}
			{#if issue.road_type}
				<span class="road-type">· {issue.road_type}</span>
			{/if}
		</div>

		<div class="card-meta">
			<span>📸 {issue.submission_count} laporan</span>
			{#if issue.casualty_count > 0}
				<span class="casualty">🚑 {issue.casualty_count} korban</span>
			{/if}
			<span>· {relativeTime(issue.last_seen_at)}</span>
		</div>
	</div>
</a>

<style>
	.issue-card {
		display: block;
		text-decoration: none;
		color: inherit;
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 12px;
		padding: 14px 16px;
		transition: box-shadow 0.15s;
	}
	.issue-card:hover {
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
	}
	.card-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 8px;
	}
	.severity-badge {
		font-size: 0.75rem;
		font-weight: 600;
		color: #fff;
		padding: 2px 10px;
		border-radius: 999px;
	}
	.status-badge {
		font-size: 0.75rem;
		color: #718096;
		background: #edf2f7;
		padding: 2px 10px;
		border-radius: 999px;
	}
	.card-location {
		margin-bottom: 6px;
		font-size: 0.95rem;
	}
	.coords {
		color: #718096;
		font-size: 0.85rem;
	}
	.road-type {
		color: #a0aec0;
		font-size: 0.85rem;
	}
	.card-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		font-size: 0.8rem;
		color: #718096;
	}
	.casualty {
		color: #e53e3e;
		font-weight: 500;
	}
</style>
