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
	const severityColor = ['', '#38a169', '#d69e2e', '#dd6b20', '#e53e3e', '#9b2c2c'];
	const statusLabel: Record<string, string> = {
		open: 'Terbuka',
		fixed: 'Diperbaiki',
		archived: 'Diarsipkan'
	};

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
			<div class="sheet-handle"></div>

			<div class="sheet-content">
				<!-- Badges -->
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

				<!-- Location -->
				<div class="location">
					{#if issue.road_name}
						<strong>{issue.road_name}</strong>
					{:else}
						<span class="coords">{issue.latitude.toFixed(4)}, {issue.longitude.toFixed(4)}</span>
					{/if}
					{#if issue.road_type}
						<span class="road-type">· {issue.road_type}</span>
					{/if}
				</div>

				<!-- Stats -->
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
					<a href="/lapor" class="btn btn-secondary">📸 Lapor di Sini</a>
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

	/* Desktop: side panel style */
	@media (min-width: 768px) {
		.sheet-overlay {
			align-items: stretch;
			justify-content: flex-end;
			background: transparent;
			pointer-events: none;
		}
		.sheet {
			pointer-events: auto;
			width: 360px !important;
			max-height: 100% !important;
			border-radius: 0 !important;
			border-left: 1px solid #e2e8f0;
			animation: slideInRight 0.2s ease-out !important;
		}
	}

	.sheet {
		background: #fff;
		width: 100%;
		max-height: 60vh;
		border-radius: 16px 16px 0 0;
		box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.12);
		animation: slideUp 0.2s ease-out;
		overflow-y: auto;
	}

	@keyframes slideUp {
		from {
			transform: translateY(100%);
		}
		to {
			transform: translateY(0);
		}
	}

	@keyframes slideInRight {
		from {
			transform: translateX(100%);
		}
		to {
			transform: translateX(0);
		}
	}

	.sheet-handle {
		width: 36px;
		height: 4px;
		background: #cbd5e0;
		border-radius: 2px;
		margin: 10px auto 0;
	}

	@media (min-width: 768px) {
		.sheet-handle {
			display: none;
		}
	}

	.sheet-content {
		padding: 14px 18px 20px;
	}

	.badges {
		display: flex;
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

	.location {
		margin-bottom: 10px;
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

	.stats {
		display: flex;
		gap: 12px;
		margin-bottom: 14px;
		flex-wrap: wrap;
	}

	.stat {
		text-align: center;
		min-width: 50px;
	}

	.stat-value {
		display: block;
		font-size: 0.9rem;
		font-weight: 600;
		color: #2d3748;
	}

	.stat-label {
		display: block;
		font-size: 0.65rem;
		color: #a0aec0;
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.stat-danger .stat-value {
		color: #e53e3e;
	}

	.actions {
		display: flex;
		gap: 8px;
	}

	.btn {
		flex: 1;
		display: block;
		text-align: center;
		text-decoration: none;
		padding: 10px 14px;
		font-size: 0.85rem;
		font-weight: 600;
		border-radius: 10px;
		border: none;
		cursor: pointer;
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
