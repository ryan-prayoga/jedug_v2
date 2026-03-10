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
		onmaperror,
		onmapready
	}: {
		issues: Issue[];
		selectedIssue: Issue | null;
		onbboxchange: (bbox: BBox) => void;
		onissueselect: (issue: Issue | null) => void;
		onmaperror?: (error: string) => void;
		onmapready?: () => void;
	} = $props();

	let mapContainer: HTMLDivElement;
	let map: maplibregl.Map | null = null;
	let mapReady = $state(false); // reactive flag so $effect can track it
	let markersOnMap: Map<
		string,
		{ marker: maplibregl.Marker; el: HTMLDivElement; visual: HTMLDivElement; issue: Issue }
	> = new Map();
	let didAutoCenter = false;

	const DEFAULT_CENTER: [number, number] = [110.4, -7.0];
	const DEFAULT_ZOOM = 7;
	const USER_ZOOM = 15;

	// ── Severity → color ──────────────────────────────────────────
	function getSeverityColor(severity: number, status: string): string {
		if (status === 'fixed' || status === 'archived') return '#94A3B8';
		switch (severity) {
			case 1: return '#F6C453';
			case 2: return '#F97316';
			case 3: return '#DC2626';
			case 4: return '#DC2626';
			case 5: return '#991B1B';
			default: return '#F97316';
		}
	}

	// ── Create marker DOM element ─────────────────────────────────
	function createMarkerElement(issue: Issue): HTMLDivElement {
		const wrapper = document.createElement('div');
		const color = getSeverityColor(issue.severity_current, issue.status);
		const isFixed = issue.status === 'fixed' || issue.status === 'archived';
		// Larger touch-target, inner visible dot
		const size = 36;
		const dotSize = issue.severity_current >= 3 ? 20 : issue.severity_current >= 2 ? 16 : 14;

		wrapper.className = 'jedug-marker';
		wrapper.style.cssText = `
			width: ${size}px;
			height: ${size}px;
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;
			border-radius: 50%;
			pointer-events: auto;
		`;

		const visual = document.createElement('div');
		visual.className = 'jedug-marker-visual';
		visual.style.cssText = `
			width: 100%;
			height: 100%;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 50%;
			transition: transform 0.15s ease, filter 0.15s ease;
			transform-origin: center center;
		`;

		const dot = document.createElement('div');
		dot.style.cssText = `
			width: ${dotSize}px;
			height: ${dotSize}px;
			background: ${color};
			border: 2.5px solid #fff;
			border-radius: 50%;
			box-shadow: 0 2px 6px rgba(0,0,0,0.35);
			${isFixed ? 'opacity: 0.45;' : ''}
		`;
		visual.appendChild(dot);
		wrapper.appendChild(visual);

		wrapper.addEventListener('mouseenter', () => {
			visual.style.transform = 'scale(1.25)';
			wrapper.style.zIndex = '10';
		});
		wrapper.addEventListener('mouseleave', () => {
			if (!wrapper.classList.contains('selected')) {
				visual.style.transform = 'scale(1)';
				wrapper.style.zIndex = '';
			}
		});
		(wrapper as HTMLDivElement & { __visual?: HTMLDivElement }).__visual = visual;

		return wrapper;
	}

	// ── Highlight selected marker ─────────────────────────────────
	function updateSelectedMarkerStyle(selectedId: string | null) {
		for (const [id, entry] of markersOnMap) {
			const el = entry.el;
			const visual = entry.visual;
			if (id === selectedId) {
				el.classList.add('selected');
				visual.style.transform = 'scale(1.5)';
				el.style.zIndex = '20';
				visual.style.filter = 'drop-shadow(0 0 6px rgba(229,72,77,0.5))';
			} else {
				el.classList.remove('selected');
				visual.style.transform = 'scale(1)';
				el.style.zIndex = '';
				visual.style.filter = '';
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
	function syncMarkers(issueList: Issue[]) {
		if (!map) return;

		console.log('[IssueMap] syncMarkers called, count:', issueList.length);
		if (issueList.length > 0) {
			const sample = issueList[0];
			console.log('[IssueMap] sample issue:', sample.id, 'lat:', sample.latitude, 'lng:', sample.longitude, 'sev:', sample.severity_current);
		}

		const currentIds = new Set(issueList.map((i) => i.id));

		// Remove markers no longer in dataset
		for (const [id, entry] of markersOnMap) {
			if (!currentIds.has(id)) {
				entry.marker.remove();
				markersOnMap.delete(id);
			}
		}

		let added = 0;
		for (const issue of issueList) {
			if (issue.status === 'hidden' || issue.status === 'rejected') continue;
			// Use typeof check: lat/lng could be 0 which is falsy but valid
			if (typeof issue.latitude !== 'number' || typeof issue.longitude !== 'number') continue;
			if (markersOnMap.has(issue.id)) continue;

			const el = createMarkerElement(issue);
			const visual = (el as HTMLDivElement & { __visual?: HTMLDivElement }).__visual ?? el;
			el.addEventListener('click', (e) => {
				e.stopPropagation();
				onissueselect(issue);
			});

			const marker = new maplibregl.Marker({ element: el, anchor: 'center' })
				.setLngLat([issue.longitude, issue.latitude])
				.addTo(map!);

			markersOnMap.set(issue.id, { marker, el, visual, issue });
			added++;
		}
		console.log('[IssueMap] markers added:', added, 'total on map:', markersOnMap.size);
	}

	// ── React to issues changes ───────────────────────────────────
	// CRITICAL: read both reactive deps (mapReady, issues) BEFORE any
	// conditional so Svelte 5 $effect always tracks them.
	$effect(() => {
		const ready = mapReady;
		const issueList = issues;
		if (ready && map) {
			syncMarkers(issueList);
		}
	});

	// ── React to selectedIssue ────────────────────────────────────
	$effect(() => {
		const sel = selectedIssue;
		const ready = mapReady;
		updateSelectedMarkerStyle(sel?.id ?? null);
		if (ready && map && sel) {
			map.flyTo({
				center: [sel.longitude, sel.latitude],
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
				console.log('[IssueMap] map style loaded');
				mapReady = true; // ← triggers $effect
				onmapready?.();
				emitBBox();
				tryAutoCenter();
			});

			map.on('moveend', emitBBox);

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
