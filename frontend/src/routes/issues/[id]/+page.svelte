<script lang="ts">
	import { navigating } from '$app/state';
	import { ApiError } from '$lib/api/client';
	import { getIssue } from '$lib/api/issues';
	import type { IssueDetail } from '$lib/api/types';
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
		getIssueRegionOrCoordinates,
		getPrimaryMedia,
		getPublicIssueNote,
		getSeverityColor,
		getSeverityLabel,
		getStatusLabel,
		getStatusTone,
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
	const fallbackOgImageUrl = (() => data.seo.fallback_og_image_url)();

	const seo = $derived(buildIssueDetailSeo(issue, { canonicalUrl, fallbackOgImageUrl }));
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
	const publicNote = $derived(issue ? getPublicIssueNote(issue) : null);
	const isRouteNavigating = $derived(navigating.to?.route?.id === '/issues/[id]');

	const visibleMedia = $derived.by(() => {
		if (!issue) return [];
		return issue.media.filter((item) => !failedMediaIDs.has(item.id));
	});

	const heroMedia = $derived.by(() => {
		if (!issue) return null;
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
				return;
			}

			issue = result.data;
			failedMediaIDs = new Set();
		} catch (err) {
			if (err instanceof ApiError && (err.status === 400 || err.status === 404)) {
				notFound = true;
				issue = null;
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
</script>

<svelte:head>
	<title>{seo.title}</title>
	<meta name="description" content={seo.description} />
	<link rel="canonical" href={seo.canonical_url} />

	<meta property="og:type" content="article" />
	<meta property="og:site_name" content="JEDUG" />
	<meta property="og:title" content={seo.title} />
	<meta property="og:description" content={seo.description} />
	<meta property="og:url" content={seo.canonical_url} />
	<meta property="og:image" content={seo.og_image_url} />
	<meta property="og:image:alt" content={seo.og_image_alt} />

	<meta name="twitter:card" content={seo.twitter_card} />
	<meta name="twitter:title" content={seo.title} />
	<meta name="twitter:description" content={seo.description} />
	<meta name="twitter:image" content={seo.og_image_url} />
</svelte:head>

<div class="issue-detail-page">
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
		<ErrorState message={errorMessage} onretry={retryFetchIssue} />
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
			onHeroSelect={() => heroMedia && openPreview(heroMedia.id)}
			onHeroError={() => heroMedia && markMediaFailed(heroMedia.id)}
		/>

		<IssueStats
			submissionCount={issue.submission_count}
			photoCount={issue.photo_count}
			casualtyCount={issue.casualty_count}
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
					<h2>Detail Issue</h2>
					<dl class="detail-list">
						<div class="detail-row">
							<dt>Nama jalan</dt>
							<dd>{issue.road_name || 'Tidak tersedia'}</dd>
						</div>
						<div class="detail-row">
							<dt>Tipe jalan</dt>
							<dd>{issue.road_type || 'Tidak tersedia'}</dd>
						</div>
						<div class="detail-row">
							<dt>Region</dt>
							<dd>{issue.region_name || 'Tidak tersedia'}</dd>
						</div>
						<div class="detail-row">
							<dt>Koordinat</dt>
							<dd>{coordinatesLabel}</dd>
						</div>
						<div class="detail-row">
							<dt>Status</dt>
							<dd>{statusLabel}</dd>
						</div>
						<div class="detail-row">
							<dt>Verifikasi</dt>
							<dd>{verificationLabel}</dd>
						</div>
					</dl>

					{#if publicNote}
						<div class="public-note">
							<span class="note-label">Catatan publik ringkas</span>
							<p>{publicNote}</p>
						</div>
					{/if}
				</section>

				{#if issue.recent_submissions.length > 0}
					<section class="detail-card">
						<div class="section-header">
							<div>
								<h2>Laporan Terbaru</h2>
								<p>Ringkasan aktivitas publik terbaru pada titik ini.</p>
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
									{#if submission.note}
										<p class="submission-note">{submission.note}</p>
									{/if}
									{#if submission.has_casualty}
										<p class="submission-flag">Laporan ini mencatat korban.</p>
									{/if}
								</article>
							{/each}
						</div>
					</section>
				{/if}
			</div>

			<div class="aside-column">
				<ShareActions
					title={seo.title}
					shareText={seo.share_text}
					shareUrl={seo.canonical_url}
					externalMapUrl={externalMapUrl}
				/>
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
	.aside-column {
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
		background: #f8fafc;
		color: #0f172a;
		font-size: 13px;
		font-weight: 700;
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

	.public-note {
		margin-top: 16px;
		padding: 14px;
		border-radius: 12px;
		border: 1px solid #e2e8f0;
		background: #f8fafc;
	}

	.note-label {
		display: block;
		margin-bottom: 6px;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		color: #64748b;
	}

	.public-note p {
		margin: 0;
		font-size: 14px;
		line-height: 1.55;
		color: #0f172a;
	}

	.submission-list {
		display: grid;
		gap: 10px;
		margin-top: 16px;
	}

	.submission-item {
		padding: 12px;
		border-radius: 12px;
		border: 1px solid #e2e8f0;
		background: #fff;
	}

	.submission-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
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
		margin-top: 8px;
		font-size: 14px;
		line-height: 1.5;
		color: #334155;
	}

	.submission-flag {
		margin-top: 8px;
		font-size: 12px;
		font-weight: 700;
		color: #b91c1c;
	}

	.lightbox-overlay {
		position: fixed;
		inset: 0;
		z-index: 1200;
		display: grid;
		place-items: center;
		padding: 16px;
		background: rgba(2, 6, 23, 0.88);
	}

	.lightbox-content {
		position: relative;
		width: 100%;
		max-width: 960px;
	}

	.lightbox-close {
		position: absolute;
		top: 12px;
		right: 12px;
		z-index: 2;
		padding: 6px 10px;
		border-radius: 999px;
		border: 1px solid rgba(255, 255, 255, 0.28);
		background: rgba(15, 23, 42, 0.72);
		color: #fff;
		font-size: 12px;
		font-weight: 700;
		cursor: pointer;
	}

	.lightbox-content img {
		display: block;
		width: 100%;
		max-height: 84vh;
		object-fit: contain;
		border-radius: 16px;
	}

	@media (min-width: 900px) {
		.issue-detail-page {
			padding-top: 20px;
		}

		.detail-layout {
			grid-template-columns: minmax(0, 1.15fr) minmax(320px, 0.85fr);
			align-items: start;
		}

		.aside-column {
			position: sticky;
			top: 76px;
		}

		.detail-row {
			flex-direction: row;
			align-items: flex-start;
			justify-content: space-between;
			gap: 20px;
		}

		dt {
			flex: 0 0 160px;
		}

		dd {
			text-align: right;
		}
	}
</style>
