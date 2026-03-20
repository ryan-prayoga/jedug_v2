<script lang="ts">
	import { browserPushState } from '$lib/stores/browser-push';
	import { notificationPreferencesState } from '$lib/stores/notification-preferences';
	import {
		BellIcon,
		CheckCircleIcon,
		DangerIcon,
		InfoIcon,
		RefreshIcon
	} from '$lib/icons';
	import { resetAnonymousBrowserIdentity } from '$lib/utils/storage';

	type Variant = 'card' | 'compact';

	let {
		variant = 'card',
		title = 'Notifikasi Browser',
		lead = 'Aktifkan update browser agar perubahan issue tetap masuk saat JEDUG tidak sedang kamu buka.',
		blocked = false,
		blockedMessage = 'Ikuti issue ini dulu agar browser ini punya target notifikasi yang jelas.'
	}: {
		variant?: Variant;
		title?: string;
		lead?: string;
		blocked?: boolean;
		blockedMessage?: string;
	} = $props();

	const pushState = $derived($browserPushState);
	const requiresRepair = $derived(pushState.needsFollowerRebind);

	const canEnable = $derived(
		!blocked &&
			!requiresRepair &&
			(pushState.status === 'default' || pushState.status === 'granted')
	);
	const canDisable = $derived(!requiresRepair && pushState.status === 'subscribed');
	const statusTone = $derived.by(() => {
		if (requiresRepair) return 'warning';
		switch (pushState.status) {
			case 'subscribed':
				return 'success';
			case 'ios_browser_tab':
			case 'denied':
				return 'warning';
			case 'unsupported':
				return 'muted';
			default:
				return 'default';
		}
	});

	const statusLabel = $derived.by(() => {
		if (pushState.status === 'subscribed') return 'Aktif';
		if (pushState.status === 'ios_browser_tab') return 'Home Screen';
		if (pushState.status === 'denied') return 'Ditolak';
		if (pushState.status === 'unsupported') return 'Tidak didukung';
		if (pushState.status === 'granted') return 'Siap';
		return 'Belum aktif';
	});

	const toneClasses = $derived.by(() => {
		if (statusTone === 'success') return 'border-emerald-200 bg-emerald-50 text-emerald-700';
		if (statusTone === 'warning') return 'border-amber-200 bg-amber-50 text-amber-800';
		if (statusTone === 'muted') return 'border-slate-200 bg-slate-100 text-slate-600';
		return 'border-brand-200 bg-brand-50 text-brand-700';
	});

	const iconToneClasses = $derived.by(() => {
		if (statusTone === 'success') return 'bg-emerald-50 text-emerald-600';
		if (statusTone === 'warning') return 'bg-amber-50 text-amber-700';
		if (statusTone === 'muted') return 'bg-slate-100 text-slate-600';
		return 'bg-brand-50 text-brand-600';
	});

	const summary = $derived.by(() => {
		if (requiresRepair) {
			return (
				pushState.followerAuthMessage ||
				'Sesi notifikasi browser ini perlu di-reset dari browser yang sama sebelum bisa dipakai lagi.'
			);
		}

		if (blocked && pushState.status !== 'subscribed') {
			return blockedMessage;
		}

		switch (pushState.status) {
			case 'ios_browser_tab':
				return 'Di iPhone, notifikasi browser untuk JEDUG hanya aktif jika aplikasi web ini ditambahkan ke Home Screen lalu dibuka dari ikon app.';
			case 'unsupported':
				return 'Browser ini belum mendukung Web Push.';
			case 'denied':
				return 'Izin notifikasi ditolak. Ubah lewat pengaturan browser jika ingin mengaktifkannya lagi.';
			case 'granted':
				return 'Izin browser sudah ada. Tinggal selesaikan subscription agar update tetap masuk saat tab ditutup.';
			case 'subscribed':
				return 'Notifikasi browser aktif di perangkat ini.';
			default:
				return lead;
		}
	});

	const installSteps = [
		'Buka menu Share di Safari.',
		'Pilih Add to Home Screen.',
		'Buka JEDUG dari ikon yang terpasang, lalu aktifkan notifikasi lagi.'
	];

	const buttonLabel = $derived.by(() => {
		if (pushState.busy) {
			return pushState.status === 'subscribed' ? 'Mematikan...' : 'Mengaktifkan...';
		}
		if (requiresRepair) return 'Reset browser ini';
		if (pushState.status === 'subscribed') return 'Matikan di perangkat ini';
		return 'Aktifkan notifikasi browser';
	});

	const statusIcon = $derived.by(() => {
		if (requiresRepair) return RefreshIcon;
		switch (pushState.status) {
			case 'subscribed':
				return CheckCircleIcon;
			case 'denied':
			case 'ios_browser_tab':
				return DangerIcon;
			case 'unsupported':
				return InfoIcon;
			default:
				return BellIcon;
		}
	});

	const StatusIcon = $derived(statusIcon);

	async function handleAction() {
		if (requiresRepair) {
			resetAnonymousBrowserIdentity();
			window.location.reload();
			return;
		}
		if (blocked && pushState.status !== 'subscribed') return;
		if (pushState.status === 'subscribed') {
			const disabled = await browserPushState.disable();
			if (disabled) {
				await notificationPreferencesState.refresh();
			}
			return;
		}
		const enabled = await browserPushState.enable();
		if (enabled) {
			await notificationPreferencesState.refresh();
		}
	}
</script>

<section
	class={variant === 'compact'
		? 'jedug-card-soft mx-1 mb-3 p-4'
		: 'jedug-card p-5'}
	aria-live="polite"
>
	<div class="flex flex-col gap-4">
		<div class="flex items-start gap-3">
			<div class={`flex size-11 shrink-0 items-center justify-center rounded-[18px] ${iconToneClasses}`}>
				<StatusIcon class="size-5" />
			</div>
			<div class="min-w-0 flex-1 space-y-2">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<h3 class="text-sm font-bold text-slate-950">{title}</h3>
					<span class={`inline-flex items-center gap-1 rounded-full border px-3 py-1 text-[11px] font-bold uppercase tracking-[0.16em] ${toneClasses}`}>
						{statusLabel}
					</span>
				</div>
				<p class="text-sm leading-6 text-slate-600">{summary}</p>
			</div>
		</div>

		{#if requiresRepair}
			<ol class="space-y-2 pl-5 text-xs leading-5 text-slate-600">
				<li>Tekan tombol reset di bawah.</li>
				<li>Setujui ulang popup consent JEDUG setelah halaman reload.</li>
				<li>Ikuti laporan lagi bila diperlukan, lalu aktifkan notifikasi browser.</li>
			</ol>
		{:else if pushState.status === 'ios_browser_tab'}
			<ol class="space-y-2 pl-5 text-xs leading-5 text-slate-600">
				{#each installSteps as step}
					<li>{step}</li>
				{/each}
			</ol>
		{/if}

		{#if pushState.error}
			<p class="error-panel border-0 bg-rose-50/80">{pushState.error}</p>
		{/if}
		{#if pushState.success}
			<p class="rounded-[18px] border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
				{pushState.success}
			</p>
		{/if}

		{#if requiresRepair || canEnable || canDisable}
			<button
				class={pushState.status === 'subscribed' && !requiresRepair ? 'btn-secondary w-full' : 'btn-primary w-full'}
				disabled={pushState.busy}
				type="button"
				onclick={handleAction}
			>
				<StatusIcon class="size-[18px]" />
				{buttonLabel}
			</button>
		{/if}
	</div>
</section>
