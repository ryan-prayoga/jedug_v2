<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import BrowserPushCard from '$lib/components/BrowserPushCard.svelte';
	import { requestIssueDetailRefresh } from '$lib/utils/issue-detail-refresh';
	import {
		notificationsState,
		unreadNotificationCount,
	} from '$lib/stores/notifications';

	const pathname = $derived($page.url.pathname);
	const unreadCount = $derived($unreadNotificationCount);
	const notifState = $derived($notificationsState);

	let openNotif = $state(false);

	function isLaporActive(path: string): boolean {
		return path === '/lapor' || path.startsWith('/lapor/');
	}

	function isPetaActive(path: string): boolean {
		return path === '/issues' || path.startsWith('/issues/');
	}

	function isStatsActive(path: string): boolean {
		return path === '/stats' || path.startsWith('/stats/');
	}

	function formatNotifTime(input: string): string {
		const date = new Date(input);
		if (Number.isNaN(date.getTime())) return '';
		return date.toLocaleString('id-ID', {
			day: '2-digit',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit',
		});
	}

	async function handleNotificationClick(id: string, issueID: string) {
		await notificationsState.markRead(id);
		openNotif = false;

		const targetPath = `/issues/${issueID}`;
		if (pathname === targetPath) {
			requestIssueDetailRefresh({ issueID, source: 'notification' });
			return;
		}

		await goto(targetPath);
	}

	async function handleNotificationDelete(event: MouseEvent, id: string) {
		event.preventDefault();
		event.stopPropagation();
		await notificationsState.delete(id);
	}
</script>

<nav class="app-header">
	<a href="/" class="logo-group">
		<span class="logo-text">JEDUG</span>
		<span class="logo-sub">Pantau Jalan Rusak</span>
	</a>
	<div class="nav-links">
		<a
			href="/lapor"
			class="nav-link"
			class:active={isLaporActive(pathname)}
			aria-current={isLaporActive(pathname) ? 'page' : undefined}
		>
			Lapor
		</a>
		<a
			href="/issues"
			class="nav-link"
			class:active={isPetaActive(pathname)}
			aria-current={isPetaActive(pathname) ? 'page' : undefined}
		>
			Peta
		</a>
		<a
			href="/stats"
			class="nav-link"
			class:active={isStatsActive(pathname)}
			aria-current={isStatsActive(pathname) ? 'page' : undefined}
		>
			Statistik
		</a>
		<div class="notif-wrap">
			<button
				type="button"
				class="notif-button"
				onclick={() => (openNotif = !openNotif)}
				aria-label="Lihat notifikasi"
			>
				🔔
				{#if unreadCount > 0}
					<span class="notif-badge">{unreadCount}</span>
				{/if}
			</button>

			{#if openNotif}
				<div class="notif-panel">
					<div class="notif-title">Notifikasi</div>
					<BrowserPushCard
						variant="compact"
						lead="Aktifkan notifikasi browser agar update issue tetap masuk walau tab JEDUG tidak sedang dibuka."
					/>
					{#if notifState.loading}
						<div class="notif-empty">Memuat...</div>
					{:else if notifState.error}
						<div class="notif-empty">{notifState.error}</div>
					{:else if notifState.items.length === 0}
						<div class="notif-empty">Belum ada notifikasi.</div>
					{:else}
						<ul class="notif-list">
							{#each notifState.items as item (item.id)}
								<li class="notif-row">
									<button
										type="button"
										class="notif-item"
										class:unread={!item.read_at}
										onclick={() => handleNotificationClick(item.id, item.issue_id)}
									>
										<div class="notif-item-title">{item.title}</div>
										<div class="notif-item-message">{item.message}</div>
										<div class="notif-item-time">{formatNotifTime(item.created_at)}</div>
									</button>
									<button
										type="button"
										class="notif-delete"
										aria-label="Hapus notifikasi"
										disabled={notifState.deletingIDs.includes(item.id)}
										onclick={(event) => handleNotificationDelete(event, item.id)}
									>
										{notifState.deletingIDs.includes(item.id) ? '...' : 'Hapus'}
									</button>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</nav>

<style>
	.app-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		background: #fff;
		border-bottom: 1px solid #E2E8F0;
		position: sticky;
		top: 0;
		z-index: 100;
	}
	.logo-group {
		display: flex;
		align-items: baseline;
		gap: 8px;
		text-decoration: none;
	}
	.logo-text {
		font-size: 1.25rem;
		font-weight: 800;
		color: #E5484D;
		letter-spacing: -0.5px;
		line-height: 1;
	}
	.logo-sub {
		font-size: 11px;
		color: #64748B;
		font-weight: 400;
		letter-spacing: 0;
		display: none;
	}
	@media (min-width: 360px) {
		.logo-sub {
			display: inline;
		}
	}
	.nav-links {
		display: flex;
		gap: 4px;
		align-items: center;
	}
	.nav-link {
		font-size: 14px;
		color: #64748B;
		text-decoration: none;
		font-weight: 500;
		padding: 6px 12px;
		border-radius: 8px;
		transition: background 0.15s, color 0.15s;
	}
	.nav-link:hover {
		color: #E5484D;
		background: #FEF2F2;
	}
	.nav-link.active {
		color: #B42318;
		background: #FEE4E2;
		font-weight: 700;
	}

	.notif-wrap {
		position: relative;
	}

	.notif-button {
		position: relative;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 999px;
		border: 1px solid #E2E8F0;
		background: #fff;
		cursor: pointer;
		font-size: 16px;
	}

	.notif-badge {
		position: absolute;
		top: -4px;
		right: -4px;
		min-width: 18px;
		height: 18px;
		padding: 0 4px;
		border-radius: 999px;
		background: #DC2626;
		color: #fff;
		font-size: 10px;
		font-weight: 700;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.notif-panel {
		position: absolute;
		right: 0;
		top: calc(100% + 8px);
		width: min(92vw, 340px);
		max-height: 420px;
		overflow: auto;
		background: #fff;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		box-shadow: 0 14px 30px rgba(15, 23, 42, 0.15);
		padding: 8px;
		z-index: 120;
	}

	.notif-title {
		font-size: 13px;
		font-weight: 700;
		color: #0F172A;
		padding: 8px;
	}

	.notif-list {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.notif-row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		gap: 8px;
		align-items: stretch;
	}

	.notif-item {
		width: 100%;
		text-align: left;
		border: 1px solid #E2E8F0;
		background: #fff;
		border-radius: 10px;
		padding: 10px;
		cursor: pointer;
	}

	.notif-item.unread {
		border-color: #FCA5A5;
		background: #FEF2F2;
	}

	.notif-item-title {
		font-size: 12px;
		font-weight: 700;
		color: #0F172A;
	}

	.notif-item-message {
		font-size: 12px;
		color: #475569;
		margin-top: 4px;
	}

	.notif-item-time {
		font-size: 11px;
		color: #94A3B8;
		margin-top: 6px;
	}

	.notif-delete {
		min-width: 60px;
		padding: 0 12px;
		border-radius: 10px;
		border: 1px solid #FECACA;
		background: #FFF5F5;
		color: #B42318;
		font-size: 12px;
		font-weight: 700;
		cursor: pointer;
	}

	.notif-delete:disabled {
		opacity: 0.65;
		cursor: default;
	}

	.notif-empty {
		padding: 12px;
		font-size: 12px;
		color: #64748B;
	}

	@media (max-width: 420px) {
		.nav-links {
			gap: 2px;
		}
		.nav-link {
			padding: 6px 8px;
			font-size: 13px;
		}
	}
</style>
