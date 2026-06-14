<script lang="ts">
	import type { MediaItem } from '$lib/api/types';
	import { CameraIcon, GalleryIcon } from '$lib/icons';

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

<section class="jedug-card p-5" aria-label="Galeri issue">
	<div class="flex items-start justify-between gap-4">
		<div class="flex min-w-0 items-start gap-3">
			<div class="flex size-11 shrink-0 items-center justify-center rounded-[4px] bg-brand-tint text-brand">
				<GalleryIcon class="size-6" />
			</div>
			<div class="min-w-0">
				<h2 class="text-lg font-bold text-ink">Galeri Foto</h2>
				<p class="mt-1 text-sm leading-6 text-muted">{helperText}</p>
			</div>
		</div>
		<span class="badge-tint h-10 min-w-10 justify-center px-3">{totalPhotoCount}</span>
	</div>

	{#if media.length > 0}
		<div class="mt-5 grid grid-cols-2 gap-3 md:grid-cols-3">
			{#each media as item (item.id)}
				<button
					type="button"
					class="group overflow-hidden rounded-[4px] border border-hairline bg-sunken text-left transition hover:border-hairline-strong"
					onclick={() => onSelectMedia(item.id)}
					aria-label={`Buka foto issue di ${locationLabel}`}
				>
					<img
						src={item.public_url}
						alt={`Foto issue jalan rusak di ${locationLabel}`}
						class="h-36 w-full object-cover transition duration-300 group-hover:scale-[1.03] md:h-44"
						loading="lazy"
						decoding="async"
						onerror={() => onMediaError(item.id)}
					/>
				</button>
			{/each}
		</div>
	{:else}
		<div class="mt-5 rounded-[4px] border border-dashed border-hairline-strong bg-sunken px-4 py-8 text-center">
			<div class="mx-auto flex size-12 items-center justify-center rounded-[8px] bg-surface text-muted">
				<CameraIcon class="size-6" />
			</div>
			{#if totalPhotoCount > 0}
				<strong class="mt-4 block text-sm font-bold text-ink">Foto tidak berhasil dimuat</strong>
				<p class="mt-2 text-sm leading-6 text-muted">
					Coba buka ulang halaman atau jaringan lain untuk melihat foto publik issue ini.
				</p>
			{:else}
				<strong class="mt-4 block text-sm font-bold text-ink">Galeri masih kosong</strong>
				<p class="mt-2 text-sm leading-6 text-muted">
					Belum ada foto tambahan yang layak tampil untuk issue ini.
				</p>
			{/if}
		</div>
	{/if}
</section>
