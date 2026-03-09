<script lang="ts">
	import { onMount } from 'svelte';
	import { adminListIssues } from '$lib/api/admin';
	import type { AdminIssue } from '$lib/api/types';
	import { relativeTime } from '$lib/utils/date';

	let issues = $state<AdminIssue[]>([]);
	let loading = $state(true);
	let error = $state('');
	let statusFilter = $state('');
	let offset = $state(0);
	const limit = 20;

	async function loadIssues() {
		loading = true;
		error = '';
		try {
			const params: { limit: number; offset: number; status?: string } = { limit, offset };
			if (statusFilter) params.status = statusFilter;
			const res = await adminListIssues(params);
			issues = res.data ?? [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Gagal memuat issues';
		} finally {
			loading = false;
		}
	}

	onMount(loadIssues);

	function handleFilter() {
		offset = 0;
		loadIssues();
	}

	function nextPage() {
		offset += limit;
		loadIssues();
	}

	function prevPage() {
		offset = Math.max(0, offset - limit);
		loadIssues();
	}

	function statusColor(status: string): string {
		switch (status) {
			case 'open': return '#38a169';
			case 'fixed': return '#3182ce';
			case 'rejected': return '#e53e3e';
			case 'archived': return '#718096';
			default: return '#718096';
		}
	}

	function severityLabel(s: number): string {
		switch (s) {
			case 1: return 'Ringan';
			case 2: return 'Sedang';
			case 3: return 'Berat';
			case 4: return 'Parah';
			case 5: return 'Kritis';
			default: return String(s);
		}
	}
</script>

<div class="page">
	<div class="page-header">
		<h1>Issues</h1>
		<div class="filters">
			<select bind:value={statusFilter} onchange={handleFilter}>
				<option value="">Semua Status</option>
				<option value="open">Open</option>
				<option value="fixed">Fixed</option>
				<option value="rejected">Rejected</option>
				<option value="archived">Archived</option>
			</select>
		</div>
	</div>

	{#if loading}
		<p class="info">Memuat...</p>
	{:else if error}
		<div class="error-msg">{error}</div>
	{:else if issues.length === 0}
		<p class="info">Tidak ada issue ditemukan.</p>
	{:else}
		<div class="table-wrap">
			<table>
				<thead>
					<tr>
						<th>Status</th>
						<th>Severity</th>
						<th>Lokasi</th>
						<th>Laporan</th>
						<th>Foto</th>
						<th>Hidden</th>
						<th>Terakhir</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each issues as issue (issue.id)}
						<tr class:hidden-row={issue.is_hidden}>
							<td>
								<span class="badge" style="background:{statusColor(issue.status)}">
									{issue.status}
								</span>
							</td>
							<td>{severityLabel(issue.severity_current)}</td>
							<td class="location">
								{#if issue.road_name}
									{issue.road_name}
								{:else}
									{issue.latitude.toFixed(5)}, {issue.longitude.toFixed(5)}
								{/if}
							</td>
							<td>{issue.submission_count}</td>
							<td>{issue.photo_count}</td>
							<td>
								{#if issue.is_hidden}
									<span class="badge" style="background:#e53e3e">Ya</span>
								{:else}
									—
								{/if}
							</td>
							<td class="date">{relativeTime(issue.last_seen_at)}</td>
							<td>
								<a href="/admin/issues/{issue.id}" class="detail-link">Detail →</a>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<div class="pagination">
			<button onclick={prevPage} disabled={offset === 0}>← Prev</button>
			<span class="page-info">Halaman {Math.floor(offset / limit) + 1}</span>
			<button onclick={nextPage} disabled={issues.length < limit}>Next →</button>
		</div>
	{/if}
</div>

<style>
	.page {
		width: 100%;
	}
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
		flex-wrap: wrap;
		gap: 12px;
	}
	h1 {
		font-size: 1.4rem;
		margin: 0;
	}
	select {
		padding: 8px 12px;
		border: 1px solid #e2e8f0;
		border-radius: 6px;
		font-size: 0.9rem;
		background: #fff;
	}
	.table-wrap {
		overflow-x: auto;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.9rem;
	}
	th {
		text-align: left;
		padding: 10px 12px;
		border-bottom: 2px solid #e2e8f0;
		color: #4a5568;
		font-weight: 600;
		font-size: 0.8rem;
		text-transform: uppercase;
		white-space: nowrap;
	}
	td {
		padding: 10px 12px;
		border-bottom: 1px solid #edf2f7;
		vertical-align: middle;
	}
	.hidden-row {
		background: #fff5f5;
	}
	.location {
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.date {
		white-space: nowrap;
		color: #718096;
		font-size: 0.85rem;
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
	.detail-link {
		color: #3182ce;
		text-decoration: none;
		font-size: 0.85rem;
		white-space: nowrap;
	}
	.detail-link:hover {
		text-decoration: underline;
	}
	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 16px;
		margin-top: 20px;
	}
	.pagination button {
		padding: 8px 16px;
		border: 1px solid #e2e8f0;
		border-radius: 6px;
		background: #fff;
		cursor: pointer;
		font-size: 0.85rem;
	}
	.pagination button:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
	.page-info {
		color: #718096;
		font-size: 0.85rem;
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
