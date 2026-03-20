<script lang="ts">
	import type { Issue } from '$lib/api/types';
	import { relativeTime } from '$lib/utils/date';
	import { getIssueRoadOrAreaLabel } from '$lib/utils/issue-detail';

	let {
		issue,
		onclose,
		visible = false
	}: {
		issue: Issue | null;
		onclose: () => void;
		visible: boolean;
	} = $props();

	let sheetEl = $state<HTMLDivElement | null>(null);
	let dragOffsetY = $state(0);
	let isDragging = $state(false);
	let dragPointerId = $state<number | null>(null);
	let dragStartY = $state(0);
	let dragStartAt = $state(0);
	let ignoreNextOverlayClick = $state(false);

	const DRAG_CLOSE_THRESHOLD_PX = 96;
	const DRAG_CLOSE_VELOCITY = 0.7; // px/ms

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
		if (ignoreNextOverlayClick) {
			ignoreNextOverlayClick = false;
			return;
		}
		onclose();
	}

	function handleSheetClick(e: MouseEvent) {
		e.stopPropagation();
	}

	function isMobileViewport(): boolean {
		return window.matchMedia('(max-width: 767px)').matches;
	}

	function resetDragState() {
		dragOffsetY = 0;
		isDragging = false;
		dragPointerId = null;
		dragStartY = 0;
		dragStartAt = 0;
	}

	function canStartDrag(target: EventTarget | null): boolean {
		if (!target || !sheetEl) return false;
		if (!isMobileViewport()) return false;

		const element = target as HTMLElement;
		const insideHandle = element.closest('.sheet-handle-area');
		if (insideHandle) return true;

		// Allow dragging from content only when sheet is at top to avoid fighting internal scroll.
		if (sheetEl.scrollTop > 0) return false;
		if (element.closest('a, button, input, textarea, select, label')) return false;
		return true;
	}

	function handlePointerDown(e: PointerEvent) {
		if (!visible || !canStartDrag(e.target)) return;
		if (e.pointerType === 'mouse' && e.button !== 0) return;

		dragPointerId = e.pointerId;
		dragStartY = e.clientY;
		dragStartAt = performance.now();
		isDragging = true;
		ignoreNextOverlayClick = false;
		sheetEl?.setPointerCapture(e.pointerId);
	}

	function handlePointerMove(e: PointerEvent) {
		if (!isDragging || dragPointerId !== e.pointerId) return;
		const deltaY = Math.max(0, e.clientY - dragStartY);
		if (deltaY > 8) {
			ignoreNextOverlayClick = true;
		}
		dragOffsetY = deltaY;

		if (e.cancelable) {
			e.preventDefault();
		}
	}

	function finishDrag(shouldClose: boolean) {
		if (shouldClose) {
			resetDragState();
			onclose();
			return;
		}

		isDragging = false;
		dragOffsetY = 0;
		dragPointerId = null;
	}

	function handlePointerUp(e: PointerEvent) {
		if (!isDragging || dragPointerId !== e.pointerId) return;

		const deltaY = Math.max(0, e.clientY - dragStartY);
		const durationMs = Math.max(1, performance.now() - dragStartAt);
		const velocity = deltaY / durationMs;

		const shouldClose =
			deltaY >= DRAG_CLOSE_THRESHOLD_PX ||
			(deltaY >= 32 && velocity >= DRAG_CLOSE_VELOCITY);

		finishDrag(shouldClose);

		if (ignoreNextOverlayClick) {
			setTimeout(() => {
				ignoreNextOverlayClick = false;
			}, 120);
		}
	}

	function handlePointerCancel(e: PointerEvent) {
		if (!isDragging || dragPointerId !== e.pointerId) return;
		finishDrag(false);
	}

	$effect(() => {
		if (!visible) {
			resetDragState();
		}
	});
</script>

{#if visible && issue}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="sheet-overlay" onclick={handleOverlayClick}>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="sheet"
			class:dragging={isDragging}
			bind:this={sheetEl}
			style="transform: translateY({dragOffsetY}px);"
			onclick={handleSheetClick}
			onpointerdown={handlePointerDown}
			onpointermove={handlePointerMove}
			onpointerup={handlePointerUp}
			onpointercancel={handlePointerCancel}
		>
			<div class="sheet-handle-area">
				<div class="sheet-handle"></div>
			</div>

			<div class="sheet-content">
				<div class="sheet-close-row">
					<button class="sheet-close-btn" type="button" onclick={onclose} aria-label="Tutup detail issue">
						✕
					</button>
				</div>

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
					{#if getIssueRoadOrAreaLabel(issue)}
						<span class="road-name">{getIssueRoadOrAreaLabel(issue)}</span>
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
		touch-action: pan-y;
		transition: transform 0.2s ease-out;
		will-change: transform;
	}

	.sheet.dragging {
		transition: none;
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
		touch-action: none;
		cursor: grab;
	}

	/* Desktop/web: hide swipe affordance, keep close button only */
	@media (hover: hover) and (pointer: fine) {
		.sheet-handle-area {
			display: none;
		}
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

	.sheet-close-row {
		display: flex;
		justify-content: flex-end;
		margin-bottom: 6px;
	}

	.sheet-close-btn {
		width: 28px;
		height: 28px;
		border: 1px solid #E2E8F0;
		border-radius: 999px;
		background: #fff;
		color: #64748B;
		font-size: 14px;
		line-height: 1;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		transition: background 0.15s, color 0.15s;
	}

	.sheet-close-btn:hover {
		background: #F8FAFC;
		color: #0F172A;
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
		font-weight: 700;
		border-radius: 12px;
		border: none;
		cursor: pointer;
		min-height: 50px;
		transition: opacity 0.15s, transform 0.1s;
	}

	.btn:active {
		transform: scale(0.97);
	}

	.btn-primary {
		background: linear-gradient(180deg, #EB5960 0%, #E5484D 100%);
		color: #fff;
		border: 1px solid rgba(173, 40, 45, 0.35);
		box-shadow: 0 4px 12px rgba(229, 72, 77, 0.2), 0 1px 2px rgba(0,0,0,0.06);
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
