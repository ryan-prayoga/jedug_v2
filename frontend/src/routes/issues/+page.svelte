<script lang="ts">
	import { onMount } from 'svelte';
	import { listIssues } from '$lib/api/issues';
	import type { Issue } from '$lib/api/types';
	import IssueCard from '$lib/components/IssueCard.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';

	let issues = $state<Issue[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function fetchIssues() {
		loading = true;
		error = null;
		try {
			const res = await listIssues({ limit: 20, offset: 0 });
			issues = res.data || [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat data';
		} finally {
			loading = false;
		}
	}

	onMount(fetchIssues);
</script>

<div class="issues-page">
	<div class="page-header">
		<h1>Laporan Publik</h1>
		<button class="refresh-btn" onclick={fetchIssues} disabled={loading}>
			🔄
		</button>
	</div>

	{#if loading}
		<LoadingState message="Memuat laporan..." />
	{:else if error}
		<ErrorState message={error} onretry={fetchIssues} />
	{:else if issues.length === 0}
		<EmptyState
			message="Belum ada laporan. Jadilah yang pertama melaporkan!"
			icon="🚧"
		/>
	{:else}
		<div class="issue-list">
			{#each issues as issue (issue.id)}
				<IssueCard {issue} />
			{/each}
		</div>
	{/if}

	<div class="bottom-cta">
		<a href="/lapor" class="report-cta">📸 Laporkan Jalan Rusak</a>
	</div>
</div>

<style>
	.issues-page {
		padding-top: 1.5rem;
	}
	.page-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}
	h1 {
		font-size: 1.3rem;
		font-weight: 700;
	}
	.refresh-btn {
		background: none;
		border: 1px solid #e2e8f0;
		border-radius: 8px;
		padding: 6px 12px;
		cursor: pointer;
		font-size: 1rem;
	}
	.refresh-btn:disabled {
		opacity: 0.5;
	}
	.issue-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.bottom-cta {
		margin-top: 2rem;
		text-align: center;
	}
	.report-cta {
		display: inline-block;
		padding: 12px 24px;
		font-size: 1rem;
		font-weight: 600;
		color: #fff;
		background: #e53e3e;
		border-radius: 12px;
		text-decoration: none;
	}
	.report-cta:hover {
		opacity: 0.9;
	}
</style>
