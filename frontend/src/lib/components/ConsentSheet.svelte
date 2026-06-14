<script lang="ts">
	let { onaccept }: { onaccept: () => void } = $props();
</script>

<div class="consent-overlay">
	<div class="consent-sheet">
		<p class="kicker mb-2">Sebelum lanjut</p>
		<h2>Ketentuan singkat</h2>
		<div class="consent-body">
			<p>Dengan menggunakan JEDUG, kamu setuju bahwa:</p>
			<ul>
				<li>Laporan yang kamu kirim bersifat publik dan bisa dilihat semua orang</li>
				<li>Kami menyimpan data lokasi dan foto yang kamu kirim untuk keperluan pelaporan</li>
				<li>Identitasmu tetap anonim — tidak ada login, tidak ada data pribadi</li>
				<li>Kamu tidak akan mengirim konten yang melanggar hukum</li>
			</ul>
		</div>
		<button class="btn-primary w-full" onclick={onaccept}> Setuju & Lanjutkan </button>
	</div>
</div>

<style>
	.consent-overlay {
		position: fixed;
		inset: 0;
		background: rgba(26, 26, 26, 0.45);
		display: flex;
		align-items: flex-end;
		justify-content: center;
		z-index: 1000;
	}
	.consent-sheet {
		background: var(--surface);
		border-top: 1px solid var(--hairline);
		border-radius: 12px 12px 0 0;
		padding: 28px 24px calc(24px + env(safe-area-inset-bottom));
		max-width: 480px;
		width: 100%;
	}
	@media (prefers-reduced-motion: no-preference) {
		.consent-sheet {
			animation: slide-up 0.25s ease-out;
		}
		@keyframes slide-up {
			from {
				transform: translateY(100%);
			}
			to {
				transform: translateY(0);
			}
		}
	}
	h2 {
		margin: 0 0 14px;
		font-family: var(--font-serif);
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.01em;
		color: var(--ink);
	}
	.consent-body {
		font-size: 14px;
		color: var(--muted);
		line-height: 1.6;
		margin-bottom: 24px;
	}
	.consent-body p {
		margin: 0 0 10px;
	}
	ul {
		margin: 0;
		padding-left: 18px;
		list-style: none;
	}
	li {
		position: relative;
		margin-bottom: 8px;
		padding-left: 4px;
	}
	li::before {
		content: "—";
		position: absolute;
		left: -18px;
		color: var(--subtle);
	}
</style>
