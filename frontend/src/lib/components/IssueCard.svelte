<script lang="ts">
	import type { Issue } from '$lib/api/types';
	import { CameraIcon, DangerIcon, DocumentIcon } from '$lib/icons';
	import { relativeTime } from '$lib/utils/date';
	import { getIssueRoadOrAreaLabel, getStatusLabel } from '$lib/utils/issue-detail';

	type IssueCardMode = 'link' | 'static';

	let { issue, mode = 'link' }: { issue: Issue; mode?: IssueCardMode } = $props();

	const severityLabel = ['', 'Ringan', 'Sedang', 'Berat', 'Parah', 'Kritis'];
	const statusLabel: Record<string, string> = {
		closed: 'Selesai',
		in_progress: 'Diproses',
		verified: 'Terverifikasi',
		rejected: 'Ditolak',
		merged: 'Digabung'
	};
	const statusData: Record<string, string> = {
		fixed: 'fixed',
		closed: 'fixed',
		archived: 'archived',
		merged: 'archived',
		rejected: 'rejected'
	};

	const rootTag = $derived(mode === 'link' ? 'a' : 'article');
	const rootHref = $derived(mode === 'link' ? `/issues/${issue.id}` : undefined);
	const rootClass = $derived(
		mode === 'link'
			? 'group jedug-card block p-5 text-inherit transition-colors hover:border-ink'
			: 'jedug-card block p-5 text-inherit'
	);
</script>

<svelte:element this={rootTag} href={rootHref} class={rootClass}>
	<div class="space-y-4">
		<div class="flex flex-wrap items-center gap-2">
			<span class="severity-pill" data-sev={issue.severity_current}>
				{severityLabel[issue.severity_current] || `Level ${issue.severity_current}`}
			</span>
			<span class="status-pill" data-status={statusData[issue.status] ?? 'open'}>
				{statusLabel[issue.status] || getStatusLabel(issue.status)}
			</span>
		</div>

		<div class="space-y-2">
			<p class="font-serif text-lg font-semibold leading-tight text-ink">
				{getIssueRoadOrAreaLabel(issue) || `${issue.latitude.toFixed(4)}, ${issue.longitude.toFixed(4)}`}
			</p>
			{#if issue.road_type}
				<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-subtle">
					{issue.road_type}
				</p>
			{/if}
			<p class="text-sm leading-6 text-muted">
				Update terakhir {relativeTime(issue.last_seen_at)}.
			</p>
		</div>

		<div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
			<div class="metric-tile">
				<div class="flex items-center gap-2 text-subtle">
					<DocumentIcon class="size-[18px]" />
					<span class="text-[11px] font-semibold uppercase tracking-[0.16em]">Laporan</span>
				</div>
				<p class="mt-2 font-serif text-xl font-semibold tabular-nums text-ink">{issue.submission_count}</p>
			</div>
			<div class="metric-tile">
				<div class="flex items-center gap-2 text-subtle">
					<CameraIcon class="size-[18px]" />
					<span class="text-[11px] font-semibold uppercase tracking-[0.16em]">Foto</span>
				</div>
				<p class="mt-2 font-serif text-xl font-semibold tabular-nums text-ink">{issue.photo_count}</p>
			</div>
			<div class="metric-tile" class:border-brand={issue.casualty_count > 0}>
				<div class="flex items-center gap-2 text-subtle">
					<DangerIcon class="size-[18px]" />
					<span class="text-[11px] font-semibold uppercase tracking-[0.16em]">Korban</span>
				</div>
				<p class="mt-2 font-serif text-xl font-semibold tabular-nums" class:text-brand={issue.casualty_count > 0} class:text-ink={issue.casualty_count === 0}>
					{issue.casualty_count}
				</p>
			</div>
			<div class="metric-tile">
				<div class="text-[11px] font-semibold uppercase tracking-[0.16em] text-subtle">Status</div>
				<p class="mt-2 text-sm font-semibold text-ink">{statusLabel[issue.status] || getStatusLabel(issue.status)}</p>
			</div>
		</div>
	</div>
</svelte:element>
