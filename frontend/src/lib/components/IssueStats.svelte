<script lang="ts">
	import { formatDate, relativeTime } from '$lib/utils/date';

	const numberFormatter = new Intl.NumberFormat('id-ID');

	let {
		submissionCount,
		photoCount,
		casualtyCount,
		reactionCount,
		updatedAt
	}: {
		submissionCount: number;
		photoCount: number;
		casualtyCount: number;
		reactionCount: number;
		updatedAt: string;
	} = $props();
</script>

<section class="issue-stats" aria-label="Statistik ringkas issue">
	<article class="stat-card">
		<span class="label">Laporan</span>
		<strong>{numberFormatter.format(submissionCount)}</strong>
		<small>Total laporan terkumpul</small>
	</article>
	<article class="stat-card">
		<span class="label">Foto</span>
		<strong>{numberFormatter.format(photoCount)}</strong>
		<small>Media publik tersedia</small>
	</article>
	<article class="stat-card" class:alert={casualtyCount > 0}>
		<span class="label">Korban</span>
		<strong>{numberFormatter.format(casualtyCount)}</strong>
		<small>{casualtyCount > 0 ? 'Laporan korban tercatat' : 'Belum ada korban tercatat'}</small>
	</article>
	<article class="stat-card">
		<span class="label">Reaksi</span>
		<strong>{numberFormatter.format(reactionCount)}</strong>
		<small>{reactionCount > 0 ? 'Dukungan publik tercatat' : 'Belum ada reaksi publik'}</small>
	</article>
	<article class="stat-card">
		<span class="label">Diperbarui</span>
		<strong>{relativeTime(updatedAt)}</strong>
		<small>{formatDate(updatedAt)}</small>
	</article>
</section>

<style>
	.issue-stats {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}

	.stat-card {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 16px;
		padding: 14px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
	}

	.stat-card.alert {
		border-color: #fecaca;
		background: #fff7f7;
	}

	.label {
		display: block;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		color: #64748b;
	}

	strong {
		display: block;
		margin-top: 8px;
		font-size: 18px;
		line-height: 1.2;
		color: #0f172a;
	}

	small {
		display: block;
		margin-top: 6px;
		font-size: 12px;
		line-height: 1.4;
		color: #64748b;
	}

	@media (min-width: 768px) {
		.issue-stats {
			grid-template-columns: repeat(5, minmax(0, 1fr));
		}
	}
</style>
