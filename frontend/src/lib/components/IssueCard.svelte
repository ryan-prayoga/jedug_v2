<script lang="ts">
	import type { Issue } from '$lib/api/types';
	import { CameraIcon, DangerIcon, DocumentIcon, LocationIcon } from '$lib/icons';
	import { relativeTime } from '$lib/utils/date';
	import { getIssueRoadOrAreaLabel, getStatusLabel } from '$lib/utils/issue-detail';

	type IssueCardMode = 'link' | 'static';

	let { issue, mode = 'link' }: { issue: Issue; mode?: IssueCardMode } = $props();

	const severityLabel = ['', 'Ringan', 'Sedang', 'Berat', 'Parah', 'Kritis'];
	const severityColor = ['', '#F6C453', '#F97316', '#DC2626', '#DC2626', '#991B1B'];
	const statusLabel: Record<string, string> = {
		closed: 'Selesai',
		in_progress: 'Diproses',
		verified: 'Terverifikasi',
		rejected: 'Ditolak',
		merged: 'Digabung'
	};
	const statusTone: Record<string, string> = {
		open: 'border-blue-200 bg-blue-50 text-blue-700',
		fixed: 'border-slate-200 bg-slate-100 text-slate-600',
		archived: 'border-slate-200 bg-slate-100 text-slate-600',
		closed: 'border-slate-200 bg-slate-100 text-slate-600',
		in_progress: 'border-amber-200 bg-amber-50 text-amber-700',
		verified: 'border-emerald-200 bg-emerald-50 text-emerald-700',
		rejected: 'border-rose-200 bg-rose-50 text-rose-700',
		merged: 'border-slate-200 bg-slate-50 text-slate-500'
	};

	const rootTag = $derived(mode === 'link' ? 'a' : 'article');
	const rootHref = $derived(mode === 'link' ? `/issues/${issue.id}` : undefined);
	const rootClass = $derived(
		mode === 'link'
			? 'group jedug-card block overflow-hidden p-4 text-inherit transition hover:-translate-y-1 hover:shadow-[0_22px_46px_rgba(15,23,42,0.12)]'
			: 'jedug-card block overflow-hidden p-4 text-inherit'
	);
</script>

<svelte:element this={rootTag} href={rootHref} class={rootClass}>
	<div class="space-y-4">
		<div class="flex flex-wrap items-center gap-2">
			<span
				class="inline-flex items-center rounded-full px-3 py-1 text-xs font-bold text-white"
				style={`background: ${severityColor[issue.severity_current] || '#94A3B8'}`}
			>
				{severityLabel[issue.severity_current] || `Level ${issue.severity_current}`}
			</span>
			<span
				class={`inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold ${statusTone[issue.status] || statusTone.open}`}
			>
				{statusLabel[issue.status] || getStatusLabel(issue.status)}
			</span>
		</div>

		<div class="space-y-2">
			<div class="flex items-start gap-3">
				<div class="mt-0.5 flex size-10 shrink-0 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
					<LocationIcon class="size-5" />
				</div>
				<div class="min-w-0">
					<p class="text-base font-bold leading-6 text-slate-950">
						{getIssueRoadOrAreaLabel(issue) || `${issue.latitude.toFixed(4)}, ${issue.longitude.toFixed(4)}`}
					</p>
					{#if issue.road_type}
						<p class="mt-1 text-xs font-semibold uppercase tracking-[0.16em] text-slate-400">
							{issue.road_type}
						</p>
					{/if}
				</div>
			</div>
			<p class="text-sm leading-6 text-slate-500">
				Update terakhir {relativeTime(issue.last_seen_at)}. Cocok untuk scan cepat dari daftar atau panel peta.
			</p>
		</div>

		<div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
			<div class="rounded-[18px] border border-slate-200 bg-slate-50 px-3 py-3">
				<div class="flex items-center gap-2 text-slate-500">
					<DocumentIcon class="size-[18px]" />
					<span class="text-[11px] font-bold uppercase tracking-[0.16em]">Laporan</span>
				</div>
				<p class="mt-2 text-lg font-[800] tracking-[-0.03em] text-slate-950">{issue.submission_count}</p>
			</div>
			<div class="rounded-[18px] border border-slate-200 bg-slate-50 px-3 py-3">
				<div class="flex items-center gap-2 text-slate-500">
					<CameraIcon class="size-[18px]" />
					<span class="text-[11px] font-bold uppercase tracking-[0.16em]">Foto</span>
				</div>
				<p class="mt-2 text-lg font-[800] tracking-[-0.03em] text-slate-950">{issue.photo_count}</p>
			</div>
			<div
				class={`rounded-[18px] border px-3 py-3 ${issue.casualty_count > 0 ? 'border-rose-200 bg-rose-50/70' : 'border-slate-200 bg-slate-50'}`}
			>
				<div class="flex items-center gap-2 text-slate-500">
					<DangerIcon class="size-[18px]" />
					<span class="text-[11px] font-bold uppercase tracking-[0.16em]">Korban</span>
				</div>
				<p class:text-rose-700={issue.casualty_count > 0} class="mt-2 text-lg font-[800] tracking-[-0.03em] text-slate-950">
					{issue.casualty_count}
				</p>
			</div>
			<div class="rounded-[18px] border border-slate-200 bg-slate-50 px-3 py-3">
				<div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-500">Status</div>
				<p class="mt-2 text-sm font-semibold text-slate-800">{statusLabel[issue.status] || getStatusLabel(issue.status)}</p>
			</div>
		</div>
	</div>
</svelte:element>
