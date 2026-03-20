<script lang="ts">
	import { goto, afterNavigate } from '$app/navigation';
	import { page } from '$app/stores';
	import { adminLogout, adminMe } from '$lib/api/admin';

	let { children } = $props();

	let ready = $state(false);
	let username = $state('');

	const isLoginPage = $derived($page.url.pathname === '/admin/login');

	afterNavigate(({ from }) => {
		void syncAdminSession(from?.url?.pathname);
	});

	async function syncAdminSession(fromPathname?: string) {
		if (isLoginPage) {
			ready = true;
			return;
		}

		// Skip re-check when navigating within admin (already authenticated)
		if (ready && fromPathname !== '/admin/login') return;

		ready = false;

		try {
			const res = await adminMe();
			username = res.data?.username ?? '';
			ready = true;
		} catch {
			goto('/admin/login');
		}
	}

	async function handleLogout() {
		try {
			await adminLogout();
		} catch {
			// Session may already be expired; navigation still needs to happen.
		}
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
