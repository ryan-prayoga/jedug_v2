<script lang="ts">
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

<div class="image-picker">
	{#if preview}
		<button class="preview-container" onclick={openPicker} type="button">
			<img src={preview} alt="Preview foto" class="preview-img" />
			<span class="change-label">Ganti Foto</span>
		</button>
	{:else}
		<button class="picker-placeholder" onclick={openPicker} type="button">
			<span class="picker-icon">📷</span>
			<span>Ambil / Pilih Foto</span>
		</button>
	{/if}
	<input
		bind:this={fileInput}
		type="file"
		accept="image/jpeg,image/png,image/webp,image/heic,image/heif"
		capture="environment"
		onchange={handleFile}
		class="hidden-input"
	/>
</div>

<style>
	.image-picker {
		width: 100%;
	}
	.hidden-input {
		display: none;
	}
	.picker-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		width: 100%;
		min-height: 180px;
		background: #F8FAFC;
		border: 2px dashed #E2E8F0;
		border-radius: 12px;
		cursor: pointer;
		color: #64748B;
		font-size: 14px;
		gap: 8px;
		transition: border-color 0.15s, background 0.15s;
	}
	.picker-placeholder:hover {
		border-color: #CBD5E1;
		background: #F1F5F9;
	}
	.picker-icon {
		font-size: 36px;
	}
	.preview-container {
		position: relative;
		display: block;
		width: 100%;
		border: none;
		padding: 0;
		background: none;
		cursor: pointer;
		border-radius: 12px;
		overflow: hidden;
	}
	.preview-img {
		width: 100%;
		max-height: 300px;
		object-fit: cover;
		border-radius: 12px;
	}
	.change-label {
		position: absolute;
		bottom: 8px;
		right: 8px;
		background: rgba(0, 0, 0, 0.55);
		color: #fff;
		font-size: 12px;
		font-weight: 500;
		padding: 4px 12px;
		border-radius: 8px;
	}
</style>
