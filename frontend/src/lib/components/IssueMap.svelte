<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import maplibregl from 'maplibre-gl';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import type { Issue } from '$lib/api/types';
	import type { BBox } from '$lib/utils/bbox';

	let {
		issues = [],
		selectedIssue = null,
		onbboxchange,
		onissueselect,
		onmaperror
	}: {
		issues: Issue[];
		selectedIssue: Issue | null;
		onbboxchange: (bbox: BBox) => void;
		onissueselect: (issue: Issue | null) => void;
		onmaperror?: (error: string) => void;
	} = $props();

	let mapContainer: HTMLDivElement;
	let map: maplibregl.Map | null = null;
	let markersOnMap: Map<string, maplibregl.Marker> = new Map();

	// Severity color mapping
	function getSeverityColor(severity: number, status: string): string {
		if (status === 'fixed' || status === 'archived') return '#a0aec0';
		switch (severity) {
			case 1: return '#d69e2e';
			case 2: return '#dd6b20';
			case 3: return '#e53e3e';
			case 4: return '#c53030';
			case 5: return '#9b2c2c';
			default: return '#dd6b20';
		}
	}

	function createMarkerElement(issue: Issue): HTMLDivElement {
		const el = document.createElement('div');
		const color = getSeverityColor(issue.severity_current, issue.status);
		const size = issue.severity_current >= 4 ? 16 : issue.severity_current >= 2 ? 13 : 11;
		const isFixed = issue.status === 'fixed' || issue.status === 'archived';

		el.style.cssText = `
			width: ${size}px;
			height: ${size}px;
			background: ${color};
			border: 2px solid #fff;
			border-radius: 50%;
			cursor: pointer;
			box-shadow: 0 1px 4px rgba(0,0,0,0.3);
			transition: transform 0.15s;
			${isFixed ? 'opacity: 0.5;' : ''}
		`;
		el.addEventListener('mouseenter', () => {
			el.style.transform = 'scale(1.3)';
		});
		el.addEventListener('mouseleave', () => {
			el.style.transform = 'scale(1)';
		});
		return el;
	}

	function emitBBox() {
		if (!map) return;
		const bounds = map.getBounds();
		const bbox: BBox = [
			bounds.getWest(),
			bounds.getSouth(),
			bounds.getEast(),
			bounds.getNorth()
		];
		onbboxchange(bbox);
	}

	function syncMarkers() {
		if (!map) return;

		const currentIds = new Set(issues.map((i) => i.id));
		// Remove markers no longer in dataset
		for (const [id, marker] of markersOnMap) {
			if (!currentIds.has(id)) {
				marker.remove();
				markersOnMap.delete(id);
			}
		}

		// Add/update markers
		for (const issue of issues) {
			if (markersOnMap.has(issue.id)) continue;

			const el = createMarkerElement(issue);
			el.addEventListener('click', (e) => {
				e.stopPropagation();
				onissueselect(issue);
			});

			const marker = new maplibregl.Marker({ element: el })
				.setLngLat([issue.longitude, issue.latitude])
				.addTo(map!);

			markersOnMap.set(issue.id, marker);
		}
	}

	// React to issues changes
	$effect(() => {
		if (map && issues) {
			syncMarkers();
		}
	});

	// React to selectedIssue to highlight / fly to
	$effect(() => {
		if (map && selectedIssue) {
			map.flyTo({
				center: [selectedIssue.longitude, selectedIssue.latitude],
				zoom: Math.max(map.getZoom(), 14),
				duration: 500
			});
		}
	});

	onMount(() => {
		try {
			map = new maplibregl.Map({
				container: mapContainer,
				style: 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json',
				center: [110.4, -7.0], // Central Java default
				zoom: 7,
				attributionControl: false
			});

			map.addControl(
				new maplibregl.AttributionControl({ compact: true }),
				'bottom-right'
			);

			map.addControl(
				new maplibregl.NavigationControl({ showCompass: false }),
				'top-right'
			);

			map.addControl(
				new maplibregl.GeolocateControl({
					positionOptions: { enableHighAccuracy: true },
					trackUserLocation: false
				}),
				'top-right'
			);

			map.on('load', emitBBox);
			map.on('moveend', emitBBox);

			// Click on map background deselects
			map.on('click', () => {
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
		markersOnMap.clear();
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
