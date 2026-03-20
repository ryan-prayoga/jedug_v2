<script lang="ts">
	import { getLocation } from '$lib/utils/geolocation';
	import { nearbyAlertsState } from '$lib/stores/nearby-alerts';
	import {
		AddCircleIcon,
		ArrowUpIcon,
		CompassIcon,
		LocationIcon,
		TrashIcon
	} from '$lib/icons';

	const nearbyState = $derived($nearbyAlertsState);

	let open = $state(false);
	let locating = $state(false);
	let formLabel = $state('');
	let formLatitude = $state('');
	let formLongitude = $state('');
	let formRadius = $state('500');
	let formError = $state<string | null>(null);
	let drafts = $state<Record<string, { label: string; radius: string }>>({});

	const limitReached = $derived(nearbyState.items.length >= 10);

	$effect(() => {
		if (!open) return;
		void nearbyAlertsState.init();
	});

	$effect(() => {
		const activeIDs = new Set(nearbyState.items.map((item) => item.id));
		for (const item of nearbyState.items) {
			if (!drafts[item.id]) {
				drafts[item.id] = {
					label: item.label ?? '',
					radius: String(item.radius_m)
				};
			}
		}
		for (const id of Object.keys(drafts)) {
			if (!activeIDs.has(id)) {
				delete drafts[id];
			}
		}
	});

	function isSaving(key: string): boolean {
		return nearbyState.savingKeys.includes(key);
	}

	function updateDraftLabel(id: string, value: string) {
		drafts[id] = {
			...(drafts[id] ?? { label: '', radius: '500' }),
			label: value
		};
	}

	function updateDraftRadius(id: string, value: string) {
		drafts[id] = {
			...(drafts[id] ?? { label: '', radius: '500' }),
			radius: value
		};
	}

	async function useCurrentLocation() {
		locating = true;
		formError = null;
		try {
			const location = await getLocation({ forceFresh: true });
			formLatitude = location.latitude.toFixed(6);
			formLongitude = location.longitude.toFixed(6);
		} catch (error) {
			formError = error instanceof Error ? error.message : 'Belum bisa mengambil lokasi browser.';
		} finally {
			locating = false;
		}
	}

	async function createAlert() {
		formError = null;
		const latitude = Number.parseFloat(formLatitude);
		const longitude = Number.parseFloat(formLongitude);
		const radius = Number.parseInt(formRadius, 10);

		if (Number.isNaN(latitude) || Number.isNaN(longitude)) {
			formError = 'Isi latitude dan longitude dulu, atau pakai lokasi browser.';
			return;
		}
		if (latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) {
			formError = 'Koordinat belum valid. Pastikan latitude dan longitude benar.';
			return;
		}
		if (!Number.isFinite(radius) || radius < 100 || radius > 5000) {
			formError = 'Radius harus antara 100 sampai 5000 meter.';
			return;
		}

		const created = await nearbyAlertsState.create({
			label: formLabel.trim() || undefined,
			latitude,
			longitude,
			radius_m: radius,
			enabled: true
		});
		if (!created) return;

		formLabel = '';
		formLatitude = '';
		formLongitude = '';
		formRadius = '500';
	}

	async function saveItem(id: string) {
		const draft = drafts[id];
		if (!draft) return;

		const radius = Number.parseInt(draft.radius, 10);
		if (!Number.isFinite(radius) || radius < 100 || radius > 5000) {
			formError = 'Radius harus antara 100 sampai 5000 meter.';
			return;
		}

		formError = null;
		await nearbyAlertsState.update(
			id,
			{
				label: draft.label.trim(),
				radius_m: radius
			},
			`save:${id}`
		);
	}

	async function toggleItem(id: string, enabled: boolean) {
		await nearbyAlertsState.update(id, { enabled }, `toggle:${id}`);
	}

	async function deleteItem(id: string) {
		formError = null;
		await nearbyAlertsState.remove(id);
	}

	function formatCoordinates(latitude: number, longitude: number): string {
		return `${latitude.toFixed(5)}, ${longitude.toFixed(5)}`;
	}
</script>

<section class="jedug-card-soft mx-1 mb-3 overflow-hidden">
	<button
		type="button"
		class="flex w-full items-center justify-between gap-3 px-4 py-4 text-left"
		aria-expanded={open}
		onclick={() => (open = !open)}
	>
		<div class="flex min-w-0 items-start gap-3">
			<div class="flex size-10 shrink-0 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
				<LocationIcon class="size-5" />
			</div>
			<div class="min-w-0">
				<div class="text-sm font-bold text-slate-950">Pantau area sekitar rumah</div>
				<p class="mt-1 text-xs leading-5 text-slate-500">
					Beri tahu saya jika ada laporan baru di sekitar lokasi ini.
				</p>
			</div>
		</div>
		<span class="badge-muted gap-2">
			{open ? 'Tutup' : 'Atur'}
			<ArrowUpIcon class={`size-4 transition ${open ? '' : 'rotate-180'}`} />
		</span>
	</button>

	{#if open}
		<div class="space-y-4 border-t border-slate-100 px-4 pb-4 pt-3">
			<div class="rounded-[22px] border border-slate-200 bg-white px-4 py-3 text-sm leading-6 text-slate-600">
				Simpan beberapa lokasi seperti rumah, kantor, atau area yang ingin kamu pantau. Nearby Alerts tetap mengikuti preferensi notifikasi dan channel yang kamu aktifkan.
			</div>

			<div class="jedug-panel space-y-4 p-4">
				<div class="flex items-center gap-2">
					<AddCircleIcon class="size-5 text-brand-500" />
					<div>
						<p class="text-sm font-bold text-slate-900">Tambah lokasi pantauan</p>
						<p class="text-xs leading-5 text-slate-500">Maksimum 10 lokasi agar notifikasi tetap ringan.</p>
					</div>
				</div>

				<div class="grid gap-3 sm:grid-cols-2">
					<label class="input-shell">
						<span class="input-label">Label lokasi</span>
						<input
							class="input-field"
							type="text"
							placeholder="Mis. Rumah / Kantor"
							bind:value={formLabel}
							maxlength="80"
						/>
					</label>
					<label class="input-shell">
						<span class="input-label">Radius (meter)</span>
						<input class="input-field" type="number" min="100" max="5000" step="50" bind:value={formRadius} />
					</label>
					<label class="input-shell">
						<span class="input-label">Latitude</span>
						<input
							class="input-field"
							type="number"
							step="0.000001"
							placeholder="-6.200000"
							bind:value={formLatitude}
						/>
					</label>
					<label class="input-shell">
						<span class="input-label">Longitude</span>
						<input
							class="input-field"
							type="number"
							step="0.000001"
							placeholder="106.816666"
							bind:value={formLongitude}
						/>
					</label>
				</div>

				<div class="flex flex-col gap-2 sm:flex-row">
					<button
						type="button"
						class="btn-secondary flex-1"
						disabled={locating || nearbyState.creating}
						onclick={useCurrentLocation}
					>
						<CompassIcon class="size-[18px]" />
						{locating ? 'Mengambil lokasi...' : 'Gunakan lokasi saya'}
					</button>
					<button
						type="button"
						class="btn-primary flex-1"
						disabled={limitReached || nearbyState.creating}
						onclick={createAlert}
					>
						<AddCircleIcon class="size-[18px]" />
						{nearbyState.creating ? 'Menyimpan...' : 'Tambah lokasi'}
					</button>
				</div>

				{#if limitReached}
					<div class="notice-panel">Maksimum 10 lokasi pantauan per browser agar notifikasi tetap ringan dan tidak spammy.</div>
				{/if}
			</div>

			{#if formError}
				<div class="error-panel">{formError}</div>
			{/if}

			{#if nearbyState.loading}
				<div class="state-panel border-0 bg-white/70 px-4 py-6">
					<div class="mx-auto size-9 animate-spin rounded-full border-[3px] border-slate-200 border-t-brand-500"></div>
					<p class="mt-3 text-sm text-slate-500">Memuat lokasi pantauan...</p>
				</div>
			{:else if nearbyState.unavailableMessage}
				<div class="notice-panel">{nearbyState.unavailableMessage}</div>
			{:else if nearbyState.error}
				<div class="error-panel">{nearbyState.error}</div>
			{/if}

			{#if nearbyState.items.length > 0}
				<div class="space-y-3">
					{#each nearbyState.items as item (item.id)}
						<article
							class="jedug-card space-y-4 p-4"
							aria-busy={isSaving(`save:${item.id}`) || isSaving(`toggle:${item.id}`)}
						>
							<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
								<div>
									<p class="text-sm font-bold text-slate-950">{item.label || 'Area pantauan'}</p>
									<p class="mt-1 text-xs leading-5 text-slate-500">
										{formatCoordinates(item.latitude, item.longitude)}
									</p>
								</div>
								<label class="flex items-center gap-2 text-xs font-semibold text-slate-500">
									<span>{item.enabled ? 'Aktif' : 'Nonaktif'}</span>
									<input
										type="checkbox"
										class="h-4 w-4 accent-[#e5484d]"
										checked={item.enabled}
										disabled={isSaving(`toggle:${item.id}`)}
										onchange={(event) =>
											toggleItem(item.id, (event.currentTarget as HTMLInputElement).checked)}
									/>
								</label>
							</div>

							<div class="grid gap-3 sm:grid-cols-2">
								<label class="input-shell">
									<span class="input-label">Label</span>
									<input
										class="input-field"
										type="text"
										value={drafts[item.id]?.label ?? item.label ?? ''}
										maxlength="80"
										oninput={(event) =>
											updateDraftLabel(item.id, (event.currentTarget as HTMLInputElement).value)}
									/>
								</label>
								<label class="input-shell">
									<span class="input-label">Radius (meter)</span>
									<input
										class="input-field"
										type="number"
										min="100"
										max="5000"
										step="50"
										value={drafts[item.id]?.radius ?? String(item.radius_m)}
										oninput={(event) =>
											updateDraftRadius(item.id, (event.currentTarget as HTMLInputElement).value)}
									/>
								</label>
							</div>

							<div class="flex flex-col gap-2 sm:flex-row">
								<button
									type="button"
									class="btn-danger flex-1"
									disabled={nearbyState.deletingIDs.includes(item.id)}
									onclick={() => deleteItem(item.id)}
								>
									<TrashIcon class="size-[18px]" />
									{nearbyState.deletingIDs.includes(item.id) ? 'Menghapus...' : 'Hapus'}
								</button>
								<button
									type="button"
									class="btn-primary flex-1"
									disabled={isSaving(`save:${item.id}`)}
									onclick={() => saveItem(item.id)}
								>
									<AddCircleIcon class="size-[18px]" />
									{isSaving(`save:${item.id}`) ? 'Menyimpan...' : 'Simpan perubahan'}
								</button>
							</div>
						</article>
					{/each}
				</div>
			{:else if !nearbyState.loading && !nearbyState.error && !nearbyState.unavailableMessage}
				<div class="state-panel border-0 bg-white/70 px-4 py-6">
					<div class="mx-auto flex size-12 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
						<LocationIcon class="size-6" />
					</div>
					<p class="mt-3 text-sm font-semibold text-slate-700">Belum ada area pantauan</p>
					<p class="mt-1 text-xs leading-5 text-slate-500">
						Tambahkan lokasi penting buatmu supaya laporan baru di sekitarnya bisa langsung masuk ke notifikasi.
					</p>
				</div>
			{/if}
		</div>
	{/if}
</section>
