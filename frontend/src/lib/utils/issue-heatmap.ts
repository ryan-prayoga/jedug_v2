import type { Issue } from '$lib/api/types';

export type MapVisualMode = 'marker' | 'heatmap';

const SEVERITY_BASE_WEIGHT: Record<number, number> = {
	1: 0.18,
	2: 0.34,
	3: 0.58,
	4: 0.78,
	5: 0.92
};

function clampSeverity(level: number): number {
	if (!Number.isFinite(level)) return 1;
	return Math.min(5, Math.max(1, Math.round(level)));
}

export function getIssueHeatWeight(
	issue: Pick<Issue, 'severity_current' | 'casualty_count' | 'submission_count' | 'status'>
): number {
	const severity = clampSeverity(issue.severity_current);
	const baseWeight = SEVERITY_BASE_WEIGHT[severity] ?? SEVERITY_BASE_WEIGHT[1];
	const casualtyBonus = Math.min(Math.max(issue.casualty_count, 0), 3) * 0.06;
	const submissionBonus = Math.min(Math.max(issue.submission_count - 1, 0), 4) * 0.02;
	const statusMultiplier = issue.status === 'fixed' || issue.status === 'archived' ? 0.45 : 1;
	const weight = Math.min(1, Math.max(0.08, (baseWeight + casualtyBonus + submissionBonus) * statusMultiplier));

	return Math.round(weight * 100) / 100;
}
