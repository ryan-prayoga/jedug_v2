<script lang="ts">
	import type { Issue } from '$lib/api/types';
	import {
		CameraIcon,
		CloseCircleIcon,
		DangerIcon,
		DocumentIcon,
		LocationIcon,
		MapIcon
	} from '$lib/icons';
	import { relativeTime } from '$lib/utils/date';
	import { getIssueRoadOrAreaLabel, getStatusLabel } from '$lib/utils/issue-detail';

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
	const DRAG_CLOSE_VELOCITY = 0.7;

	const severityLabel = ['', 'Ringan', 'Sedang', 'Berat', 'Parah', 'Kritis'];
	const statusData: Record<string, string> = {
		fixed: 'fixed',
		closed: 'fixed',
		archived: 'archived',
		merged: 'archived',
		rejected: 'rejected'
	};

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
			class="sheet rounded-t-[12px] border border-hairline bg-surface md:rounded-none md:border-l md:border-r-0 md:border-t-0"
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
				<div class="h-1 w-11 rounded-full bg-hairline-strong"></div>
			</div>

			<div class="space-y-5 px-5 pb-6 pt-2">
				<div class="flex items-center justify-between">
					<div class="flex flex-wrap gap-2">
						<span class="severity-pill" data-sev={issue.severity_current}>
							{severityLabel[issue.severity_current] || `Level ${issue.severity_current}`}
						</span>
						<span class="status-pill" data-status={statusData[issue.status] ?? 'open'}>
							{getStatusLabel(issue.status)}
						</span>
					</div>
					<button class="btn-icon size-10" type="button" onclick={onclose} aria-label="Tutup detail issue">
						<CloseCircleIcon class="size-5" />
					</button>
				</div>

				<div class="jedug-panel p-4">
					<div class="flex items-start gap-3">
						<LocationIcon class="mt-0.5 size-5 shrink-0 text-muted" />
						<div class="min-w-0">
							<p class="font-serif text-base font-semibold leading-snug text-ink">
								{getIssueRoadOrAreaLabel(issue) || `${issue.latitude.toFixed(4)}, ${issue.longitude.toFixed(4)}`}
							</p>
							<div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted nums">
								<span>{issue.latitude.toFixed(4)}, {issue.longitude.toFixed(4)}</span>
								{#if issue.road_type}
									<span>• {issue.road_type}</span>
								{/if}
							</div>
						</div>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div class="metric-tile">
						<div class="flex items-center gap-2 text-subtle">
							<DocumentIcon class="size-[18px]" />
							<span class="text-[11px] font-semibold uppercase tracking-[0.16em]">Laporan</span>
						</div>
						<p class="mt-2 font-serif text-xl font-semibold tabular-nums text-ink">{issue.submission_count}</p>
					</div>
					<div class="metric-tile">
						<div class="flex items-center gap-2 text-subtle">
							<CameraIcon class="size-[18px]" />
							<span class="text-[11px] font-semibold uppercase tracking-[0.16em]">Foto</span>
						</div>
						<p class="mt-2 font-serif text-xl font-semibold tabular-nums text-ink">{issue.photo_count}</p>
					</div>
					<div class="metric-tile" class:border-brand={issue.casualty_count > 0}>
						<div class="flex items-center gap-2 text-subtle">
							<DangerIcon class="size-[18px]" />
							<span class="text-[11px] font-semibold uppercase tracking-[0.16em]">Korban</span>
						</div>
						<p class="mt-2 font-serif text-xl font-semibold tabular-nums" class:text-brand={issue.casualty_count > 0} class:text-ink={issue.casualty_count === 0}>
							{issue.casualty_count}
						</p>
					</div>
					<div class="metric-tile">
						<div class="text-[11px] font-semibold uppercase tracking-[0.16em] text-subtle">Terakhir</div>
						<p class="mt-2 text-sm font-semibold text-ink">{relativeTime(issue.last_seen_at)}</p>
					</div>
				</div>

				<div class="flex flex-col gap-2 sm:flex-row">
					<a href="/issues/{issue.id}" class="btn-primary flex-1">
						<MapIcon class="size-[18px]" />
						Lihat Detail
					</a>
					<a href="/lapor" class="btn-secondary flex-1">
						<LocationIcon class="size-[18px]" />
						Lapor di Sini
					</a>
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
		background: rgba(26, 26, 26, 0.14);
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
			width: 392px !important;
			max-height: 100% !important;
			animation: slideInRight 0.2s ease-out !important;
		}

		.sheet-handle-area {
			display: none;
		}
	}

	.sheet {
		width: 100%;
		max-height: 62vh;
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

	.sheet-handle-area {
		padding: 12px 0 4px;
		display: flex;
		justify-content: center;
		touch-action: none;
		cursor: grab;
	}

	@media (hover: hover) and (pointer: fine) {
		.sheet-handle-area {
			display: none;
		}
	}
</style>
