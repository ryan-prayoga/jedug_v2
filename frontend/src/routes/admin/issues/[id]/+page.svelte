<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import {
		adminGetIssue,
		adminHideIssue,
		adminUnhideIssue,
		adminFixIssue,
		adminRejectIssue,
		adminBanDevice,
	} from '$lib/api/admin';
	import type { AdminIssueDetail } from '$lib/api/types';
	import { formatDate, relativeTime } from '$lib/utils/date';

	let detail = $state<AdminIssueDetail | null>(null);
	let loading = $state(true);
	let error = $state('');
	let actionLoading = $state('');
	let actionError = $state('');
	let reasonInput = $state('');

	const issueId = $derived($page.params.id ?? '');

	async function loadIssue() {
		if (!issueId) return;
		loading = true;
		error = '';
		try {
			const res = await adminGetIssue(issueId);
			detail = res.data ?? null;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Gagal memuat issue';
		} finally {
			loading = false;
		}
	}

	onMount(loadIssue);

	async function doAction(action: string, fn: (id: string, reason?: string) => Promise<unknown>) {
		if (actionLoading) return;
		actionLoading = action;
		actionError = '';
		try {
			await fn(issueId, reasonInput || undefined);
			reasonInput = '';
			await loadIssue();
		} catch (err) {
			actionError = err instanceof Error ? err.message : 'Aksi gagal';
		} finally {
			actionLoading = '';
		}
	}

	async function handleBanDevice(deviceId: string) {
		if (!confirm(`Ban device ${deviceId.slice(0, 8)}...? Aksi ini tidak bisa dibatalkan.`)) return;
		actionLoading = 'ban';
		actionError = '';
		try {
			await adminBanDevice(deviceId, reasonInput || undefined);
			reasonInput = '';
			await loadIssue();
		} catch (err) {
			actionError = err instanceof Error ? err.message : 'Ban device gagal';
		} finally {
			actionLoading = '';
		}
	}

	function severityLabel(s: number): string {
		const labels: Record<number, string> = { 1: 'Ringan', 2: 'Sedang', 3: 'Berat', 4: 'Parah', 5: 'Kritis' };
		return labels[s] ?? String(s);
	}

	function actionTypeLabel(a: string): string {
		const labels: Record<string, string> = {
			hide_issue: 'Sembunyikan',
			unhide_issue: 'Tampilkan',
			mark_fixed: 'Tandai Selesai',
			reject_issue: 'Tolak',
			ban_device: 'Ban Device',
		};
		return labels[a] ?? a;
	}
</script>

{#if loading}
	<p class="info">Memuat...</p>
{:else if error}
	<div class="error-msg">{error}</div>
{:else if detail}
	<div class="detail-page">
		<!-- Header -->
		<div class="back-row">
			<a href="/admin/issues">← Kembali ke daftar</a>
		</div>

		<div class="issue-header">
			<h1>
				Issue
				<span class="badge" style="background:{detail.status === 'open' ? '#38a169' : detail.status === 'fixed' ? '#3182ce' : detail.status === 'rejected' ? '#e53e3e' : '#718096'}">
					{detail.status}
				</span>
				{#if detail.is_hidden}
					<span class="badge" style="background:#e53e3e">HIDDEN</span>
				{/if}
			</h1>
			<code class="issue-id">{detail.id}</code>
		</div>

		<!-- Metadata Grid -->
		<div class="meta-grid">
			<div class="meta-item">
				<span class="meta-label">Severity</span>
				<span class="meta-value">{severityLabel(detail.severity_current)} (max: {severityLabel(detail.severity_max)})</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Lokasi</span>
				<span class="meta-value">
					{#if detail.road_name}{detail.road_name}{:else}{detail.latitude.toFixed(6)}, {detail.longitude.toFixed(6)}{/if}
				</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Laporan</span>
				<span class="meta-value">{detail.submission_count}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Foto</span>
				<span class="meta-value">{detail.photo_count}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Korban</span>
				<span class="meta-value">{detail.casualty_count}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Pertama Terlihat</span>
				<span class="meta-value">{formatDate(detail.first_seen_at)}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Terakhir Terlihat</span>
				<span class="meta-value">{relativeTime(detail.last_seen_at)}</span>
			</div>
		</div>

		<!-- Media Gallery -->
		{#if detail.media.length > 0}
			<section class="section">
				<h2>Foto ({detail.media.length})</h2>
				<div class="gallery">
					{#each detail.media as media}
						<a href={media.public_url} target="_blank" rel="noopener noreferrer">
							<img src={media.public_url} alt="Evidence" loading="lazy" />
						</a>
					{/each}
				</div>
			</section>
		{/if}

		<!-- Moderation Actions -->
		<section class="section">
			<h2>Aksi Moderasi</h2>

			{#if actionError}
				<div class="error-msg" style="margin-bottom:12px">{actionError}</div>
			{/if}

			<div class="reason-row">
				<input
					type="text"
					placeholder="Alasan (opsional)"
					bind:value={reasonInput}
					disabled={!!actionLoading}
				/>
			</div>

			<div class="actions-row">
				{#if detail.is_hidden}
					<button
						class="action-btn unhide"
						onclick={() => doAction('unhide', adminUnhideIssue)}
						disabled={!!actionLoading}
					>
						{actionLoading === 'unhide' ? '...' : '👁 Tampilkan'}
					</button>
				{:else}
					<button
						class="action-btn hide"
						onclick={() => doAction('hide', adminHideIssue)}
						disabled={!!actionLoading}
					>
						{actionLoading === 'hide' ? '...' : '🙈 Sembunyikan'}
					</button>
				{/if}

				{#if detail.status === 'open'}
					<button
						class="action-btn fix"
						onclick={() => doAction('fix', adminFixIssue)}
						disabled={!!actionLoading}
					>
						{actionLoading === 'fix' ? '...' : '✅ Tandai Selesai'}
					</button>
					<button
						class="action-btn reject"
						onclick={() => doAction('reject', adminRejectIssue)}
						disabled={!!actionLoading}
					>
						{actionLoading === 'reject' ? '...' : '❌ Tolak'}
					</button>
				{/if}
			</div>
		</section>

		<!-- Submissions -->
		{#if detail.submissions.length > 0}
			<section class="section">
				<h2>Submissions ({detail.submissions.length})</h2>
				<div class="table-wrap">
					<table>
						<thead>
							<tr>
								<th>Device</th>
								<th>Status</th>
								<th>Severity</th>
								<th>Korban</th>
								<th>Note</th>
								<th>Waktu</th>
								<th></th>
							</tr>
						</thead>
						<tbody>
							{#each detail.submissions as sub}
								<tr class:banned-row={sub.device_is_banned}>
									<td>
										<code class="device-id">{sub.device_id.slice(0, 8)}</code>
										{#if sub.device_is_banned}
											<span class="badge" style="background:#e53e3e;font-size:0.65rem">BANNED</span>
										{/if}
									</td>
									<td>{sub.status}</td>
									<td>{severityLabel(sub.severity)}</td>
									<td>{sub.has_casualty ? sub.has_casualty : '—'}</td>
									<td class="note-cell">{sub.note ?? '—'}</td>
									<td class="date">{relativeTime(sub.reported_at)}</td>
									<td>
										{#if !sub.device_is_banned}
											<button
												class="ban-btn"
												onclick={() => handleBanDevice(sub.device_id)}
												disabled={!!actionLoading}
											>
												Ban Device
											</button>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>
		{/if}

		<!-- Moderation Log -->
		{#if detail.moderation_log.length > 0}
			<section class="section">
				<h2>Log Moderasi</h2>
				<div class="log-list">
					{#each detail.moderation_log as log}
						<div class="log-item">
							<span class="log-action">{actionTypeLabel(log.action_type)}</span>
							<span class="log-by">oleh {log.admin_username ?? 'system'}</span>
							{#if log.note}
								<span class="log-note">— {log.note}</span>
							{/if}
							<span class="log-time">{formatDate(log.created_at)}</span>
						</div>
					{/each}
				</div>
			</section>
		{/if}
	</div>
{/if}

<style>
	.detail-page {
		width: 100%;
	}
	.back-row {
		margin-bottom: 16px;
	}
	.back-row a {
		color: #3182ce;
		text-decoration: none;
		font-size: 0.9rem;
	}
	.issue-header {
		margin-bottom: 20px;
	}
	.issue-header h1 {
		font-size: 1.3rem;
		margin: 0 0 4px;
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.issue-id {
		font-size: 0.75rem;
		color: #a0aec0;
		word-break: break-all;
	}
	.badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 4px;
		color: #fff;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}
	.meta-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 12px;
		margin-bottom: 24px;
	}
	.meta-item {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 8px;
		padding: 12px;
	}
	.meta-label {
		display: block;
		font-size: 0.75rem;
		color: #718096;
		text-transform: uppercase;
		font-weight: 600;
		margin-bottom: 4px;
	}
	.meta-value {
		font-size: 0.95rem;
		font-weight: 500;
	}
	.section {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 8px;
		padding: 16px;
		margin-bottom: 16px;
	}
	.section h2 {
		font-size: 1rem;
		margin: 0 0 12px;
		color: #2d3748;
	}
	.gallery {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.gallery img {
		width: 120px;
		height: 120px;
		object-fit: cover;
		border-radius: 6px;
		border: 1px solid #e2e8f0;
	}
	.reason-row {
		margin-bottom: 12px;
	}
	.reason-row input {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid #e2e8f0;
		border-radius: 6px;
		font-size: 0.9rem;
	}
	.actions-row {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.action-btn {
		padding: 8px 16px;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		font-size: 0.85rem;
		font-weight: 600;
		color: #fff;
	}
	.action-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.action-btn.hide { background: #d69e2e; }
	.action-btn.unhide { background: #38a169; }
	.action-btn.fix { background: #3182ce; }
	.action-btn.reject { background: #e53e3e; }
	.table-wrap {
		overflow-x: auto;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}
	th {
		text-align: left;
		padding: 8px 10px;
		border-bottom: 2px solid #e2e8f0;
		color: #4a5568;
		font-weight: 600;
		font-size: 0.75rem;
		text-transform: uppercase;
		white-space: nowrap;
	}
	td {
		padding: 8px 10px;
		border-bottom: 1px solid #edf2f7;
	}
	.device-id {
		font-size: 0.8rem;
		color: #4a5568;
	}
	.banned-row {
		background: #fff5f5;
	}
	.note-cell {
		max-width: 160px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.date {
		white-space: nowrap;
		color: #718096;
		font-size: 0.8rem;
	}
	.ban-btn {
		padding: 4px 10px;
		background: #e53e3e;
		color: #fff;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.75rem;
		white-space: nowrap;
	}
	.ban-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.log-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.log-item {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
		font-size: 0.85rem;
		padding: 8px 0;
		border-bottom: 1px solid #edf2f7;
	}
	.log-item:last-child {
		border-bottom: none;
	}
	.log-action {
		font-weight: 600;
		color: #2d3748;
	}
	.log-by {
		color: #718096;
	}
	.log-note {
		color: #4a5568;
		font-style: italic;
	}
	.log-time {
		color: #a0aec0;
		font-size: 0.8rem;
		margin-left: auto;
	}
	.info {
		color: #718096;
		text-align: center;
		padding: 40px 0;
	}
	.error-msg {
		background: #fff5f5;
		border: 1px solid #fed7d7;
		color: #c53030;
		padding: 12px;
		border-radius: 6px;
	}
</style>
