<script lang="ts">
	import { browser } from '$app/environment';
	import { navigating } from '$app/state';
	import { onMount } from 'svelte';
	import { ApiError } from '$lib/api/client';
	import {
		followIssue,
		getIssue,
		getIssueFollowStatus,
		getIssueFollowerCount,
		getIssueTimeline,
		unfollowIssue
	} from '$lib/api/issues';
	import type { IssueDetail, IssueTimelineEvent } from '$lib/api/types';
	import BrowserPushCard from '$lib/components/BrowserPushCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import IssueGallery from '$lib/components/IssueGallery.svelte';
	import IssueHeader from '$lib/components/IssueHeader.svelte';
	import IssueStats from '$lib/components/IssueStats.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import ShareActions from '$lib/components/ShareActions.svelte';
	import {
		AddCircleIcon,
		ArrowLeftIcon,
		CameraIcon,
		CheckCircleIcon,
		ClockIcon,
		CloseCircleIcon,
		DangerIcon,
		DocumentIcon,
		HistoryIcon,
		InfoIcon,
		LocationIcon,
		NotificationIcon,
		UsersGroupIcon
	} from '$lib/icons';
	import { formatDate, relativeTime, relativeTimeLabel } from '$lib/utils/date';
	import { persistFollowerAuthFromIssueState } from '$lib/utils/follower-auth';
	import { onIssueDetailRefresh } from '$lib/utils/issue-detail-refresh';
	import { getOrCreateIssueFollowerId } from '$lib/utils/storage';
	import {
		buildIssueDetailSeo,
		formatCoordinates,
		getIssueLocationLabel,
		getIssuePrimaryMedia,
		getIssueRegionLabel,
		getIssueRegionOrCoordinates,
		getIssueRoadOrAreaLabel,
		getIssueRoadTypeLabel,
		getIssueSecondaryLocationLine,
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

	const canonicalUrl = $derived(data.seo.canonical_url);
	const ogImageUrl = $derived(data.seo.og_image_url);

	const seo = $derived(buildIssueDetailSeo(issue, { canonicalUrl, ogImageUrl }));
	const timelinePageSize = 100;
	let timelineEvents = $state<IssueTimelineEvent[]>([]);
	let timelineLoading = $state(false);
	let timelineLoadingMore = $state(false);
	let timelineError = $state<string | null>(null);
	let timelineOffset = $state(0);
	let timelineHasMore = $state(false);
	let timelineIssueID = $state<string | null>(null);
	let followerID = $state<string | null>(null);
	let followLoading = $state(false);
	let followMutating = $state(false);
	let followErrorMessage = $state<string | null>(null);
	let isFollowing = $state(false);
	let followersCount = $state(0);
	let notificationRefreshLoading = $state(false);
	let notificationRefreshMessage = $state<string | null>(null);
	let refreshMessageTimer: ReturnType<typeof setTimeout> | null = null;

	const locationLabel = $derived(issue ? getIssueLocationLabel(issue) : '-');
	const locationContext = $derived(issue ? getIssueRegionOrCoordinates(issue) : '-');
	const roadOrAreaLabel = $derived(issue ? getIssueRoadOrAreaLabel(issue) : null);
	const regionLabel = $derived(issue ? getIssueRegionLabel(issue) : null);
	const roadTypeLabel = $derived(issue ? getIssueRoadTypeLabel(issue) : null);
	const secondaryLocationLine = $derived(issue ? getIssueSecondaryLocationLine(issue) : null);
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
	const followButtonLabel = $derived.by(() => {
		if (followMutating) {
			return isFollowing ? 'Memproses berhenti mengikuti...' : 'Memproses mengikuti...';
		}

		if (followLoading && !followerID) {
			return 'Menyiapkan...';
		}

		return isFollowing ? 'Berhenti mengikuti' : 'Ikuti laporan ini';
	});
	const followerCountLabel = $derived.by(() => {
		if (followersCount === 1) {
			return '1 orang mengikuti laporan ini';
		}

		return `${followersCount} orang mengikuti laporan ini`;
	});

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

	function getTimelineEventPresentation(event: IssueTimelineEvent) {
		switch (event.type) {
			case 'issue_created':
				return {
					icon: AddCircleIcon,
					tone: 'border-brand-100 bg-brand-50 text-brand-600'
				};
			case 'photo_added':
				return {
					icon: CameraIcon,
					tone: 'border-sky-100 bg-sky-50 text-sky-600'
				};
			case 'severity_changed':
			case 'casualty_reported':
				return {
					icon: DangerIcon,
					tone: 'border-amber-100 bg-amber-50 text-amber-700'
				};
			case 'status_updated':
				return {
					icon: CheckCircleIcon,
					tone: 'border-emerald-100 bg-emerald-50 text-emerald-600'
				};
			default:
				return {
					icon: InfoIcon,
					tone: 'border-slate-200 bg-slate-100 text-slate-600'
				};
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

	function resetFollowState() {
		followerID = null;
		followLoading = false;
		followMutating = false;
		followErrorMessage = null;
		isFollowing = false;
		followersCount = 0;
	}

	function clearRefreshMessageTimer() {
		if (refreshMessageTimer !== null) {
			clearTimeout(refreshMessageTimer);
			refreshMessageTimer = null;
		}
	}

	function showNotificationRefreshMessage(message: string) {
		clearRefreshMessageTimer();
		notificationRefreshMessage = message;

		if (!browser) return;

		refreshMessageTimer = window.setTimeout(() => {
			notificationRefreshMessage = null;
			refreshMessageTimer = null;
		}, 2400);
	}

	type FollowStateSnapshot = {
		following: boolean;
		followersCount: number;
		errorMessage: string | null;
	};

	async function loadFollowState(
		issueID: string,
		currentFollowerID: string
	): Promise<FollowStateSnapshot> {
		try {
			const statusResult = await getIssueFollowStatus(issueID, currentFollowerID);
			const statusData = statusResult.data;

			if (statusData) {
				persistFollowerAuthFromIssueState(statusData);
				return {
					following: statusData.following,
					followersCount: statusData.followers_count,
					errorMessage: null
				};
			}
		} catch {
			// handled by count fallback below
		}

		try {
			const countResult = await getIssueFollowerCount(issueID);
			return {
				following: false,
				followersCount: countResult.data?.followers_count ?? 0,
				errorMessage: 'Status mengikuti belum tersedia. Kamu masih bisa coba ikuti laporan ini.'
			};
		} catch {
			return {
				following: false,
				followersCount: 0,
				errorMessage: 'Belum bisa memuat status mengikuti saat ini.'
			};
		}
	}

	async function handleFollowToggle() {
		if (!issue || !followerID || followMutating || followLoading) return;

		followMutating = true;
		followErrorMessage = null;

		try {
			const result = isFollowing
				? await unfollowIssue(issue.id, followerID)
				: await followIssue(issue.id, followerID);

			if (result.data) {
				persistFollowerAuthFromIssueState(result.data);
				isFollowing = result.data.following;
				followersCount = result.data.followers_count;
			}
		} catch {
			followErrorMessage = isFollowing
				? 'Belum bisa berhenti mengikuti. Coba lagi.'
				: 'Belum bisa mengikuti laporan ini. Coba lagi.';
		} finally {
			followMutating = false;
		}
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

	async function refreshIssueFromNotification(issueID: string) {
		if (!browser || notificationRefreshLoading) return;
		if (issue?.id !== issueID) return;

		notificationRefreshLoading = true;
		notificationRefreshMessage = null;

		try {
			const result = await getIssue(issueID);
			if (!result.data) {
				notFound = true;
				errorMessage = null;
				issue = null;
				resetTimelineState();
				timelineIssueID = null;
				resetFollowState();
				return;
			}

			issue = result.data;
			notFound = false;
			errorMessage = null;
			failedMediaIDs = new Set();

			const timelineRefresh = loadInitialTimeline(issueID);
			const currentFollowerID = getOrCreateIssueFollowerId();
			let followRefresh: Promise<void> | null = null;

			if (currentFollowerID) {
				followerID = currentFollowerID;
				followLoading = true;
				followErrorMessage = null;
				followRefresh = (async () => {
					const snapshot = await loadFollowState(issueID, currentFollowerID);
					if (issue?.id !== issueID) return;

					isFollowing = snapshot.following;
					followersCount = snapshot.followersCount;
					followErrorMessage = snapshot.errorMessage;
					followLoading = false;
				})();
			}

			await Promise.all([timelineRefresh, followRefresh]);
			showNotificationRefreshMessage('Laporan diperbarui');
		} catch (err) {
			if (err instanceof ApiError && (err.status === 400 || err.status === 404)) {
				notFound = true;
				errorMessage = null;
				issue = null;
				resetTimelineState();
				timelineIssueID = null;
				resetFollowState();
				return;
			}

			showNotificationRefreshMessage('Belum bisa memperbarui laporan.');
		} finally {
			followLoading = false;
			notificationRefreshLoading = false;
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
		const nextIssue = data.issue;
		const nextIssueID = nextIssue?.id ?? data.id;

		issue = nextIssue;
		errorMessage = data.loadError;
		notFound = data.notFound;
		loading = false;
		previewMediaID = null;
		failedMediaIDs = new Set();
		notificationRefreshLoading = false;
		notificationRefreshMessage = null;
		clearRefreshMessageTimer();

		if (!nextIssue) {
			resetTimelineState();
			timelineIssueID = null;
			resetFollowState();
			return;
		}

		if (timelineIssueID && timelineIssueID !== nextIssueID) {
			resetTimelineState();
			timelineIssueID = null;
		}
	});

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

	$effect(() => {
		const issueID = issue?.id;
		if (!browser || !issueID) {
			resetFollowState();
			return;
		}

		const currentFollowerID = getOrCreateIssueFollowerId();
		followerID = currentFollowerID;

		if (!currentFollowerID) {
			followErrorMessage = 'Identitas anonim belum siap. Muat ulang halaman lalu coba lagi.';
			followLoading = false;
			return;
		}

		followLoading = true;
		followErrorMessage = null;

		let cancelled = false;

		void (async () => {
			const snapshot = await loadFollowState(issueID, currentFollowerID);
			if (cancelled) return;

			isFollowing = snapshot.following;
			followersCount = snapshot.followersCount;
			followErrorMessage = snapshot.errorMessage;
			followLoading = false;
		})();

		return () => {
			cancelled = true;
		};
	});

	onMount(() => {
		return onIssueDetailRefresh(({ issueID }) => {
			void refreshIssueFromNotification(issueID);
		});
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

<div class="public-stack pb-10 pt-2">
	<a
		class="inline-flex w-fit items-center gap-2 rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-bold text-slate-700 shadow-[0_10px_24px_rgba(15,23,42,0.06)] transition hover:-translate-y-0.5 hover:border-slate-300 hover:text-slate-950"
		href="/issues"
	>
		<ArrowLeftIcon class="size-[18px]" />
		Kembali ke peta
	</a>

	{#if loading}
		<LoadingState message="Memuat detail issue..." />
	{:else if notFound}
		<EmptyState
			message="Issue tidak ditemukan atau tidak tersedia untuk publik."
			ctaHref="/issues"
			ctaLabel="Kembali ke Peta"
		/>
	{:else if errorMessage}
		<div class="flex flex-col items-center gap-3">
			<ErrorState message={errorMessage} onretry={retryFetchIssue} />
			<a
				class="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-300 hover:text-slate-950"
				href="/issues"
			>
				<ArrowLeftIcon class="size-[18px]" />
				Lihat issue lain di peta
			</a>
		</div>
	{:else if issue}
		{#if isRouteNavigating || notificationRefreshLoading || notificationRefreshMessage}
			<div class="sticky top-[78px] z-20 flex flex-wrap gap-2">
				{#if isRouteNavigating}
					<span class="inline-flex items-center gap-2 rounded-full bg-slate-950 px-4 py-2 text-xs font-bold text-white shadow-[0_14px_28px_rgba(15,23,42,0.18)]">
						<HistoryIcon class="size-4" />
						Memuat halaman issue...
					</span>
				{/if}
				{#if notificationRefreshLoading || notificationRefreshMessage}
					<span class="inline-flex items-center gap-2 rounded-full border border-amber-200 bg-white px-4 py-2 text-xs font-bold text-amber-800 shadow-[0_14px_28px_rgba(15,23,42,0.12)]">
						<NotificationIcon class="size-4" />
						{notificationRefreshLoading ? 'Memperbarui laporan...' : notificationRefreshMessage}
					</span>
				{/if}
			</div>
		{/if}

		<IssueHeader
			{issue}
			{locationLabel}
			{locationContext}
			{secondaryLocationLine}
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
		/>

		<section class="jedug-card p-5 md:p-6" aria-live="polite">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div class="max-w-[54ch]">
					<span class="section-kicker">
						<UsersGroupIcon class="size-4" />
						Ikuti perkembangan
					</span>
					<h2 class="mt-4 text-2xl font-[800] tracking-[-0.04em] text-slate-950">
						Simpan issue ini di browser anonim kamu.
					</h2>
					<p class="mt-3 text-sm leading-6 text-slate-500">
						Fitur follow tetap ringan dan tidak meminta login penuh. Satu browser atau device anonim dihitung sebagai satu pengikut untuk membantu notifikasi update issue.
					</p>
				</div>
				<div class="rounded-[24px] border border-brand-100 bg-brand-50 px-4 py-4 text-left lg:min-w-[240px]">
					<span class="surface-label text-brand-500">Pengikut publik</span>
					<strong class="mt-2 block text-3xl font-[800] tracking-[-0.04em] text-brand-700">
						{followersCount}
					</strong>
					<p class="mt-2 text-xs leading-5 text-brand-700/80">{followerCountLabel}</p>
				</div>
			</div>

			<div class="mt-5 grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(260px,320px)] lg:items-start">
				<div class="rounded-[26px] border border-slate-200 bg-slate-50 px-4 py-4">
					<p class="text-sm font-bold text-slate-900">
						{#if followLoading && followersCount === 0}
							Memuat status follow...
						{:else if isFollowing}
							Issue ini sudah tersimpan di browser ini.
						{:else}
							Ikuti issue ini untuk memantau perubahan berikutnya.
						{/if}
					</p>
					<p class="mt-2 text-sm leading-6 text-slate-500">
						{#if isFollowing}
							Nanti status ini bisa menjadi dasar notifikasi browser untuk issue yang kamu anggap penting.
						{:else}
							Kamu bisa berhenti mengikuti kapan saja. Password atau identitas personal tidak dipakai untuk alur ini.
						{/if}
					</p>

					{#if followErrorMessage}
						<div class="error-panel mt-4">{followErrorMessage}</div>
					{/if}
				</div>

				<div class="flex flex-col gap-3">
					<button
						type="button"
						class={isFollowing
							? 'btn-secondary w-full border-rose-200 bg-rose-50 text-rose-700 hover:bg-rose-100'
							: 'btn-primary w-full'}
						disabled={followLoading || followMutating || !followerID}
						onclick={handleFollowToggle}
					>
						<NotificationIcon class="size-[18px]" />
						{followButtonLabel}
					</button>
					<p class="text-xs leading-5 text-slate-500">
						Status follow disimpan pada identitas anonim browser saat ini.
					</p>
				</div>
			</div>

			{#if isFollowing}
				<div class="mt-5">
					<BrowserPushCard
						title="Aktifkan Notifikasi Browser"
						lead="Kalau issue ini kamu ikuti, update terbaru juga bisa muncul di browser meski halaman JEDUG sedang tidak aktif."
					/>
				</div>
			{/if}
		</section>

		<IssueGallery
			media={visibleMedia}
			locationLabel={locationLabel}
			totalPhotoCount={issue.photo_count}
			onSelectMedia={openPreview}
			onMediaError={markMediaFailed}
		/>

		<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_340px]">
			<div class="flex flex-col gap-5">
				<section class="jedug-card p-5 md:p-6">
					<div class="flex items-start gap-3">
						<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-brand-50 text-brand-600">
							<DocumentIcon class="size-6" />
						</div>
						<div>
							<h2 class="text-xl font-[800] tracking-[-0.03em] text-slate-950">Ringkasan publik</h2>
							<p class="mt-2 text-sm leading-6 text-slate-500">
								Informasi lokasi inti yang aman dibagikan untuk pressure publik, share link, dan pemindaian cepat.
							</p>
						</div>
					</div>

					{#if publicNote}
						<div class="mt-5 rounded-[24px] border border-slate-200 bg-slate-50 px-4 py-4">
							<span class="surface-label">Catatan ringkas</span>
							<p class="mt-2 text-sm leading-6 text-slate-700">{publicNote}</p>
						</div>
					{/if}

					<dl class="mt-5 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
						<div class="rounded-[22px] border border-slate-200 bg-white px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
							<dt class="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">Nama jalan / area</dt>
							<dd class="mt-2 text-sm font-bold leading-6 text-slate-900">
								{roadOrAreaLabel || 'Belum tersedia'}
							</dd>
						</div>
						{#if roadTypeLabel}
							<div class="rounded-[22px] border border-slate-200 bg-white px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
								<dt class="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">Tipe jalan</dt>
								<dd class="mt-2 text-sm font-bold leading-6 text-slate-900">{roadTypeLabel}</dd>
							</div>
						{/if}
						<div class="rounded-[22px] border border-slate-200 bg-white px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
							<dt class="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">Wilayah</dt>
							<dd class="mt-2 text-sm font-bold leading-6 text-slate-900">
								{regionLabel || 'Belum tersedia'}
							</dd>
						</div>
					</dl>
				</section>

				<section class="jedug-card p-5 md:p-6">
					<div class="flex items-start gap-3">
						<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-slate-100 text-slate-700">
							<HistoryIcon class="size-6" />
						</div>
						<div>
							<h2 class="text-xl font-[800] tracking-[-0.03em] text-slate-950">Status & jejak waktu</h2>
							<p class="mt-2 text-sm leading-6 text-slate-500">
								Ringkasan lifecycle publik issue ini tanpa mengulang informasi lokasi.
							</p>
						</div>
					</div>

					<div class="mt-5 grid gap-3 sm:grid-cols-2">
						<article class="metric-card">
							<div class="flex items-center gap-2 text-slate-500">
								<CheckCircleIcon class="size-[18px]" />
								<span class="metric-label">Status</span>
							</div>
							<strong class="mt-3 block text-lg font-[800] text-slate-950">{statusLabel}</strong>
							<p class="mt-2 text-sm leading-6 text-slate-500">Diperbarui {relativeTime(issue.updated_at)}</p>
						</article>

						<article class="metric-card">
							<div class="flex items-center gap-2 text-slate-500">
								<InfoIcon class="size-[18px]" />
								<span class="metric-label">Verifikasi</span>
							</div>
							<strong class="mt-3 block text-lg font-[800] text-slate-950">{verificationLabel}</strong>
							<p class="mt-2 text-sm leading-6 text-slate-500">{issueSnapshot}</p>
						</article>

						<article class="metric-card">
							<div class="flex items-center gap-2 text-slate-500">
								<ClockIcon class="size-[18px]" />
								<span class="metric-label">Pertama terlihat</span>
							</div>
							<strong class="mt-3 block text-lg font-[800] text-slate-950">{formatDate(issue.first_seen_at)}</strong>
							<p class="mt-2 text-sm leading-6 text-slate-500">Mulai tercatat di titik ini</p>
						</article>

						<article class="metric-card">
							<div class="flex items-center gap-2 text-slate-500">
								<HistoryIcon class="size-[18px]" />
								<span class="metric-label">Terakhir terlihat</span>
							</div>
							<strong class="mt-3 block text-lg font-[800] text-slate-950">{relativeTimeLabel(issue.last_seen_at)}</strong>
							<p class="mt-2 text-sm leading-6 text-slate-500">{formatDate(issue.last_seen_at)}</p>
						</article>
					</div>
				</section>

				{#if issue.recent_submissions.length > 0}
					<section class="jedug-card p-5 md:p-6">
						<div class="flex items-start justify-between gap-3">
							<div class="flex items-start gap-3">
								<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-sky-50 text-sky-600">
									<CameraIcon class="size-6" />
								</div>
								<div>
									<h2 class="text-xl font-[800] tracking-[-0.03em] text-slate-950">Aktivitas laporan terbaru</h2>
									<p class="mt-2 text-sm leading-6 text-slate-500">
										Ringkasan aman dari laporan publik terbaru di titik yang sama.
									</p>
								</div>
							</div>
							<span class="badge-muted">{issue.recent_submissions.length}</span>
						</div>

						<div class="mt-5 grid gap-3">
							{#each issue.recent_submissions as submission (submission.id)}
								<article class="rounded-[24px] border border-slate-200 bg-white px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
									<div class="flex flex-wrap items-start justify-between gap-3">
										<div>
											<span
												class="inline-flex rounded-full px-3 py-1 text-xs font-bold text-white"
												style={`background:${getSeverityColor(submission.severity)}`}
											>
												{getSeverityLabel(submission.severity)}
											</span>
											<p class="mt-3 text-sm font-semibold text-slate-900">
												{relativeTimeLabel(submission.reported_at)}
											</p>
											<p class="mt-1 text-xs leading-5 text-slate-500">{formatDate(submission.reported_at)}</p>
										</div>
										{#if submission.has_casualty}
											<span class="badge-tint bg-amber-50 text-amber-700">
												<DangerIcon class="size-4" />
												{submission.casualty_count > 0
													? `${submission.casualty_count} korban`
													: 'Ada korban'}
											</span>
										{/if}
									</div>

									{#if getSubmissionPublicNote(submission)}
										<p class="mt-4 text-sm leading-6 text-slate-600">{getSubmissionPublicNote(submission)}</p>
									{/if}
								</article>
							{/each}
						</div>
					</section>
				{/if}

				<section class="jedug-card p-5 md:p-6">
					<div class="flex items-start gap-3">
						<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-slate-100 text-slate-700">
							<HistoryIcon class="size-6" />
						</div>
						<div>
							<h2 class="text-xl font-[800] tracking-[-0.03em] text-slate-950">Riwayat laporan</h2>
							<p class="mt-2 text-sm leading-6 text-slate-500">
								Jejak perkembangan issue untuk transparansi publik.
							</p>
						</div>
					</div>

					{#if timelineLoading}
						<div class="mt-5">
							<LoadingState message="Memuat riwayat laporan..." />
						</div>
					{:else if timelineError}
						<div class="mt-5 rounded-[24px] border border-rose-200 bg-rose-50 px-4 py-4">
							<p class="text-sm font-semibold text-rose-700">{timelineError}</p>
							<button
								type="button"
								class="btn-secondary mt-4"
								onclick={() => issue && loadInitialTimeline(issue.id)}
							>
								Coba lagi
							</button>
						</div>
					{:else if timelineEvents.length === 0}
						<div class="mt-5 rounded-[24px] border border-dashed border-slate-200 bg-slate-50 px-4 py-5">
							<EmptyState message="Belum ada riwayat event untuk issue ini." />
						</div>
					{:else}
						<ol class="mt-5 grid gap-5" aria-label="Riwayat laporan issue">
							{#each timelineEvents as event, index (`${event.created_at}-${event.type}-${index}`)}
								{@const presentation = getTimelineEventPresentation(event)}
								{@const EventIcon = presentation.icon}
								<li class="relative pl-14">
									{#if index < timelineEvents.length - 1}
										<span class="absolute left-5 top-11 bottom-[-22px] w-px bg-slate-200"></span>
									{/if}
									<span class={`absolute left-0 top-0 flex size-10 items-center justify-center rounded-[18px] border ${presentation.tone}`}>
										<EventIcon class="size-5" />
									</span>
									<div class="rounded-[24px] border border-slate-200 bg-white px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
										<p class="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">
											{formatDate(event.created_at)}
										</p>
										<p class="mt-2 text-sm font-bold text-slate-950">{getTimelineEventTitle(event)}</p>
										{#if getTimelineEventMeta(event)}
											<p class="mt-2 text-sm leading-6 text-slate-500">{getTimelineEventMeta(event)}</p>
										{/if}
									</div>
								</li>
							{/each}
						</ol>

						{#if timelineHasMore}
							<button
								type="button"
								class="btn-secondary mt-6 w-full sm:w-auto"
								disabled={timelineLoadingMore}
								onclick={loadMoreTimeline}
							>
								<HistoryIcon class="size-[18px]" />
								{timelineLoadingMore ? 'Memuat...' : 'Muat event lebih lama'}
							</button>
						{/if}
					{/if}
				</section>
			</div>

			<div class="flex flex-col gap-5 xl:sticky xl:top-24 xl:self-start">
				<ShareActions
					title={seo.title}
					shareText={seo.share_text}
					shareUrl={seo.canonical_url}
					externalMapUrl={externalMapUrl}
				/>

				<section class="jedug-card p-5">
					<div class="flex items-start gap-3">
						<div class="flex size-11 shrink-0 items-center justify-center rounded-[20px] bg-brand-50 text-brand-600">
							<LocationIcon class="size-6" />
						</div>
						<div>
							<h2 class="text-lg font-bold text-slate-950">Lokasi</h2>
							<p class="mt-1 text-sm leading-6 text-slate-500">
								Ringkasan posisi publik untuk share, koordinat, dan orientasi area.
							</p>
						</div>
					</div>

					<div class="mt-5 space-y-3">
						<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-4">
							<span class="surface-label">Label lokasi</span>
							<p class="mt-2 text-sm font-bold leading-6 text-slate-950">{locationLabel}</p>
							{#if regionLabel && regionLabel !== locationLabel}
								<p class="mt-1 text-sm leading-6 text-slate-500">{regionLabel}</p>
							{/if}
						</div>

						<div class="rounded-[22px] border border-slate-200 bg-white px-4 py-4">
							<span class="surface-label">Koordinat</span>
							<p class="mt-2 text-sm font-bold leading-6 text-slate-950">{coordinatesLabel}</p>
							<p class="mt-1 text-xs leading-5 text-slate-500">{locationContext}</p>
						</div>
					</div>
				</section>
			</div>
		</div>
	{/if}
</div>

{#if previewMedia}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-[120] bg-slate-950/82 p-4 backdrop-blur-sm"
		role="button"
		tabindex="0"
		aria-label="Tutup preview foto issue"
		onclick={handlePreviewOverlayClick}
		onkeydown={handlePreviewOverlayKeydown}
	>
		<div class="mx-auto flex h-full max-w-5xl items-center justify-center">
			<div class="relative w-full overflow-hidden rounded-[30px] border border-white/10 bg-slate-950 shadow-[0_30px_80px_rgba(15,23,42,0.4)]">
				<button
					type="button"
					class="absolute right-4 top-4 z-10 inline-flex size-11 items-center justify-center rounded-[18px] border border-white/10 bg-slate-900/82 text-white backdrop-blur transition hover:bg-slate-800"
					onclick={closePreview}
					aria-label="Tutup preview foto"
				>
					<CloseCircleIcon class="size-5" />
				</button>
				<img
					src={previewMedia.public_url}
					alt={`Preview foto issue jalan rusak di ${locationLabel}`}
					class="max-h-[85dvh] w-full object-contain bg-slate-950"
					onerror={() => markMediaFailed(previewMedia.id)}
				/>
			</div>
		</div>
	</div>
{/if}
