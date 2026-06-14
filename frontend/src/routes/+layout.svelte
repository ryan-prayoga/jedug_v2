<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import AppHeader from '$lib/components/AppHeader.svelte';
	import ConsentSheet from '$lib/components/ConsentSheet.svelte';
	import { DangerIcon } from '$lib/icons';
	import { getAnonToken, isConsentGiven, setConsentGiven } from '$lib/utils/storage';
	import { recordConsent } from '$lib/api/device';
	import { ensureDeviceBootstrap } from '$lib/utils/device-init';
	import { notificationsState } from '$lib/stores/notifications';
	import { getInitialTheme, useThemeSync } from '$lib/stores/theme';

	let { children } = $props();

	let ready = $state(false);
	let showConsent = $state(false);
	let initError = $state<string | null>(null);

	const isAdmin = $derived($page.url.pathname.startsWith('/admin'));
	const isMapPage = $derived($page.url.pathname === '/issues');
	const isIssueDetailPage = $derived(
		$page.url.pathname.startsWith('/issues/') && $page.url.pathname !== '/issues'
	);

	// Apply theme on init
	if (typeof window !== 'undefined') {
		const initialTheme = getInitialTheme();
		// Will be applied by the store
		import('$lib/stores/theme').then(({ setTheme }) => setTheme(initialTheme));
	}

	onMount(async () => {
		if (isAdmin) {
			ready = true;
			return;
		}

		// Setup theme sync with system preference
		const cleanup = useThemeSync();

		try {
			await ensureDeviceBootstrap({ retry: 1 });
			await notificationsState.init();

			if (!isConsentGiven()) {
				showConsent = true;
			}

			ready = true;

			return cleanup;
		} catch (e) {
			initError = 'Inisialisasi perangkat belum selesai. Mohon tunggu sebentar lalu muat ulang halaman.';
			ready = true;
			
			return cleanup;
		}
	});

	async function handleConsent() {
		const token = getAnonToken();
		if (token) {
			try {
				await recordConsent(token);
			} catch {
				// still allow consent locally
			}
		}
		setConsentGiven();
		showConsent = false;
	}
</script>

<svelte:head>
	<title>JEDUG — Laporkan Jalan Rusak</title>
	<meta name="description" content="Platform pelaporan jalan rusak berbasis partisipasi publik" />
</svelte:head>

{#if isAdmin}
	{@render children()}
{:else}
	<div class="app-shell">
		<AppHeader />
		<main class="app-main" class:app-main-full={isMapPage} class:app-main-wide={isIssueDetailPage}>
			{@render children()}
		</main>

		{#if initError}
			<div class="fixed inset-x-4 bottom-4 z-[200] mx-auto flex max-w-xl items-start gap-3 rounded-[12px] border border-brand/30 bg-surface px-4 py-3 text-sm text-brand">
				<DangerIcon class="mt-0.5 size-5 shrink-0" />
				<span>{initError}</span>
			</div>
		{/if}

		{#if showConsent}
			<ConsentSheet onaccept={handleConsent} />
		{/if}
	</div>
{/if}
