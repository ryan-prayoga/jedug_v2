<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { getAdminToken, clearAdminToken } from '$lib/utils/storage';
	import { adminMe } from '$lib/api/admin';

	let { children } = $props();

	let ready = $state(false);
	let username = $state('');

	const isLoginPage = $derived($page.url.pathname === '/admin/login');

	onMount(async () => {
		if (isLoginPage) {
			ready = true;
			return;
		}

		const token = getAdminToken();
		if (!token) {
			goto('/admin/login');
			return;
		}

		try {
			const res = await adminMe();
			username = res.data?.username ?? '';
			ready = true;
		} catch {
			clearAdminToken();
			goto('/admin/login');
		}
	});

	function handleLogout() {
		clearAdminToken();
		goto('/admin/login');
	}
</script>

<svelte:head>
	<title>JEDUG Admin</title>
</svelte:head>

{#if isLoginPage}
	{@render children()}
{:else if ready}
	<div class="admin-shell">
		<header class="admin-header">
			<div class="header-inner">
				<a href="/admin" class="logo">JEDUG Admin</a>
				<nav class="nav">
					<a href="/admin/issues">Issues</a>
				</nav>
				<div class="user-area">
					<span class="username">{username}</span>
					<button class="logout-btn" onclick={handleLogout}>Logout</button>
				</div>
			</div>
		</header>
		<main class="admin-main">
			{@render children()}
		</main>
	</div>
{:else}
	<div class="loading">Memuat...</div>
{/if}

<style>
	:global(.admin-shell *) {
		box-sizing: border-box;
	}
	.admin-shell {
		min-height: 100dvh;
		background: #f7fafc;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		color: #1a202c;
	}
	.admin-header {
		background: #1a202c;
		color: #fff;
		padding: 0 16px;
		position: sticky;
		top: 0;
		z-index: 100;
	}
	.header-inner {
		max-width: 1200px;
		margin: 0 auto;
		display: flex;
		align-items: center;
		height: 56px;
		gap: 24px;
	}
	.logo {
		font-weight: 700;
		font-size: 1.1rem;
		color: #e53e3e;
		text-decoration: none;
	}
	.nav {
		flex: 1;
		display: flex;
		gap: 16px;
	}
	.nav a {
		color: #cbd5e0;
		text-decoration: none;
		font-size: 0.9rem;
	}
	.nav a:hover {
		color: #fff;
	}
	.user-area {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.username {
		font-size: 0.85rem;
		color: #a0aec0;
	}
	.logout-btn {
		background: none;
		border: 1px solid #4a5568;
		color: #e2e8f0;
		padding: 4px 12px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.8rem;
	}
	.logout-btn:hover {
		border-color: #e53e3e;
		color: #e53e3e;
	}
	.admin-main {
		max-width: 1200px;
		margin: 0 auto;
		padding: 24px 16px;
	}
	.loading {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100dvh;
		color: #718096;
	}
</style>
