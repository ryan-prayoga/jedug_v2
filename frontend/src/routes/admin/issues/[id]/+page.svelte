<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import {
		adminBanDevice,
		adminFixIssue,
		adminGetIssue,
		adminHideIssue,
		adminRejectIssue,
		adminUnhideIssue
	} from '$lib/api/admin';
	import type { AdminIssueDetail } from '$lib/api/types';
	import {
		ArrowLeftIcon,
		DangerIcon,
		DocumentIcon,
		EyeClosedIcon,
		EyeIcon,
		MapIcon,
		ShieldCheckIcon,
		TrashIcon
	} from '$lib/icons';
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
		const labels: Record<number, string> = {
			1: 'Ringan',
			2: 'Sedang',
			3: 'Berat',
			4: 'Parah',
			5: 'Kritis'
		};
		return labels[s] ?? String(s);
	}

	function actionTypeLabel(a: string): string {
		const labels: Record<string, string> = {
			hide_issue: 'Sembunyikan',
			unhide_issue: 'Tampilkan',
			mark_fixed: 'Tandai selesai',
			reject_issue: 'Tolak',
			ban_device: 'Ban device'
		};
		return labels[a] ?? a;
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
</script>

{#if loading}
	<div class="state-panel">
		<div class="mx-auto size-11 animate-spin rounded-full border-[3px] border-slate-200 border-t-brand-500"></div>
		<p class="mt-4 text-sm font-semibold text-slate-700">Memuat issue...</p>
	</div>
{:else if error}
	<div class="error-panel">{error}</div>
{:else if detail}
	<div class="space-y-5">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<a href="/admin/issues" class="btn-secondary min-h-10 px-4 py-2">
				<ArrowLeftIcon class="size-[18px]" />
				Kembali ke daftar
			</a>
			<span class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-400">
				Issue ID {detail.id}
			</span>
		</div>

		<section class="admin-card overflow-hidden">
			<div class="grid gap-5 px-5 py-5 lg:grid-cols-[minmax(0,1fr)_320px]">
				<div class="space-y-4">
					<div class="flex flex-wrap items-center gap-2">
						<span class={`inline-flex rounded-full border px-3 py-1 text-xs font-semibold ${statusTone(detail.status)}`}>
							{detail.status}
						</span>
						{#if detail.is_hidden}
							<span class="inline-flex rounded-full border border-rose-200 bg-rose-50 px-3 py-1 text-xs font-semibold text-rose-700">
								Hidden
							</span>
						{/if}
						<span class="inline-flex rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-600">
							Severity {severityLabel(detail.severity_current)}
						</span>
					</div>

					<div class="space-y-2">
						<h1 class="text-3xl font-[800] tracking-[-0.05em] text-slate-950">
							{detail.road_name || `${detail.latitude.toFixed(6)}, ${detail.longitude.toFixed(6)}`}
						</h1>
						<p class="text-sm leading-6 text-slate-500">
							Issue ini menggabungkan laporan publik, media, dan jejak moderasi pada satu titik yang sama.
						</p>
					</div>

					<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
						<div class="metric-card">
							<div class="flex items-center gap-2 text-slate-500">
								<DocumentIcon class="size-[18px]" />
								<span class="metric-label">Laporan</span>
							</div>
							<p class="metric-value">{detail.submission_count}</p>
							<p class="metric-copy">Total submission masuk</p>
						</div>
						<div class="metric-card">
							<div class="flex items-center gap-2 text-slate-500">
								<MapIcon class="size-[18px]" />
								<span class="metric-label">Foto</span>
							</div>
							<p class="metric-value">{detail.photo_count}</p>
							<p class="metric-copy">Media bukti tersedia</p>
						</div>
						<div class="metric-card">
							<div class="flex items-center gap-2 text-slate-500">
								<DangerIcon class="size-[18px]" />
								<span class="metric-label">Korban</span>
							</div>
							<p class="metric-value">{detail.casualty_count}</p>
							<p class="metric-copy">Jumlah korban tercatat</p>
						</div>
						<div class="metric-card">
							<div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Terakhir</div>
							<p class="mt-2 text-lg font-[800] tracking-[-0.03em] text-slate-950">
								{relativeTime(detail.last_seen_at)}
							</p>
							<p class="metric-copy">{formatDate(detail.last_seen_at)}</p>
						</div>
					</div>
				</div>

				<div class="jedug-panel space-y-4 p-4">
					<div class="flex items-center gap-3">
						<div class="flex size-10 items-center justify-center rounded-2xl bg-white text-brand-600 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
							<ShieldCheckIcon class="size-5" />
						</div>
						<div>
							<p class="text-sm font-bold text-slate-900">Aksi moderasi</p>
							<p class="text-xs leading-5 text-slate-500">Pisahkan alasan operasional dan tindakan utama.</p>
						</div>
					</div>

					{#if actionError}
						<div class="error-panel">{actionError}</div>
					{/if}

					<label class="input-shell">
						<span class="input-label">Alasan moderasi</span>
						<input
							type="text"
							class="input-field"
							placeholder="Opsional, tapi disarankan"
							bind:value={reasonInput}
							disabled={!!actionLoading}
						/>
					</label>

					<div class="grid gap-2">
						{#if detail.is_hidden}
							<button class="btn-secondary w-full justify-start" onclick={() => doAction('unhide', adminUnhideIssue)} disabled={!!actionLoading}>
								<EyeIcon class="size-[18px]" />
								{actionLoading === 'unhide' ? 'Memproses...' : 'Tampilkan kembali ke publik'}
							</button>
						{:else}
							<button class="btn-danger w-full justify-start" onclick={() => doAction('hide', adminHideIssue)} disabled={!!actionLoading}>
								<EyeClosedIcon class="size-[18px]" />
								{actionLoading === 'hide' ? 'Memproses...' : 'Sembunyikan dari publik'}
							</button>
						{/if}

						{#if detail.status === 'open'}
							<button class="btn-primary w-full justify-start" onclick={() => doAction('fix', adminFixIssue)} disabled={!!actionLoading}>
								<ShieldCheckIcon class="size-[18px]" />
								{actionLoading === 'fix' ? 'Memproses...' : 'Tandai selesai'}
							</button>
							<button class="btn-secondary w-full justify-start border-rose-200 bg-rose-50 text-rose-700 hover:bg-rose-100" onclick={() => doAction('reject', adminRejectIssue)} disabled={!!actionLoading}>
								<TrashIcon class="size-[18px]" />
								{actionLoading === 'reject' ? 'Memproses...' : 'Tolak issue'}
							</button>
						{/if}
					</div>
				</div>
			</div>
		</section>

		<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
			<div class="space-y-5">
				<section class="admin-card p-5">
					<h2 class="text-lg font-bold text-slate-950">Metadata issue</h2>
					<div class="mt-4 grid gap-3 sm:grid-cols-2">
						<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-4">
							<p class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Severity</p>
							<p class="mt-2 text-sm font-semibold text-slate-900">
								{severityLabel(detail.severity_current)} (max: {severityLabel(detail.severity_max)})
							</p>
						</div>
						<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-4">
							<p class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Lokasi</p>
							<p class="mt-2 text-sm font-semibold text-slate-900">
								{detail.road_name || `${detail.latitude.toFixed(6)}, ${detail.longitude.toFixed(6)}`}
							</p>
						</div>
						<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-4">
							<p class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Pertama terlihat</p>
							<p class="mt-2 text-sm font-semibold text-slate-900">{formatDate(detail.first_seen_at)}</p>
						</div>
						<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-4">
							<p class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Terakhir terlihat</p>
							<p class="mt-2 text-sm font-semibold text-slate-900">{relativeTime(detail.last_seen_at)}</p>
						</div>
					</div>
				</section>

				{#if detail.media.length > 0}
					<section class="admin-card p-5">
						<h2 class="text-lg font-bold text-slate-950">Foto bukti</h2>
						<div class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
							{#each detail.media as media}
								<a href={media.public_url} target="_blank" rel="noopener noreferrer" class="overflow-hidden rounded-[22px] border border-slate-200 bg-slate-100 shadow-[0_10px_24px_rgba(15,23,42,0.05)] transition hover:-translate-y-0.5 hover:border-slate-300">
									<img src={media.public_url} alt="Evidence" class="h-48 w-full object-cover" loading="lazy" />
								</a>
							{/each}
						</div>
					</section>
				{/if}

				{#if detail.submissions.length > 0}
					<section class="admin-card overflow-hidden">
						<div class="px-5 py-5">
							<h2 class="text-lg font-bold text-slate-950">Submission terkait</h2>
						</div>
						<div class="overflow-x-auto">
							<table class="min-w-full text-left">
								<thead class="bg-slate-50/90">
									<tr class="border-b border-slate-200">
										<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Device</th>
										<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Status</th>
										<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Severity</th>
										<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Korban</th>
										<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Catatan</th>
										<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Waktu</th>
										<th class="px-4 py-3 text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Aksi</th>
									</tr>
								</thead>
								<tbody>
									{#each detail.submissions as sub}
										<tr class={`border-b border-slate-100 align-top ${sub.device_is_banned ? 'bg-rose-50/60' : ''}`}>
											<td class="px-4 py-4">
												<div class="flex flex-wrap items-center gap-2">
													<code class="rounded-full bg-slate-100 px-2 py-1 text-xs text-slate-600">{sub.device_id.slice(0, 8)}</code>
													{#if sub.device_is_banned}
														<span class="inline-flex rounded-full border border-rose-200 bg-rose-50 px-3 py-1 text-xs font-semibold text-rose-700">Banned</span>
													{/if}
												</div>
											</td>
											<td class="px-4 py-4 text-sm text-slate-700">{sub.status}</td>
											<td class="px-4 py-4 text-sm text-slate-700">{severityLabel(sub.severity)}</td>
											<td class="px-4 py-4 text-sm text-slate-700">{sub.has_casualty ? 'Ya' : 'Tidak'}</td>
											<td class="px-4 py-4 text-sm text-slate-500">{sub.note ?? '—'}</td>
											<td class="px-4 py-4 text-sm text-slate-500">{relativeTime(sub.reported_at)}</td>
											<td class="px-4 py-4">
												{#if !sub.device_is_banned}
													<button class="btn-danger min-h-10 px-4 py-2" onclick={() => handleBanDevice(sub.device_id)} disabled={!!actionLoading}>
														Ban device
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
			</div>

			{#if detail.moderation_log.length > 0}
				<section class="admin-card p-5">
					<h2 class="text-lg font-bold text-slate-950">Log moderasi</h2>
					<div class="mt-4 space-y-3">
						{#each detail.moderation_log as log}
							<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-4">
								<p class="text-sm font-semibold text-slate-900">{actionTypeLabel(log.action_type)}</p>
								<p class="mt-1 text-xs leading-5 text-slate-500">oleh {log.admin_username ?? 'system'}</p>
								{#if log.note}
									<p class="mt-2 text-sm leading-6 text-slate-600">{log.note}</p>
								{/if}
								<p class="mt-3 text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">
									{formatDate(log.created_at)}
								</p>
							</div>
						{/each}
					</div>
				</section>
			{/if}
		</div>
	</div>
{/if}
