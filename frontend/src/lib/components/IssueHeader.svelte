<script lang="ts">
	import type { IssueDetail, MediaItem } from '$lib/api/types';
	import { formatDate, relativeTime } from '$lib/utils/date';

	type Tone = {
		bg: string;
		text: string;
	};

	let {
		issue,
		locationLabel,
		locationContext,
		coordinatesLabel,
		severityLabel,
		severityColor,
		statusLabel,
		statusTone,
		verificationLabel,
		verificationTone,
		heroMedia,
		onHeroSelect = () => {},
		onHeroError = () => {}
	}: {
		issue: IssueDetail;
		locationLabel: string;
		locationContext: string;
		coordinatesLabel: string;
		severityLabel: string;
		severityColor: string;
		statusLabel: string;
		statusTone: Tone;
		verificationLabel: string;
		verificationTone: Tone;
		heroMedia: MediaItem | null;
		onHeroSelect?: () => void;
		onHeroError?: () => void;
	} = $props();
</script>

<section class="issue-header">
	<div class="hero-shell">
		{#if heroMedia}
			<button
				type="button"
				class="hero-media"
				onclick={onHeroSelect}
				aria-label={`Buka foto utama issue di ${locationLabel}`}
			>
				<img
					src={heroMedia.public_url}
					alt={`Foto issue jalan rusak di ${locationLabel}`}
					loading="eager"
					decoding="async"
					onerror={onHeroError}
				/>
				<div class="hero-overlay"></div>
				<div class="hero-caption">
					<span>{locationContext}</span>
					<strong>{locationLabel}</strong>
				</div>
			</button>
		{:else}
			<div class="hero-placeholder">
				<span class="placeholder-kicker">Belum ada foto utama</span>
				<strong>{locationLabel}</strong>
				<p>Laporan ini tetap tampil publik agar kondisi jalan bisa terus dipantau.</p>
			</div>
		{/if}
	</div>

	<div class="summary-card">
		<div class="badge-row">
			<span class="badge severity" style={`background: ${severityColor}`}>{severityLabel}</span>
			<span class="badge" style={`background: ${statusTone.bg}; color: ${statusTone.text}`}>
				{statusLabel}
			</span>
			<span class="badge" style={`background: ${verificationTone.bg}; color: ${verificationTone.text}`}>
				{verificationLabel}
			</span>
		</div>

		<h1>{locationLabel}</h1>
		<p class="lede">
			{#if issue.road_type}
				{issue.road_type}
			{:else}
				Titik laporan publik
			{/if}
			<span aria-hidden="true">·</span>
			{coordinatesLabel}
		</p>

		<div class="meta-grid">
			<article class="meta-item">
				<span class="meta-label">Lokasi</span>
				<strong>{locationContext}</strong>
			</article>
			<article class="meta-item">
				<span class="meta-label">Pertama terlihat</span>
				<strong>{formatDate(issue.first_seen_at)}</strong>
			</article>
			<article class="meta-item">
				<span class="meta-label">Terakhir terlihat</span>
				<strong>{relativeTime(issue.last_seen_at)}</strong>
				<small>{formatDate(issue.last_seen_at)}</small>
			</article>
		</div>
	</div>
</section>

<style>
	.issue-header {
		display: grid;
		gap: 16px;
	}

	.hero-shell {
		border-radius: 18px;
		overflow: hidden;
		border: 1px solid #e2e8f0;
		background: #0f172a;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
	}

	.hero-media {
		position: relative;
		display: block;
		width: 100%;
		border: 0;
		padding: 0;
		background: transparent;
		cursor: zoom-in;
	}

	.hero-media img {
		display: block;
		width: 100%;
		height: min(60vw, 360px);
		object-fit: cover;
	}

	.hero-overlay {
		position: absolute;
		inset: 0;
		background: linear-gradient(180deg, rgba(15, 23, 42, 0.04) 0%, rgba(15, 23, 42, 0.72) 100%);
	}

	.hero-caption {
		position: absolute;
		left: 16px;
		right: 16px;
		bottom: 16px;
		display: grid;
		gap: 2px;
		color: #fff;
		text-align: left;
	}

	.hero-caption span {
		font-size: 12px;
		font-weight: 600;
		color: rgba(255, 255, 255, 0.78);
	}

	.hero-caption strong {
		font-size: 20px;
		line-height: 1.2;
	}

	.hero-placeholder {
		display: grid;
		place-content: center;
		gap: 8px;
		min-height: 240px;
		padding: 24px;
		background:
			radial-gradient(circle at top right, rgba(229, 72, 77, 0.18), transparent 40%),
			linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
		color: #fff;
		text-align: center;
	}

	.placeholder-kicker {
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: #fca5a5;
	}

	.hero-placeholder strong {
		font-size: 22px;
		line-height: 1.2;
	}

	.hero-placeholder p {
		max-width: 34ch;
		margin: 0 auto;
		color: #cbd5e1;
	}

	.summary-card {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 16px;
		padding: 16px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
	}

	.badge-row {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 12px;
	}

	.badge {
		display: inline-flex;
		align-items: center;
		padding: 6px 10px;
		border-radius: 999px;
		font-size: 11px;
		font-weight: 700;
		line-height: 1;
	}

	.severity {
		color: #fff;
	}

	h1 {
		margin: 0;
		font-size: 24px;
		line-height: 1.15;
		color: #0f172a;
	}

	.lede {
		margin-top: 8px;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		color: #64748b;
		font-size: 14px;
	}

	.meta-grid {
		display: grid;
		grid-template-columns: repeat(1, minmax(0, 1fr));
		gap: 10px;
		margin-top: 16px;
	}

	.meta-item {
		padding: 12px;
		border-radius: 12px;
		background: #f8fafc;
		border: 1px solid #e2e8f0;
	}

	.meta-label {
		display: block;
		margin-bottom: 4px;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		color: #64748b;
	}

	.meta-item strong {
		display: block;
		font-size: 14px;
		line-height: 1.4;
		color: #0f172a;
	}

	.meta-item small {
		display: block;
		margin-top: 4px;
		font-size: 12px;
		color: #64748b;
	}

	@media (min-width: 768px) {
		.issue-header {
			grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.8fr);
			align-items: stretch;
		}

		.hero-media img,
		.hero-placeholder {
			min-height: 100%;
			height: 100%;
		}

		.hero-caption strong {
			font-size: 24px;
		}

		.summary-card {
			padding: 20px;
		}

		.meta-grid {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
	}
</style>
