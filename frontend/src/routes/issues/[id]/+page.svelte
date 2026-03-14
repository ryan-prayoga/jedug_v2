<script lang="ts">
	import { navigating } from '$app/state';
	import { ApiError } from '$lib/api/client';
	import { getIssue, getIssueTimeline } from '$lib/api/issues';
	import type { IssueDetail, IssueTimelineEvent } from '$lib/api/types';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import IssueGallery from '$lib/components/IssueGallery.svelte';
	import IssueHeader from '$lib/components/IssueHeader.svelte';
	import IssueStats from '$lib/components/IssueStats.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import ShareActions from '$lib/components/ShareActions.svelte';
	import { formatDate, relativeTime } from '$lib/utils/date';
	import {
		buildIssueDetailSeo,
		formatCoordinates,
		getIssueLocationLabel,
		getIssuePrimaryMedia,
		getIssueRegionOrCoordinates,
		getIssueSnapshot,
		getPrimaryMedia,
		getPublicIssueNote,
		getSeverityColor,
		getSeverityLabel,
		getStatusLabel,
		getStatusTone,
		getSubmissionPublicNote,
		getVerificationLabel,
		getVerificationTone
	} from '$lib/utils/issue-detail';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let issue = $state<IssueDetail | null>((() => data.issue)());
	let loading = $state(false);
	let errorMessage = $state<string | null>((() => data.loadError)());
	let notFound = $state((() => data.notFound)());
	let previewMediaID = $state<string | null>(null);
	let failedMediaIDs = $state<Set<string>>(new Set());

	const canonicalUrl = (() => data.seo.canonical_url)();
	const ogImageUrl = (() => data.seo.og_image_url)();

	const seo = $derived(buildIssueDetailSeo(issue, { canonicalUrl, ogImageUrl }));
	const timelinePageSize = 100;
	let timelineEvents = $state<IssueTimelineEvent[]>([]);
	let timelineLoading = $state(false);
	let timelineLoadingMore = $state(false);
	let timelineError = $state<string | null>(null);
	let timelineOffset = $state(0);
	let timelineHasMore = $state(false);
	let timelineIssueID = $state<string | null>(null);

	const locationLabel = $derived(issue ? getIssueLocationLabel(issue) : '-');
	const locationContext = $derived(issue ? getIssueRegionOrCoordinates(issue) : '-');
	const coordinatesLabel = $derived(
		issue ? formatCoordinates(issue.latitude, issue.longitude, 5) : '-'
	);
	const severityLabel = $derived(issue ? getSeverityLabel(issue.severity_current) : '-');
	const statusLabel = $derived(issue ? getStatusLabel(issue.status) : '-');
	const verificationLabel = $derived(
		issue ? getVerificationLabel(issue.verification_status) : 'Belum Diverifikasi'
	);
	const statusTone = $derived(issue ? getStatusTone(issue.status) : getStatusTone('open'));
	const verificationTone = $derived(
		issue ? getVerificationTone(issue.verification_status) : getVerificationTone('unverified')
	);
	const severityColor = $derived(issue ? getSeverityColor(issue.severity_current) : '#94A3B8');
	const issueSnapshot = $derived(issue ? getIssueSnapshot(issue) : '');
	const publicNote = $derived(issue ? getPublicIssueNote(issue) : null);
	const isRouteNavigating = $derived(navigating.to?.route?.id === '/issues/[id]');

	const visibleMedia = $derived.by(() => {
		if (!issue) return [];
		return issue.media.filter((item) => !failedMediaIDs.has(item.id));
	});

	const heroMedia = $derived.by(() => {
		if (!issue) return null;

		const primaryMedia = getIssuePrimaryMedia(issue);
		if (primaryMedia && !failedMediaIDs.has(primaryMedia.id)) {
			return primaryMedia;
		}

		return getPrimaryMedia(visibleMedia) || null;
	});

	const previewMedia = $derived.by(() => {
		if (!previewMediaID) return null;
		return visibleMedia.find((item) => item.id === previewMediaID) || null;
	});

	const externalMapUrl = $derived.by(() => {
		if (!issue) return '#';
		return `https://www.google.com/maps?q=${issue.latitude},${issue.longitude}`;
	});

	function resetTimelineState() {
		timelineEvents = [];
		timelineOffset = 0;
		timelineHasMore = false;
		timelineError = null;
		timelineLoading = false;
		timelineLoadingMore = false;
	}

	function readNumberFromData(data: Record<string, unknown>, key: string): number | null {
		const value = data[key];
		return typeof value === 'number' ? value : null;
	}

	function readStringFromData(data: Record<string, unknown>, key: string): string | null {
		const value = data[key];
		return typeof value === 'string' ? value : null;
	}

	function getTimelineEventTitle(event: IssueTimelineEvent): string {
		switch (event.type) {
			case 'issue_created':
				return 'Laporan pertama dibuat';
			case 'photo_added':
				return 'Foto tambahan ditambahkan';
			case 'severity_changed':
				return 'Severity diperbarui';
			case 'casualty_reported':
				return 'Korban dilaporkan';
			case 'status_updated': {
				const toStatus = readStringFromData(event.data, 'to_status');
				if (!toStatus) return 'Status issue diperbarui';
				return `Status diperbarui: ${getStatusLabel(toStatus)}`;
			}
			default:
				return 'Aktivitas issue diperbarui';
		}
	}

	function getTimelineEventMeta(event: IssueTimelineEvent): string | null {
		switch (event.type) {
			case 'photo_added': {
				const photoCount = readNumberFromData(event.data, 'photo_count');
				if (!photoCount || photoCount <= 0) return null;
				return `${photoCount} foto ditambahkan.`;
			}
			case 'severity_changed': {
				const from = readNumberFromData(event.data, 'from');
				const to = readNumberFromData(event.data, 'to');
				if (from == null || to == null) return null;
				return `${getSeverityLabel(from)} → ${getSeverityLabel(to)}.`;
			}
			case 'casualty_reported': {
				const to = readNumberFromData(event.data, 'to');
				if (to == null) return null;
				return `Jumlah korban terlapor: ${to}.`;
			}
			case 'status_updated': {
				const fromStatus = readStringFromData(event.data, 'from_status');
				const toStatus = readStringFromData(event.data, 'to_status');
				if (!fromStatus || !toStatus) return null;
				return `${getStatusLabel(fromStatus)} → ${getStatusLabel(toStatus)}.`;
			}
			default:
				return null;
		}
	}

	async function fetchTimeline(issueID: string, append: boolean) {
		if (append) {
			timelineLoadingMore = true;
		} else {
			timelineLoading = true;
			timelineError = null;
		}

		const requestOffset = append ? timelineOffset : 0;

		try {
			const result = await getIssueTimeline(issueID, {
				limit: timelinePageSize,
				offset: requestOffset
			});

			if (timelineIssueID !== issueID) return;

			const incoming = result.data || [];
			timelineEvents = append ? [...timelineEvents, ...incoming] : incoming;
			timelineOffset = requestOffset + incoming.length;
			timelineHasMore = incoming.length === timelinePageSize;
		} catch (err) {
			if (timelineIssueID !== issueID) return;
			timelineError =
				err instanceof Error ? err.message : 'Gagal memuat riwayat laporan issue saat ini.';
		} finally {
			if (timelineIssueID === issueID) {
				timelineLoading = false;
				timelineLoadingMore = false;
			}
		}
	}

	async function loadInitialTimeline(issueID: string) {
		timelineIssueID = issueID;
		resetTimelineState();
		await fetchTimeline(issueID, false);
	}

	async function loadMoreTimeline() {
		if (!issue || timelineLoadingMore || !timelineHasMore) return;
		await fetchTimeline(issue.id, true);
	}

	function markMediaFailed(mediaID: string) {
		const next = new Set(failedMediaIDs);
		next.add(mediaID);
		failedMediaIDs = next;

		if (previewMediaID === mediaID) {
			previewMediaID = null;
		}
	}

	function openPreview(mediaID: string) {
		if (!visibleMedia.some((item) => item.id === mediaID)) return;
		previewMediaID = mediaID;
	}

	function closePreview() {
		previewMediaID = null;
	}

	async function retryFetchIssue() {
		loading = true;
		errorMessage = null;
		notFound = false;

		try {
			const result = await getIssue(data.id);
			if (!result.data) {
				notFound = true;
				issue = null;
				resetTimelineState();
				timelineIssueID = null;
				return;
			}

			issue = result.data;
			failedMediaIDs = new Set();
			timelineIssueID = null;
		} catch (err) {
			if (err instanceof ApiError && (err.status === 400 || err.status === 404)) {
				notFound = true;
				issue = null;
				resetTimelineState();
				timelineIssueID = null;
				return;
			}

			errorMessage = err instanceof Error ? err.message : 'Gagal memuat detail issue.';
		} finally {
			loading = false;
		}
	}

	function handlePreviewOverlayClick(event: MouseEvent) {
		if (event.currentTarget !== event.target) return;
		closePreview();
	}

	function handlePreviewOverlayKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' || event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			closePreview();
		}
	}

	$effect(() => {
		const issueID = issue?.id;
		if (!issueID) {
			resetTimelineState();
			timelineIssueID = null;
			return;
		}
		if (timelineIssueID === issueID) return;

		void loadInitialTimeline(issueID);
	});
</script>

<svelte:head>
	<title>{seo.title}</title>
	<meta name="description" content={seo.description} />
	<meta name="robots" content="index,follow,max-image-preview:large" />
	<link rel="canonical" href={seo.canonical_url} />

	<meta property="og:type" content="article" />
	<meta property="og:site_name" content="JEDUG" />
	<meta property="og:locale" content="id_ID" />
	<meta property="og:title" content={seo.title} />
	<meta property="og:description" content={seo.description} />
	<meta property="og:url" content={seo.canonical_url} />
	<meta property="og:image" content={seo.og_image_url} />
	<meta property="og:image:alt" content={seo.og_image_alt} />
	{#if seo.og_image_width}
		<meta property="og:image:width" content={seo.og_image_width} />
	{/if}
	{#if seo.og_image_height}
		<meta property="og:image:height" content={seo.og_image_height} />
	{/if}
	{#if issue}
		<meta property="article:published_time" content={issue.first_seen_at} />
		<meta property="article:modified_time" content={issue.updated_at} />
	{/if}

	<meta name="twitter:card" content={seo.twitter_card} />
	<meta name="twitter:title" content={seo.title} />
	<meta name="twitter:description" content={seo.description} />
	<meta name="twitter:image" content={seo.og_image_url} />
	<meta name="twitter:image:alt" content={seo.og_image_alt} />
</svelte:head>

<div class="issue-detail-page">
	<a class="page-back" href="/issues">Kembali ke Peta</a>

	{#if loading}
		<LoadingState message="Memuat detail issue..." />
	{:else if notFound}
		<EmptyState
			icon="🛣️"
			message="Issue tidak ditemukan atau tidak tersedia untuk publik."
			ctaHref="/issues"
			ctaLabel="Kembali ke Peta"
		/>
	{:else if errorMessage}
		<div class="state-shell">
			<ErrorState message={errorMessage} onretry={retryFetchIssue} />
			<a class="secondary-link" href="/issues">Lihat issue lain di peta</a>
		</div>
	{:else if issue}
		{#if isRouteNavigating}
			<div class="page-loading-indicator">Memuat halaman issue...</div>
		{/if}

		<IssueHeader
			{issue}
			{locationLabel}
			{locationContext}
			{coordinatesLabel}
			{severityLabel}
			{severityColor}
			{statusLabel}
			{statusTone}
			{verificationLabel}
			{verificationTone}
			{heroMedia}
			snapshot={issueSnapshot}
			onHeroSelect={() => heroMedia && openPreview(heroMedia.id)}
			onHeroError={() => heroMedia && markMediaFailed(heroMedia.id)}
		/>

		<IssueStats
			submissionCount={issue.submission_count}
			photoCount={issue.photo_count}
			casualtyCount={issue.casualty_count}
			reactionCount={issue.reaction_count}
			updatedAt={issue.updated_at}
		/>

		<IssueGallery
			media={visibleMedia}
			locationLabel={locationLabel}
			totalPhotoCount={issue.photo_count}
			onSelectMedia={openPreview}
			onMediaError={markMediaFailed}
		/>

		<div class="detail-layout">
			<div class="content-column">
				<section class="detail-card">
					<div class="section-header">
						<div>
							<h2>Ringkasan Publik</h2>
							<p>Informasi utama yang aman ditampilkan untuk pressure publik dan share link.</p>
						</div>
					</div>

					{#if publicNote}
						<div class="public-note">
							<span class="note-label">Catatan ringkas</span>
							<p>{publicNote}</p>
						</div>
					{/if}

					<dl class="detail-list">
						<div class="detail-row">
							<dt>Nama jalan</dt>
							<dd>{issue.road_name || 'Belum tersedia'}</dd>
						</div>
						<div class="detail-row">
							<dt>Tipe jalan</dt>
							<dd>{issue.road_type || 'Belum tersedia'}</dd>
						</div>
						<div class="detail-row">
							<dt>Wilayah</dt>
							<dd>{issue.region_name || 'Belum tersedia'}</dd>
						</div>
						<div class="detail-row">
							<dt>Koordinat</dt>
							<dd>{coordinatesLabel}</dd>
						</div>
						<div class="detail-row">
							<dt>Pertama terlihat</dt>
							<dd>{formatDate(issue.first_seen_at)}</dd>
						</div>
						<div class="detail-row">
							<dt>Terakhir terlihat</dt>
							<dd>{formatDate(issue.last_seen_at)}</dd>
						</div>
					</dl>
				</section>

				<section class="detail-card">
					<div class="section-header">
						<div>
							<h2>Status & Jejak Waktu</h2>
							<p>Ringkasan lifecycle publik issue ini.</p>
						</div>
					</div>

					<div class="timeline-grid">
						<article class="timeline-item">
							<span class="timeline-label">Status</span>
							<strong>{statusLabel}</strong>
							<small>Diperbarui {relativeTime(issue.updated_at)}</small>
						</article>
						<article class="timeline-item">
							<span class="timeline-label">Verifikasi</span>
							<strong>{verificationLabel}</strong>
							<small>{issueSnapshot}</small>
						</article>
						<article class="timeline-item">
							<span class="timeline-label">Terakhir terlihat</span>
							<strong>{relativeTime(issue.last_seen_at)}</strong>
							<small>{formatDate(issue.last_seen_at)}</small>
						</article>
					</div>
				</section>

				{#if issue.recent_submissions.length > 0}
					<section class="detail-card">
						<div class="section-header">
							<div>
								<h2>Aktivitas Laporan Terbaru</h2>
								<p>Ringkasan aman dari laporan publik terbaru di titik yang sama.</p>
							</div>
							<span>{issue.recent_submissions.length}</span>
						</div>

						<div class="submission-list">
							{#each issue.recent_submissions as submission (submission.id)}
								<article class="submission-item">
									<div class="submission-head">
										<strong>{getSeverityLabel(submission.severity)}</strong>
										<span>{relativeTime(submission.reported_at)}</span>
									</div>
									<p class="submission-meta">{formatDate(submission.reported_at)}</p>
									{#if getSubmissionPublicNote(submission)}
										<p class="submission-note">{getSubmissionPublicNote(submission)}</p>
									{/if}
									{#if submission.has_casualty && submission.casualty_count > 0}
										<p class="submission-flag">
											Laporan ini mencatat {submission.casualty_count} korban.
										</p>
									{:else if submission.has_casualty}
										<p class="submission-flag">Laporan ini mencatat korban.</p>
									{/if}
								</article>
							{/each}
						</div>
					</section>
				{/if}

				<section class="detail-card">
					<div class="section-header">
						<div>
							<h2>Riwayat Laporan</h2>
							<p>Jejak perkembangan issue untuk transparansi publik.</p>
						</div>
					</div>

					{#if timelineLoading}
						<p class="timeline-state">Memuat riwayat laporan...</p>
					{:else if timelineError}
						<div class="timeline-state timeline-state-error">
							<p>{timelineError}</p>
							<button type="button" class="timeline-button" onclick={() => issue && loadInitialTimeline(issue.id)}>
								Coba lagi
							</button>
						</div>
					{:else if timelineEvents.length === 0}
						<p class="timeline-state">Belum ada riwayat event untuk issue ini.</p>
					{:else}
						<ol class="issue-timeline" aria-label="Riwayat laporan issue">
							{#each timelineEvents as event, index (`${event.created_at}-${event.type}-${index}`)}
								<li class="timeline-event">
									<div class="timeline-dot" aria-hidden="true"></div>
									<div class="timeline-content">
										<p class="timeline-date">{formatDate(event.created_at)}</p>
										<p class="timeline-title">{getTimelineEventTitle(event)}</p>
										{#if getTimelineEventMeta(event)}
											<p class="timeline-meta">{getTimelineEventMeta(event)}</p>
										{/if}
									</div>
								</li>
							{/each}
						</ol>

						{#if timelineHasMore}
							<button
								type="button"
								class="timeline-button"
								disabled={timelineLoadingMore}
								onclick={loadMoreTimeline}
							>
								{timelineLoadingMore ? 'Memuat...' : 'Muat event lebih lama'}
							</button>
						{/if}
					{/if}
				</section>
			</div>

			<div class="aside-column">
				<div class="aside-stack">
					<ShareActions
						title={seo.title}
						shareText={seo.share_text}
						shareUrl={seo.canonical_url}
						externalMapUrl={externalMapUrl}
					/>

					<section class="detail-card compact-card">
						<h2>Lokasi Singkat</h2>
						<p class="compact-text">{locationContext}</p>
						<p class="compact-text muted">{coordinatesLabel}</p>
					</section>
				</div>
			</div>
		</div>
	{/if}
</div>

{#if previewMedia}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="lightbox-overlay"
		role="button"
		tabindex="0"
		aria-label="Tutup preview foto issue"
		onclick={handlePreviewOverlayClick}
		onkeydown={handlePreviewOverlayKeydown}
	>
		<div class="lightbox-content">
			<button type="button" class="lightbox-close" onclick={closePreview} aria-label="Tutup preview foto">
				Tutup
			</button>
			<img
				src={previewMedia.public_url}
				alt={`Preview foto issue jalan rusak di ${locationLabel}`}
				onerror={() => markMediaFailed(previewMedia.id)}
			/>
		</div>
	</div>
{/if}

<style>
	.issue-detail-page {
		padding-top: 16px;
		padding-bottom: 40px;
		display: grid;
		gap: 16px;
	}

	.page-back,
	.secondary-link {
		width: fit-content;
		font-size: 13px;
		font-weight: 700;
		color: #e5484d;
		text-decoration: none;
	}

	.page-back:hover,
	.secondary-link:hover {
		text-decoration: underline;
	}

	.state-shell {
		display: grid;
		justify-items: center;
		gap: 12px;
	}

	.page-loading-indicator {
		position: sticky;
		top: 58px;
		z-index: 12;
		width: fit-content;
		background: rgba(15, 23, 42, 0.92);
		color: #fff;
		font-size: 12px;
		font-weight: 700;
		padding: 8px 12px;
		border-radius: 999px;
	}

	.detail-layout,
	.content-column,
	.aside-column,
	.aside-stack {
		display: grid;
		gap: 16px;
	}

	.detail-card {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 16px;
		padding: 16px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
	}

	h2 {
		margin: 0;
		font-size: 18px;
		color: #0f172a;
	}

	.section-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
	}

	.section-header p {
		margin-top: 4px;
		font-size: 13px;
		line-height: 1.5;
		color: #64748b;
	}

	.section-header span {
		min-width: 36px;
		height: 36px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: #fff1f2;
		color: #e5484d;
		font-size: 13px;
		font-weight: 700;
	}

	.public-note {
		margin-top: 16px;
		padding: 14px;
		border-radius: 12px;
		border: 1px solid #e2e8f0;
		background: #f8fafc;
	}

	.note-label,
	.timeline-label {
		display: block;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		color: #64748b;
	}

	.public-note p {
		margin-top: 8px;
		font-size: 14px;
		line-height: 1.6;
		color: #0f172a;
	}

	.detail-list {
		display: grid;
		gap: 10px;
		margin-top: 16px;
	}

	.detail-row {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding-bottom: 10px;
		border-bottom: 1px solid #f1f5f9;
	}

	.detail-row:last-child {
		padding-bottom: 0;
		border-bottom: none;
	}

	dt {
		font-size: 12px;
		font-weight: 700;
		color: #64748b;
	}

	dd {
		margin: 0;
		font-size: 14px;
		line-height: 1.5;
		color: #0f172a;
	}

	.timeline-grid {
		margin-top: 16px;
		display: grid;
		gap: 10px;
	}

	.timeline-state {
		margin-top: 16px;
		font-size: 13px;
		line-height: 1.5;
		color: #475569;
	}

	.timeline-state-error {
		display: grid;
		gap: 10px;
	}

	.issue-timeline {
		margin: 16px 0 0;
		padding: 0;
		list-style: none;
		display: grid;
		gap: 0;
	}

	.timeline-event {
		position: relative;
		display: grid;
		grid-template-columns: 20px minmax(0, 1fr);
		column-gap: 12px;
		padding-bottom: 14px;
	}

	.timeline-event:last-child {
		padding-bottom: 0;
	}

	.timeline-event:not(:last-child)::after {
		content: '';
		position: absolute;
		left: 9px;
		top: 12px;
		bottom: -2px;
		width: 2px;
		background: #e2e8f0;
	}

	.timeline-dot {
		width: 10px;
		height: 10px;
		margin-top: 3px;
		border-radius: 999px;
		background: #e5484d;
		box-shadow: 0 0 0 3px #ffe4e6;
	}

	.timeline-content {
		padding: 0 0 0 2px;
	}

	.timeline-date {
		margin: 0;
		font-size: 12px;
		font-weight: 700;
		color: #64748b;
	}

	.timeline-title {
		margin: 5px 0 0;
		font-size: 14px;
		font-weight: 700;
		line-height: 1.5;
		color: #0f172a;
	}

	.timeline-meta {
		margin: 6px 0 0;
		font-size: 12px;
		line-height: 1.5;
		color: #64748b;
	}

	.timeline-button {
		margin-top: 14px;
		border: 1px solid #fecdd3;
		background: #fff1f2;
		color: #9f1239;
		font-size: 13px;
		font-weight: 700;
		border-radius: 10px;
		padding: 9px 12px;
		cursor: pointer;
	}

	.timeline-button:disabled {
		opacity: 0.7;
		cursor: default;
	}

	.timeline-item {
		padding: 14px;
		border-radius: 12px;
		background: #f8fafc;
		border: 1px solid #e2e8f0;
	}

	.timeline-item strong {
		display: block;
		margin-top: 8px;
		font-size: 16px;
		color: #0f172a;
	}

	.timeline-item small {
		display: block;
		margin-top: 6px;
		font-size: 12px;
		line-height: 1.5;
		color: #64748b;
	}

	.submission-list {
		margin-top: 16px;
		display: grid;
		gap: 12px;
	}

	.submission-item {
		padding: 14px;
		border-radius: 12px;
		border: 1px solid #e2e8f0;
		background: #fff;
	}

	.submission-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}

	.submission-head strong {
		font-size: 14px;
		color: #0f172a;
	}

	.submission-head span,
	.submission-meta {
		font-size: 12px;
		color: #64748b;
	}

	.submission-meta {
		margin-top: 4px;
	}

	.submission-note {
		margin-top: 10px;
		font-size: 14px;
		line-height: 1.6;
		color: #0f172a;
	}

	.submission-flag {
		margin-top: 10px;
		padding: 9px 10px;
		border-radius: 10px;
		background: #fff7ed;
		color: #9a3412;
		font-size: 12px;
		font-weight: 700;
	}

	.compact-card {
		gap: 8px;
	}

	.compact-text {
		margin-top: 8px;
		font-size: 14px;
		line-height: 1.5;
		color: #0f172a;
	}

	.compact-text.muted {
		margin-top: 4px;
		color: #64748b;
	}

	.lightbox-overlay {
		position: fixed;
		inset: 0;
		z-index: 1200;
		display: grid;
		place-items: center;
		padding: 20px;
		background: rgba(15, 23, 42, 0.86);
	}

	.lightbox-content {
		position: relative;
		max-width: min(960px, 100%);
		max-height: 100%;
	}

	.lightbox-content img {
		display: block;
		max-width: 100%;
		max-height: calc(100dvh - 80px);
		border-radius: 16px;
		object-fit: contain;
	}

	.lightbox-close {
		position: absolute;
		top: 12px;
		right: 12px;
		border: 0;
		border-radius: 999px;
		padding: 10px 14px;
		background: rgba(15, 23, 42, 0.78);
		color: #fff;
		font-size: 12px;
		font-weight: 700;
		cursor: pointer;
	}

	@media (min-width: 768px) {
		.issue-detail-page {
			padding-top: 24px;
			gap: 20px;
		}

		.timeline-grid {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
	}

	@media (min-width: 960px) {
		.detail-layout {
			grid-template-columns: minmax(0, 1.5fr) minmax(320px, 0.8fr);
			align-items: start;
		}

		.aside-stack {
			position: sticky;
			top: 78px;
		}
	}
</style>
