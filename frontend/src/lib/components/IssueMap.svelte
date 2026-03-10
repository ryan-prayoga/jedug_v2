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
	let markersOnMap: Map<string, { marker: maplibregl.Marker; el: HTMLDivElement; issue: Issue }> = new Map();
	let didAutoCenter = false;

	// Default center: Jakarta / Central Java
	const DEFAULT_CENTER: [number, number] = [110.4, -7.0];
	const DEFAULT_ZOOM = 7;
	const USER_ZOOM = 15;

	// ── Severity → color ──────────────────────────────────────────
	function getSeverityColor(severity: number, status: string): string {
		if (status === 'fixed' || status === 'archived') return '#94a3b8'; // slate-400
		switch (severity) {
			case 1: return '#eab308'; // yellow-500
			case 2: return '#f97316'; // orange-500
			case 3: return '#ef4444'; // red-500
			case 4: return '#dc2626'; // red-600
			case 5: return '#991b1b'; // red-900
			default: return '#f97316';
		}
	}

	function getSeverityRing(severity: number): string {
		if (severity >= 4) return 'rgba(220,38,38,0.25)';
		if (severity >= 3) return 'rgba(239,68,68,0.2)';
		return 'transparent';
	}

	// ── Create marker DOM element ─────────────────────────────────
	function createMarkerElement(issue: Issue): HTMLDivElement {
		const wrapper = document.createElement('div');
		const color = getSeverityColor(issue.severity_current, issue.status);
		const isFixed = issue.status === 'fixed' || issue.status === 'archived';
		const ring = getSeverityRing(issue.severity_current);
		const size = 28; // touch-friendly minimum
		const dotSize = issue.severity_current >= 4 ? 18 : issue.severity_current >= 2 ? 15 : 13;

		wrapper.className = 'jedug-marker';
		wrapper.dataset.issueId = issue.id;
		wrapper.style.cssText = `
			width: ${size}px;
			height: ${size}px;
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;
			background: ${ring};
			border-radius: 50%;
			transition: transform 0.15s ease;
		`;

		const dot = document.createElement('div');
		dot.style.cssText = `
			width: ${dotSize}px;
			height: ${dotSize}px;
			background: ${color};
			border: 2.5px solid #fff;
			border-radius: 50%;
			box-shadow: 0 1px 6px rgba(0,0,0,0.35);
			${isFixed ? 'opacity: 0.55;' : ''}
		`;
		wrapper.appendChild(dot);

		// Hover/touch feedback
		wrapper.addEventListener('mouseenter', () => {
			wrapper.style.transform = 'scale(1.25)';
			wrapper.style.zIndex = '10';
		});
		wrapper.addEventListener('mouseleave', () => {
			if (!wrapper.classList.contains('selected')) {
				wrapper.style.transform = 'scale(1)';
				wrapper.style.zIndex = '';
			}
		});

		return wrapper;
	}

	// ── Highlight selected marker ─────────────────────────────────
	function updateSelectedMarkerStyle(selectedId: string | null) {
		for (const [id, entry] of markersOnMap) {
			const el = entry.el;
			if (id === selectedId) {
				el.classList.add('selected');
				el.style.transform = 'scale(1.35)';
				el.style.zIndex = '20';
				// Add selected ring
				el.style.boxShadow = '0 0 0 3px rgba(239,68,68,0.4)';
			} else {
				el.classList.remove('selected');
				el.style.transform = 'scale(1)';
				el.style.zIndex = '';
				el.style.boxShadow = '';
			}
		}
	}

	// ── Emit bounding box to parent ───────────────────────────────
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

	// ── Sync markers with current issues array ────────────────────
	function syncMarkers() {
		if (!map) return;

		const currentIds = new Set(issues.map((i) => i.id));

		// Remove markers no longer in dataset
		for (const [id, entry] of markersOnMap) {
			if (!currentIds.has(id)) {
				entry.marker.remove();
				markersOnMap.delete(id);
			}
		}

		// Add new markers
		for (const issue of issues) {
			// Skip hidden/rejected (safety — backend already filters)
			if (issue.status === 'hidden' || issue.status === 'rejected') continue;
			if (!issue.latitude || !issue.longitude) continue;

			if (markersOnMap.has(issue.id)) continue;

			const el = createMarkerElement(issue);
			el.addEventListener('click', (e) => {
				e.stopPropagation();
				onissueselect(issue);
			});

			const marker = new maplibregl.Marker({ element: el, anchor: 'center' })
				.setLngLat([issue.longitude, issue.latitude])
				.addTo(map!);

			markersOnMap.set(issue.id, { marker, el, issue });
		}

		// Ensure selected state is applied
		updateSelectedMarkerStyle(selectedIssue?.id ?? null);
	}

	// ── React to issues changes ───────────────────────────────────
	$effect(() => {
		if (map && issues) {
			syncMarkers();
		}
	});

	// ── React to selectedIssue ────────────────────────────────────
	$effect(() => {
		const selId = selectedIssue?.id ?? null;
		updateSelectedMarkerStyle(selId);
		if (map && selectedIssue) {
			map.flyTo({
				center: [selectedIssue.longitude, selectedIssue.latitude],
				zoom: Math.max(map.getZoom(), 14),
				duration: 500
			});
		}
	});

	// ── Auto-geolocation on mount ─────────────────────────────────
	function tryAutoCenter() {
		if (didAutoCenter || !map) return;
		if (!navigator.geolocation) return;

		navigator.geolocation.getCurrentPosition(
			(pos) => {
				if (didAutoCenter || !map) return;
				didAutoCenter = true;
				map.flyTo({
					center: [pos.coords.longitude, pos.coords.latitude],
					zoom: USER_ZOOM,
					duration: 1200
				});
			},
			() => {
				// Permission denied or error — stay at default center
				didAutoCenter = true;
			},
			{ enableHighAccuracy: true, timeout: 8000, maximumAge: 60000 }
		);
	}

	// ── Lifecycle ─────────────────────────────────────────────────
	onMount(() => {
		try {
			map = new maplibregl.Map({
				container: mapContainer,
				style: 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json',
				center: DEFAULT_CENTER,
				zoom: DEFAULT_ZOOM,
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

			map.on('load', () => {
				emitBBox();
				// Auto-center to user location once the map is ready
				tryAutoCenter();
			});

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

	/* Ensure markers are above map tiles */
	.map-wrapper :global(.maplibregl-marker) {
		z-index: 1;
	}

	.map-wrapper :global(.maplibregl-marker:hover) {
		z-index: 10 !important;
	}
</style>
