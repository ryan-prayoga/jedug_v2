<script lang="ts">
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
	const telegramShareUrl = $derived(
		`https://t.me/share/url?url=${encodedUrl}&text=${encodedText}`
	);
	const twitterShareUrl = $derived(
		`https://twitter.com/intent/tweet?text=${encodedText}&url=${encodedUrl}`
	);
	const facebookShareUrl = $derived(
		`https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`
	);

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

<section class="share-card" aria-label="Aksi publik issue">
	<div class="card-header">
		<h2>Bagikan Issue</h2>
		<p>Siap dibagikan ke WhatsApp, Telegram, Twitter/X, Facebook, atau link biasa.</p>
	</div>

	<div class="cta-grid">
		<a class="btn btn-secondary" href={backHref}>Kembali ke Peta</a>
		<button type="button" class="btn btn-primary" onclick={handleShare}>Bagikan Issue</button>
		<a class="btn btn-secondary" href={reportHref}>Lapor di Sekitar Sini</a>
		<a class="btn btn-secondary" href={externalMapUrl} target="_blank" rel="noopener noreferrer">
			Buka Lokasi di Peta
		</a>
	</div>

	<div class="social-links">
		<a href={whatsappShareUrl} target="_blank" rel="noopener noreferrer">WhatsApp</a>
		<a href={telegramShareUrl} target="_blank" rel="noopener noreferrer">Telegram</a>
		<a href={twitterShareUrl} target="_blank" rel="noopener noreferrer">Twitter/X</a>
		<a href={facebookShareUrl} target="_blank" rel="noopener noreferrer">Facebook</a>
		<button type="button" onclick={copyLink} disabled={copyingLink}>
			{copyingLink ? 'Menyalin...' : 'Salin Link'}
		</button>
	</div>

	{#if shareMessage}
		<p class="share-message">{shareMessage}</p>
	{/if}
</section>

<style>
	.share-card {
		background: #fff;
		border: 1px solid #e2e8f0;
		border-radius: 16px;
		padding: 16px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
	}

	.card-header h2 {
		margin: 0;
		font-size: 18px;
		color: #0f172a;
	}

	.card-header p {
		margin-top: 6px;
		font-size: 13px;
		line-height: 1.5;
		color: #64748b;
	}

	.cta-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 8px;
		margin-top: 16px;
	}

	.btn {
		min-height: 48px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 10px 16px;
		border-radius: 12px;
		border: 1px solid #e2e8f0;
		text-decoration: none;
		font-size: 14px;
		font-weight: 700;
		text-align: center;
		transition: opacity 0.18s ease, transform 0.12s ease, background-color 0.18s ease;
		cursor: pointer;
	}

	.btn:active {
		transform: scale(0.98);
	}

	.btn-primary {
		background: #e5484d;
		border-color: #e5484d;
		color: #fff;
	}

	.btn-primary:hover {
		opacity: 0.92;
	}

	.btn-secondary {
		background: #fff;
		color: #0f172a;
	}

	.btn-secondary:hover {
		background: #f8fafc;
	}

	.social-links {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: 14px;
	}

	.social-links a,
	.social-links button {
		padding: 8px 10px;
		border-radius: 999px;
		border: 1px solid #e2e8f0;
		background: #fff;
		color: #334155;
		font-size: 12px;
		font-weight: 700;
		text-decoration: none;
		cursor: pointer;
	}

	.social-links button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.share-message {
		margin-top: 12px;
		padding: 9px 10px;
		border-radius: 10px;
		border: 1px solid #bbf7d0;
		background: #f0fdf4;
		color: #166534;
		font-size: 12px;
		font-weight: 700;
	}

	@media (min-width: 768px) {
		.cta-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
