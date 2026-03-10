import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { ApiResponse, IssueDetail } from '$lib/api/types';
import { buildIssueDetailSeo, type IssueDetailSeo } from '$lib/utils/issue-detail';
import type { PageLoad } from './$types';

export const ssr = true;

type IssueDetailPageData = {
	id: string;
	issue: IssueDetail | null;
	notFound: boolean;
	loadError: string | null;
	seo: IssueDetailSeo;
};

function getCanonicalUrl(origin: string, issueID: string): string {
	return `${origin}/issues/${encodeURIComponent(issueID)}`;
}

function getFallbackOgImageUrl(origin: string): string {
	return `${origin}/og/issue-fallback.svg`;
}

function getApiIssueDetailUrl(origin: string, issueID: string): string {
	const baseUrl = PUBLIC_API_BASE_URL || origin;
	return new URL(`/api/v1/issues/${encodeURIComponent(issueID)}`, baseUrl).toString();
}

export const load: PageLoad = async ({ params, fetch, url }) => {
	const id = params.id;
	const canonicalUrl = getCanonicalUrl(url.origin, id);
	const fallbackOgImageUrl = getFallbackOgImageUrl(url.origin);

	const data: IssueDetailPageData = {
		id,
		issue: null,
		notFound: false,
		loadError: null,
		seo: buildIssueDetailSeo(null, { canonicalUrl, fallbackOgImageUrl })
	};

	let response: Response;
	try {
		response = await fetch(getApiIssueDetailUrl(url.origin, id));
	} catch {
		return {
			...data,
			loadError: 'Gagal memuat detail issue saat ini. Coba beberapa saat lagi.'
		};
	}

	if (response.status === 400 || response.status === 404) {
		return {
			...data,
			notFound: true
		};
	}

	if (!response.ok) {
		return {
			...data,
			loadError: 'Layanan issue detail sedang bermasalah. Coba lagi nanti.'
		};
	}

	try {
		const json = (await response.json()) as ApiResponse<IssueDetail>;
		if (!json.success || !json.data) {
			return {
				...data,
				loadError: json.message || 'Data issue tidak valid.'
			};
		}

		return {
			...data,
			issue: json.data,
			seo: buildIssueDetailSeo(json.data, { canonicalUrl, fallbackOgImageUrl })
		};
	} catch {
		return {
			...data,
			loadError: 'Terjadi kesalahan saat membaca data issue.'
		};
	}
};
