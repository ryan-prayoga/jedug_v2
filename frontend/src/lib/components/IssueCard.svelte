<script lang="ts">
	import { relativeTime } from '$lib/utils/date';
	import type { Issue } from '$lib/api/types';

	let { issue }: { issue: Issue } = $props();

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
</script>

<a href="/issues/{issue.id}" class="issue-card">
	<div class="card-body">
		<div class="card-header">
			<span
				class="severity-badge"
				style="background: {severityColor[issue.severity_current] || '#94A3B8'}"
			>
				{severityLabel[issue.severity_current] || `Level ${issue.severity_current}`}
			</span>
			<span class="status-badge" style="{statusStyle[issue.status] || statusStyle['open']}">
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
		border: 1px solid #E2E8F0;
		border-radius: 16px;
		padding: 16px;
		transition: box-shadow 0.15s;
	}
	.issue-card:hover {
		box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04);
	}
	.card-header {
		display: flex;
		align-items: center;
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
	.card-location {
		margin-bottom: 8px;
		font-size: 14px;
		color: #0F172A;
	}
	.coords {
		color: #64748B;
		font-size: 13px;
	}
	.road-type {
		color: #94A3B8;
		font-size: 13px;
	}
	.card-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		font-size: 12px;
		color: #64748B;
	}
	.casualty {
		color: #DC2626;
		font-weight: 500;
	}
</style>
