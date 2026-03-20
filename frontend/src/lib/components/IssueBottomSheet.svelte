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
	const DRAG_CLOSE_VELOCITY = 0.7;

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
		const sc = statusColor[status] || statusColor.open;
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
			class="sheet rounded-t-[30px] border border-white/80 bg-white/96 shadow-[0_-12px_40px_rgba(15,23,42,0.18)] backdrop-blur-xl md:rounded-none md:border-l md:border-r-0 md:border-t-0"
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
				<div class="h-1 w-11 rounded-full bg-slate-300"></div>
			</div>

			<div class="space-y-5 px-5 pb-6 pt-2">
				<div class="flex items-center justify-between">
					<div class="flex flex-wrap gap-2">
						<span
							class="inline-flex items-center rounded-full px-3 py-1 text-xs font-bold text-white"
							style={`background: ${severityColor[issue.severity_current] || '#94A3B8'}`}
						>
							{severityLabel[issue.severity_current] || `Level ${issue.severity_current}`}
						</span>
						<span
							class="inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold"
							style={getStatusStyle(issue.status)}
						>
							{statusLabel[issue.status] || issue.status}
						</span>
					</div>
					<button class="btn-icon size-10" type="button" onclick={onclose} aria-label="Tutup detail issue">
						<CloseCircleIcon class="size-5" />
					</button>
				</div>

				<div class="rounded-[24px] border border-slate-200 bg-slate-50/85 p-4">
					<div class="flex items-start gap-3">
						<div class="flex size-10 shrink-0 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
							<LocationIcon class="size-5" />
						</div>
						<div class="min-w-0">
							<p class="text-base font-bold leading-6 text-slate-950">
								{getIssueRoadOrAreaLabel(issue) || `${issue.latitude.toFixed(4)}, ${issue.longitude.toFixed(4)}`}
							</p>
							<div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-slate-500">
								<span>{issue.latitude.toFixed(4)}, {issue.longitude.toFixed(4)}</span>
								{#if issue.road_type}
									<span>• {issue.road_type}</span>
								{/if}
							</div>
						</div>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div class="rounded-[20px] border border-slate-200 bg-white px-4 py-3 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
						<div class="flex items-center gap-2 text-slate-500">
							<DocumentIcon class="size-[18px]" />
							<span class="text-[11px] font-bold uppercase tracking-[0.16em]">Laporan</span>
						</div>
						<p class="mt-2 text-lg font-[800] tracking-[-0.03em] text-slate-950">{issue.submission_count}</p>
					</div>
					<div class="rounded-[20px] border border-slate-200 bg-white px-4 py-3 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
						<div class="flex items-center gap-2 text-slate-500">
							<CameraIcon class="size-[18px]" />
							<span class="text-[11px] font-bold uppercase tracking-[0.16em]">Foto</span>
						</div>
						<p class="mt-2 text-lg font-[800] tracking-[-0.03em] text-slate-950">{issue.photo_count}</p>
					</div>
					<div
						class={`rounded-[20px] border px-4 py-3 shadow-[0_10px_24px_rgba(15,23,42,0.04)] ${issue.casualty_count > 0 ? 'border-rose-200 bg-rose-50/70' : 'border-slate-200 bg-white'}`}
					>
						<div class="flex items-center gap-2 text-slate-500">
							<DangerIcon class="size-[18px]" />
							<span class="text-[11px] font-bold uppercase tracking-[0.16em]">Korban</span>
						</div>
						<p class:text-rose-700={issue.casualty_count > 0} class="mt-2 text-lg font-[800] tracking-[-0.03em] text-slate-950">
							{issue.casualty_count}
						</p>
					</div>
					<div class="rounded-[20px] border border-slate-200 bg-white px-4 py-3 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
						<div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-500">Terakhir</div>
						<p class="mt-2 text-sm font-bold text-slate-900">{relativeTime(issue.last_seen_at)}</p>
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
		background: linear-gradient(180deg, rgba(15, 23, 42, 0.04), rgba(15, 23, 42, 0.18));
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
