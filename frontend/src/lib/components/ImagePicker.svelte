<script lang="ts">
	import { CameraIcon, GalleryIcon } from '$lib/icons';

	let { onchange }: { onchange: (file: File) => void } = $props();

	let preview = $state<string | null>(null);
	let fileInput: HTMLInputElement;

	function handleFile(e: Event) {
		const target = e.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;

		if (preview) URL.revokeObjectURL(preview);
		preview = URL.createObjectURL(file);
		onchange(file);
	}

	function openPicker() {
		fileInput.click();
	}
</script>

<div class="w-full">
	{#if preview}
		<button
			class="group relative block w-full overflow-hidden rounded-[4px] border border-hairline bg-ink text-left"
			onclick={openPicker}
			type="button"
		>
			<img src={preview} alt="Preview foto" class="max-h-[360px] w-full object-cover transition duration-300 group-hover:scale-[1.02]" />
			<div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-ink via-ink to-transparent px-4 py-4">
				<div class="flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-white/65">Foto bukti</p>
						<p class="mt-1 text-sm font-semibold text-white">Tap untuk mengganti foto</p>
					</div>
					<span class="inline-flex items-center gap-2 rounded-full border border-hairline bg-surface px-3 py-2 text-xs font-semibold text-white">
						<GalleryIcon class="size-[18px]" />
						Ganti
					</span>
				</div>
			</div>
		</button>
	{:else}
		<button
			class="group flex min-h-[220px] w-full flex-col items-center justify-center gap-4 rounded-[4px] border border-dashed border-hairline-strong bg-[linear-gradient(180deg,rgba(255,255,255,0.86)_0%,rgba(248,250,252,0.92)_100%)] px-6 py-8 text-center transition hover:border-brand-300 hover:bg-surface"
			onclick={openPicker}
			type="button"
		>
			<div class="flex size-16 items-center justify-center rounded-[4px] bg-brand-tint text-brand transition group-hover:scale-105">
				<CameraIcon class="size-8" />
			</div>
			<div class="space-y-2">
				<p class="text-base font-bold text-ink">Ambil atau pilih foto jalan rusak</p>
				<p class="mx-auto max-w-[28ch] text-sm leading-6 text-muted">
					Foto yang jelas membantu moderasi dan membuat issue lebih meyakinkan di peta publik.
				</p>
			</div>
			<span class="btn-secondary min-w-[11rem]">Pilih foto</span>
		</button>
	{/if}
	<input
		bind:this={fileInput}
		type="file"
		accept="image/jpeg,image/png,image/webp,image/heic,image/heif"
		capture="environment"
		onchange={handleFile}
		class="hidden"
	/>
</div>
