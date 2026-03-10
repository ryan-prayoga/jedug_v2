<script lang="ts">
	import { navigating } from '$app/state';
	import { ApiError } from '$lib/api/client';
	import { getIssue } from '$lib/api/issues';
	import type { IssueDetail } from '$lib/api/types';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import { formatDate, relativeTime } from '$lib/utils/date';
	import {
		buildIssueDetailSeo,
		formatCoordinates,
		getIssueLocationLabel,
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
	let shareMessage = $state<string | null>(null);
	let copyingLink = $state(false);
	let lightboxIndex = $state<number | null>(null);
	let failedMediaIDs = $state<Set<string>>(new Set());

	const canonicalUrl = (() => data.seo.canonical_url)();
	const fallbackOgImageUrl = (() => data.seo.fallback_og_image_url)();

	const seo = $derived(buildIssueDetailSeo(issue, { canonicalUrl, fallbackOgImageUrl }));
	const locationLabel = $derived(issue ? getIssueLocationLabel(issue) : '-');
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
		return issue.media.filter((media) => !failedMediaIDs.has(media.id));
	});

	const primaryMedia = $derived.by(() => {
		if (!issue) return null;
		const preferred = getPrimaryMedia(issue.media);
		if (preferred && !failedMediaIDs.has(preferred.id)) return preferred;
		return issue.media.find((media) => !failedMediaIDs.has(media.id)) || null;
	});

	const externalMapURL = $derived.by(() => {
		if (!issue) return '#';
		return `https://www.google.com/maps?q=${issue.latitude},${issue.longitude}`;
	});

	const shareUrlEncoded = $derived(encodeURIComponent(seo.canonical_url));
	const shareTextEncoded = $derived(encodeURIComponent(seo.share_text));
	const whatsappShareURL = $derived(`https://wa.me/?text=${shareTextEncoded}%20${shareUrlEncoded}`);
	const telegramShareURL = $derived(
		`https://t.me/share/url?url=${shareUrlEncoded}&text=${shareTextEncoded}`
	);
	const twitterShareURL = $derived(
		`https://twitter.com/intent/tweet?text=${shareTextEncoded}&url=${shareUrlEncoded}`
	);
	const facebookShareURL = $derived(
		`https://www.facebook.com/sharer/sharer.php?u=${shareUrlEncoded}`
	);

	const lightboxMedia = $derived.by(() => {
		if (lightboxIndex === null) return null;
		return visibleMedia[lightboxIndex] || null;
	});

	function markMediaFailed(mediaID: string) {
		const next = new Set(failedMediaIDs);
		next.add(mediaID);
		failedMediaIDs = next;
	}

	function openLightbox(mediaID: string) {
		const index = visibleMedia.findIndex((media) => media.id === mediaID);
		if (index === -1) return;
		lightboxIndex = index;
	}

	function closeLightbox() {
		lightboxIndex = null;
	}

	function showShareMessage(message: string) {
		shareMessage = message;
		setTimeout(() => {
			shareMessage = null;
		}, 2400);
	}

	async function handleShare() {
		if (!issue) return;

		if (
			typeof navigator !== 'undefined' &&
			'share' in navigator &&
			typeof navigator.share === 'function'
		) {
			try {
				await navigator.share({
					title: seo.title,
					text: seo.share_text,
					url: seo.canonical_url
				});
				showShareMessage('Link issue berhasil dibagikan.');
				return;
			} catch {
				// Fallback to copy.
			}
		}

		await copyLink();
	}

	async function copyLink() {
		if (!issue || copyingLink) return;

		copyingLink = true;
		try {
			if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
				await navigator.clipboard.writeText(seo.canonical_url);
				showShareMessage('Link issue berhasil disalin.');
			} else {
				showShareMessage('Gagal menyalin link secara otomatis.');
			}
		} catch {
			showShareMessage('Gagal menyalin link issue.');
		} finally {
			copyingLink = false;
		}
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
			if (err instanceof ApiError && err.status === 404) {
				notFound = true;
				issue = null;
				return;
			}
			errorMessage = err instanceof Error ? err.message : 'Gagal memuat detail issue';
		} finally {
			loading = false;
		}
	}

	function handleLightboxOverlayClick(event: MouseEvent) {
		if (event.currentTarget !== event.target) return;
		closeLightbox();
	}

	function handleLightboxOverlayKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' || event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			closeLightbox();
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

		<section class="hero-section">
			{#if primaryMedia}
				<button class="hero-image-button" onclick={() => openLightbox(primaryMedia.id)}>
					<img
						src={primaryMedia.public_url}
						alt={`Foto issue jalan rusak di ${locationLabel}`}
						loading="eager"
						decoding="async"
						onerror={() => markMediaFailed(primaryMedia.id)}
					/>
				</button>
			{:else}
				<div class="hero-placeholder">
					<div class="hero-placeholder-title">Belum ada foto utama</div>
					<div class="hero-placeholder-subtitle">Laporan ini tetap dapat dipantau publik.</div>
				</div>
			{/if}
		</section>

		<section class="headline-card">
			<div class="badge-row">
				<span class="badge severity" style="background: {severityColor}">{severityLabel}</span>
				<span class="badge status" style="background: {statusTone.bg}; color: {statusTone.text}">
					{statusLabel}
				</span>
				<span
					class="badge verification"
					style="background: {verificationTone.bg}; color: {verificationTone.text}"
				>
					{verificationLabel}
				</span>
			</div>
			<h1>{locationLabel}</h1>
			<p class="subline">
				{#if issue.road_type}
					{issue.road_type}
				{:else}
					Koordinat titik issue
				{/if}
				· {formatCoordinates(issue.latitude, issue.longitude, 5)}
			</p>
		</section>

		<section class="metrics-grid">
			<article class="metric-card">
				<span class="metric-label">Laporan</span>
				<span class="metric-value">{issue.submission_count}</span>
			</article>
			<article class="metric-card">
				<span class="metric-label">Foto</span>
				<span class="metric-value">{issue.photo_count}</span>
			</article>
			<article class="metric-card">
				<span class="metric-label">Korban</span>
				<span class="metric-value">{issue.casualty_count}</span>
			</article>
			<article class="metric-card">
				<span class="metric-label">Reaksi</span>
				<span class="metric-value">{issue.reaction_count}</span>
			</article>
			<article class="metric-card">
				<span class="metric-label">Visibilitas</span>
				<span class="metric-value">Tampil Publik</span>
			</article>
		</section>

		<section class="info-card">
			<h2>Informasi Utama</h2>
			<dl>
				<div class="info-row">
					<dt>Lokasi</dt>
					<dd>{locationLabel}</dd>
				</div>
				<div class="info-row">
					<dt>Pertama Terlihat</dt>
					<dd>{formatDate(issue.first_seen_at)}</dd>
				</div>
				<div class="info-row">
					<dt>Terakhir Terlihat</dt>
					<dd>{relativeTime(issue.last_seen_at)} ({formatDate(issue.last_seen_at)})</dd>
				</div>
				<div class="info-row">
					<dt>Region</dt>
					<dd>
						{#if issue.region_name}
							{issue.region_name}
						{:else if issue.region_id}
							ID {issue.region_id}
						{:else}
							Tidak tersedia
						{/if}
					</dd>
				</div>
				<div class="info-row">
					<dt>Status Verifikasi</dt>
					<dd>{verificationLabel}</dd>
				</div>
			</dl>
		</section>

		<section class="info-card">
			<h2>Informasi Tambahan</h2>
			<dl>
				<div class="info-row">
					<dt>Nama Jalan</dt>
					<dd>{issue.road_name || 'Tidak tersedia'}</dd>
				</div>
				<div class="info-row">
					<dt>Tipe Jalan</dt>
					<dd>{issue.road_type || 'Tidak tersedia'}</dd>
				</div>
			</dl>

			{#if publicNote}
				<div class="public-note">
					<p>{publicNote}</p>
				</div>
			{/if}
		</section>

		<section class="gallery-card">
			<div class="section-header">
				<h2>Galeri Issue</h2>
				<span>{visibleMedia.length} foto</span>
			</div>

			{#if visibleMedia.length > 0}
				<div class="gallery-grid">
					{#each visibleMedia as media (media.id)}
						<button class="gallery-item" onclick={() => openLightbox(media.id)}>
							<img
								src={media.public_url}
								alt={`Media issue ${locationLabel}`}
								loading="lazy"
								decoding="async"
								onerror={() => markMediaFailed(media.id)}
							/>
						</button>
					{/each}
				</div>
			{:else}
				<div class="gallery-empty">Belum ada foto tambahan untuk issue ini.</div>
			{/if}
		</section>

		{#if issue.recent_submissions.length > 0}
			<section class="activity-card">
				<div class="section-header">
					<h2>Aktivitas Terbaru</h2>
					<span>{issue.recent_submissions.length} laporan</span>
				</div>
				<div class="activity-list">
					{#each issue.recent_submissions as submission (submission.id)}
						<article class="activity-item">
							<div class="activity-meta">
								<span class="activity-severity">{getSeverityLabel(submission.severity)}</span>
								<span class="activity-time">{relativeTime(submission.reported_at)}</span>
							</div>
							{#if submission.note}
								<p>{submission.note}</p>
							{/if}
						</article>
					{/each}
				</div>
			</section>
		{/if}

		<section class="cta-card">
			<div class="cta-grid">
				<a class="btn btn-secondary" href="/issues">Kembali ke Peta</a>
				<button class="btn btn-primary" onclick={handleShare}>Bagikan Issue</button>
				<a class="btn btn-secondary" href="/lapor">Lapor di Sekitar Sini</a>
				<a class="btn btn-secondary" href={externalMapURL} target="_blank" rel="noopener noreferrer">
					Buka Lokasi di Peta
				</a>
			</div>

			<div class="social-share-links">
				<a href={whatsappShareURL} target="_blank" rel="noopener noreferrer">WhatsApp</a>
				<a href={telegramShareURL} target="_blank" rel="noopener noreferrer">Telegram</a>
				<a href={twitterShareURL} target="_blank" rel="noopener noreferrer">Twitter/X</a>
				<a href={facebookShareURL} target="_blank" rel="noopener noreferrer">Facebook</a>
				<button onclick={copyLink} disabled={copyingLink}>
					{copyingLink ? 'Menyalin...' : 'Salin Link'}
				</button>
			</div>

			{#if shareMessage}
				<p class="share-message">{shareMessage}</p>
			{/if}
		</section>
	{/if}
</div>

{#if lightboxMedia}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="lightbox-overlay"
		role="button"
		tabindex="0"
		aria-label="Tutup preview media issue"
		onclick={handleLightboxOverlayClick}
		onkeydown={handleLightboxOverlayKeydown}
	>
		<div class="lightbox-content">
			<button class="lightbox-close" onclick={closeLightbox} aria-label="Tutup preview media">
				Tutup
			</button>
			<img src={lightboxMedia.public_url} alt={`Preview media issue di ${locationLabel}`} />
		</div>
	</div>
{/if}

<style>
	.issue-detail-page {
		padding-top: 8px;
		padding-bottom: 28px;
		display: grid;
		gap: 14px;
	}

	.page-loading-indicator {
		position: sticky;
		top: 58px;
		z-index: 12;
		background: rgba(15, 23, 42, 0.9);
		color: #fff;
		font-size: 12px;
		font-weight: 600;
		padding: 8px 12px;
		border-radius: 10px;
	}

	.hero-section {
		margin: 0 -16px;
		background: #0f172a;
	}

	.hero-image-button {
		display: block;
		width: 100%;
		border: 0;
		padding: 0;
		background: #0f172a;
		cursor: zoom-in;
	}

	.hero-image-button img {
		display: block;
		width: 100%;
		height: min(58vw, 360px);
		object-fit: cover;
	}

	.hero-placeholder {
		display: grid;
		place-content: center;
		gap: 6px;
		min-height: 220px;
		padding: 20px;
		text-align: center;
		color: #e2e8f0;
		background: linear-gradient(135deg, #334155 0%, #0f172a 100%);
	}

	.hero-placeholder-title {
		font-size: 16px;
		font-weight: 700;
	}

	.hero-placeholder-subtitle {
		font-size: 13px;
		color: #cbd5e1;
	}

	.headline-card,
	.info-card,
	.gallery-card,
	.activity-card,
	.cta-card {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 16px;
		padding: 14px;
	}

	.badge-row {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 10px;
	}

	.badge {
		display: inline-flex;
		align-items: center;
		padding: 4px 10px;
		font-size: 11px;
		font-weight: 700;
		line-height: 1;
		border-radius: 999px;
	}

	.badge.severity {
		color: #fff;
	}

	.headline-card h1 {
		margin: 0;
		color: #0f172a;
		font-size: 22px;
		line-height: 1.2;
	}

	.subline {
		margin-top: 6px;
		font-size: 13px;
		color: #64748b;
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px;
	}

	.metric-card {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 12px;
		padding: 12px;
		text-align: center;
	}

	.metric-label {
		display: block;
		font-size: 11px;
		color: #94a3b8;
		text-transform: uppercase;
		letter-spacing: 0.45px;
	}

	.metric-value {
		display: block;
		margin-top: 4px;
		font-size: 16px;
		font-weight: 700;
		color: #0f172a;
	}

	.info-card h2,
	.gallery-card h2,
	.activity-card h2 {
		margin: 0;
		font-size: 16px;
		color: #0f172a;
	}

	dl {
		display: grid;
		gap: 10px;
		margin-top: 12px;
	}

	.info-row {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding-bottom: 10px;
		border-bottom: 1px solid #f1f5f9;
	}

	.info-row:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}

	dt {
		font-size: 12px;
		color: #64748b;
	}

	dd {
		margin: 0;
		color: #0f172a;
		font-size: 14px;
		line-height: 1.45;
	}

	.public-note {
		margin-top: 12px;
		padding: 12px;
		background: #f8fafc;
		border-radius: 12px;
		border: 1px solid #e2e8f0;
	}

	.public-note p {
		margin: 0;
		font-size: 14px;
		line-height: 1.5;
		color: #0f172a;
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.section-header span {
		font-size: 12px;
		color: #64748b;
	}

	.gallery-grid {
		margin-top: 12px;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px;
	}

	.gallery-item {
		padding: 0;
		border: 0;
		border-radius: 12px;
		overflow: hidden;
		background: #f1f5f9;
		cursor: zoom-in;
	}

	.gallery-item img {
		display: block;
		width: 100%;
		height: 132px;
		object-fit: cover;
	}

	.gallery-empty {
		margin-top: 12px;
		font-size: 13px;
		color: #64748b;
		padding: 12px;
		border: 1px dashed #cbd5e1;
		border-radius: 12px;
	}

	.activity-list {
		margin-top: 12px;
		display: grid;
		gap: 8px;
	}

	.activity-item {
		border: 1px solid #e2e8f0;
		border-radius: 12px;
		padding: 10px;
	}

	.activity-meta {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.activity-severity {
		font-size: 12px;
		font-weight: 700;
		color: #0f172a;
	}

	.activity-time {
		font-size: 12px;
		color: #64748b;
	}

	.activity-item p {
		margin: 8px 0 0;
		font-size: 13px;
		color: #334155;
		line-height: 1.45;
	}

	.cta-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 8px;
	}

	.btn {
		min-height: 48px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		text-align: center;
		border-radius: 12px;
		padding: 10px 16px;
		font-size: 14px;
		font-weight: 700;
		text-decoration: none;
		border: 1px solid #e2e8f0;
		transition: opacity 0.2s, transform 0.1s, background 0.2s;
		cursor: pointer;
	}

	.btn:active {
		transform: scale(0.98);
	}

	.btn-primary {
		background: #e5484d;
		color: #fff;
		border-color: #e5484d;
	}

	.btn-primary:hover {
		opacity: 0.9;
	}

	.btn-secondary {
		background: #fff;
		color: #0f172a;
	}

	.btn-secondary:hover {
		background: #f8fafc;
	}

	.social-share-links {
		margin-top: 12px;
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.social-share-links a,
	.social-share-links button {
		font-size: 12px;
		font-weight: 600;
		padding: 8px 10px;
		border-radius: 999px;
		border: 1px solid #e2e8f0;
		background: #fff;
		color: #334155;
		text-decoration: none;
		cursor: pointer;
	}

	.social-share-links button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.share-message {
		margin-top: 10px;
		color: #166534;
		background: #f0fdf4;
		border: 1px solid #bbf7d0;
		border-radius: 10px;
		padding: 8px 10px;
		font-size: 12px;
		font-weight: 600;
	}

	.lightbox-overlay {
		position: fixed;
		inset: 0;
		z-index: 1200;
		background: rgba(2, 6, 23, 0.86);
		display: grid;
		place-items: center;
		padding: 16px;
	}

	.lightbox-content {
		position: relative;
		max-width: 860px;
		width: 100%;
	}

	.lightbox-close {
		position: absolute;
		top: 10px;
		right: 10px;
		z-index: 2;
		border: 1px solid rgba(255, 255, 255, 0.3);
		background: rgba(15, 23, 42, 0.72);
		color: #fff;
		font-size: 12px;
		font-weight: 700;
		padding: 6px 10px;
		border-radius: 999px;
		cursor: pointer;
	}

	.lightbox-content img {
		display: block;
		width: 100%;
		max-height: 84vh;
		object-fit: contain;
		border-radius: 14px;
	}

	@media (min-width: 768px) {
		.issue-detail-page {
			padding-top: 16px;
			padding-bottom: 36px;
			grid-template-columns: 1fr;
			gap: 16px;
		}

		.hero-section {
			margin: 0;
			border-radius: 18px;
			overflow: hidden;
			border: 1px solid #e2e8f0;
		}

		.hero-image-button img {
			height: 360px;
		}

		.metrics-grid {
			grid-template-columns: repeat(5, minmax(0, 1fr));
		}

		.info-row {
			flex-direction: row;
			justify-content: space-between;
			align-items: flex-start;
			gap: 16px;
		}

		dt {
			flex: 0 0 180px;
		}

		dd {
			text-align: right;
		}

		.gallery-grid {
			grid-template-columns: repeat(4, minmax(0, 1fr));
		}

		.gallery-item img {
			height: 150px;
		}

		.cta-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
