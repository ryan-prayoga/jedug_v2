<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { listIssues } from '$lib/api/issues';
	import type { Issue } from '$lib/api/types';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import IssueCard from '$lib/components/IssueCard.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import {
		AddCircleIcon,
		ChartIcon,
		CloseCircleIcon,
		LayersIcon,
		ListCheckIcon,
		LocationIcon,
		MapIcon
	} from '$lib/icons';
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
	const modeSummary = $derived.by(() => {
		if (mapVisualMode === 'heatmap') {
			return {
				icon: ChartIcon,
				label: 'Heatmap',
				copy: 'Pola kepadatan severity, korban, dan laporan di viewport aktif.'
			};
		}

		return {
			icon: LocationIcon,
			label: 'Marker',
			copy: 'Titik issue individual untuk review cepat per lokasi.'
		};
	});
	const mapCountSummary = $derived.by(() => {
		if (mapVisualMode === 'heatmap') {
			return `${issues.length} hotspot / titik sumber`;
		}

		return `${issues.length} laporan publik`;
	});

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

<svelte:head>
	<title>Peta Laporan Publik | JEDUG</title>
	<meta
		name="description"
		content="Peta dan daftar laporan publik JEDUG untuk melihat titik jalan rusak dan pola area rawan."
	/>
</svelte:head>

<div class="flex min-h-0 flex-1 flex-col bg-paper">
	<div class="px-5 pb-5 pt-6 lg:px-6">
		<div class="flex flex-col gap-5">
			<div class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
				<div class="space-y-3">
					<p class="kicker">Peta laporan publik</p>
					<div class="space-y-2">
						<h1 class="section-title text-balance">
							Amati issue jalan rusak dari peta atau daftar.
						</h1>
						<p class="section-copy max-w-[60ch]">
							Mode marker untuk review per titik, heatmap untuk membaca pola area rawan, dan panel daftar untuk membandingkan lokasi di viewport aktif.
						</p>
					</div>
				</div>

				<div class="flex flex-col gap-3 xl:max-w-[520px] xl:items-end">
					<div class="flex flex-wrap gap-2">
						{#if viewMode === 'map' && !mapError}
							<div
								class="inline-flex flex-wrap items-center gap-1 rounded-[8px] border border-hairline bg-surface p-1"
								aria-label="Mode visual peta"
							>
								{#each MAP_VISUAL_MODES as mode}
									<button
										type="button"
										class={`inline-flex min-h-10 items-center gap-2 rounded-[6px] px-4 py-2 text-sm font-semibold transition-colors ${
											mapVisualMode === mode.id
												? 'bg-ink text-paper'
												: 'text-muted hover:text-ink'
										}`}
										onclick={() => handleVisualModeChange(mode.id)}
									>
										{#if mode.id === 'heatmap'}
											<ChartIcon class="size-[18px]" />
										{:else}
											<LocationIcon class="size-[18px]" />
										{/if}
										{mode.label}
									</button>
								{/each}
							</div>

							<button
								type="button"
								class={showList ? 'btn-primary' : 'btn-secondary'}
								onclick={toggleListPanel}
								title="Daftar laporan pada area aktif"
							>
								<ListCheckIcon class="size-[18px]" />
								{showList ? 'Tutup panel' : 'Buka panel'}
							</button>
						{/if}

						<button
							type="button"
							class="btn-secondary"
							onclick={toggleView}
							title={viewMode === 'map' ? 'Pindah ke mode daftar' : 'Kembali ke mode peta'}
						>
							{#if viewMode === 'map'}
								<ListCheckIcon class="size-[18px]" />
								Mode daftar
							{:else}
								<MapIcon class="size-[18px]" />
								Mode peta
							{/if}
						</button>
					</div>

					<div class="grid gap-3 sm:grid-cols-3 xl:w-full">
						<article class="jedug-card px-4 py-4">
							<div class="flex items-center gap-2 text-subtle">
								{#if mapVisualMode === 'heatmap'}
									<ChartIcon class="size-[18px]" />
								{:else}
									<LocationIcon class="size-[18px]" />
								{/if}
								<span class="surface-label">Mode aktif</span>
							</div>
							<strong class="mt-2 block text-sm font-semibold text-ink">{modeSummary.label}</strong>
							<p class="mt-1 text-xs leading-5 text-muted">{modeSummary.copy}</p>
						</article>

						<article class="jedug-card px-4 py-4">
							<div class="flex items-center gap-2 text-subtle">
								<LayersIcon class="size-[18px]" />
								<span class="surface-label">Viewport</span>
							</div>
							<strong class="mt-2 block text-sm font-semibold text-ink">{mapCountSummary}</strong>
							<p class="mt-1 text-xs leading-5 text-muted">
								Data menyesuaikan saat peta digeser atau daftar dimuat penuh.
							</p>
						</article>

						<article class="jedug-card px-4 py-4">
							<div class="flex items-center gap-2 text-subtle">
								<AddCircleIcon class="size-[18px]" />
								<span class="surface-label">Aksi cepat</span>
							</div>
							<strong class="mt-2 block text-sm font-semibold text-ink">Laporkan area sekitar</strong>
							<p class="mt-1 text-xs leading-5 text-muted">
								Gunakan CTA di bawah untuk menambah laporan tanpa keluar dari alur publik.
							</p>
						</article>
					</div>
				</div>
			</div>
		</div>
	</div>

	{#if viewMode === 'map' && !mapError}
		<div class="flex min-h-0 flex-1 px-0 pb-28 lg:px-6 lg:pb-6">
			<div class="relative flex min-h-0 flex-1 overflow-hidden border-y border-hairline bg-surface lg:rounded-[12px] lg:border">
				{#if IssueMap}
					<div class="relative min-h-0 flex-1 bg-sunken">
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
							<div class="absolute left-4 top-4 z-10 inline-flex items-center gap-2 rounded-[999px] border border-hairline bg-surface px-4 py-2 text-xs font-semibold text-muted">
								<span class="size-2 animate-pulse rounded-full bg-brand"></span>
								Memuat titik laporan...
							</div>
						{/if}

						{#if showMapInfo}
							<div class="absolute left-4 top-4 z-10 flex max-w-[min(360px,calc(100%-2rem))] flex-wrap items-center gap-2 rounded-[8px] border border-hairline bg-surface px-4 py-3 text-sm">
								<span class="inline-flex items-center gap-2 rounded-[999px] border border-hairline px-3 py-1 text-xs font-semibold text-ink">
									{#if mapVisualMode === 'heatmap'}
										<ChartIcon class="size-4" />
									{:else}
										<LocationIcon class="size-4" />
									{/if}
									{modeSummary.label}
								</span>
								<span class="text-sm font-semibold text-ink nums">{issues.length} titik</span>
								<span class="text-subtle">•</span>
								<span class="text-sm text-muted">{mapInfoText}</span>
							</div>
						{/if}

						{#if heatmapNotice}
							<div class="absolute left-4 right-4 top-20 z-10 flex items-start justify-between gap-3 rounded-[8px] border border-hairline bg-surface px-4 py-3 text-sm text-muted">
								<div class="min-w-0">
									<p class="font-semibold text-ink">Mode visual disesuaikan</p>
									<p class="mt-1 leading-6">{heatmapNotice}</p>
								</div>
								<button
									type="button"
									class="btn-ghost min-h-0 shrink-0 px-2 py-1"
									onclick={() => {
										heatmapNotice = null;
									}}
								>
									Tutup
								</button>
							</div>
						{/if}

						{#if error}
							<div class="absolute right-4 top-4 z-10 flex max-w-[min(360px,calc(100%-2rem))] items-start gap-3 rounded-[8px] border border-brand/30 bg-surface px-4 py-3 text-sm text-brand">
								<CloseCircleIcon class="mt-0.5 size-5 shrink-0" />
								<div class="min-w-0 flex-1">
									<p class="font-semibold">Data viewport belum siap</p>
									<p class="mt-1 leading-6">{error}</p>
								</div>
								<button
									type="button"
									class="btn-ghost min-h-0 shrink-0 px-2 py-1 text-brand"
									onclick={() => {
										error = null;
									}}
								>
									Tutup
								</button>
							</div>
						{/if}

						{#if showHeatmapLegend}
							<div class="absolute bottom-24 left-4 z-10 w-[min(280px,calc(100%-2rem))] rounded-[12px] border border-hairline bg-surface px-4 py-4">
								<div class="flex items-center gap-2 text-ink">
									<ChartIcon class="size-[18px] text-brand" />
									<strong class="text-sm font-semibold">Intensitas area</strong>
								</div>
								<p class="mt-2 text-xs leading-5 text-muted">
									Heatmap membaca kombinasi severity, korban, dan volume laporan di area aktif.
								</p>
								<div class="mt-3 h-2.5 rounded-full bg-[linear-gradient(90deg,#f0b847_0%,#e0732b_52%,#c5363a_82%,#7f1f22_100%)]"></div>
								<div class="mt-2 flex justify-between text-[11px] font-semibold text-muted">
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
					<div class="grid flex-1 place-items-center bg-sunken px-4">
						<LoadingState message="Menyiapkan peta..." />
					</div>
				{/if}

				{#if showList && IssueMap}
					<aside class="absolute inset-0 z-20 flex flex-col border-l border-hairline bg-surface md:static md:w-[380px] md:max-w-[42vw]">
						<div class="flex items-start justify-between gap-3 border-b border-hairline px-4 py-4">
							<div>
								<p class="surface-label">Panel laporan</p>
								<h2 class="mt-1 font-serif text-base font-semibold text-ink">Daftar lokasi pada area aktif</h2>
								<p class="mt-1 text-xs leading-5 text-muted">{issues.length} laporan siap dipindai.</p>
							</div>
							<button type="button" class="btn-icon size-10" onclick={toggleListPanel}>
								<CloseCircleIcon class="size-5" />
							</button>
						</div>

						<div class="flex-1 overflow-y-auto p-3">
							{#if issues.length === 0}
								<div class="rounded-[12px] border border-dashed border-hairline bg-sunken px-4 py-5">
									<EmptyState message="Tidak ada laporan di area ini" />
								</div>
							{:else}
								<div class="flex flex-col gap-3">
									{#each issues as issue (issue.id)}
										<button
											type="button"
											class={`block w-full rounded-[12px] text-left transition ${
												mapVisualMode === 'marker' && selectedIssue?.id === issue.id
													? 'ring-2 ring-ink ring-offset-2 ring-offset-surface'
													: ''
											}`}
											onclick={() => handleIssueSelect(issue)}
										>
											<IssueCard {issue} mode="static" />
										</button>
									{/each}
								</div>
							{/if}
						</div>
					</aside>
				{/if}
			</div>
		</div>
	{:else}
		<div class="mx-auto flex w-full max-w-[1180px] flex-1 flex-col gap-4 px-4 pb-28 pt-2 lg:px-6">
			{#if mapError}
				<div class="notice-panel flex items-start gap-3">
					<CloseCircleIcon class="mt-0.5 size-5 shrink-0 text-muted" />
					<div>
						<p class="font-semibold text-ink">Mode peta tidak tersedia</p>
						<p class="mt-1 leading-6">{mapError}. Daftar publik tetap ditampilkan agar alur pelaporan tidak terputus.</p>
					</div>
				</div>
			{/if}

			<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
				<article class="jedug-card p-5">
					<div class="flex items-center gap-2 text-subtle">
						<ListCheckIcon class="size-[18px]" />
						<span class="surface-label">Mode aktif</span>
					</div>
					<strong class="mt-2 block text-sm font-semibold text-ink">Daftar laporan publik</strong>
					<p class="mt-3 text-sm leading-6 text-muted">
						Cocok untuk memindai issue satu per satu, terutama saat jaringan atau komponen peta tidak ideal.
					</p>
				</article>

				<article class="jedug-card p-5">
					<div class="flex items-center gap-2 text-subtle">
						<LayersIcon class="size-[18px]" />
						<span class="surface-label">Ringkasan</span>
					</div>
					<strong class="mt-2 block text-sm font-semibold text-ink">{issues.length} laporan dimuat</strong>
					<p class="mt-3 text-sm leading-6 text-muted">
						Memakai endpoint publik yang sama dan menjaga shape data issue yang dipakai area lain.
					</p>
				</article>

				<article class="jedug-card p-5">
					<div class="flex items-center gap-2 text-subtle">
						<AddCircleIcon class="size-[18px]" />
						<span class="surface-label">Aksi warga</span>
					</div>
					<strong class="mt-2 block text-sm font-semibold text-ink">Tambah laporan baru</strong>
					<p class="mt-3 text-sm leading-6 text-muted">
						Tetap mobile-first dan ringan, warga bisa langsung melapor setelah meninjau daftar issue.
					</p>
				</article>
			</div>

			{#if loading}
				<LoadingState message="Memuat laporan..." />
			{:else if error}
				<ErrorState message={error} onretry={fetchList} />
			{:else if issues.length === 0}
				<EmptyState
					message="Belum ada laporan. Jadilah yang pertama melaporkan!"
					ctaHref="/lapor"
					ctaLabel="Laporkan Jalan Rusak"
				/>
			{:else}
				<div class="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3">
					{#each issues as issue (issue.id)}
						<IssueCard {issue} />
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	<div class="pointer-events-none fixed inset-x-0 bottom-0 z-30 px-4 pb-4 lg:bottom-6 lg:left-auto lg:right-6 lg:w-auto lg:px-0 lg:pb-0">
		<a href="/lapor" class="btn-primary pointer-events-auto mx-auto flex w-full max-w-[420px] lg:w-auto">
			<AddCircleIcon class="size-[18px]" />
			Laporkan Jalan Rusak
		</a>
	</div>
</div>
