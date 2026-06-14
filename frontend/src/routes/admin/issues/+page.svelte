<script lang="ts">
	import { onMount } from 'svelte';
	import { adminListIssues } from '$lib/api/admin';
	import type { AdminIssue } from '$lib/api/types';
	import { ArrowLeftIcon, ArrowRightIcon, FilterIcon, MapIcon, SearchIcon } from '$lib/icons';
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

	function statusTone(status: string): string {
		switch (status) {
			case 'open':
				return 'border-hairline bg-sunken text-verify-community';
			case 'fixed':
				return 'border-hairline bg-sunken text-ink';
			case 'rejected':
				return 'border-brand/30 bg-brand-tint text-brand';
			case 'archived':
				return 'border-hairline bg-sunken text-muted';
			default:
				return 'border-hairline bg-sunken text-muted';
		}
	}

	function severityLabel(s: number): string {
		switch (s) {
			case 1:
				return 'Ringan';
			case 2:
				return 'Sedang';
			case 3:
				return 'Berat';
			case 4:
				return 'Parah';
			case 5:
				return 'Kritis';
			default:
				return String(s);
		}
	}
</script>

<div class="space-y-5">
	<section class="admin-card overflow-hidden">
		<div class="grid gap-4 px-5 py-5 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
			<div class="space-y-3">
				<span class="section-kicker">
					<MapIcon class="size-4" />
					Moderasi issue
				</span>
				<div>
					<h1 class="font-serif text-3xl font-semibold tracking-[-0.02em] text-ink">Daftar issue publik</h1>
					<p class="mt-2 max-w-[60ch] text-sm leading-6 text-muted">
						Area ini dipoles untuk memindai status, severity, visibilitas, dan akses detail moderasi lebih cepat.
					</p>
				</div>
			</div>

			<div class="rounded-[4px] border border-hairline bg-sunken px-4 py-4">
				<label class="input-shell min-w-[220px]">
					<span class="input-label">Filter status</span>
					<div class="relative">
						<span class="pointer-events-none absolute inset-y-0 left-4 flex items-center text-subtle">
							<FilterIcon class="size-5" />
						</span>
						<select class="select-field w-full pl-12" bind:value={statusFilter} onchange={handleFilter}>
							<option value="">Semua status</option>
							<option value="open">Open</option>
							<option value="fixed">Fixed</option>
							<option value="rejected">Rejected</option>
							<option value="archived">Archived</option>
						</select>
					</div>
				</label>
			</div>
		</div>
	</section>

	{#if loading}
		<div class="state-panel">
			<div class="mx-auto size-11 animate-spin rounded-full border-[3px] border-hairline border-t-brand-500"></div>
			<p class="mt-4 text-sm font-semibold text-ink">Memuat daftar issue...</p>
		</div>
	{:else if error}
		<div class="error-panel">{error}</div>
	{:else if issues.length === 0}
		<div class="state-panel">
			<div class="mx-auto flex size-12 items-center justify-center rounded-[8px] bg-sunken text-muted">
				<SearchIcon class="size-6" />
			</div>
			<p class="mt-4 text-sm font-semibold text-ink">Tidak ada issue ditemukan</p>
			<p class="mt-1 text-xs leading-5 text-muted">Ubah filter atau kembali lagi saat ada issue baru.</p>
		</div>
	{:else}
		<section class="admin-card overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full text-left">
					<thead class="bg-sunken">
						<tr class="border-b border-hairline">
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-subtle">Status</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-subtle">Severity</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-subtle">Lokasi</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-subtle">Laporan</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-subtle">Foto</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-subtle">Visibility</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-subtle">Terakhir</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-subtle">Aksi</th>
						</tr>
					</thead>
					<tbody>
						{#each issues as issue (issue.id)}
							<tr class={`border-b border-hairline align-top ${issue.is_hidden ? 'bg-brand-tint' : ''}`}>
								<td class="px-4 py-4">
									<span class={`inline-flex rounded-full border px-3 py-1 text-xs font-semibold ${statusTone(issue.status)}`}>
										{issue.status}
									</span>
								</td>
								<td class="px-4 py-4 text-sm font-semibold text-ink">
									{severityLabel(issue.severity_current)}
								</td>
								<td class="px-4 py-4">
									<div class="max-w-[280px]">
										<p class="truncate text-sm font-semibold text-ink">
											{issue.road_name || `${issue.latitude.toFixed(5)}, ${issue.longitude.toFixed(5)}`}
										</p>
										<p class="mt-1 text-xs text-muted">
											ID {issue.id.slice(0, 8)}...
										</p>
									</div>
								</td>
								<td class="px-4 py-4 text-sm text-ink">{issue.submission_count}</td>
								<td class="px-4 py-4 text-sm text-ink">{issue.photo_count}</td>
								<td class="px-4 py-4">
									{#if issue.is_hidden}
										<span class="inline-flex rounded-full border border-brand/30 bg-brand-tint px-3 py-1 text-xs font-semibold text-brand">Hidden</span>
									{:else}
										<span class="inline-flex rounded-full border border-hairline bg-sunken px-3 py-1 text-xs font-semibold text-muted">Public</span>
									{/if}
								</td>
								<td class="px-4 py-4 text-sm text-muted">{relativeTime(issue.last_seen_at)}</td>
								<td class="px-4 py-4">
									<a href="/admin/issues/{issue.id}" class="btn-secondary min-h-10 px-4 py-2">
										Buka detail
									</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>

		<div class="flex flex-col items-center justify-between gap-3 rounded-[4px] border border-hairline bg-surface px-4 py-4 sm:flex-row">
			<p class="text-sm text-muted">Halaman {Math.floor(offset / limit) + 1}</p>
			<div class="flex gap-2">
				<button class="btn-secondary min-h-10 px-4 py-2" onclick={prevPage} disabled={offset === 0}>
					<ArrowLeftIcon class="size-[18px]" />
					Prev
				</button>
				<button class="btn-secondary min-h-10 px-4 py-2" onclick={nextPage} disabled={issues.length < limit}>
					Next
					<ArrowRightIcon class="size-[18px]" />
				</button>
			</div>
		</div>
	{/if}
</div>
