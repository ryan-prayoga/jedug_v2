<script lang="ts">
	import { onMount } from 'svelte';
	import { browserPushState } from '$lib/stores/browser-push';
	import { notificationPreferencesState } from '$lib/stores/notification-preferences';
	import { ArrowUpIcon, BellIcon, SettingsIcon } from '$lib/icons';

	const prefsState = $derived($notificationPreferencesState);
	const pushState = $derived($browserPushState);

	let open = $state(false);

	const preferences = $derived(prefsState.preferences);
	const channelsDisabled = $derived(!preferences || !preferences.notifications_enabled);
	const pushAvailable = $derived(pushState.status === 'subscribed');
	const pushToggleDisabled = $derived(
		channelsDisabled ||
			pushState.busy ||
			pushState.needsFollowerRebind ||
			(!pushAvailable && !preferences?.push_enabled)
	);
	const pushHint = $derived.by(() => {
		if (!preferences) return '';
		if (pushState.needsFollowerRebind) {
			return (
				pushState.followerAuthMessage ||
				'Browser ini perlu di-reset lalu consent JEDUG perlu diulang sebelum channel push bisa dipakai lagi.'
			);
		}
		if (pushAvailable) {
			return 'Update akan dikirim ke browser ini saat langganan push masih aktif.';
		}
		if (pushState.status === 'ios_browser_tab') {
			return 'Di iPhone, buka JEDUG dari Home Screen dulu sebelum channel push bisa dipakai.';
		}
		if (pushState.status === 'denied') {
			return 'Izin browser ditolak. Aktifkan lagi dari pengaturan browser jika perlu.';
		}
		if (pushState.status === 'unsupported') {
			return 'Browser ini belum mendukung push notification.';
		}
		return 'Aktifkan notifikasi browser di kartu atas dulu untuk memakai channel ini.';
	});

	const preferenceSections = $derived.by(() => {
		if (!preferences) return [];
		return [
			{
				title: 'Umum',
				items: [
					{
						key: 'notifications_enabled',
						label: 'Semua notifikasi',
						description: 'Matikan seluruh update untuk laporan yang kamu ikuti di browser ini.',
						checked: preferences.notifications_enabled,
						disabled: isSaving('notifications_enabled')
					}
				]
			},
			{
				title: 'Channel',
				items: [
					{
						key: 'in_app_enabled',
						label: 'Notifikasi di dalam aplikasi',
						description: 'Update baru muncul di panel lonceng saat kamu membuka JEDUG.',
						checked: preferences.in_app_enabled,
						disabled: channelsDisabled || isSaving('in_app_enabled')
					},
					{
						key: 'push_enabled',
						label: 'Notifikasi browser',
						description: pushHint,
						checked: preferences.push_enabled,
						disabled: pushToggleDisabled || isSaving('push_enabled')
					}
				]
			},
			{
				title: 'Jenis update',
				items: [
					{
						key: 'notify_on_photo_added',
						label: 'Foto baru pada laporan yang kamu ikuti',
						description: 'Dipakai saat ada bukti foto tambahan pada issue yang sama.',
						checked: preferences.notify_on_photo_added,
						disabled: channelsDisabled || isSaving('notify_on_photo_added')
					},
					{
						key: 'notify_on_status_updated',
						label: 'Perubahan status laporan',
						description: 'Misalnya saat issue ditandai selesai, ditolak, atau diarsipkan.',
						checked: preferences.notify_on_status_updated,
						disabled: channelsDisabled || isSaving('notify_on_status_updated')
					},
					{
						key: 'notify_on_severity_changed',
						label: 'Perubahan tingkat keparahan',
						description: 'Dipakai saat tingkat kerusakan issue dinaikkan oleh laporan baru.',
						checked: preferences.notify_on_severity_changed,
						disabled: channelsDisabled || isSaving('notify_on_severity_changed')
					},
					{
						key: 'notify_on_casualty_reported',
						label: 'Laporan korban baru',
						description: 'Dipakai saat ada laporan korban baru atau jumlah korban meningkat.',
						checked: preferences.notify_on_casualty_reported,
						disabled: channelsDisabled || isSaving('notify_on_casualty_reported')
					},
					{
						key: 'notify_on_nearby_issue_created',
						label: 'Laporan baru di area pantauan',
						description: 'Dipakai saat ada issue baru yang masuk ke radius lokasi Nearby Alerts milikmu.',
						checked: preferences.notify_on_nearby_issue_created,
						disabled: channelsDisabled || isSaving('notify_on_nearby_issue_created')
					}
				]
			}
		];
	});

	function isSaving(key: string): boolean {
		return prefsState.savingKeys.includes(key);
	}

	onMount(() => {
		void browserPushState.init();
		void notificationPreferencesState.init();
	});

	async function toggle(key: string, value: boolean) {
		await notificationPreferencesState.update({ [key]: value }, key);
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
			<div class="flex size-10 shrink-0 items-center justify-center rounded-2xl bg-slate-100 text-slate-700">
				<SettingsIcon class="size-5" />
			</div>
			<div class="min-w-0">
				<div class="text-sm font-bold text-slate-950">Preferensi Notifikasi</div>
				<p class="mt-1 text-xs leading-5 text-slate-500">
					Atur update mana yang masih ingin kamu terima.
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
			{#if prefsState.loading}
				<div class="state-panel border-0 bg-white/70 px-4 py-6">
					<div class="mx-auto size-9 animate-spin rounded-full border-[3px] border-slate-200 border-t-brand-500"></div>
					<p class="mt-3 text-sm text-slate-500">Memuat preferensi...</p>
				</div>
			{:else if prefsState.unavailableMessage}
				<div class="notice-panel">{prefsState.unavailableMessage}</div>
			{:else if prefsState.error}
				<div class="error-panel">{prefsState.error}</div>
			{:else if preferences}
				{#each preferenceSections as section}
					<div class="space-y-3">
						<div class="flex items-center gap-2 px-1">
							<BellIcon class="size-4 text-brand-500" />
							<div class="text-[11px] font-bold uppercase tracking-[0.18em] text-slate-400">
								{section.title}
							</div>
						</div>

						<div class="space-y-2">
							{#each section.items as item}
								<label
									class={`flex items-start justify-between gap-4 rounded-[22px] border px-4 py-3 shadow-[0_10px_22px_rgba(15,23,42,0.04)] ${item.checked ? 'border-brand-100 bg-brand-50/50' : 'border-slate-200 bg-white'}`}
								>
									<div class="space-y-1">
										<div class="text-sm font-semibold text-slate-900">{item.label}</div>
										<p class="text-xs leading-5 text-slate-500">{item.description}</p>
									</div>
									<input
										type="checkbox"
										class="mt-1 h-4 w-4 shrink-0 accent-[#e5484d]"
										checked={item.checked}
										disabled={item.disabled}
										onchange={(event) =>
											toggle(item.key, (event.currentTarget as HTMLInputElement).checked)}
									/>
								</label>
							{/each}
						</div>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</section>
