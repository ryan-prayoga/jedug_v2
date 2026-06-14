<script lang="ts">
	import type { IssueDetail, MediaItem } from '$lib/api/types';
	import { formatDate, relativeTimeLabel } from '$lib/utils/date';
	import { CheckCircleIcon, InfoIcon, LocationIcon } from '$lib/icons';

	type Tone = {
		bg: string;
		text: string;
	};

	let {
		issue,
		locationLabel,
		locationContext,
		secondaryLocationLine,
		severityLabel,
		severityColor,
		statusLabel,
		statusTone,
		verificationLabel,
		verificationTone,
		heroMedia,
		snapshot,
		onHeroSelect = () => {},
		onHeroError = () => {}
	}: {
		issue: IssueDetail;
		locationLabel: string;
		locationContext: string;
		secondaryLocationLine: string | null;
		severityLabel: string;
		severityColor: string;
		statusLabel: string;
		statusTone: Tone;
		verificationLabel: string;
		verificationTone: Tone;
		heroMedia: MediaItem | null;
		snapshot: string;
		onHeroSelect?: () => void;
		onHeroError?: () => void;
	} = $props();

	const statusData: Record<string, string> = {
		fixed: 'fixed',
		closed: 'fixed',
		archived: 'archived',
		merged: 'archived',
		rejected: 'rejected'
	};

	const metaItems = $derived.by(() => [
		{
			label: 'Wilayah',
			value: locationContext,
			description: 'Konteks area publik'
		},
		{
			label: 'Pertama terlihat',
			value: formatDate(issue.first_seen_at),
			description: 'Laporan paling awal'
		},
		{
			label: 'Terakhir terlihat',
			value: relativeTimeLabel(issue.last_seen_at),
			description: formatDate(issue.last_seen_at)
		}
	]);
</script>

<section class="grid gap-4 md:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)]">
	<div class="overflow-hidden rounded-[4px] border border-hairline bg-ink">
		{#if heroMedia}
			<button
				type="button"
				class="group relative block w-full cursor-zoom-in border-0 bg-transparent p-0 text-left"
				onclick={onHeroSelect}
				aria-label={`Buka foto utama issue di ${locationLabel}`}
			>
				<img
					src={heroMedia.public_url}
					alt={`Foto issue jalan rusak di ${locationLabel}`}
					class="block h-[280px] w-full object-cover transition duration-300 group-hover:scale-[1.02] md:h-full md:min-h-[360px]"
					loading="eager"
					decoding="async"
					onerror={onHeroError}
				/>
				<div class="absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent"></div>
				<div class="absolute inset-x-0 bottom-0 flex items-end justify-between gap-4 px-5 py-5 text-left">
					<div class="min-w-0">
						<span class="inline-flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.18em] text-white/90">
							<LocationIcon class="size-4" />
							{locationContext}
						</span>
						<strong class="mt-3 block font-serif text-[1.65rem] font-semibold leading-tight tracking-[-0.02em] text-white md:text-[2rem]">
							{locationLabel}
						</strong>
					</div>
				</div>
			</button>
		{:else}
			<div class="grid min-h-[280px] place-content-center gap-4 bg-sunken px-6 py-8 text-center md:min-h-[360px]">
				<span class="kicker mx-auto">
					<InfoIcon class="size-4" />
					Belum ada foto utama
				</span>
				<strong class="font-serif text-[1.8rem] font-semibold leading-tight tracking-[-0.02em] text-ink">{locationLabel}</strong>
				<p class="mx-auto max-w-[34ch] text-sm leading-6 text-muted">
					Laporan ini tetap tampil publik agar kondisi jalan bisa terus dipantau sambil menunggu bukti visual tambahan.
				</p>
			</div>
		{/if}
	</div>

	<div class="jedug-card flex flex-col justify-between p-5">
		<div>
			<div class="flex flex-wrap gap-2">
				<span class="severity-pill" data-sev={issue.severity_current}>
					{severityLabel}
				</span>
				<span class="status-pill" data-status={statusData[issue.status] ?? 'open'}>
					{statusLabel}
				</span>
				<span class="badge-muted">
					<CheckCircleIcon class="size-4" />
					{verificationLabel}
				</span>
			</div>

			<h1 class="mt-4 font-serif text-[1.85rem] font-semibold leading-tight tracking-[-0.02em] text-ink">
				{locationLabel}
			</h1>
			{#if secondaryLocationLine}
				<p class="mt-2 text-sm leading-6 text-muted">{secondaryLocationLine}</p>
			{/if}
			<p class="mt-4 rounded-[4px] border border-hairline bg-sunken px-4 py-3 text-sm leading-6 text-muted">
				{snapshot}
			</p>
		</div>

		<div class="mt-5 grid gap-3 md:grid-cols-3">
			{#each metaItems as item}
				<article class="rounded-[4px] border border-hairline bg-surface px-4 py-4">
					<span class="text-[11px] font-bold uppercase tracking-[0.16em] text-subtle">{item.label}</span>
					<strong class="mt-2 block text-sm font-bold leading-6 text-ink">{item.value}</strong>
					<small class="mt-1 block text-xs leading-5 text-muted">{item.description}</small>
				</article>
			{/each}
		</div>
	</div>
</section>
