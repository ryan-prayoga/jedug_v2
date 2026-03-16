<script lang="ts">
	import { browserPushState } from '$lib/stores/browser-push';

	type Variant = 'card' | 'compact';

	let {
		variant = 'card',
		title = 'Notifikasi Browser',
		lead = 'Aktifkan update browser agar perubahan issue tetap masuk saat JEDUG tidak sedang kamu buka.',
		blocked = false,
		blockedMessage = 'Ikuti issue ini dulu agar browser ini punya target notifikasi yang jelas.',
	}: {
		variant?: Variant;
		title?: string;
		lead?: string;
		blocked?: boolean;
		blockedMessage?: string;
	} = $props();

	const pushState = $derived($browserPushState);

	const canEnable = $derived(
		!blocked &&
			(pushState.status === 'default' || pushState.status === 'granted')
	);
	const canDisable = $derived(pushState.status === 'subscribed');
	const statusTone = $derived.by(() => {
		switch (pushState.status) {
			case 'subscribed':
				return 'success';
			case 'denied':
				return 'warning';
			case 'unsupported':
				return 'muted';
			default:
				return 'default';
		}
	});

	const summary = $derived.by(() => {
		if (blocked && pushState.status !== 'subscribed') {
			return blockedMessage;
		}

		switch (pushState.status) {
			case 'unsupported':
				return 'Browser ini belum mendukung Web Push.';
			case 'denied':
				return 'Izin notifikasi ditolak. Ubah lewat pengaturan browser jika ingin mengaktifkannya lagi.';
			case 'granted':
				return 'Izin browser sudah ada. Selesaikan subscription agar update tetap masuk saat tab ditutup.';
			case 'subscribed':
				return 'Notifikasi browser aktif di perangkat ini.';
			default:
				return lead;
		}
	});

	const buttonLabel = $derived.by(() => {
		if (pushState.busy) {
			return pushState.status === 'subscribed' ? 'Mematikan...' : 'Mengaktifkan...';
		}
		if (pushState.status === 'subscribed') {
			return 'Matikan di perangkat ini';
		}
		return 'Aktifkan notifikasi browser';
	});

	async function handleAction() {
		if (blocked && pushState.status !== 'subscribed') return;
		if (pushState.status === 'subscribed') {
			await browserPushState.disable();
			return;
		}
		await browserPushState.enable();
	}
</script>

<section class:compact={variant === 'compact'} class="push-card" aria-live="polite">
	<div class="push-copy">
		<div class="push-header">
			<h3>{title}</h3>
			<span class:success={statusTone === 'success'} class:warning={statusTone === 'warning'}>
				{pushState.status === 'subscribed'
					? 'Aktif'
					: pushState.status === 'denied'
						? 'Ditolak'
						: pushState.status === 'unsupported'
							? 'Tidak didukung'
							: pushState.status === 'granted'
								? 'Siap'
								: 'Belum aktif'}
			</span>
		</div>

		<p>{summary}</p>

		{#if pushState.error}
			<p class="push-error">{pushState.error}</p>
		{/if}
		{#if pushState.success}
			<p class="push-success">{pushState.success}</p>
		{/if}
	</div>

	{#if canEnable || canDisable}
		<button class="push-button" disabled={pushState.busy} type="button" onclick={handleAction}>
			{buttonLabel}
		</button>
	{/if}
</section>

<style>
	.push-card {
		display: grid;
		gap: 12px;
		padding: 14px;
		border-radius: 14px;
		border: 1px solid #E2E8F0;
		background: linear-gradient(180deg, #FFF7ED 0%, #FFFFFF 100%);
	}

	.push-card.compact {
		padding: 12px;
		border-radius: 12px;
		background: #FFF7ED;
		margin: 4px 8px 12px;
	}

	.push-copy {
		display: grid;
		gap: 8px;
	}

	.push-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}

	h3 {
		font-size: 14px;
		font-weight: 700;
		color: #0F172A;
	}

	span {
		flex-shrink: 0;
		font-size: 11px;
		font-weight: 700;
		padding: 4px 8px;
		border-radius: 999px;
		background: #F1F5F9;
		color: #475569;
	}

	span.success {
		background: #DCFCE7;
		color: #166534;
	}

	span.warning {
		background: #FEF3C7;
		color: #92400E;
	}

	p {
		font-size: 13px;
		color: #475569;
		line-height: 1.5;
	}

	.push-button {
		min-height: 48px;
		border: none;
		border-radius: 12px;
		background: #E5484D;
		color: #FFFFFF;
		font-size: 14px;
		font-weight: 700;
		cursor: pointer;
		transition: transform 0.16s ease, box-shadow 0.16s ease, opacity 0.16s ease;
	}

	.push-button:hover:enabled {
		transform: translateY(-1px);
		box-shadow: 0 10px 24px rgba(229, 72, 77, 0.18);
	}

	.push-button:active:enabled {
		transform: scale(0.98);
	}

	.push-button:disabled {
		opacity: 0.55;
		cursor: default;
		box-shadow: none;
	}

	.push-error {
		color: #B42318;
	}

	.push-success {
		color: #166534;
	}
</style>
