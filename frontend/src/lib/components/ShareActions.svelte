<script lang="ts">
	import {
		ArrowLeftIcon,
		CopyIcon,
		LocationIcon,
		MapIcon,
		ShareIcon
	} from '$lib/icons';

	let {
		title,
		shareText,
		shareUrl,
		externalMapUrl,
		backHref = '/issues',
		reportHref = '/lapor'
	}: {
		title: string;
		shareText: string;
		shareUrl: string;
		externalMapUrl: string;
		backHref?: string;
		reportHref?: string;
	} = $props();

	let shareMessage = $state<string | null>(null);
	let copyingLink = $state(false);

	const encodedUrl = $derived(encodeURIComponent(shareUrl));
	const encodedText = $derived(encodeURIComponent(shareText));
	const whatsappShareUrl = $derived(`https://wa.me/?text=${encodedText}%20${encodedUrl}`);
	const telegramShareUrl = $derived(`https://t.me/share/url?url=${encodedUrl}&text=${encodedText}`);
	const twitterShareUrl = $derived(`https://twitter.com/intent/tweet?text=${encodedText}&url=${encodedUrl}`);
	const facebookShareUrl = $derived(`https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`);

	const socialLinks = $derived.by(() => [
		{ label: 'WhatsApp', href: whatsappShareUrl },
		{ label: 'Telegram', href: telegramShareUrl },
		{ label: 'Twitter/X', href: twitterShareUrl },
		{ label: 'Facebook', href: facebookShareUrl }
	]);

	function showShareMessage(message: string) {
		shareMessage = message;
		setTimeout(() => {
			shareMessage = null;
		}, 2400);
	}

	async function copyLink() {
		if (copyingLink) return;

		copyingLink = true;
		try {
			if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
				await navigator.clipboard.writeText(shareUrl);
				showShareMessage('Link issue berhasil disalin.');
			} else {
				showShareMessage('Browser ini tidak mendukung salin link otomatis.');
			}
		} catch {
			showShareMessage('Gagal menyalin link issue.');
		} finally {
			copyingLink = false;
		}
	}

	async function handleShare() {
		if (
			typeof navigator !== 'undefined' &&
			'share' in navigator &&
			typeof navigator.share === 'function'
		) {
			try {
				await navigator.share({
					title,
					text: shareText,
					url: shareUrl
				});
				showShareMessage('Link issue berhasil dibagikan.');
				return;
			} catch (error) {
				if (error instanceof DOMException && error.name === 'AbortError') {
					return;
				}
			}
		}

		await copyLink();
	}
</script>

<section class="jedug-card p-5" aria-label="Aksi publik issue">
	<div class="flex items-start gap-3">
		<div class="flex size-11 shrink-0 items-center justify-center rounded-[18px] bg-brand-50 text-brand-600">
			<ShareIcon class="size-6" />
		</div>
		<div class="min-w-0">
			<h2 class="text-lg font-bold text-slate-950">Bagikan Issue</h2>
			<p class="mt-1 text-sm leading-6 text-slate-500">
				Siap dibagikan ke WhatsApp, Telegram, Twitter/X, Facebook, atau link biasa.
			</p>
		</div>
	</div>

	<div class="mt-5 grid grid-cols-1 gap-3 md:grid-cols-2">
		<a class="btn-secondary" href={backHref}>
			<ArrowLeftIcon class="size-[18px]" />
			Kembali ke Peta
		</a>
		<button type="button" class="btn-primary" onclick={handleShare}>
			<ShareIcon class="size-[18px]" />
			Bagikan Issue
		</button>
		<a class="btn-secondary" href={reportHref}>
			<MapIcon class="size-[18px]" />
			Lapor di Sekitar Sini
		</a>
		<a class="btn-secondary" href={externalMapUrl} target="_blank" rel="noopener noreferrer">
			<LocationIcon class="size-[18px]" />
			Buka Lokasi di Peta
		</a>
	</div>

	<div class="mt-4 flex flex-wrap gap-2">
		{#each socialLinks as item}
			<a
				href={item.href}
				target="_blank"
				rel="noopener noreferrer"
				class="inline-flex items-center justify-center rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-xs font-bold text-slate-600 transition hover:border-slate-300 hover:bg-white hover:text-slate-900"
			>
				{item.label}
			</a>
		{/each}
		<button
			type="button"
			class="inline-flex items-center justify-center gap-2 rounded-full border border-brand-200 bg-brand-50 px-3 py-2 text-xs font-bold text-brand-700 transition hover:bg-brand-100 disabled:cursor-not-allowed disabled:opacity-60"
			onclick={copyLink}
			disabled={copyingLink}
		>
			<CopyIcon class="size-4" />
			{copyingLink ? 'Menyalin...' : 'Salin Link'}
		</button>
	</div>

	{#if shareMessage}
		<p class="mt-4 rounded-[18px] border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-semibold text-emerald-700">
			{shareMessage}
		</p>
	{/if}
</section>
