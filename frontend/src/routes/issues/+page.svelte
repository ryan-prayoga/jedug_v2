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
		loading = true;
		fetchIssuesByBBox(bbox, { limit: 100 }, (data, err) => {
			if (err) {
				error = err;
			} else {
				error = null;
				issues = data;
			}
			loading = false;
		});
	}

	function handleIssueSelect(issue: Issue | null) {
		selectedIssue = issue;
	}

	function handleMapError(msg: string) {
		mapError = msg;
		viewMode = 'list';
		fetchList();
	}

	function handleCloseSheet() {
		selectedIssue = null;
	}

	function toggleView() {
		if (viewMode === 'map') {
			viewMode = 'list';
			if (issues.length === 0) fetchList();
		} else {
			viewMode = 'map';
		}
	}

	function toggleListPanel() {
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

	{#if viewMode === 'map' && IssueMap && !mapError}
		<!-- Map View -->
		<div class="map-container">
			<div class="map-area">
				<IssueMap
					{issues}
					{selectedIssue}
					onbboxchange={handleBBoxChange}
					onissueselect={handleIssueSelect}
					onmaperror={handleMapError}
				/>

				<!-- Loading overlay -->
				{#if loading}
					<div class="map-loading">
						<span class="loading-dot"></span>
						Memuat...
					</div>
				{/if}

				<!-- Issue count badge -->
				{#if !loading && issues.length > 0}
					<div class="issue-count">{issues.length} titik</div>
				{/if}

				<!-- Empty state overlay -->
				{#if !loading && issues.length === 0 && !error}
					<div class="map-empty">
						Tidak ada laporan di area ini
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

			<!-- Side list panel (desktop) / slide panel -->
			{#if showList}
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
		<a href="/lapor" class="report-cta">📸 Laporkan Jalan Rusak</a>
	</div>
</div>

<style>
	.issues-page {
		display: flex;
		flex-direction: column;
		height: calc(100dvh - 49px); /* subtract header height */
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
		border-bottom: 1px solid #e2e8f0;
		flex-shrink: 0;
		z-index: 10;
	}

	.toolbar h1 {
		font-size: 1.1rem;
		font-weight: 700;
		margin: 0;
	}

	.toolbar-actions {
		display: flex;
		gap: 8px;
	}

	.tool-btn {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 8px;
		padding: 5px 10px;
		font-size: 0.85rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.tool-btn:hover {
		background: #f7fafc;
	}

	.btn-label {
		font-size: 0.8rem;
		color: #4a5568;
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

	/* Map overlays */
	.map-loading {
		position: absolute;
		top: 12px;
		left: 12px;
		background: #fff;
		padding: 5px 12px;
		border-radius: 8px;
		font-size: 0.8rem;
		box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
		display: flex;
		align-items: center;
		gap: 6px;
		z-index: 5;
	}

	.loading-dot {
		width: 8px;
		height: 8px;
		background: #e53e3e;
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
		padding: 4px 10px;
		border-radius: 8px;
		font-size: 0.75rem;
		font-weight: 600;
		color: #4a5568;
		box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
		z-index: 5;
	}

	.map-empty {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: rgba(255, 255, 255, 0.92);
		padding: 12px 20px;
		border-radius: 10px;
		font-size: 0.9rem;
		color: #718096;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
		z-index: 5;
		pointer-events: none;
	}

	.map-error-overlay {
		position: absolute;
		top: 12px;
		left: 12px;
		right: 12px;
		background: #fff5f5;
		border: 1px solid #fed7d7;
		padding: 8px 12px;
		border-radius: 8px;
		font-size: 0.8rem;
		color: #c53030;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		z-index: 5;
	}

	.map-error-overlay button {
		background: none;
		border: none;
		color: #c53030;
		font-weight: 600;
		cursor: pointer;
		font-size: 0.8rem;
		white-space: nowrap;
	}

	/* Side panel */
	.side-panel {
		width: 340px;
		background: #fff;
		border-left: 1px solid #e2e8f0;
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
		z-index: 15;
	}

	.side-panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 14px;
		border-bottom: 1px solid #e2e8f0;
	}

	.side-panel-header h2 {
		font-size: 0.9rem;
		font-weight: 600;
		margin: 0;
	}

	.close-panel {
		background: none;
		border: none;
		font-size: 1rem;
		cursor: pointer;
		color: #718096;
		padding: 4px;
	}

	.side-panel-list {
		flex: 1;
		overflow-y: auto;
		padding: 8px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.list-item-btn {
		all: unset;
		cursor: pointer;
		display: block;
		width: 100%;
		border-radius: 12px;
		transition: outline 0.1s;
	}

	.list-item-btn.selected {
		outline: 2px solid #e53e3e;
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
		padding: 0 16px;
		max-width: 480px;
		margin: 0 auto;
		width: 100%;
	}

	.map-error-banner {
		background: #fff5f5;
		border: 1px solid #fed7d7;
		color: #c53030;
		font-size: 0.85rem;
		padding: 8px 12px;
		border-radius: 8px;
		margin-bottom: 12px;
	}

	.issue-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	/* Bottom CTA */
	.bottom-cta {
		padding: 10px 16px;
		text-align: center;
		background: #fff;
		border-top: 1px solid #e2e8f0;
		flex-shrink: 0;
	}

	.bottom-cta.cta-over-map {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		z-index: 10;
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(4px);
	}

	.report-cta {
		display: inline-block;
		padding: 10px 22px;
		font-size: 0.9rem;
		font-weight: 600;
		color: #fff;
		background: #e53e3e;
		border-radius: 12px;
		text-decoration: none;
	}

	.report-cta:hover {
		opacity: 0.9;
	}
</style>
