<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import AppHeader from '$lib/components/AppHeader.svelte';
	import ConsentSheet from '$lib/components/ConsentSheet.svelte';
	import { getAnonToken, isConsentGiven, setConsentGiven } from '$lib/utils/storage';
	import { recordConsent } from '$lib/api/device';
	import { ensureDeviceBootstrap } from '$lib/utils/device-init';

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

			if (!isConsentGiven()) {
				showConsent = true;
			}

			ready = true;
		} catch (e) {
			console.error('[layout] bootstrap init failed', e);
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
		<main
			class="app-main"
			class:app-main-full={isMapPage}
			class:app-main-wide={isIssueDetailPage}
		>
			{@render children()}
		</main>

		{#if initError}
			<div class="init-toast">⚠️ {initError}</div>
		{/if}

		{#if showConsent}
			<ConsentSheet onaccept={handleConsent} />
		{/if}
	</div>
{/if}

<style>
	:global(*) {
		margin: 0;
		padding: 0;
		box-sizing: border-box;
	}
	:global(body) {
		font-family: 'Inter', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
		background: #F8FAFC;
		color: #0F172A;
		font-size: 14px;
		line-height: 1.5;
		-webkit-font-smoothing: antialiased;
		-moz-osx-font-smoothing: grayscale;
	}
	.app-shell {
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
	}
	.app-main {
		flex: 1;
		max-width: 480px;
		width: 100%;
		margin: 0 auto;
		padding: 0 16px 24px;
	}
	.app-main-full {
		max-width: none;
		padding: 0;
		overflow: hidden;
	}
	.app-main-wide {
		max-width: 1120px;
		padding-bottom: 40px;
	}
	.init-toast {
		position: fixed;
		bottom: 1rem;
		left: 50%;
		transform: translateX(-50%);
		background: #FFF5F5;
		border: 1px solid #FED7D7;
		color: #DC2626;
		font-size: 12px;
		padding: 8px 16px;
		border-radius: 12px;
		box-shadow: 0 4px 16px rgba(0,0,0,0.10);
		z-index: 200;
	}

	@media (min-width: 768px) {
		.app-main-wide {
			padding-left: 24px;
			padding-right: 24px;
		}
	}
</style>
