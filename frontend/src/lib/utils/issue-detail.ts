import type { IssueDetail, MediaItem } from '$lib/api/types';

export interface IssueDetailSeo {
	title: string;
	description: string;
	canonical_url: string;
	og_image_url: string;
	og_image_alt: string;
	twitter_card: 'summary_large_image';
	share_text: string;
	fallback_og_image_url: string;
}

type Tone = {
	bg: string;
	text: string;
};

const STATUS_LABELS: Record<string, string> = {
	open: 'Terbuka',
	verified: 'Terverifikasi',
	in_progress: 'Diproses',
	fixed: 'Selesai',
	archived: 'Diarsipkan',
	rejected: 'Ditolak',
	merged: 'Digabung'
};

const STATUS_TONES: Record<string, Tone> = {
	open: { bg: '#EFF6FF', text: '#2563EB' },
	verified: { bg: '#F0FDF4', text: '#16A34A' },
	in_progress: { bg: '#FEF3C7', text: '#B45309' },
	fixed: { bg: '#F1F5F9', text: '#64748B' },
	archived: { bg: '#F1F5F9', text: '#64748B' },
	rejected: { bg: '#FEF2F2', text: '#DC2626' },
	merged: { bg: '#F8FAFC', text: '#64748B' }
};

const VERIFICATION_LABELS: Record<string, string> = {
	verified: 'Terverifikasi',
	pending: 'Menunggu Verifikasi',
	unverified: 'Belum Diverifikasi',
	rejected: 'Verifikasi Ditolak'
};

const VERIFICATION_TONES: Record<string, Tone> = {
	verified: { bg: '#DCFCE7', text: '#166534' },
	pending: { bg: '#FEF3C7', text: '#92400E' },
	unverified: { bg: '#F1F5F9', text: '#475569' },
	rejected: { bg: '#FEE2E2', text: '#991B1B' }
};

const SEVERITY_LABELS = ['', 'Ringan', 'Sedang', 'Berat', 'Parah', 'Kritis'];
const SEVERITY_COLORS = ['', '#F6C453', '#F97316', '#DC2626', '#DC2626', '#991B1B'];

export function getSeverityLabel(value: number): string {
	return SEVERITY_LABELS[value] || `Level ${value}`;
}

export function getSeverityColor(value: number): string {
	return SEVERITY_COLORS[value] || '#94A3B8';
}

export function getStatusLabel(status: string): string {
	return STATUS_LABELS[status] || status;
}

export function getStatusTone(status: string): Tone {
	return STATUS_TONES[status] || STATUS_TONES.open;
}

export function getVerificationLabel(verificationStatus: string): string {
	return VERIFICATION_LABELS[verificationStatus] || verificationStatus;
}

export function getVerificationTone(verificationStatus: string): Tone {
	return VERIFICATION_TONES[verificationStatus] || VERIFICATION_TONES.unverified;
}

export function getPrimaryMedia(media: MediaItem[]): MediaItem | null {
	if (media.length === 0) return null;
	return media.find((item) => item.is_primary) || media[0];
}

export function getIssueLocationLabel(issue: IssueDetail): string {
	const parts: string[] = [];

	if (issue.road_name) {
		parts.push(issue.road_name);
	}

	if (issue.region_name && issue.region_name !== issue.road_name) {
		parts.push(issue.region_name);
	}

	if (parts.length > 0) {
		return parts.join(', ');
	}

	if (issue.road_type) {
		return issue.road_type;
	}

	return formatCoordinates(issue.latitude, issue.longitude, 5);
}

export function getIssueRegionOrCoordinates(issue: IssueDetail): string {
	if (issue.region_name) {
		return issue.region_name;
	}

	return `Koordinat ${formatCoordinates(issue.latitude, issue.longitude, 5)}`;
}

export function formatCoordinates(latitude: number, longitude: number, precision = 6): string {
	return `${latitude.toFixed(precision)}, ${longitude.toFixed(precision)}`;
}

export function getPublicIssueNote(issue: IssueDetail, maxLength = 220): string | null {
	const firstNote = issue.recent_submissions.find((item) => item.note && item.note.trim().length > 0)?.note;
	if (!firstNote) return null;

	const normalized = firstNote.replace(/\s+/g, ' ').trim();
	if (normalized.length <= maxLength) return normalized;
	return `${normalized.slice(0, maxLength - 1)}…`;
}

type BuildSeoOptions = {
	canonicalUrl: string;
	fallbackOgImageUrl: string;
};

export function buildIssueDetailSeo(issue: IssueDetail | null, options: BuildSeoOptions): IssueDetailSeo {
	if (!issue) {
		return {
			title: 'Detail Issue Jalan Rusak | JEDUG',
			description:
				'Lihat detail issue jalan rusak publik di JEDUG: lokasi, tingkat keparahan, foto, jumlah laporan, dan status terbaru.',
			canonical_url: options.canonicalUrl,
			og_image_url: options.fallbackOgImageUrl,
			og_image_alt: 'Ilustrasi laporan jalan rusak JEDUG',
			twitter_card: 'summary_large_image',
			share_text: 'Lihat detail issue jalan rusak di JEDUG',
			fallback_og_image_url: options.fallbackOgImageUrl
		};
	}

	const locationLabel = getIssueLocationLabel(issue);
	const severityLabel = getSeverityLabel(issue.severity_current);
	const statusLabel = getStatusLabel(issue.status);
	const primaryMedia = getPrimaryMedia(issue.media);

	const title =
		issue.status === 'open'
			? `Jalan Rusak ${severityLabel} di ${locationLabel} | JEDUG`
			: `Issue Jalan Rusak ${statusLabel} di ${locationLabel} | JEDUG`;

	const description =
		`Lihat detail laporan jalan rusak di ${locationLabel}: ` +
		`keparahan ${severityLabel.toLowerCase()}, ${issue.submission_count} laporan, ` +
		`${issue.photo_count} foto, dan status terbaru ${statusLabel.toLowerCase()}.`;

	return {
		title,
		description,
		canonical_url: options.canonicalUrl,
		og_image_url: primaryMedia?.public_url || options.fallbackOgImageUrl,
		og_image_alt: primaryMedia
			? `Foto issue jalan rusak di ${locationLabel}`
			: 'Ilustrasi laporan jalan rusak JEDUG',
		twitter_card: 'summary_large_image',
		share_text: `Pantau issue jalan rusak ${severityLabel.toLowerCase()} di ${locationLabel}`,
		fallback_og_image_url: options.fallbackOgImageUrl
	};
}
