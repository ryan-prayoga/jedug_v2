<script lang="ts">
	import type { MediaItem } from '$lib/api/types';

	let {
		media,
		locationLabel,
		totalPhotoCount,
		onSelectMedia = () => {},
		onMediaError = () => {}
	}: {
		media: MediaItem[];
		locationLabel: string;
		totalPhotoCount: number;
		onSelectMedia?: (mediaID: string) => void;
		onMediaError?: (mediaID: string) => void;
	} = $props();

	const helperText = $derived.by(() => {
		if (totalPhotoCount === 0) {
			return 'Belum ada media publik untuk issue ini.';
		}

		if (media.length === 0) {
			return 'Media publik ada, tetapi tidak berhasil dimuat di perangkat ini.';
		}

		if (totalPhotoCount > media.length) {
			return `Menampilkan ${media.length} foto terbaru dari total ${totalPhotoCount} foto publik.`;
		}

		return `${media.length} foto publik tersedia untuk issue ini.`;
	});
</script>

<section class="gallery-card" aria-label="Galeri issue">
	<div class="section-header">
		<div>
			<h2>Galeri Foto</h2>
			<p>{helperText}</p>
		</div>
		<span class="section-count">{totalPhotoCount}</span>
	</div>

	{#if media.length > 0}
		<div class="gallery-grid">
			{#each media as item (item.id)}
				<button
					type="button"
					class="gallery-item"
					onclick={() => onSelectMedia(item.id)}
					aria-label={`Buka foto issue di ${locationLabel}`}
				>
					<img
						src={item.public_url}
						alt={`Foto issue jalan rusak di ${locationLabel}`}
						loading="lazy"
						decoding="async"
						onerror={() => onMediaError(item.id)}
					/>
				</button>
			{/each}
		</div>
	{:else}
		<div class="gallery-empty">
			{#if totalPhotoCount > 0}
				<strong>Foto tidak berhasil dimuat</strong>
				<p>Coba buka ulang halaman atau jaringan lain untuk melihat foto publik issue ini.</p>
			{:else}
				<strong>Galeri masih kosong</strong>
				<p>Belum ada foto tambahan yang layak tampil untuk issue ini.</p>
			{/if}
		</div>
	{/if}
</section>

<style>
	.gallery-card {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 16px;
		padding: 16px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
	}

	.section-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
	}

	h2 {
		margin: 0;
		font-size: 18px;
		color: #0f172a;
	}

	p {
		margin-top: 4px;
		font-size: 13px;
		line-height: 1.5;
		color: #64748b;
	}

	.section-count {
		min-width: 40px;
		height: 40px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: #fff1f2;
		color: #e5484d;
		font-size: 14px;
		font-weight: 700;
	}

	.gallery-grid {
		margin-top: 16px;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px;
	}

	.gallery-item {
		display: block;
		padding: 0;
		border: 1px solid #e2e8f0;
		border-radius: 12px;
		overflow: hidden;
		background: #f8fafc;
		cursor: zoom-in;
	}

	.gallery-item img {
		display: block;
		width: 100%;
		height: 140px;
		object-fit: cover;
	}

	.gallery-empty {
		margin-top: 16px;
		padding: 18px;
		border-radius: 12px;
		border: 1px dashed #cbd5e1;
		background: #f8fafc;
		text-align: center;
	}

	.gallery-empty strong {
		display: block;
		font-size: 15px;
		color: #0f172a;
	}

	.gallery-empty p {
		margin-top: 6px;
	}

	@media (min-width: 768px) {
		.gallery-grid {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}

		.gallery-item img {
			height: 168px;
		}
	}
</style>
