<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import maplibregl, { type MapLayerMouseEvent } from 'maplibre-gl';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import type { Issue } from '$lib/api/types';
	import type { BBox } from '$lib/utils/bbox';
	import { getIssueHeatWeight, type MapVisualMode } from '$lib/utils/issue-heatmap';

	let {
		issues = [],
		selectedIssue = null,
		visualMode = 'marker',
		onbboxchange,
		onissueselect,
		onmaperror,
		onmapready,
		onvisualmodefallback
	}: {
		issues: Issue[];
		selectedIssue: Issue | null;
		visualMode: MapVisualMode;
		onbboxchange: (bbox: BBox) => void;
		onissueselect: (issue: Issue | null) => void;
		onmaperror?: (error: string) => void;
		onmapready?: () => void;
		onvisualmodefallback?: (mode: MapVisualMode, message: string) => void;
	} = $props();

	type IssueFeature = {
		type: 'Feature';
		geometry: {
			type: 'Point';
			coordinates: [number, number];
		};
		properties: {
			id: string;
			status: string;
			severity_current: number;
			casualty_count: number;
			submission_count: number;
			heat_weight: number;
		};
	};

	type IssueFeatureCollection = {
		type: 'FeatureCollection';
		features: IssueFeature[];
	};

	let mapContainer: HTMLDivElement;
	let map: maplibregl.Map | null = null;
	let geolocateControl: maplibregl.GeolocateControl | null = null;
	let mapReady = $state(false);
	let didAutoGeolocate = false;
	let clusteringEnabled = $state(true);
	let heatmapAvailable = $state(true);
	let issueByID = new Map<string, Issue>();

	const DEFAULT_CENTER: [number, number] = [110.4, -7.0];
	const DEFAULT_ZOOM = 7;
	const USER_ZOOM = 15;

	const ISSUE_SOURCE_ID = 'jedug-issues-source';
	const HEATMAP_SOURCE_ID = 'jedug-heatmap-source';
	const CLUSTER_CIRCLE_LAYER_ID = 'jedug-cluster-circles';
	const CLUSTER_COUNT_LAYER_ID = 'jedug-cluster-counts';
	const UNCLUSTERED_HIT_LAYER_ID = 'jedug-unclustered-hit';
	const UNCLUSTERED_BASE_LAYER_ID = 'jedug-unclustered-markers';
	const SELECTED_GLOW_LAYER_ID = 'jedug-selected-glow';
	const SELECTED_CORE_LAYER_ID = 'jedug-selected-core';
	const HEATMAP_DENSITY_LAYER_ID = 'jedug-heatmap-density';
	const HEATMAP_POINT_LAYER_ID = 'jedug-heatmap-points';

	const EMPTY_FEATURE_COLLECTION: IssueFeatureCollection = {
		type: 'FeatureCollection',
		features: []
	};

	const markerColorExpression: any = [
		'case',
		[
			'any',
			['==', ['get', 'status'], 'fixed'],
			['==', ['get', 'status'], 'archived']
		],
		'#94A3B8',
		['>=', ['to-number', ['get', 'severity_current']], 5],
		'#991B1B',
		['>=', ['to-number', ['get', 'severity_current']], 3],
		'#DC2626',
		['>=', ['to-number', ['get', 'severity_current']], 2],
		'#F97316',
		'#F6C453'
	];

	const markerRadiusExpression: any = [
		'case',
		['>=', ['to-number', ['get', 'severity_current']], 3],
		10,
		['>=', ['to-number', ['get', 'severity_current']], 2],
		8,
		7
	];

	const markerOpacityExpression: any = [
		'case',
		[
			'any',
			['==', ['get', 'status'], 'fixed'],
			['==', ['get', 'status'], 'archived']
		],
		0.45,
		1
	];

	const clusterCircleColorExpression: any = [
		'step',
		['get', 'point_count'],
		'#FCA5A5',
		20,
		'#FB7185',
		60,
		'#E5484D',
		120,
		'#BE123C'
	];

	const clusterCircleRadiusExpression: any = [
		'step',
		['get', 'point_count'],
		18,
		20,
		24,
		60,
		30,
		120,
		36
	];

	const clusterTextSizeExpression: any = [
		'step',
		['get', 'point_count'],
		12,
		20,
		13,
		60,
		14,
		120,
		15
	];

	const heatmapWeightExpression: any = ['coalesce', ['to-number', ['get', 'heat_weight']], 0.1];

	const heatmapPointColorExpression: any = [
		'interpolate',
		['linear'],
		heatmapWeightExpression,
		0.1,
		'rgba(246, 196, 83, 0.24)',
		0.55,
		'rgba(249, 115, 22, 0.34)',
		1,
		'rgba(229, 72, 77, 0.48)'
	];

	function buildFeatureCollection(issueList: Issue[]): IssueFeatureCollection {
		const features: IssueFeature[] = [];
		const issueMap = new Map<string, Issue>();

		for (const issue of issueList) {
			if (issue.status === 'hidden' || issue.status === 'rejected') continue;
			if (typeof issue.latitude !== 'number' || typeof issue.longitude !== 'number') continue;

			issueMap.set(issue.id, issue);
			features.push({
				type: 'Feature',
				geometry: {
					type: 'Point',
					coordinates: [issue.longitude, issue.latitude]
				},
				properties: {
					id: issue.id,
					status: issue.status,
					severity_current: issue.severity_current,
					casualty_count: issue.casualty_count,
					submission_count: issue.submission_count,
					heat_weight: getIssueHeatWeight(issue)
				}
			});
		}

		issueByID = issueMap;
		return {
			type: 'FeatureCollection',
			features
		};
	}

	function markerLayerIDs(): string[] {
		return [
			CLUSTER_CIRCLE_LAYER_ID,
			CLUSTER_COUNT_LAYER_ID,
			UNCLUSTERED_HIT_LAYER_ID,
			UNCLUSTERED_BASE_LAYER_ID,
			SELECTED_GLOW_LAYER_ID,
			SELECTED_CORE_LAYER_ID
		];
	}

	function heatmapLayerIDs(): string[] {
		return [HEATMAP_DENSITY_LAYER_ID, HEATMAP_POINT_LAYER_ID];
	}

	function getSource(sourceID: string): maplibregl.GeoJSONSource | null {
		if (!map) return null;
		const source = map.getSource(sourceID);
		if (!source) return null;
		return source as maplibregl.GeoJSONSource;
	}

	function setSourceData(sourceID: string, featureCollection: IssueFeatureCollection) {
		const source = getSource(sourceID);
		if (!source) return;
		source.setData(featureCollection as any);
	}

	function setIssueSourceData(issueList: Issue[]) {
		const featureCollection = buildFeatureCollection(issueList);
		setSourceData(ISSUE_SOURCE_ID, featureCollection);
		setSourceData(HEATMAP_SOURCE_ID, featureCollection);
	}

	function unclusteredFilter(): any {
		return ['!', ['has', 'point_count']];
	}

	function buildUnclusteredBaseFilter(selectedID: string | null): any {
		const base = unclusteredFilter();
		if (!selectedID) return base;
		return ['all', base, ['!=', ['get', 'id'], selectedID]];
	}

	function buildSelectedFilter(selectedID: string | null): any {
		const base = unclusteredFilter();
		if (!selectedID) {
			return ['all', base, ['==', ['get', 'id'], '__no_selected_issue__']];
		}
		return ['all', base, ['==', ['get', 'id'], selectedID]];
	}

	function removeLayerGroup(layerIDs: string[]) {
		if (!map) return;
		for (const layerID of layerIDs) {
			if (map.getLayer(layerID)) {
				map.removeLayer(layerID);
			}
		}
	}

	function removeMarkerLayersAndSource() {
		if (!map) return;
		removeLayerGroup(markerLayerIDs());
		if (map.getSource(ISSUE_SOURCE_ID)) {
			map.removeSource(ISSUE_SOURCE_ID);
		}
	}

	function removeHeatmapLayersAndSource() {
		if (!map) return;
		removeLayerGroup(heatmapLayerIDs());
		if (map.getSource(HEATMAP_SOURCE_ID)) {
			map.removeSource(HEATMAP_SOURCE_ID);
		}
	}

	function addMarkerLayers(enableClustering: boolean) {
		if (!map) return;

		map.addSource(ISSUE_SOURCE_ID, {
			type: 'geojson',
			data: EMPTY_FEATURE_COLLECTION as any,
			cluster: enableClustering,
			clusterMaxZoom: 13,
			clusterRadius: 52
		});

		if (enableClustering) {
			map.addLayer({
				id: CLUSTER_CIRCLE_LAYER_ID,
				type: 'circle',
				source: ISSUE_SOURCE_ID,
				filter: ['has', 'point_count'],
				paint: {
					'circle-color': clusterCircleColorExpression,
					'circle-radius': clusterCircleRadiusExpression,
					'circle-opacity': 0.92,
					'circle-stroke-color': '#FFFFFF',
					'circle-stroke-width': 2
				}
			});

			map.addLayer({
				id: CLUSTER_COUNT_LAYER_ID,
				type: 'symbol',
				source: ISSUE_SOURCE_ID,
				filter: ['has', 'point_count'],
				layout: {
					'text-field': ['get', 'point_count_abbreviated'],
					'text-size': clusterTextSizeExpression,
					'text-font': ['Open Sans Bold']
				},
				paint: {
					'text-color': '#FFFFFF',
					'text-halo-color': '#7F1D1D',
					'text-halo-width': 1
				}
			});
		}

		map.addLayer({
			id: UNCLUSTERED_HIT_LAYER_ID,
			type: 'circle',
			source: ISSUE_SOURCE_ID,
			filter: unclusteredFilter(),
			paint: {
				'circle-radius': 18,
				'circle-color': '#000000',
				'circle-opacity': 0
			}
		});

		map.addLayer({
			id: UNCLUSTERED_BASE_LAYER_ID,
			type: 'circle',
			source: ISSUE_SOURCE_ID,
			filter: buildUnclusteredBaseFilter(null),
			paint: {
				'circle-color': markerColorExpression,
				'circle-radius': markerRadiusExpression,
				'circle-opacity': markerOpacityExpression,
				'circle-stroke-color': '#FFFFFF',
				'circle-stroke-width': 2.5
			}
		});

		map.addLayer({
			id: SELECTED_GLOW_LAYER_ID,
			type: 'circle',
			source: ISSUE_SOURCE_ID,
			filter: buildSelectedFilter(null),
			paint: {
				'circle-color': '#E5484D',
				'circle-radius': ['*', markerRadiusExpression, 2],
				'circle-opacity': 0.35,
				'circle-blur': 0.7
			}
		});

		map.addLayer({
			id: SELECTED_CORE_LAYER_ID,
			type: 'circle',
			source: ISSUE_SOURCE_ID,
			filter: buildSelectedFilter(null),
			paint: {
				'circle-color': markerColorExpression,
				'circle-radius': ['*', markerRadiusExpression, 1.5],
				'circle-opacity': markerOpacityExpression,
				'circle-stroke-color': '#FFFFFF',
				'circle-stroke-width': 3
			}
		});
	}

	function addHeatmapLayers() {
		if (!map) return;

		map.addSource(HEATMAP_SOURCE_ID, {
			type: 'geojson',
			data: EMPTY_FEATURE_COLLECTION as any
		});

		map.addLayer({
			id: HEATMAP_DENSITY_LAYER_ID,
			type: 'heatmap',
			source: HEATMAP_SOURCE_ID,
			maxzoom: 15,
			paint: {
				'heatmap-weight': heatmapWeightExpression,
				'heatmap-intensity': ['interpolate', ['linear'], ['zoom'], 4, 0.55, 8, 0.9, 12, 1.35, 15, 1.7],
				'heatmap-radius': ['interpolate', ['linear'], ['zoom'], 4, 18, 8, 28, 12, 40, 15, 54],
				'heatmap-opacity': ['interpolate', ['linear'], ['zoom'], 4, 0.72, 10, 0.88, 15, 0.52],
				'heatmap-color': [
					'interpolate',
					['linear'],
					['heatmap-density'],
					0,
					'rgba(246, 196, 83, 0)',
					0.14,
					'rgba(246, 196, 83, 0.34)',
					0.4,
					'rgba(249, 115, 22, 0.56)',
					0.72,
					'rgba(229, 72, 77, 0.82)',
					1,
					'rgba(153, 27, 27, 0.94)'
				]
			}
		} as any);

		map.addLayer({
			id: HEATMAP_POINT_LAYER_ID,
			type: 'circle',
			source: HEATMAP_SOURCE_ID,
			minzoom: 11,
			paint: {
				'circle-color': heatmapPointColorExpression,
				'circle-radius': ['interpolate', ['linear'], ['zoom'], 11, 4, 13, 6, 15, 9],
				'circle-opacity': ['interpolate', ['linear'], ['zoom'], 11, 0, 12, 0.26, 15, 0.42],
				'circle-stroke-color': 'rgba(255, 255, 255, 0.55)',
				'circle-stroke-width': ['interpolate', ['linear'], ['zoom'], 11, 0, 13, 0.8, 15, 1.2]
			}
		} as any);
	}

	function setLayerVisibility(layerIDs: string[], visible: boolean) {
		if (!map) return;
		for (const layerID of layerIDs) {
			if (!map.getLayer(layerID)) continue;
			map.setLayoutProperty(layerID, 'visibility', visible ? 'visible' : 'none');
		}
	}

	function updateVisualMode(mode: MapVisualMode) {
		if (!map) return;
		const showHeatmap = mode === 'heatmap' && heatmapAvailable;
		setLayerVisibility(heatmapLayerIDs(), showHeatmap);
		setLayerVisibility(markerLayerIDs(), !showHeatmap);
		if (showHeatmap) {
			clearPointerCursor();
		}
	}

	function updateSelectedLayer(selectedID: string | null) {
		if (!map) return;
		if (!map.getSource(ISSUE_SOURCE_ID)) return;

		if (map.getLayer(UNCLUSTERED_BASE_LAYER_ID)) {
			map.setFilter(UNCLUSTERED_BASE_LAYER_ID, buildUnclusteredBaseFilter(selectedID));
		}
		const selectedFilter = buildSelectedFilter(selectedID);
		if (map.getLayer(SELECTED_GLOW_LAYER_ID)) {
			map.setFilter(SELECTED_GLOW_LAYER_ID, selectedFilter);
		}
		if (map.getLayer(SELECTED_CORE_LAYER_ID)) {
			map.setFilter(SELECTED_CORE_LAYER_ID, selectedFilter);
		}
	}

	function resolveIssueFromEvent(event: MapLayerMouseEvent): Issue | null {
		const feature = event.features?.[0];
		const props = feature?.properties as Record<string, unknown> | undefined;
		const rawID = props?.id;
		const issueID = typeof rawID === 'string' ? rawID : typeof rawID === 'number' ? String(rawID) : null;
		if (!issueID) return null;
		return issueByID.get(issueID) ?? null;
	}

	function handleIssueClick(event: MapLayerMouseEvent) {
		if (visualMode !== 'marker') return;
		const issue = resolveIssueFromEvent(event);
		if (!issue) return;
		onissueselect(issue);
	}

	function handleClusterClick(event: MapLayerMouseEvent) {
		if (!map || !clusteringEnabled || visualMode !== 'marker') return;
		const feature = event.features?.[0];
		if (!feature) return;

		const props = feature.properties as Record<string, unknown> | undefined;
		const rawClusterID = props?.cluster_id;
		const clusterID =
			typeof rawClusterID === 'number'
				? rawClusterID
				: typeof rawClusterID === 'string'
					? Number(rawClusterID)
					: NaN;
		if (!Number.isFinite(clusterID)) return;

		const geometry = feature.geometry;
		if (!geometry || geometry.type !== 'Point') return;
		const [longitude, latitude] = geometry.coordinates as [number, number];

		const source = getSource(ISSUE_SOURCE_ID);
		if (!source) return;

		void source
			.getClusterExpansionZoom(clusterID)
			.then((zoom) => {
				if (!map) return;
				map.easeTo({
					center: [longitude, latitude],
					zoom: Math.min(zoom + 0.35, 18),
					duration: 450
				});
			})
			.catch(() => {
				// Keep map interaction resilient; if expansion zoom fails we do nothing.
			});
	}

	function interactiveLayerIDs(mode: MapVisualMode): string[] {
		if (mode !== 'marker') return [];
		const layerIDs = [UNCLUSTERED_HIT_LAYER_ID, SELECTED_CORE_LAYER_ID];
		if (clusteringEnabled) {
			layerIDs.unshift(CLUSTER_CIRCLE_LAYER_ID, CLUSTER_COUNT_LAYER_ID);
		}
		return layerIDs;
	}

	function isInteractiveFeatureClick(event: maplibregl.MapMouseEvent): boolean {
		if (!map) return false;
		const activeLayers = interactiveLayerIDs(visualMode).filter((layerID) => Boolean(map?.getLayer(layerID)));
		if (activeLayers.length === 0) return false;
		return map.queryRenderedFeatures(event.point, { layers: activeLayers }).length > 0;
	}

	function setPointerCursor() {
		if (!map) return;
		if (visualMode !== 'marker') return;
		map.getCanvas().style.cursor = 'pointer';
	}

	function clearPointerCursor() {
		if (!map) return;
		map.getCanvas().style.cursor = '';
	}

	function registerLayerInteractions() {
		if (!map) return;

		const pointerLayers = [UNCLUSTERED_HIT_LAYER_ID, SELECTED_CORE_LAYER_ID];
		if (clusteringEnabled) {
			pointerLayers.push(CLUSTER_CIRCLE_LAYER_ID, CLUSTER_COUNT_LAYER_ID);
		}

		for (const layerID of pointerLayers) {
			if (!map.getLayer(layerID)) continue;
			map.on('mouseenter', layerID, setPointerCursor);
			map.on('mouseleave', layerID, clearPointerCursor);
		}

		if (map.getLayer(UNCLUSTERED_HIT_LAYER_ID)) {
			map.on('click', UNCLUSTERED_HIT_LAYER_ID, handleIssueClick);
		}
		if (map.getLayer(SELECTED_CORE_LAYER_ID)) {
			map.on('click', SELECTED_CORE_LAYER_ID, handleIssueClick);
		}

		if (clusteringEnabled) {
			if (map.getLayer(CLUSTER_CIRCLE_LAYER_ID)) {
				map.on('click', CLUSTER_CIRCLE_LAYER_ID, handleClusterClick);
			}
			if (map.getLayer(CLUSTER_COUNT_LAYER_ID)) {
				map.on('click', CLUSTER_COUNT_LAYER_ID, handleClusterClick);
			}
		}
	}

	function setupIssueRendering(): boolean {
		if (!map) return false;

		removeMarkerLayersAndSource();
		removeHeatmapLayersAndSource();

		heatmapAvailable = true;
		try {
			addHeatmapLayers();
		} catch (heatmapError) {
			heatmapAvailable = false;
			console.error('[IssueMap] heatmap setup failed, keeping marker mode available', heatmapError);
		}

		try {
			addMarkerLayers(true);
			clusteringEnabled = true;
			registerLayerInteractions();
			updateVisualMode(visualMode);
			return true;
		} catch (clusterError) {
			console.error('[IssueMap] clustering setup failed, falling back to unclustered markers', clusterError);
		}

		removeMarkerLayersAndSource();
		try {
			addMarkerLayers(false);
			clusteringEnabled = false;
			registerLayerInteractions();
			updateVisualMode(visualMode);
			return true;
		} catch (fallbackError) {
			onmaperror?.(fallbackError instanceof Error ? fallbackError.message : 'Peta gagal menyiapkan marker');
			return false;
		}
	}

	function emitBBox() {
		if (!map) return;
		const bounds = map.getBounds();
		const bbox: BBox = [bounds.getWest(), bounds.getSouth(), bounds.getEast(), bounds.getNorth()];
		onbboxchange(bbox);
	}

	function tryInitialGeolocate() {
		if (didAutoGeolocate || !geolocateControl) return;
		didAutoGeolocate = true;

		// Trigger once at initial load so blue dot appears without manual click.
		void geolocateControl.trigger();
	}

	$effect(() => {
		const ready = mapReady;
		const issueList = issues;
		if (!ready) return;
		setIssueSourceData(issueList);
	});

	$effect(() => {
		const ready = mapReady;
		const mode = visualMode;
		if (!ready) return;
		if (mode === 'heatmap' && !heatmapAvailable) {
			onvisualmodefallback?.('marker', 'Heatmap gagal dimuat di perangkat ini. Mode marker tetap aktif.');
			return;
		}
		updateVisualMode(mode);
	});

	$effect(() => {
		const ready = mapReady;
		const selected = selectedIssue;
		const mode = visualMode;

		updateSelectedLayer(mode === 'marker' ? selected?.id ?? null : null);
		if (!ready || !map || !selected || mode !== 'marker') return;

		map.flyTo({
			center: [selected.longitude, selected.latitude],
			zoom: Math.max(map.getZoom(), 14),
			duration: 500
		});
	});

	onMount(() => {
		try {
			map = new maplibregl.Map({
				container: mapContainer,
				style: 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json',
				center: DEFAULT_CENTER,
				zoom: DEFAULT_ZOOM,
				attributionControl: false
			});

			map.addControl(new maplibregl.AttributionControl({ compact: true }), 'bottom-right');
			map.addControl(new maplibregl.NavigationControl({ showCompass: false }), 'top-right');
			geolocateControl = new maplibregl.GeolocateControl({
				positionOptions: { enableHighAccuracy: true, timeout: 12000, maximumAge: 60000 },
				trackUserLocation: false,
				showUserLocation: true,
				fitBoundsOptions: {
					maxZoom: USER_ZOOM,
					duration: 1000
				}
			});
			map.addControl(geolocateControl, 'top-right');

			map.on('load', () => {
				if (!setupIssueRendering()) {
					return;
				}

				mapReady = true;
				onmapready?.();
				setIssueSourceData(issues);
				updateSelectedLayer(visualMode === 'marker' ? selectedIssue?.id ?? null : null);
				updateVisualMode(visualMode);
				emitBBox();
				tryInitialGeolocate();
			});

			map.on('moveend', emitBBox);
			map.on('click', (event) => {
				if (isInteractiveFeatureClick(event)) {
					return;
				}
				onissueselect(null);
			});
		} catch (e) {
			onmaperror?.(e instanceof Error ? e.message : 'Peta gagal dimuat');
		}
	});

	onDestroy(() => {
		if (map) {
			map.remove();
			map = null;
		}
		geolocateControl = null;
		issueByID.clear();
	});
</script>

<div class="map-wrapper" bind:this={mapContainer}></div>

<style>
	.map-wrapper {
		width: 100%;
		height: 100%;
		min-height: 300px;
	}
</style>
