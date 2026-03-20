<script lang="ts">
	import { goto } from '$app/navigation';
	import { adminLogin, adminMe } from '$lib/api/admin';
	import { onMount } from 'svelte';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	onMount(() => {
		void redirectIfSessionExists();
	});

	async function redirectIfSessionExists() {
		try {
			await adminMe();
			goto('/admin');
		} catch {
			// No active admin session; stay on login page.
		}
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			await adminLogin(username, password);
			goto('/admin');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login gagal';
		} finally {
			loading = false;
		}
	}
</script>

<div class="login-page">
	<div class="login-card">
		<h1>JEDUG Admin</h1>
		<p class="subtitle">Masuk untuk moderasi</p>

		<form onsubmit={handleSubmit}>
			{#if error}
				<div class="error-msg">{error}</div>
			{/if}

			<label>
				<span>Username</span>
				<input
					type="text"
					bind:value={username}
					required
					autocomplete="username"
					disabled={loading}
				/>
			</label>

			<label>
				<span>Password</span>
				<input
					type="password"
					bind:value={password}
					required
					autocomplete="current-password"
					disabled={loading}
				/>
			</label>

			<button type="submit" disabled={loading}>
				{loading ? 'Memproses...' : 'Masuk'}
			</button>
		</form>
	</div>
</div>

<style>
	.login-page {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 100dvh;
		background: #f7fafc;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	}
	.login-card {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 12px;
		padding: 40px 32px;
		width: 100%;
		max-width: 380px;
		margin: 16px;
	}
	h1 {
		font-size: 1.5rem;
		color: #e53e3e;
		margin: 0 0 4px;
	}
	.subtitle {
		color: #718096;
		font-size: 0.9rem;
		margin: 0 0 24px;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	label span {
		font-size: 0.85rem;
		font-weight: 500;
		color: #4a5568;
	}
	input {
		padding: 10px 12px;
		border: 1px solid #e2e8f0;
		border-radius: 6px;
		font-size: 0.95rem;
		outline: none;
	}
	input:focus {
		border-color: #e53e3e;
	}
	button {
		background: #e53e3e;
		color: #fff;
		border: none;
		padding: 12px;
		border-radius: 6px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		margin-top: 8px;
	}
	button:hover:not(:disabled) {
		background: #c53030;
	}
	button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
	.error-msg {
		background: #fff5f5;
		border: 1px solid #fed7d7;
		color: #c53030;
		padding: 10px 12px;
		border-radius: 6px;
		font-size: 0.85rem;
	}
</style>
