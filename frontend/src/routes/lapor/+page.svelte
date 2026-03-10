<script lang="ts">
	import { goto } from '$app/navigation';
	import ImagePicker from '$lib/components/ImagePicker.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import { getAnonToken, isConsentGiven } from '$lib/utils/storage';
	import { getLocation, type GeoResult } from '$lib/utils/geolocation';
	import { compressImage } from '$lib/utils/image';
	import { presignUpload, uploadFile } from '$lib/api/uploads';
	import { submitReport } from '$lib/api/reports';
	import { ApiError } from '$lib/api/client';

	// Idempotency key: one per form session, prevents double-submit
	let clientRequestId = crypto.randomUUID();

	// Form state
	let selectedFile = $state<File | null>(null);
	let severity = $state(3);
	let note = $state('');
	let hasCasualty = $state(false);
	let casualtyCount = $state(0);

	// Process state
	type Step =
		| 'idle'
		| 'getting-location'
		| 'compressing'
		| 'preparing-upload'
		| 'uploading'
		| 'submitting'
		| 'done';
	let currentStep = $state<Step>('idle');
	let geo = $state<GeoResult | null>(null);
	let geoError = $state<string | null>(null);
	let locationLoading = $state(false);
	let manualLatitude = $state('');
	let manualLongitude = $state('');
	let submitting = $state(false);
	let error = $state<string | null>(null);

	const stepLabels: Record<Step, string> = {
		idle: '',
		'getting-location': '📍 Mengambil lokasi...',
		compressing: '🗜️ Mengompres gambar...',
		'preparing-upload': '📦 Menyiapkan upload...',
		uploading: '⬆️ Mengunggah foto...',
		submitting: '📤 Mengirim laporan...',
		done: '✅ Selesai!'
	};

	const severityOptions = [
		{ value: 1, label: 'Ringan', desc: 'Lubang kecil, masih bisa dihindari' },
		{ value: 2, label: 'Sedang', desc: 'Cukup mengganggu, perlu hati-hati' },
		{ value: 3, label: 'Berat', desc: 'Bahaya, sulit dihindari' },
		{ value: 4, label: 'Parah', desc: 'Sangat berbahaya, pernah ada kejadian' },
		{ value: 5, label: 'Kritis', desc: 'Darurat, harus segera ditangani' }
	];

	// Get location on mount
	import { onMount } from 'svelte';
	onMount(async () => {
		await loadLocation();
	});

	async function loadLocation(forceFresh = false) {
		locationLoading = true;
		geoError = null;
		error = null;

		try {
			geo = await getLocation({ forceFresh });
		} catch (e) {
			geo = null;
			geoError = e instanceof Error ? e.message : 'Gagal mengambil lokasi';
		} finally {
			locationLoading = false;
		}
	}

	function handleFileChange(file: File) {
		selectedFile = file;
		error = null;
	}

	function getResolvedLocation(): GeoResult | null {
		if (geo) return geo;

		const latitude = Number.parseFloat(manualLatitude);
		const longitude = Number.parseFloat(manualLongitude);
		if (Number.isNaN(latitude) || Number.isNaN(longitude)) {
			return null;
		}

		return {
			latitude,
			longitude,
			accuracy: 0
		};
	}

	function validate(): string | null {
		if (!selectedFile) return 'Foto wajib dipilih';
		const location = getResolvedLocation();
		if (!location) return 'Lokasi belum tersedia. Aktifkan izin lokasi atau isi koordinat manual.';
		if (location.latitude < -90 || location.latitude > 90) return 'Latitude harus antara -90 sampai 90';
		if (location.longitude < -180 || location.longitude > 180) return 'Longitude harus antara -180 sampai 180';
		if (severity < 1 || severity > 5) return 'Pilih tingkat keparahan';
		if (hasCasualty && casualtyCount < 1) return 'Jumlah korban minimal 1 jika ada korban';
		if (note.length > 500) return 'Catatan maksimal 500 karakter';
		return null;
	}

	async function handleSubmit() {
		if (!getResolvedLocation() && !locationLoading) {
			currentStep = 'getting-location';
			await loadLocation(true);
			currentStep = 'idle';
		}

		const validationError = validate();
		if (validationError) {
			error = validationError;
			return;
		}

		const token = getAnonToken();
		if (!token) {
			error = 'Token tidak tersedia. Coba refresh halaman.';
			return;
		}

		if (!isConsentGiven()) {
			error = 'Kamu harus menyetujui syarat penggunaan terlebih dahulu.';
			return;
		}

		submitting = true;
		error = null;

		try {
			const location = getResolvedLocation();
			if (!location) {
				throw new Error('Lokasi belum tersedia. Aktifkan izin lokasi atau isi koordinat manual.');
			}

			// Step 1: Compress image
			currentStep = 'compressing';
			const compressed = await compressImage(selectedFile!);

			// Step 2: Presign upload
			currentStep = 'preparing-upload';
			const presignRes = await presignUpload(
				'photo.webp',
				'image/webp',
				compressed.blob.size
			);
			if (!presignRes.data) throw new Error('Gagal menyiapkan upload');

			const { object_key, upload_url, upload_method, headers } = presignRes.data;

			// Step 3: Upload file
			currentStep = 'uploading';
			try {
				await uploadFile(upload_url, compressed.blob, 'image/webp', upload_method ?? 'POST', headers ?? {});
			} catch (uploadErr) {
				if (presignRes.data.upload_mode !== 'r2') {
					throw uploadErr;
				}

				await uploadFile(`/api/v1/uploads/file/${object_key}`, compressed.blob, 'image/webp', 'POST');
			}

			// Step 4: Submit report
			currentStep = 'submitting';
			const reportRes = await submitReport({
				client_request_id: clientRequestId,
				anon_token: token,
				latitude: location.latitude,
				longitude: location.longitude,
				gps_accuracy_m: geo ? geo.accuracy : undefined,
				severity,
				note: note.trim() || undefined,
				has_casualty: hasCasualty,
				casualty_count: hasCasualty ? casualtyCount : 0,
				captured_at: new Date().toISOString(),
				media: [
					{
						object_key,
						mime_type: 'image/webp',
						size_bytes: compressed.blob.size,
						width: compressed.width,
						height: compressed.height,
						sha256: null,
						is_primary: true
					}
				]
			});

			if (!reportRes.data) throw new Error('Gagal mengirim laporan');

			currentStep = 'done';

			// Redirect to issue page
			setTimeout(() => {
				goto(`/issues/${reportRes.data!.issue_id}`);
			}, 500);
		} catch (e) {
			if (e instanceof ApiError && e.status === 429) {
				error = e.message || 'Terlalu banyak permintaan. Coba lagi nanti.';
			} else if (e instanceof ApiError && e.status === 403) {
				error = e.message || 'Akun tidak diizinkan mengirim laporan saat ini.';
			} else {
				error = e instanceof Error ? e.message : 'Terjadi kesalahan';
			}
			currentStep = 'idle';
		} finally {
			if (currentStep !== 'done') {
				submitting = false;
			}
		}
	}
</script>

<div class="report-page">
	<h1 class="page-title">Laporkan Jalan Rusak</h1>

	<!-- Location status -->
	<div class="section">
		<div class="section-label">📍 Lokasi</div>
		{#if geo}
			<div class="location-info">
				<span>{geo.latitude.toFixed(6)}, {geo.longitude.toFixed(6)}</span>
				<span class="accuracy">± {Math.round(geo.accuracy)}m</span>
			</div>
			<button
				type="button"
				class="location-retry"
				onclick={() => loadLocation(true)}
				disabled={locationLoading || submitting}
			>
				Perbarui lokasi
			</button>
		{:else if locationLoading}
			<div class="location-loading">Mengambil lokasi...</div>
		{:else}
			<div class="location-error">{geoError ?? 'Lokasi belum tersedia.'}</div>
			<div class="manual-location">
				<p class="manual-location-help">
					Jika lokasi otomatis gagal di laptop, isi koordinat manual dari Google Maps.
				</p>
				<div class="manual-location-grid">
					<input
						type="number"
						inputmode="decimal"
						step="any"
						placeholder="Latitude, mis. -6.200000"
						bind:value={manualLatitude}
						disabled={submitting}
					/>
					<input
						type="number"
						inputmode="decimal"
						step="any"
						placeholder="Longitude, mis. 106.816666"
						bind:value={manualLongitude}
						disabled={submitting}
					/>
				</div>
			</div>
			<button
				type="button"
				class="location-retry"
				onclick={() => loadLocation(true)}
				disabled={locationLoading || submitting}
			>
				Coba ambil lokasi lagi
			</button>
		{/if}
	</div>

	<!-- Image picker -->
	<div class="section">
		<div class="section-label">📷 Foto <span class="required">*</span></div>
		<ImagePicker onchange={handleFileChange} />
	</div>

	<!-- Severity -->
	<div class="section">
		<div class="section-label">⚠️ Tingkat Keparahan <span class="required">*</span></div>
		<div class="severity-options">
			{#each severityOptions as opt}
				<label class="severity-option" class:selected={severity === opt.value}>
					<input type="radio" name="severity" value={opt.value} bind:group={severity} />
					<span class="severity-value">{opt.value}</span>
					<span class="severity-detail">
						<strong>{opt.label}</strong>
						<small>{opt.desc}</small>
					</span>
				</label>
			{/each}
		</div>
	</div>

	<!-- Casualty -->
	<div class="section">
		<div class="section-label">🚑 Ada korban?</div>
		<label class="toggle-label">
			<input type="checkbox" bind:checked={hasCasualty} />
			<span>Ya, ada korban</span>
		</label>
		{#if hasCasualty}
			<div class="casualty-input">
				<label>
					Jumlah korban
					<input
						type="number"
						bind:value={casualtyCount}
						min="1"
						max="999"
						class="number-input"
					/>
				</label>
			</div>
		{/if}
	</div>

	<!-- Note -->
	<div class="section">
		<div class="section-label">📝 Catatan</div>
		<textarea
			bind:value={note}
			placeholder="Deskripsi singkat kondisi jalan (opsional)"
			rows="3"
			maxlength="500"
			class="note-input"
		></textarea>
		<div class="char-count">{note.length}/500</div>
	</div>

	<!-- Progress -->
	{#if submitting && currentStep !== 'idle'}
		<div class="progress-status">
			{stepLabels[currentStep]}
		</div>
	{/if}

	<!-- Error -->
	{#if error}
		<div class="error-msg">⚠️ {error}</div>
	{/if}

	<!-- Submit -->
	<button
		class="submit-btn"
		onclick={handleSubmit}
		disabled={submitting || !selectedFile || locationLoading}
	>
		{#if submitting}
			Mengirim...
		{:else}
			Kirim Laporan
		{/if}
	</button>
</div>

<style>
	.report-page {
		padding-top: 24px;
		padding-bottom: 32px;
	}
	.page-title {
		font-size: 20px;
		font-weight: 700;
		margin-bottom: 24px;
		color: #0F172A;
	}
	.section {
		margin-bottom: 24px;
	}
	.section-label {
		font-size: 14px;
		font-weight: 600;
		color: #64748B;
		margin-bottom: 8px;
	}
	.required {
		color: #E5484D;
	}

	/* Location */
	.location-info {
		font-size: 13px;
		color: #0F172A;
		background: #F0FDF4;
		border: 1px solid #BBF7D0;
		padding: 8px 12px;
		border-radius: 10px;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.accuracy {
		color: #16A34A;
		font-size: 12px;
	}
	.location-error {
		font-size: 13px;
		color: #DC2626;
		background: #FEF2F2;
		border: 1px solid #FECACA;
		padding: 8px 12px;
		border-radius: 10px;
	}
	.location-loading {
		font-size: 13px;
		color: #64748B;
		padding: 8px 12px;
		background: #F8FAFC;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
	}
	.location-retry {
		margin-top: 8px;
		border: 1px solid #E2E8F0;
		background: #fff;
		color: #0F172A;
		padding: 8px 12px;
		border-radius: 10px;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}
	.location-retry:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}
	.manual-location {
		margin-top: 10px;
	}
	.manual-location-help {
		font-size: 12px;
		color: #64748B;
		margin: 0 0 8px;
	}
	.manual-location-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 8px;
	}
	.manual-location-grid input {
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		padding: 10px 12px;
		font-size: 14px;
		background: #fff;
	}

	/* Severity */
	.severity-options {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.severity-option {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		cursor: pointer;
		background: #fff;
		transition: border-color 0.15s;
	}
	.severity-option.selected {
		border-color: #E5484D;
		background: #FEF2F2;
	}
	.severity-option input {
		display: none;
	}
	.severity-value {
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		background: #F1F5F9;
		font-weight: 700;
		font-size: 13px;
		flex-shrink: 0;
	}
	.severity-option.selected .severity-value {
		background: #E5484D;
		color: #fff;
	}
	.severity-detail {
		display: flex;
		flex-direction: column;
	}
	.severity-detail strong {
		font-size: 14px;
	}
	.severity-detail small {
		font-size: 12px;
		color: #94A3B8;
	}

	/* Casualty */
	.toggle-label {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		cursor: pointer;
	}
	.toggle-label input {
		width: 18px;
		height: 18px;
		accent-color: #E5484D;
	}
	.casualty-input {
		margin-top: 10px;
	}
	.casualty-input label {
		font-size: 13px;
		color: #64748B;
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.number-input {
		width: 80px;
		padding: 6px 10px;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		font-size: 14px;
	}

	/* Note */
	.note-input {
		width: 100%;
		padding: 10px 12px;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		font-size: 14px;
		font-family: inherit;
		resize: vertical;
		line-height: 1.5;
	}
	.note-input:focus {
		outline: none;
		border-color: #E5484D;
	}
	.char-count {
		text-align: right;
		font-size: 12px;
		color: #94A3B8;
		margin-top: 4px;
	}

	/* Progress */
	.progress-status {
		text-align: center;
		font-size: 14px;
		color: #64748B;
		padding: 12px;
		background: #EFF6FF;
		border: 1px solid #BFDBFE;
		border-radius: 12px;
		margin-bottom: 12px;
	}

	/* Error */
	.error-msg {
		text-align: center;
		font-size: 14px;
		color: #DC2626;
		padding: 10px 12px;
		background: #FEF2F2;
		border: 1px solid #FECACA;
		border-radius: 12px;
		margin-bottom: 12px;
	}

	/* Submit */
	.submit-btn {
		display: block;
		width: 100%;
		padding: 16px;
		font-size: 16px;
		font-weight: 700;
		color: #fff;
		background: #E5484D;
		border: none;
		border-radius: 12px;
		cursor: pointer;
		transition: opacity 0.15s, transform 0.1s;
		min-height: 48px;
	}
	.submit-btn:hover:not(:disabled) {
		opacity: 0.88;
	}
	.submit-btn:active:not(:disabled) {
		transform: scale(0.97);
	}
	.submit-btn:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}
</style>
