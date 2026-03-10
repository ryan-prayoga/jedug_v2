<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { listIssues } from '$lib/api/issues';
	import type { Issue } from '$lib/api/types';
	import IssueCard from '$lib/components/IssueCard.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import { fetchIssuesByBBox, resetBBoxFetcher, type BBox } from '$lib/utils/bbox';

	let IssueMap: typeof import('$lib/components/IssueMap.svelte').default | null = $state(null);
	let IssueBottomSheet: typeof import('$lib/components/IssueBottomSheet.svelte').default | null =
		$state(null);

	let issues = $state<Issue[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let mapError = $state<string | null>(null);
	let selectedIssue = $state<Issue | null>(null);
	let viewMode = $state<'map' | 'list'>('map');
	let showList = $state(false);
	let mapReady = $state(false);
	let mapFetching = $state(false);
	let mapHasFetchedViewport = $state(false);

	const showMapLoading = $derived(mapFetching && !mapHasFetchedViewport && issues.length === 0);
	const showMapEmpty = $derived(
		mapReady &&
			mapHasFetchedViewport &&
			!mapFetching &&
			issues.length === 0 &&
			!error
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
			// Fallback: load list data
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
		if (issue && showList && window.matchMedia('(min-width: 768px)').matches) {
			showList = false;
		}
		selectedIssue = issue;
	}

	function handleMapError(msg: string) {
		mapError = msg;
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
	<!-- Toolbar -->
	<div class="toolbar">
		<h1>Laporan Publik</h1>
		<div class="toolbar-actions">
			{#if viewMode === 'map' && !mapError}
				<button class="tool-btn" onclick={toggleListPanel} title="Daftar">
					{showList ? '🗺️' : '📋'}
				</button>
			{/if}
			<button class="tool-btn" onclick={toggleView} title={viewMode === 'map' ? 'Mode List' : 'Mode Peta'}>
				{viewMode === 'map' ? '📋' : '🗺️'}
				<span class="btn-label">{viewMode === 'map' ? 'List' : 'Peta'}</span>
			</button>
		</div>
	</div>

	{#if viewMode === 'map' && !mapError}
		<!-- Map View -->
		<div class="map-container">
			{#if IssueMap}
				<div class="map-area">
					<IssueMap
						{issues}
						{selectedIssue}
						onbboxchange={handleBBoxChange}
						onissueselect={handleIssueSelect}
						onmaperror={handleMapError}
						onmapready={handleMapReady}
					/>

					<!-- Loading overlay -->
					{#if showMapLoading}
						<div class="map-loading">
							<span class="loading-dot"></span>
							Memuat titik laporan...
						</div>
					{/if}

					<!-- Issue count badge -->
					{#if !showMapLoading && issues.length > 0}
						<div class="issue-count">{issues.length} titik</div>
					{/if}

					<!-- Empty state overlay -->
					{#if showMapEmpty}
						<div class="map-empty">
							<span class="map-empty-icon">🚧</span>
							<span class="map-empty-text">Tidak ada laporan di area ini</span>
							<a href="/lapor" class="map-empty-cta">Laporkan Jalan Rusak</a>
						</div>
					{/if}

					<!-- Error overlay -->
					{#if error}
						<div class="map-error-overlay">
							{error}
							<button onclick={() => { error = null; }}>Tutup</button>
						</div>
					{/if}

					<!-- Bottom sheet for selected issue -->
					{#if IssueBottomSheet}
						<IssueBottomSheet
							issue={selectedIssue}
							visible={selectedIssue !== null}
							onclose={handleCloseSheet}
						/>
					{/if}
				</div>
			{:else}
				<div class="map-bootstrap">
					<LoadingState message="Menyiapkan peta..." />
				</div>
			{/if}

			<!-- Side list panel (desktop) / slide panel -->
			{#if showList && IssueMap}
				<div class="side-panel">
					<div class="side-panel-header">
						<h2>Daftar Laporan ({issues.length})</h2>
						<button class="close-panel" onclick={toggleListPanel}>✕</button>
					</div>
					<div class="side-panel-list">
						{#if issues.length === 0}
							<EmptyState message="Tidak ada laporan di area ini" icon="🚧" />
						{:else}
							{#each issues as issue (issue.id)}
								<button
									class="list-item-btn"
									class:selected={selectedIssue?.id === issue.id}
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
		<!-- List Fallback View -->
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

	<!-- Bottom CTA -->
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

	/* Toolbar */
	.toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 16px;
		background: #fff;
		border-bottom: 1px solid #E2E8F0;
		flex-shrink: 0;
		z-index: 10;
	}

	.toolbar h1 {
		font-size: 16px;
		font-weight: 600;
		margin: 0;
		color: #0F172A;
	}

	.toolbar-actions {
		display: flex;
		gap: 6px;
	}

	.tool-btn {
		background: #fff;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		padding: 6px 12px;
		font-size: 13px;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 4px;
		transition: background 0.15s;
	}

	.tool-btn:hover {
		background: #F8FAFC;
	}

	.btn-label {
		font-size: 13px;
		color: #64748B;
		font-weight: 500;
	}

	/* Map container */
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

	/* Map overlays */
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
		box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04);
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
		0%, 100% { opacity: 1; }
		50% { opacity: 0.3; }
	}

	.issue-count {
		position: absolute;
		top: 12px;
		left: 12px;
		background: #fff;
		padding: 6px 14px;
		border-radius: 10px;
		font-size: 13px;
		font-weight: 600;
		color: #0F172A;
		box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04);
		z-index: 5;
	}

	.map-empty {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: rgba(255, 255, 255, 0.96);
		padding: 24px 32px;
		border-radius: 16px;
		font-size: 14px;
		color: #64748B;
		box-shadow: 0 4px 16px rgba(0,0,0,0.10);
		z-index: 5;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		text-align: center;
	}

	.map-empty-icon {
		font-size: 32px;
		line-height: 1;
	}

	.map-empty-text {
		font-size: 14px;
		color: #64748B;
	}

	.map-empty-cta {
		margin-top: 4px;
		font-size: 13px;
		font-weight: 600;
		color: #E5484D;
		text-decoration: none;
		padding: 8px 16px;
		border: 1px solid #E5484D;
		border-radius: 10px;
		transition: background 0.15s;
		pointer-events: auto;
	}

	.map-empty-cta:hover {
		background: #FEF2F2;
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

	/* Side panel */
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

	/* Mobile: side panel overlays */
	@media (max-width: 767px) {
		.side-panel {
			position: absolute;
			inset: 0;
			width: 100%;
			z-index: 25;
			animation: slideUp 0.2s ease-out;
		}
	}

	@keyframes slideUp {
		from { transform: translateY(100%); }
		to { transform: translateY(0); }
	}

	/* List view (fallback) */
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

	/* Bottom CTA */
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
		box-shadow: 0 6px 16px rgba(229, 72, 77, 0.22), 0 1px 3px rgba(0,0,0,0.08);
	}

	.report-cta:hover {
		opacity: 0.94;
		box-shadow: 0 8px 20px rgba(229, 72, 77, 0.28), 0 1px 3px rgba(0,0,0,0.1);
	}

	.report-cta:active {
		transform: scale(0.97);
		box-shadow: 0 4px 12px rgba(229, 72, 77, 0.2);
	}
</style>
