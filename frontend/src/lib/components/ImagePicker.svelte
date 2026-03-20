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
			class="group relative block w-full overflow-hidden rounded-[28px] border border-slate-200 bg-slate-950 text-left shadow-[0_18px_38px_rgba(15,23,42,0.12)]"
			onclick={openPicker}
			type="button"
		>
			<img src={preview} alt="Preview foto" class="max-h-[360px] w-full object-cover transition duration-300 group-hover:scale-[1.02]" />
			<div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-slate-950/80 via-slate-950/20 to-transparent px-4 py-4">
				<div class="flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-white/65">Foto bukti</p>
						<p class="mt-1 text-sm font-semibold text-white">Tap untuk mengganti foto</p>
					</div>
					<span class="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/12 px-3 py-2 text-xs font-semibold text-white backdrop-blur">
						<GalleryIcon class="size-[18px]" />
						Ganti
					</span>
				</div>
			</div>
		</button>
	{:else}
		<button
			class="group flex min-h-[220px] w-full flex-col items-center justify-center gap-4 rounded-[28px] border border-dashed border-slate-300 bg-[linear-gradient(180deg,rgba(255,255,255,0.86)_0%,rgba(248,250,252,0.92)_100%)] px-6 py-8 text-center shadow-[inset_0_1px_0_rgba(255,255,255,0.8)] transition hover:-translate-y-0.5 hover:border-brand-300 hover:bg-white"
			onclick={openPicker}
			type="button"
		>
			<div class="flex size-16 items-center justify-center rounded-[24px] bg-brand-50 text-brand-600 transition group-hover:scale-105">
				<CameraIcon class="size-8" />
			</div>
			<div class="space-y-2">
				<p class="text-base font-bold text-slate-900">Ambil atau pilih foto jalan rusak</p>
				<p class="mx-auto max-w-[28ch] text-sm leading-6 text-slate-500">
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
