<script lang="ts">
	import type { Issue } from '$lib/api/types';
	import { relativeTime } from '$lib/utils/date';

	let {
		issue,
		onclose,
		visible = false
	}: {
		issue: Issue | null;
		onclose: () => void;
		visible: boolean;
	} = $props();

	const severityLabel = ['', 'Ringan', 'Sedang', 'Berat', 'Parah', 'Kritis'];
	const severityColor = ['', '#F6C453', '#F97316', '#DC2626', '#DC2626', '#991B1B'];
	const statusLabel: Record<string, string> = {
		open: 'Terbuka',
		fixed: 'Selesai',
		archived: 'Diarsipkan'
	};
	const statusColor: Record<string, { bg: string; text: string }> = {
		open: { bg: '#EFF6FF', text: '#2563EB' },
		fixed: { bg: '#F1F5F9', text: '#64748B' },
		archived: { bg: '#F1F5F9', text: '#64748B' }
	};

	function getStatusStyle(status: string) {
		const sc = statusColor[status] || statusColor['open'];
		return `background: ${sc.bg}; color: ${sc.text}`;
	}

	function handleOverlayClick() {
		onclose();
	}

	function handleSheetClick(e: MouseEvent) {
		e.stopPropagation();
	}
</script>

{#if visible && issue}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="sheet-overlay" onclick={handleOverlayClick}>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="sheet" onclick={handleSheetClick}>
			<div class="sheet-handle-area">
				<div class="sheet-handle"></div>
			</div>

			<div class="sheet-content">
				<!-- Severity as dominant element -->
				<div class="sheet-top-row">
					<span
						class="severity-badge"
						style="background: {severityColor[issue.severity_current] || '#94A3B8'}"
					>
						{severityLabel[issue.severity_current] || `Level ${issue.severity_current}`}
					</span>
					<span
						class="status-badge"
						style={getStatusStyle(issue.status)}
					>
						{statusLabel[issue.status] || issue.status}
					</span>
				</div>

				<!-- Location -->
				<div class="location">
					{#if issue.road_name}
						<span class="road-name">{issue.road_name}</span>
					{:else}
						<span class="coords">{issue.latitude.toFixed(4)}, {issue.longitude.toFixed(4)}</span>
					{/if}
					{#if issue.road_type}
						<span class="road-type">· {issue.road_type}</span>
					{/if}
				</div>

				<!-- Stats Grid -->
				<div class="stats">
					<div class="stat">
						<span class="stat-value">{issue.submission_count}</span>
						<span class="stat-label">Laporan</span>
					</div>
					<div class="stat">
						<span class="stat-value">{issue.photo_count}</span>
						<span class="stat-label">Foto</span>
					</div>
					{#if issue.casualty_count > 0}
						<div class="stat stat-danger">
							<span class="stat-value">{issue.casualty_count}</span>
							<span class="stat-label">Korban</span>
						</div>
					{/if}
					<div class="stat">
						<span class="stat-value">{relativeTime(issue.last_seen_at)}</span>
						<span class="stat-label">Terakhir</span>
					</div>
				</div>

				<!-- Actions -->
				<div class="actions">
					<a href="/issues/{issue.id}" class="btn btn-primary">Lihat Detail</a>
					<a href="/lapor" class="btn btn-secondary">Lapor di Sini</a>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.sheet-overlay {
		position: absolute;
		inset: 0;
		z-index: 20;
		display: flex;
		align-items: flex-end;
		justify-content: center;
	}

	@media (min-width: 768px) {
		.sheet-overlay {
			align-items: stretch;
			justify-content: flex-end;
			background: transparent;
			pointer-events: none;
		}
		.sheet {
			pointer-events: auto;
			width: 380px !important;
			max-height: 100% !important;
			border-radius: 0 !important;
			border-left: 1px solid #E2E8F0;
			animation: slideInRight 0.2s ease-out !important;
		}
		.sheet-handle-area {
			display: none;
		}
	}

	.sheet {
		background: #fff;
		width: 100%;
		max-height: 55vh;
		border-radius: 20px 20px 0 0;
		box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.10);
		animation: slideUp 0.2s ease-out;
		overflow-y: auto;
	}

	@keyframes slideUp {
		from { transform: translateY(100%); }
		to { transform: translateY(0); }
	}

	@keyframes slideInRight {
		from { transform: translateX(100%); }
		to { transform: translateX(0); }
	}

	.sheet-handle-area {
		padding: 12px 0 4px;
		display: flex;
		justify-content: center;
	}

	.sheet-handle {
		width: 40px;
		height: 4px;
		background: #CBD5E1;
		border-radius: 2px;
	}

	.sheet-content {
		padding: 12px 20px 24px;
	}

	.sheet-top-row {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 12px;
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

	.location {
		margin-bottom: 16px;
	}

	.road-name {
		font-size: 15px;
		font-weight: 600;
		color: #0F172A;
	}

	.coords {
		color: #64748B;
		font-size: 13px;
		font-family: 'SF Mono', 'Fira Code', monospace;
	}

	.road-type {
		color: #94A3B8;
		font-size: 13px;
		margin-left: 4px;
	}

	.stats {
		display: flex;
		gap: 16px;
		margin-bottom: 16px;
		flex-wrap: wrap;
	}

	.stat {
		text-align: center;
		min-width: 48px;
	}

	.stat-value {
		display: block;
		font-size: 14px;
		font-weight: 600;
		color: #0F172A;
	}

	.stat-label {
		display: block;
		font-size: 11px;
		color: #94A3B8;
		text-transform: uppercase;
		letter-spacing: 0.3px;
		margin-top: 2px;
	}

	.stat-danger .stat-value {
		color: #DC2626;
	}

	.actions {
		display: flex;
		gap: 8px;
	}

	.btn {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		text-align: center;
		text-decoration: none;
		padding: 12px 16px;
		font-size: 14px;
		font-weight: 600;
		border-radius: 12px;
		border: none;
		cursor: pointer;
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
