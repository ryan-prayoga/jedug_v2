<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import AppHeader from '$lib/components/AppHeader.svelte';
	import ConsentSheet from '$lib/components/ConsentSheet.svelte';
	import { DangerIcon } from '$lib/icons';
	import { browserPushState } from '$lib/stores/browser-push';
	import { notificationPreferencesState } from '$lib/stores/notification-preferences';
	import { getAnonToken, isConsentGiven, setConsentGiven } from '$lib/utils/storage';
	import { recordConsent } from '$lib/api/device';
	import { ensureDeviceBootstrap } from '$lib/utils/device-init';
	import { notificationsState } from '$lib/stores/notifications';

	let { children } = $props();

	let ready = $state(false);
	let showConsent = $state(false);
	let initError = $state<string | null>(null);

	const isAdmin = $derived($page.url.pathname.startsWith('/admin'));
	const isMapPage = $derived($page.url.pathname === '/issues');
	const isIssueDetailPage = $derived(
		$page.url.pathname.startsWith('/issues/') && $page.url.pathname !== '/issues'
	);

	onMount(async () => {
		if (isAdmin) {
			ready = true;
			return;
		}

		try {
			await ensureDeviceBootstrap({ retry: 1 });
			await notificationsState.init();
			await browserPushState.init();
			await notificationPreferencesState.init();

			if (!isConsentGiven()) {
				showConsent = true;
			}

			ready = true;
		} catch (e) {
			initError = 'Inisialisasi perangkat belum selesai. Mohon tunggu sebentar lalu muat ulang halaman.';
			ready = true;
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
			<div class="fixed inset-x-4 bottom-4 z-[200] mx-auto flex max-w-xl items-start gap-3 rounded-[24px] border border-rose-200 bg-white/95 px-4 py-3 text-sm text-rose-700 shadow-[0_18px_40px_rgba(15,23,42,0.14)] backdrop-blur">
				<DangerIcon class="mt-0.5 size-5 shrink-0" />
				<span>{initError}</span>
			</div>
		{/if}

		{#if showConsent}
			<ConsentSheet onaccept={handleConsent} />
		{/if}
	</div>
{/if}
