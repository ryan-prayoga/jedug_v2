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
				return 'border-emerald-200 bg-emerald-50 text-emerald-700';
			case 'fixed':
				return 'border-blue-200 bg-blue-50 text-blue-700';
			case 'rejected':
				return 'border-rose-200 bg-rose-50 text-rose-700';
			case 'archived':
				return 'border-slate-200 bg-slate-100 text-slate-600';
			default:
				return 'border-slate-200 bg-slate-100 text-slate-600';
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
					<h1 class="text-3xl font-[800] tracking-[-0.05em] text-slate-950">Daftar issue publik</h1>
					<p class="mt-2 max-w-[60ch] text-sm leading-6 text-slate-500">
						Area ini dipoles untuk memindai status, severity, visibilitas, dan akses detail moderasi lebih cepat.
					</p>
				</div>
			</div>

			<div class="rounded-[24px] border border-slate-200 bg-slate-50 px-4 py-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.7)]">
				<label class="input-shell min-w-[220px]">
					<span class="input-label">Filter status</span>
					<div class="relative">
						<span class="pointer-events-none absolute inset-y-0 left-4 flex items-center text-slate-400">
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
			<div class="mx-auto size-11 animate-spin rounded-full border-[3px] border-slate-200 border-t-brand-500"></div>
			<p class="mt-4 text-sm font-semibold text-slate-700">Memuat daftar issue...</p>
		</div>
	{:else if error}
		<div class="error-panel">{error}</div>
	{:else if issues.length === 0}
		<div class="state-panel">
			<div class="mx-auto flex size-12 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
				<SearchIcon class="size-6" />
			</div>
			<p class="mt-4 text-sm font-semibold text-slate-700">Tidak ada issue ditemukan</p>
			<p class="mt-1 text-xs leading-5 text-slate-500">Ubah filter atau kembali lagi saat ada issue baru.</p>
		</div>
	{:else}
		<section class="admin-card overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full text-left">
					<thead class="bg-slate-50/90">
						<tr class="border-b border-slate-200">
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">Status</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">Severity</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">Lokasi</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">Laporan</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">Foto</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">Visibility</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">Terakhir</th>
							<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">Aksi</th>
						</tr>
					</thead>
					<tbody>
						{#each issues as issue (issue.id)}
							<tr class={`border-b border-slate-100 align-top ${issue.is_hidden ? 'bg-rose-50/55' : ''}`}>
								<td class="px-4 py-4">
									<span class={`inline-flex rounded-full border px-3 py-1 text-xs font-semibold ${statusTone(issue.status)}`}>
										{issue.status}
									</span>
								</td>
								<td class="px-4 py-4 text-sm font-semibold text-slate-700">
									{severityLabel(issue.severity_current)}
								</td>
								<td class="px-4 py-4">
									<div class="max-w-[280px]">
										<p class="truncate text-sm font-semibold text-slate-900">
											{issue.road_name || `${issue.latitude.toFixed(5)}, ${issue.longitude.toFixed(5)}`}
										</p>
										<p class="mt-1 text-xs text-slate-500">
											ID {issue.id.slice(0, 8)}...
										</p>
									</div>
								</td>
								<td class="px-4 py-4 text-sm text-slate-700">{issue.submission_count}</td>
								<td class="px-4 py-4 text-sm text-slate-700">{issue.photo_count}</td>
								<td class="px-4 py-4">
									{#if issue.is_hidden}
										<span class="inline-flex rounded-full border border-rose-200 bg-rose-50 px-3 py-1 text-xs font-semibold text-rose-700">Hidden</span>
									{:else}
										<span class="inline-flex rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-600">Public</span>
									{/if}
								</td>
								<td class="px-4 py-4 text-sm text-slate-500">{relativeTime(issue.last_seen_at)}</td>
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

		<div class="flex flex-col items-center justify-between gap-3 rounded-[24px] border border-white/80 bg-white/80 px-4 py-4 shadow-[0_12px_28px_rgba(15,23,42,0.06)] sm:flex-row">
			<p class="text-sm text-slate-500">Halaman {Math.floor(offset / limit) + 1}</p>
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
