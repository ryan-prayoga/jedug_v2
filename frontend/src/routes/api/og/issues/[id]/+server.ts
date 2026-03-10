import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { ApiResponse, IssueDetail, MediaItem } from '$lib/api/types';
import { ImageResponse } from '@vercel/og';
import { html } from 'satori-html';
import type { RequestHandler } from './$types';

const IMAGE_WIDTH = 1200;
const IMAGE_HEIGHT = 630;

const BRAND_RED = '#E5484D';

const SUCCESS_CACHE_CONTROL = 'public, max-age=300, s-maxage=900, stale-while-revalidate=86400';
const FALLBACK_CACHE_CONTROL = 'public, max-age=60, s-maxage=300, stale-while-revalidate=600';

const SEVERITY_LABELS = ['', 'Ringan', 'Sedang', 'Berat', 'Parah', 'Kritis'];

function getSeverityLabel(value: number): string {
	return SEVERITY_LABELS[value] || `Level ${value}`;
}

function truncate(text: string, maxLength: number): string {
	if (text.length <= maxLength) return text;
	return `${text.slice(0, maxLength - 1)}…`;
}

function escapeHtml(value: string): string {
	return value
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;');
}

function normalizeText(value: string | null | undefined, fallback = '-'): string {
	if (!value) return fallback;
	const normalized = value.replace(/\s+/g, ' ').trim();
	if (!normalized) return fallback;
	return normalized;
}

function getLocationText(issue: IssueDetail): string {
	const roadName = normalizeText(issue.road_name, '');
	const regionName = normalizeText(issue.region_name, '');

	if (roadName && regionName && roadName !== regionName) {
		return `${roadName} - ${regionName}`;
	}

	if (roadName) return roadName;
	if (regionName) return regionName;

	return `${issue.latitude.toFixed(5)}, ${issue.longitude.toFixed(5)}`;
}

function getMainMedia(issue: IssueDetail): MediaItem | null {
	if (issue.primary_media) return issue.primary_media;
	if (issue.media.length === 0) return null;
	return issue.media.find((item) => item.is_primary) || issue.media[0];
}

function formatRelativeTime(dateStr: string): string {
	const date = new Date(dateStr);
	if (Number.isNaN(date.getTime())) return 'baru saja';

	const diff = Date.now() - date.getTime();
	const minute = 60 * 1000;
	const hour = 60 * minute;
	const day = 24 * hour;

	if (diff < minute) return 'baru saja';
	if (diff < hour) return `${Math.max(1, Math.floor(diff / minute))} menit lalu`;
	if (diff < day) return `${Math.max(1, Math.floor(diff / hour))} jam lalu`;
	if (diff < 30 * day) return `${Math.max(1, Math.floor(diff / day))} hari lalu`;

	return date.toLocaleDateString('id-ID', {
		day: 'numeric',
		month: 'short',
		year: 'numeric'
	});
}

function getIssueDetailApiUrl(origin: string, issueId: string): string {
	const baseUrl = PUBLIC_API_BASE_URL || origin;
	return new URL(`/api/v1/issues/${encodeURIComponent(issueId)}`, baseUrl).toString();
}

async function getIssueDetail(fetchFn: typeof fetch, origin: string, issueId: string): Promise<IssueDetail | null> {
	const controller = new AbortController();
	const timeoutID = setTimeout(() => controller.abort(), 2500);

	try {
		const response = await fetchFn(getIssueDetailApiUrl(origin, issueId), {
			method: 'GET',
			headers: {
				accept: 'application/json'
			},
			signal: controller.signal
		});

		if (response.status === 400 || response.status === 404) {
			return null;
		}

		if (!response.ok) {
			throw new Error(`Issue API returned ${response.status}`);
		}

		const json = (await response.json()) as ApiResponse<IssueDetail>;
		if (!json.success || !json.data) {
			return null;
		}

		return json.data;
	} catch {
		return null;
	} finally {
		clearTimeout(timeoutID);
	}
}

type OgTemplateData = {
	title: string;
	location: string;
	stats: string;
	reported: string;
	backgroundImageUrl: string | null;
};

async function buildHardFallbackResponse(cacheControl: string): Promise<Response> {
	const image = new ImageResponse(
		html(
			`<div
				style="display:flex;width:${IMAGE_WIDTH}px;height:${IMAGE_HEIGHT}px;align-items:center;justify-content:center;background:${BRAND_RED};color:#FFFFFF;font-size:64px;font-weight:700;font-family:Arial,sans-serif;"
			>
				JEDUG
			</div>`
		),
		{
			width: IMAGE_WIDTH,
			height: IMAGE_HEIGHT,
			headers: {
				'Content-Type': 'image/png',
				'Cache-Control': cacheControl
			}
		}
	);

	const imageBuffer = await image.arrayBuffer();
	return new Response(imageBuffer, {
		status: 200,
		headers: {
			'Content-Type': 'image/png',
			'Cache-Control': cacheControl
		}
	});
}

async function renderImage(data: OgTemplateData, cacheControl: string): Promise<Response> {
	const backgroundImageMarkup = data.backgroundImageUrl
		? `<img
				src="${escapeHtml(data.backgroundImageUrl)}"
				alt=""
				style="position:absolute;top:0;right:0;bottom:0;left:0;width:100%;height:100%;object-fit:cover;"
			/>`
		: '';

	const baseBackground = data.backgroundImageUrl
		? 'transparent'
		: `linear-gradient(135deg, ${BRAND_RED} 0%, #B91C1C 50%, #7F1D1D 100%)`;

	const overlayBackground = data.backgroundImageUrl
		? 'linear-gradient(145deg, rgba(15, 23, 42, 0.86) 0%, rgba(15, 23, 42, 0.64) 45%, rgba(15, 23, 42, 0.75) 100%)'
		: 'linear-gradient(145deg, rgba(15, 23, 42, 0.28) 0%, rgba(15, 23, 42, 0.46) 100%)';

	const safeTitle = escapeHtml(data.title);
	const safeLocation = escapeHtml(data.location);
	const safeStats = escapeHtml(data.stats);
	const safeReported = escapeHtml(data.reported);

	const image = new ImageResponse(
		html(
			`<div
				style="position:relative;display:flex;width:${IMAGE_WIDTH}px;height:${IMAGE_HEIGHT}px;overflow:hidden;background:${baseBackground};font-family:'Segoe UI','Inter',Arial,sans-serif;color:#FFFFFF;"
			>
				${backgroundImageMarkup}
				<div style="position:absolute;top:0;right:0;bottom:0;left:0;background:${overlayBackground};"></div>
				<div
					style="position:relative;display:flex;flex-direction:column;justify-content:space-between;width:100%;height:100%;padding:60px;"
				>
					<div style="display:flex;flex-direction:column;gap:18px;">
						<div
							style="display:flex;align-self:flex-start;align-items:center;border-radius:999px;background:rgba(229, 72, 77, 0.95);padding:10px 20px;font-size:26px;font-weight:700;letter-spacing:2px;"
						>
							${safeTitle}
						</div>
						<div style="font-size:48px;font-weight:800;line-height:1.14;max-width:1000px;">
							${safeLocation}
						</div>
						<div style="font-size:36px;font-weight:600;color:rgba(248, 250, 252, 0.96);">
							${safeStats}
						</div>
						<div style="font-size:30px;font-weight:500;color:rgba(226, 232, 240, 0.95);">
							${safeReported}
						</div>
					</div>
					<div style="font-size:30px;font-weight:800;letter-spacing:0.08em;color:rgba(255, 255, 255, 0.95);">
						[jedug.id]
					</div>
				</div>
			</div>`
		),
		{
			width: IMAGE_WIDTH,
			height: IMAGE_HEIGHT,
			headers: {
				'Content-Type': 'image/png',
				'Cache-Control': cacheControl
			}
		}
	);

	const imageBuffer = await image.arrayBuffer();

	return new Response(imageBuffer, {
		status: image.status,
		statusText: image.statusText,
		headers: image.headers
	});
}

function buildIssueImage(issue: IssueDetail): Promise<Response> {
	const severityLabel = getSeverityLabel(issue.severity_current).toUpperCase();
	const title = truncate(`JALAN RUSAK ${severityLabel}`, 42);
	const location = truncate(getLocationText(issue), 58);
	const stats = `${issue.submission_count} laporan • ${issue.casualty_count} korban`;
	const reported = `Dilaporkan ${formatRelativeTime(issue.created_at)}`;
	const backgroundImageUrl = getMainMedia(issue)?.public_url || null;

	return renderImage(
		{
			title,
			location,
			stats,
			reported,
			backgroundImageUrl
		},
		SUCCESS_CACHE_CONTROL
	);
}

function buildFallbackImage(): Promise<Response> {
	return renderImage(
		{
			title: 'JALAN RUSAK',
			location: 'Laporan JEDUG',
			stats: 'Pantau kondisi jalan rusak di sekitarmu',
			reported: 'Issue tidak ditemukan atau belum tersedia',
			backgroundImageUrl: null
		},
		FALLBACK_CACHE_CONTROL
	);
}

export const GET: RequestHandler = async ({ params, fetch, url }) => {
	const issueID = normalizeText(params.id, '').trim();
	if (!issueID) {
		try {
			return await buildFallbackImage();
		} catch {
			return await buildHardFallbackResponse(FALLBACK_CACHE_CONTROL);
		}
	}

	const issue = await getIssueDetail(fetch, url.origin, issueID);
	if (!issue) {
		try {
			return await buildFallbackImage();
		} catch {
			return await buildHardFallbackResponse(FALLBACK_CACHE_CONTROL);
		}
	}

	try {
		return await buildIssueImage(issue);
	} catch {
		try {
			return await buildFallbackImage();
		} catch {
			return await buildHardFallbackResponse(FALLBACK_CACHE_CONTROL);
		}
	}
};
