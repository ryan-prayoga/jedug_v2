<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { listIssues } from '$lib/api/issues';
	import type { Issue } from '$lib/api/types';
	import IssueCard from '$lib/components/IssueCard.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { fetchIssuesByBBox, resetBBoxFetcher, type BBox } from '$lib/utils/bbox';
	import type { MapVisualMode } from '$lib/utils/issue-heatmap';

	const MAP_VISUAL_MODES: Array<{ id: MapVisualMode; label: string }> = [
		{ id: 'marker', label: 'Marker' },
		{ id: 'heatmap', label: 'Heatmap' }
	];

	let IssueMap: typeof import('$lib/components/IssueMap.svelte').default | null = $state(null);
	let IssueBottomSheet: typeof import('$lib/components/IssueBottomSheet.svelte').default | null =
		$state(null);

	let issues = $state<Issue[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let mapError = $state<string | null>(null);
	let heatmapNotice = $state<string | null>(null);
	let selectedIssue = $state<Issue | null>(null);
	let viewMode = $state<'map' | 'list'>('map');
	let mapVisualMode = $state<MapVisualMode>('marker');
	let showList = $state(false);
	let mapReady = $state(false);
	let mapFetching = $state(false);
	let mapHasFetchedViewport = $state(false);

	const showMapLoading = $derived(mapFetching && !mapHasFetchedViewport && issues.length === 0);
	const showMapInfo = $derived(mapReady && mapHasFetchedViewport && !showMapLoading && !error);
	const showHeatmapLegend = $derived(
		viewMode === 'map' && mapVisualMode === 'heatmap' && mapReady && !mapError
	);
	const mapInfoText = $derived(
		issues.length === 0
			? mapVisualMode === 'heatmap'
				? 'Belum ada hotspot di area ini'
				: 'Belum ada laporan di area ini'
			: mapVisualMode === 'heatmap'
				? 'Hotspot severity area ini'
				: 'Laporan di area ini'
	);

	// Dynamically import map components (graceful fallback if they fail)
	onMount(async () => {
		try {
			const [mapMod, sheetMod] = await Promise.all([
				import('$lib/components/IssueMap.svelte'),
				import('$lib/components/IssueBottomSheet.svelte')
			]);
			IssueMap = mapMod.default;
			IssueBottomSheet = sheetMod.default;
		} catch {
			mapError = 'Komponen peta gagal dimuat';
			viewMode = 'list';
			await fetchList();
		}
	});

	onDestroy(() => {
		resetBBoxFetcher();
	});

	async function fetchList() {
		loading = true;
		error = null;
		try {
			const res = await listIssues({ limit: 50, offset: 0 });
			issues = res.data || [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat data';
		} finally {
			loading = false;
		}
	}

	function handleBBoxChange(bbox: BBox) {
		mapFetching = true;
		const fetchState = fetchIssuesByBBox(bbox, { limit: 100 }, (data, err) => {
			if (err) {
				error = err;
			} else {
				error = null;
				issues = data;
				mapHasFetchedViewport = true;
				if (selectedIssue && !data.some((issue) => issue.id === selectedIssue?.id)) {
					selectedIssue = null;
				}
			}
			mapFetching = false;
			loading = false;
		});

		if (fetchState === 'skipped') {
			mapFetching = false;
			loading = false;
		}
	}

	function handleIssueSelect(issue: Issue | null) {
		heatmapNotice = null;
		if (issue && mapVisualMode === 'heatmap') {
			mapVisualMode = 'marker';
		}
		if (issue && showList && window.matchMedia('(min-width: 768px)').matches) {
			showList = false;
		}
		selectedIssue = issue;
	}

	function handleMapError(msg: string) {
		mapError = msg;
		heatmapNotice = null;
		viewMode = 'list';
		mapReady = false;
		fetchList();
	}

	function handleMapReady() {
		mapReady = true;
	}

	function handleCloseSheet() {
		selectedIssue = null;
	}

	function handleVisualModeChange(nextMode: MapVisualMode) {
		if (mapVisualMode === nextMode) return;
		heatmapNotice = null;
		mapVisualMode = nextMode;
		if (nextMode === 'heatmap') {
			selectedIssue = null;
			showList = false;
		}
	}

	function handleVisualModeFallback(nextMode: MapVisualMode, message: string) {
		heatmapNotice = message;
		mapVisualMode = nextMode;
		selectedIssue = null;
		showList = false;
	}

	function toggleView() {
		if (viewMode === 'map') {
			viewMode = 'list';
			if (issues.length === 0 && !loading) fetchList();
		} else {
			resetBBoxFetcher();
			viewMode = 'map';
			mapReady = false;
			mapFetching = false;
			mapHasFetchedViewport = issues.length > 0;
		}
	}

	function toggleListPanel() {
		if (!showList && selectedIssue && window.matchMedia('(min-width: 768px)').matches) {
			selectedIssue = null;
		}
		showList = !showList;
	}
</script>

<div class="issues-page" class:map-mode={viewMode === 'map'}>
	<div class="toolbar">
		<div class="toolbar-copy">
			<h1>Laporan Publik</h1>
			<p>Amati titik laporan atau pola area rawan secara cepat.</p>
		</div>
		<div class="toolbar-actions">
			{#if viewMode === 'map' && !mapError}
				<div class="visual-mode-toggle" aria-label="Mode visual peta">
					{#each MAP_VISUAL_MODES as mode}
						<button
							type="button"
							class="visual-mode-btn"
							class:active={mapVisualMode === mode.id}
							onclick={() => handleVisualModeChange(mode.id)}
						>
							{mode.label}
						</button>
					{/each}
				</div>
				<button class="tool-btn" class:active={showList} onclick={toggleListPanel} title="Daftar area">
					Panel
				</button>
			{/if}
			<button class="tool-btn" onclick={toggleView} title={viewMode === 'map' ? 'Mode daftar' : 'Mode peta'}>
				<span class="btn-label">{viewMode === 'map' ? 'Daftar' : 'Peta'}</span>
			</button>
		</div>
	</div>

	{#if viewMode === 'map' && !mapError}
		<div class="map-container">
			{#if IssueMap}
				<div class="map-area">
					<IssueMap
						{issues}
						{selectedIssue}
						visualMode={mapVisualMode}
						onbboxchange={handleBBoxChange}
						onissueselect={handleIssueSelect}
						onmaperror={handleMapError}
						onmapready={handleMapReady}
						onvisualmodefallback={handleVisualModeFallback}
					/>

					{#if showMapLoading}
						<div class="map-loading">
							<span class="loading-dot"></span>
							Memuat titik laporan...
						</div>
					{/if}

					{#if showMapInfo}
						<div class="map-info-badge" class:empty={issues.length === 0} class:heatmap={mapVisualMode === 'heatmap'}>
							<span class="map-info-mode">{mapVisualMode === 'heatmap' ? 'Heatmap' : 'Marker'}</span>
							<span class="map-info-count">{issues.length} titik</span>
							<span class="map-info-separator">•</span>
							<span class="map-info-text">{mapInfoText}</span>
						</div>
					{/if}

					{#if heatmapNotice}
						<div class="map-mode-notice">
							<span>{heatmapNotice}</span>
							<button type="button" onclick={() => { heatmapNotice = null; }}>Tutup</button>
						</div>
					{/if}

					{#if error}
						<div class="map-error-overlay">
							{error}
							<button type="button" onclick={() => { error = null; }}>Tutup</button>
						</div>
					{/if}

					{#if showHeatmapLegend}
						<div class="heatmap-legend">
							<div class="heatmap-legend-header">
								<strong>Intensitas area</strong>
								<span>Severity, korban, laporan</span>
							</div>
							<div class="heatmap-legend-bar"></div>
							<div class="heatmap-legend-scale">
								<span>Rendah</span>
								<span>Tinggi</span>
							</div>
						</div>
					{/if}

					{#if IssueBottomSheet}
						<IssueBottomSheet
							issue={selectedIssue}
							visible={selectedIssue !== null && mapVisualMode === 'marker'}
							onclose={handleCloseSheet}
						/>
					{/if}
				</div>
			{:else}
				<div class="map-bootstrap">
					<LoadingState message="Menyiapkan peta..." />
				</div>
			{/if}

			{#if showList && IssueMap}
				<div class="side-panel">
					<div class="side-panel-header">
						<h2>Daftar Laporan ({issues.length})</h2>
						<button class="close-panel" type="button" onclick={toggleListPanel}>✕</button>
					</div>
					<div class="side-panel-list">
						{#if issues.length === 0}
							<EmptyState message="Tidak ada laporan di area ini" icon="🚧" />
						{:else}
							{#each issues as issue (issue.id)}
								<button
									type="button"
									class="list-item-btn"
									class:selected={mapVisualMode === 'marker' && selectedIssue?.id === issue.id}
									onclick={() => handleIssueSelect(issue)}
								>
									<IssueCard {issue} />
								</button>
							{/each}
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<div class="list-view">
			{#if mapError}
				<div class="map-error-banner">⚠️ {mapError} — menampilkan mode daftar</div>
			{/if}

			{#if loading}
				<LoadingState message="Memuat laporan..." />
			{:else if error}
				<ErrorState message={error} onretry={fetchList} />
			{:else if issues.length === 0}
				<EmptyState
					message="Belum ada laporan. Jadilah yang pertama melaporkan!"
					icon="🚧"
					ctaHref="/lapor"
					ctaLabel="Laporkan Jalan Rusak"
				/>
			{:else}
				<div class="issue-list">
					{#each issues as issue (issue.id)}
						<IssueCard {issue} />
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	<div class="bottom-cta" class:cta-over-map={viewMode === 'map' && !mapError}>
		<a href="/lapor" class="report-cta">Laporkan Jalan Rusak</a>
	</div>
</div>

<style>
	.issues-page {
		display: flex;
		flex-direction: column;
		height: calc(100dvh - 49px);
	}

	.issues-page.map-mode {
		padding: 0;
		max-width: none;
	}

	.toolbar {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
		padding: 12px 16px;
		background: #fff;
		border-bottom: 1px solid #E2E8F0;
		flex-shrink: 0;
		z-index: 10;
	}

	.toolbar-copy {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}

	.toolbar h1 {
		font-size: 16px;
		font-weight: 700;
		margin: 0;
		color: #0F172A;
	}

	.toolbar p {
		margin: 0;
		font-size: 12px;
		line-height: 1.4;
		color: #64748B;
		max-width: 260px;
	}

	.toolbar-actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		flex-wrap: wrap;
		gap: 8px;
	}

	.visual-mode-toggle {
		display: inline-flex;
		align-items: center;
		padding: 3px;
		border-radius: 12px;
		background: #F8FAFC;
		border: 1px solid #E2E8F0;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
	}

	.visual-mode-btn {
		border: none;
		background: transparent;
		color: #64748B;
		font-size: 12px;
		font-weight: 700;
		padding: 8px 12px;
		border-radius: 9px;
		cursor: pointer;
		transition: background 0.15s, color 0.15s, box-shadow 0.15s;
	}

	.visual-mode-btn.active {
		background: linear-gradient(180deg, #FFFFFF 0%, #F8FAFC 100%);
		color: #0F172A;
		box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
	}

	.tool-btn {
		background: #fff;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		padding: 8px 12px;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 4px;
		color: #334155;
		transition: background 0.15s, border-color 0.15s;
	}

	.tool-btn:hover,
	.tool-btn.active {
		background: #F8FAFC;
		border-color: #CBD5E1;
	}

	.btn-label {
		font-size: 13px;
		color: inherit;
		font-weight: 700;
	}

	.map-container {
		flex: 1;
		position: relative;
		display: flex;
		overflow: hidden;
	}

	.map-area {
		flex: 1;
		position: relative;
		min-height: 0;
	}

	.map-bootstrap {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(180deg, #F8FAFC 0%, #EEF2F7 100%);
	}

	.map-loading {
		position: absolute;
		top: 12px;
		left: 12px;
		background: #fff;
		padding: 6px 14px;
		border-radius: 10px;
		font-size: 12px;
		font-weight: 500;
		color: #64748B;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
		display: flex;
		align-items: center;
		gap: 8px;
		z-index: 5;
	}

	.loading-dot {
		width: 8px;
		height: 8px;
		background: #E5484D;
		border-radius: 50%;
		animation: pulse 1s infinite;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.3;
		}
	}

	.map-info-badge {
		position: absolute;
		top: 12px;
		left: 12px;
		background: rgba(255, 255, 255, 0.96);
		padding: 8px 12px;
		border-radius: 12px;
		font-size: 12px;
		font-weight: 600;
		color: #0F172A;
		border: 1px solid #E2E8F0;
		box-shadow: 0 4px 12px rgba(15, 23, 42, 0.08);
		z-index: 5;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		max-width: min(320px, calc(100% - 24px));
	}

	.map-info-badge.empty {
		background: rgba(255, 255, 255, 0.98);
	}

	.map-info-badge.heatmap {
		border-color: rgba(249, 115, 22, 0.24);
	}

	.map-info-mode {
		display: inline-flex;
		align-items: center;
		padding: 4px 8px;
		border-radius: 999px;
		background: #F8FAFC;
		color: #334155;
		font-size: 11px;
		font-weight: 700;
	}

	.map-info-count {
		font-size: 13px;
		font-weight: 700;
		color: #0F172A;
	}

	.map-info-separator {
		color: #94A3B8;
		font-size: 11px;
	}

	.map-info-text {
		color: #64748B;
		font-size: 12px;
		font-weight: 500;
	}

	.map-mode-notice {
		position: absolute;
		top: 64px;
		left: 12px;
		right: 12px;
		background: rgba(255, 247, 237, 0.96);
		border: 1px solid #FED7AA;
		padding: 10px 12px;
		border-radius: 12px;
		font-size: 12px;
		color: #C2410C;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		z-index: 5;
		box-shadow: 0 8px 20px rgba(249, 115, 22, 0.08);
	}

	.map-mode-notice button {
		background: none;
		border: none;
		color: #C2410C;
		font-weight: 700;
		cursor: pointer;
		font-size: 12px;
		white-space: nowrap;
	}

	.map-error-overlay {
		position: absolute;
		top: 12px;
		left: 12px;
		right: 12px;
		background: #FEF2F2;
		border: 1px solid #FECACA;
		padding: 8px 12px;
		border-radius: 10px;
		font-size: 12px;
		color: #DC2626;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		z-index: 5;
	}

	.map-error-overlay button {
		background: none;
		border: none;
		color: #DC2626;
		font-weight: 600;
		cursor: pointer;
		font-size: 12px;
		white-space: nowrap;
	}

	.heatmap-legend {
		position: absolute;
		left: 12px;
		bottom: 84px;
		padding: 12px;
		border-radius: 14px;
		background: rgba(255, 255, 255, 0.96);
		border: 1px solid #E2E8F0;
		box-shadow: 0 10px 24px rgba(15, 23, 42, 0.12);
		z-index: 5;
		max-width: min(260px, calc(100% - 24px));
	}

	.heatmap-legend-header {
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin-bottom: 10px;
	}

	.heatmap-legend-header strong {
		font-size: 12px;
		color: #0F172A;
	}

	.heatmap-legend-header span {
		font-size: 11px;
		color: #64748B;
	}

	.heatmap-legend-bar {
		height: 10px;
		border-radius: 999px;
		background: linear-gradient(90deg, rgba(246, 196, 83, 0.45) 0%, rgba(249, 115, 22, 0.72) 52%, rgba(229, 72, 77, 0.92) 82%, rgba(153, 27, 27, 0.98) 100%);
	}

	.heatmap-legend-scale {
		display: flex;
		justify-content: space-between;
		margin-top: 8px;
		font-size: 11px;
		font-weight: 600;
		color: #64748B;
	}

	.side-panel {
		width: 360px;
		background: #fff;
		border-left: 1px solid #E2E8F0;
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
		z-index: 15;
	}

	.side-panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		border-bottom: 1px solid #E2E8F0;
	}

	.side-panel-header h2 {
		font-size: 14px;
		font-weight: 600;
		margin: 0;
		color: #0F172A;
	}

	.close-panel {
		background: none;
		border: none;
		font-size: 16px;
		cursor: pointer;
		color: #64748B;
		padding: 4px;
	}

	.side-panel-list {
		flex: 1;
		overflow-y: auto;
		padding: 8px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.list-item-btn {
		all: unset;
		cursor: pointer;
		display: block;
		width: 100%;
		border-radius: 16px;
		transition: outline 0.1s;
	}

	.list-item-btn.selected {
		outline: 2px solid #E5484D;
		outline-offset: -1px;
	}

	@media (max-width: 767px) {
		.toolbar {
			flex-direction: column;
			align-items: stretch;
		}

		.toolbar-copy p {
			max-width: none;
		}

		.toolbar-actions {
			justify-content: space-between;
		}

		.side-panel {
			position: absolute;
			inset: 0;
			width: 100%;
			z-index: 25;
			animation: slideUp 0.2s ease-out;
		}

		.heatmap-legend {
			bottom: 96px;
		}
	}

	@keyframes slideUp {
		from {
			transform: translateY(100%);
		}
		to {
			transform: translateY(0);
		}
	}

	.list-view {
		flex: 1;
		overflow-y: auto;
		padding: 12px 16px;
		max-width: 480px;
		margin: 0 auto;
		width: 100%;
	}

	.map-error-banner {
		background: #FEF2F2;
		border: 1px solid #FECACA;
		color: #DC2626;
		font-size: 13px;
		padding: 8px 12px;
		border-radius: 10px;
		margin-bottom: 12px;
	}

	.issue-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.bottom-cta {
		padding: 12px 16px;
		text-align: center;
		background: #fff;
		border-top: 1px solid #E2E8F0;
		flex-shrink: 0;
	}

	.bottom-cta.cta-over-map {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		z-index: 10;
		background: rgba(255, 255, 255, 0.96);
		backdrop-filter: blur(8px);
	}

	.report-cta {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 14px 24px;
		font-size: 15px;
		font-weight: 700;
		color: #fff;
		background: linear-gradient(180deg, #EB5960 0%, #E5484D 100%);
		border-radius: 12px;
		border: 1px solid rgba(173, 40, 45, 0.35);
		text-decoration: none;
		min-height: 52px;
		width: 100%;
		letter-spacing: 0.1px;
		transition: opacity 0.15s, transform 0.1s, box-shadow 0.15s;
		box-shadow: 0 6px 16px rgba(229, 72, 77, 0.22), 0 1px 3px rgba(0, 0, 0, 0.08);
	}

	.report-cta:hover {
		opacity: 0.94;
		box-shadow: 0 8px 20px rgba(229, 72, 77, 0.28), 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.report-cta:active {
		transform: scale(0.97);
		box-shadow: 0 4px 12px rgba(229, 72, 77, 0.2);
	}
</style>
