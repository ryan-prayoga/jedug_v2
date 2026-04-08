<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount, tick } from 'svelte';
	import {
		CameraIcon,
		DangerIcon,
		DocumentIcon,
		InfoIcon,
		LocationIcon,
		RefreshIcon
	} from '$lib/icons';
	import ImagePicker from '$lib/components/ImagePicker.svelte';
	import { getOrCreateIssueFollowerId, isConsentGiven } from '$lib/utils/storage';
	import { ensureDeviceBootstrap, isBootstrapMissingError } from '$lib/utils/device-init';
	import { getLocation, type GeoResult } from '$lib/utils/geolocation';
	import { compressImage } from '$lib/utils/image';
	import { resolveLocationLabel } from '$lib/api/location';
	import type { LocationLabelData } from '$lib/api/types';
	import { presignUpload, uploadFile } from '$lib/api/uploads';
	import { submitReport } from '$lib/api/reports';
	import { ApiError } from '$lib/api/client';

	let clientRequestId = crypto.randomUUID();

	let selectedFile = $state<File | null>(null);
	let severity = $state(3);
	let note = $state('');
	let hasCasualty = $state(false);
	let casualtyCount = $state(0);

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
	let locationLabel = $state<LocationLabelData | null>(null);
	let locationLabelLoading = $state(false);
	let locationLabelError = $state<string | null>(null);
	let manualLatitude = $state('');
	let manualLongitude = $state('');
	let submitting = $state(false);
	let error = $state<string | null>(null);
	const locationLabelCache = new Map<string, LocationLabelData>();
	let activeLocationLabelRequestId = 0;
	let bootstrapInitializing = $state(true);
	let errorRef = $state<HTMLElement | null>(null);

	const stepLabels: Record<Step, string> = {
		idle: '',
		'getting-location': 'Mengambil lokasi...',
		compressing: 'Mengompres gambar...',
		'preparing-upload': 'Menyiapkan upload...',
		uploading: 'Mengunggah foto...',
		submitting: 'Mengirim laporan...',
		done: 'Selesai!'
	};

	const severityOptions = [
		{ value: 1, label: 'Ringan', desc: 'Lubang kecil, masih bisa dihindari' },
		{ value: 2, label: 'Sedang', desc: 'Cukup mengganggu, perlu hati-hati' },
		{ value: 3, label: 'Berat', desc: 'Bahaya, sulit dihindari' },
		{ value: 4, label: 'Parah', desc: 'Sangat berbahaya, pernah ada kejadian' },
		{ value: 5, label: 'Kritis', desc: 'Darurat, harus segera ditangani' }
	];

	onMount(async () => {
		try {
			await ensureDeviceBootstrap({ retry: 1 });
		} finally {
			bootstrapInitializing = false;
		}

		await loadLocation();
	});

	async function loadLocation(forceFresh = false) {
		locationLoading = true;
		geoError = null;
		error = null;

		try {
			geo = await getLocation({ forceFresh });
			void resolveLocationLabelForPoint(geo, { forceRefresh: forceFresh });
		} catch (e) {
			geo = null;
			geoError = e instanceof Error ? e.message : 'Gagal mengambil lokasi';
			locationLabel = null;
			locationLabelError = null;
		} finally {
			locationLoading = false;
		}
	}

	function getPointKey(latitude: number, longitude: number): string {
		return `${latitude.toFixed(5)},${longitude.toFixed(5)}`;
	}

	function parseManualLocation(): GeoResult | null {
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

	async function resolveLocationLabelForPoint(
		point: GeoResult,
		options: { forceRefresh?: boolean } = {}
	) {
		const { forceRefresh = false } = options;
		const key = getPointKey(point.latitude, point.longitude);

		if (!forceRefresh) {
			const cached = locationLabelCache.get(key);
			if (cached) {
				locationLabel = cached;
				locationLabelError = null;
				locationLabelLoading = false;
				return;
			}
		}

		const requestId = ++activeLocationLabelRequestId;
		locationLabelLoading = true;
		locationLabelError = null;
		locationLabel = null;

		try {
			const res = await resolveLocationLabel(point.latitude, point.longitude);
			if (requestId !== activeLocationLabelRequestId) return;

			const data = res.data ?? null;
			if (data) {
				locationLabelCache.set(key, data);
			}
			locationLabel = data;
		} catch {
			if (requestId !== activeLocationLabelRequestId) return;
			locationLabel = null;
			locationLabelError = 'Nama wilayah belum tersedia. Koordinat tetap dipakai.';
		} finally {
			if (requestId === activeLocationLabelRequestId) {
				locationLabelLoading = false;
			}
		}
	}

	function handleFileChange(file: File) {
		selectedFile = file;
		error = null;
	}

	function locationPrimaryLabel(label: LocationLabelData): string {
		const parts = label.label?.split(',').map((part) => part.trim()).filter(Boolean) ?? [];
		if (parts.length === 0) {
			return label.region_name || 'Lokasi terdeteksi';
		}
		return parts[0];
	}

	function locationSecondaryLabel(label: LocationLabelData): string {
		const parts = label.label?.split(',').map((part) => part.trim()).filter(Boolean) ?? [];
		if (parts.length > 1) {
			return parts.slice(1).join(' - ');
		}

		const fromParents = [label.parent_name, label.grandparent_name].filter(
			(value): value is string => Boolean(value && value.trim() !== '')
		);
		return fromParents.join(' - ');
	}

	async function applyManualLocation() {
		const manual = parseManualLocation();
		if (!manual) {
			geoError = 'Koordinat manual belum valid. Pastikan latitude dan longitude terisi benar.';
			return;
		}

		geo = manual;
		geoError = null;
		error = null;
		await resolveLocationLabelForPoint(manual);
	}

	function getResolvedLocation(): GeoResult | null {
		if (geo) return geo;
		return parseManualLocation();
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

	function mapSubmitError(e: unknown): string {
		if (e instanceof TypeError) {
			return 'Koneksi sedang bermasalah. Periksa internet lalu coba lagi.';
		}
		if (e instanceof ApiError) {
			if (e.status === 429) {
				return e.message || 'Terlalu banyak laporan dikirim. Tunggu beberapa menit lalu coba lagi.';
			}
			if (e.status === 409) {
				return e.message || 'Laporan ini sudah pernah diproses. Muat ulang halaman sebelum membuat laporan baru.';
			}
			if (e.status === 403) {
				return 'Akun tidak diizinkan mengirim laporan saat ini.';
			}
			if (e.status === 401) {
				return 'Inisialisasi pelaporan belum selesai. Muat ulang halaman lalu coba lagi.';
			}
			if (e.status === 400) {
				return 'Data laporan belum valid. Periksa kembali isian form lalu coba lagi.';
			}
			if (e.status >= 500) {
				return 'Laporan belum bisa dikirim saat ini. Coba beberapa saat lagi.';
			}
			return 'Terjadi kesalahan saat mengirim laporan. Coba lagi.';
		}
		if (e instanceof Error) {
			if (e.message === 'Canvas context not available') {
				return 'Gagal memproses foto. Coba pilih ulang foto lalu kirim lagi.';
			}
			if (e.message === 'bootstrap token missing') {
				return 'Inisialisasi pelaporan belum selesai. Muat ulang halaman lalu coba lagi.';
			}
			return e.message;
		}
		return 'Terjadi kesalahan. Coba lagi.';
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

		let token: string;
		try {
			token = await ensureDeviceBootstrap({ retry: 1 });
		} catch {
			error = 'Perangkat belum siap untuk mengirim laporan. Coba muat ulang halaman.';
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

			currentStep = 'compressing';
			const compressed = await compressImage(selectedFile!);

			currentStep = 'preparing-upload';
			const presignRes = await presignUpload(token, 'photo.webp', 'image/webp', compressed.blob.size);
			if (!presignRes.data) throw new Error('Gagal menyiapkan upload');

			const { object_key, upload_url, upload_method, headers, upload_token } = presignRes.data;

			currentStep = 'uploading';
			const uploadHeaders =
				presignRes.data.upload_mode === 'r2'
					? (headers ?? {})
					: { ...(headers ?? {}), 'X-Upload-Token': upload_token };
			try {
				await uploadFile(
					upload_url,
					compressed.blob,
					'image/webp',
					upload_method ?? 'POST',
					uploadHeaders
				);
			} catch (uploadErr) {
				if (presignRes.data.upload_mode !== 'r2') {
					throw uploadErr;
				}

				await uploadFile(`/api/v1/uploads/file/${object_key}`, compressed.blob, 'image/webp', 'POST', {
					'X-Upload-Token': upload_token
				});
			}

			currentStep = 'submitting';
			const payload = {
				client_request_id: clientRequestId,
				anon_token: token,
				actor_follower_id: getOrCreateIssueFollowerId() ?? undefined,
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
						upload_token,
						width: compressed.width,
						height: compressed.height,
						sha256: null,
						is_primary: true
					}
				]
			};

			let reportRes;
			try {
				reportRes = await submitReport(payload);
			} catch (submitErr) {
				if (!isBootstrapMissingError(submitErr)) {
					throw submitErr;
				}
				token = await ensureDeviceBootstrap({ forceRefresh: true, retry: 1 });
				reportRes = await submitReport({
					...payload,
					anon_token: token
				});
			}

			if (!reportRes.data) throw new Error('Gagal mengirim laporan');

			currentStep = 'done';

			setTimeout(() => {
				goto(`/issues/${reportRes.data!.issue_id}`);
			}, 500);
		} catch (e) {
			error = mapSubmitError(e);
			currentStep = 'idle';
			await tick();
			errorRef?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
		} finally {
			if (currentStep !== 'done') {
				submitting = false;
			}
		}
	}
</script>

<svelte:head>
	<title>Laporkan Jalan Rusak | JEDUG</title>
	<meta
		name="description"
		content="Kirim laporan jalan rusak secara anonim lewat JEDUG dengan lokasi akurat, foto bukti, dan tingkat keparahan yang jelas."
	/>
</svelte:head>

<div class="public-stack">
	<section class="jedug-card overflow-hidden">
		<div class="grid gap-5 px-5 py-6 sm:px-6">
			<div class="space-y-4">
				<span class="section-kicker">
					<DocumentIcon class="size-4" />
					Form pelaporan cepat
				</span>
				<div class="space-y-2">
					<h1 class="section-title text-balance">
						Laporkan jalan rusak dengan bukti yang rapi dan mudah diproses.
					</h1>
					<p class="section-copy">
						Fokusnya tetap sederhana: pastikan lokasi akurat, unggah foto yang jelas, lalu pilih tingkat keparahan yang paling sesuai.
					</p>
				</div>
			</div>
			<div class="grid gap-3 sm:grid-cols-3">
				<div class="jedug-panel p-4">
					<p class="surface-label">1. Lokasi</p>
					<p class="mt-2 text-sm font-semibold text-slate-900">Pastikan titik laporan benar</p>
					<p class="mt-1 text-xs leading-5 text-slate-500">Koordinat tetap jadi acuan utama, label wilayah hanya konfirmasi UX.</p>
				</div>
				<div class="jedug-panel p-4">
					<p class="surface-label">2. Foto</p>
					<p class="mt-2 text-sm font-semibold text-slate-900">Gunakan foto yang mudah dibaca</p>
					<p class="mt-1 text-xs leading-5 text-slate-500">Foto yang fokus membantu moderasi dan membuat issue lebih kredibel.</p>
				</div>
				<div class="jedug-panel p-4">
					<p class="surface-label">3. Severity</p>
					<p class="mt-2 text-sm font-semibold text-slate-900">Pilih level yang jujur</p>
					<p class="mt-1 text-xs leading-5 text-slate-500">Gunakan deskripsi yang paling mendekati kondisi nyata di lapangan.</p>
				</div>
			</div>
		</div>
	</section>

	<section class="jedug-card p-5">
		<div class="flex items-start gap-3">
			<div class="flex size-11 shrink-0 items-center justify-center rounded-[18px] bg-brand-50 text-brand-600">
				<LocationIcon class="size-6" />
			</div>
			<div class="min-w-0">
				<h2 class="text-lg font-bold text-slate-950">Lokasi laporan</h2>
				<p class="mt-1 text-sm leading-6 text-slate-500">
					Koordinat yang akurat membantu issue masuk ke titik yang benar di peta.
				</p>
			</div>
		</div>

		<div class="mt-5 space-y-3">
			{#if geo}
				<div class="rounded-[24px] border border-emerald-200 bg-emerald-50/80 px-4 py-4">
					<div class="flex flex-wrap items-center justify-between gap-3">
						<div>
							<p class="text-[11px] font-bold uppercase tracking-[0.18em] text-emerald-700">Koordinat aktif</p>
							<p class="mt-2 text-base font-bold text-slate-950">
								{geo.latitude.toFixed(6)}, {geo.longitude.toFixed(6)}
							</p>
						</div>
						<span class="badge-muted border-emerald-200 bg-white text-emerald-700">
							{geo.accuracy > 0 ? `± ${Math.round(geo.accuracy)}m` : 'Koordinat manual'}
						</span>
					</div>
				</div>

				{#if locationLabelLoading}
					<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-500">
						Mencari nama wilayah...
					</div>
				{:else if locationLabel?.label}
					<div class="rounded-[22px] border border-blue-200 bg-blue-50/70 px-4 py-4">
						<p class="text-[11px] font-bold uppercase tracking-[0.18em] text-blue-700">Konfirmasi wilayah</p>
						<p class="mt-2 text-sm font-bold text-slate-950">{locationPrimaryLabel(locationLabel)}</p>
						{#if locationSecondaryLabel(locationLabel)}
							<p class="mt-1 text-xs leading-5 text-blue-700">{locationSecondaryLabel(locationLabel)}</p>
						{/if}
					</div>
					<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-500">
						Nama jalan akan dilengkapi otomatis saat laporan dikirim.
					</div>
				{:else if locationLabelError}
					<div class="notice-panel">{locationLabelError}</div>
				{:else}
					<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-500">
						Nama wilayah belum tersedia, koordinat tetap jadi acuan laporan.
					</div>
				{/if}

				<button type="button" class="btn-secondary" onclick={() => loadLocation(true)} disabled={locationLoading || submitting}>
					<RefreshIcon class="size-[18px]" />
					Perbarui lokasi
				</button>
			{:else if locationLoading}
				<div class="rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-4 text-sm text-slate-500">
					Mengambil lokasi...
				</div>
			{:else}
				<div class="error-panel">{geoError ?? 'Lokasi belum tersedia.'}</div>
				<div class="jedug-panel space-y-3 p-4">
					<div class="flex items-start gap-3">
						<div class="flex size-10 shrink-0 items-center justify-center rounded-2xl bg-white text-brand-600 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
							<InfoIcon class="size-5" />
						</div>
						<p class="text-sm leading-6 text-slate-600">
							Jika lokasi otomatis gagal di laptop, isi koordinat manual dari Google Maps.
						</p>
					</div>
					<div class="grid gap-3 sm:grid-cols-2">
						<input
							class="input-field"
							type="number"
							inputmode="decimal"
							step="any"
							placeholder="Latitude, mis. -6.200000"
							bind:value={manualLatitude}
							disabled={submitting}
						/>
						<input
							class="input-field"
							type="number"
							inputmode="decimal"
							step="any"
							placeholder="Longitude, mis. 106.816666"
							bind:value={manualLongitude}
							disabled={submitting}
						/>
					</div>
					<div class="flex flex-col gap-2 sm:flex-row">
						<button type="button" class="btn-primary flex-1" onclick={applyManualLocation} disabled={submitting || locationLoading}>
							<LocationIcon class="size-[18px]" />
							Gunakan koordinat ini
						</button>
						<button type="button" class="btn-secondary flex-1" onclick={() => loadLocation(true)} disabled={locationLoading || submitting}>
							<RefreshIcon class="size-[18px]" />
							Coba ambil lokasi lagi
						</button>
					</div>
				</div>
			{/if}
		</div>
	</section>

	<section class="jedug-card p-5">
		<div class="mb-4 flex items-start gap-3">
			<div class="flex size-11 shrink-0 items-center justify-center rounded-[18px] bg-brand-50 text-brand-600">
				<CameraIcon class="size-6" />
			</div>
			<div>
				<h2 class="text-lg font-bold text-slate-950">Foto bukti</h2>
				<p class="mt-1 text-sm leading-6 text-slate-500">
					Unggah minimal satu foto yang memperlihatkan kondisi kerusakan dengan jelas.
				</p>
			</div>
		</div>
		<ImagePicker onchange={handleFileChange} />
	</section>

	<section class="jedug-card p-5">
		<div class="mb-4 flex items-start gap-3">
			<div class="flex size-11 shrink-0 items-center justify-center rounded-[18px] bg-brand-50 text-brand-600">
				<DangerIcon class="size-6" />
			</div>
			<div>
				<h2 class="text-lg font-bold text-slate-950">Tingkat keparahan</h2>
				<p class="mt-1 text-sm leading-6 text-slate-500">
					Pilih level yang paling menggambarkan dampak kerusakan di lapangan.
				</p>
			</div>
		</div>

		<div class="space-y-2">
			{#each severityOptions as opt}
				<label
					class={`flex cursor-pointer items-center gap-4 rounded-[24px] border px-4 py-4 transition hover:border-slate-300 ${severity === opt.value ? 'border-brand-200 bg-brand-50/70' : 'border-slate-200 bg-white'}`}
				>
					<input class="hidden" type="radio" name="severity" value={opt.value} bind:group={severity} />
					<span
						class:bg-brand-500={severity === opt.value}
						class:text-white={severity === opt.value}
						class="flex size-10 shrink-0 items-center justify-center rounded-2xl bg-slate-100 text-sm font-bold text-slate-700"
					>
						{opt.value}
					</span>
					<span class="min-w-0">
						<strong class="block text-sm font-bold text-slate-900">{opt.label}</strong>
						<small class="mt-1 block text-xs leading-5 text-slate-500">{opt.desc}</small>
					</span>
				</label>
			{/each}
		</div>
	</section>

	<section class="grid gap-4 md:grid-cols-2">
		<div class="jedug-card p-5">
			<div class="mb-4 flex items-start gap-3">
				<div class="flex size-11 shrink-0 items-center justify-center rounded-[18px] bg-brand-50 text-brand-600">
					<DangerIcon class="size-6" />
				</div>
				<div>
					<h2 class="text-lg font-bold text-slate-950">Informasi korban</h2>
					<p class="mt-1 text-sm leading-6 text-slate-500">
						Isi hanya jika memang ada laporan korban dari kejadian terkait titik ini.
					</p>
				</div>
			</div>

			<label class="flex items-center gap-3 rounded-[22px] border border-slate-200 bg-slate-50 px-4 py-3">
				<input type="checkbox" class="h-4 w-4 accent-[#e5484d]" bind:checked={hasCasualty} />
				<span class="text-sm font-semibold text-slate-800">Ya, ada korban</span>
			</label>

			{#if hasCasualty}
				<div class="mt-3">
					<label class="input-shell">
						<span class="input-label">Jumlah korban</span>
						<input type="number" bind:value={casualtyCount} min="1" max="999" class="input-field" />
					</label>
				</div>
			{/if}
		</div>

		<div class="jedug-card p-5">
			<div class="mb-4 flex items-start gap-3">
				<div class="flex size-11 shrink-0 items-center justify-center rounded-[18px] bg-brand-50 text-brand-600">
					<DocumentIcon class="size-6" />
				</div>
				<div>
					<h2 class="text-lg font-bold text-slate-950">Catatan tambahan</h2>
					<p class="mt-1 text-sm leading-6 text-slate-500">
						Tambahkan konteks singkat yang membantu orang lain memahami situasinya.
					</p>
				</div>
			</div>

			<textarea
				bind:value={note}
				placeholder="Deskripsi singkat kondisi jalan (opsional)"
				rows="4"
				maxlength="500"
				class="textarea-field"
			></textarea>
			<div class="mt-2 text-right text-xs font-semibold text-slate-400">{note.length}/500</div>
		</div>
	</section>

	{#if submitting && currentStep !== 'idle'}
		<div class="rounded-[22px] border border-brand-200 bg-brand-50 px-4 py-3 text-sm font-semibold text-brand-700">
			{stepLabels[currentStep]}
		</div>
	{/if}

	{#if error}
		<div class="error-panel" bind:this={errorRef}>{error}</div>
	{/if}

	<div class="jedug-card p-4">
		<button
			class="btn-primary w-full"
			onclick={handleSubmit}
			disabled={submitting || bootstrapInitializing || !selectedFile || locationLoading}
		>
			<DocumentIcon class="size-[18px]" />
			{#if submitting}
				Mengirim...
			{:else}
				Kirim Laporan
			{/if}
		</button>
		<p class="mt-3 text-center text-xs leading-5 text-slate-500">
			Laporan dikirim anonim dari browser ini. Pastikan foto dan koordinat sudah benar sebelum menekan tombol kirim.
		</p>
	</div>
</div>
