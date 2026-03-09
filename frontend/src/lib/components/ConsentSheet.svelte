<script lang="ts">
	let { onaccept }: { onaccept: () => void } = $props();
</script>

<div class="consent-overlay">
	<div class="consent-sheet">
		<h2>Sebelum Lanjut</h2>
		<div class="consent-body">
			<p>
				Dengan menggunakan JEDUG, kamu setuju bahwa:
			</p>
			<ul>
				<li>Laporan yang kamu kirim bersifat publik dan bisa dilihat semua orang</li>
				<li>Kami menyimpan data lokasi dan foto yang kamu kirim untuk keperluan pelaporan</li>
				<li>Identitasmu tetap anonim — tidak ada login, tidak ada data pribadi</li>
				<li>Kamu tidak akan mengirim konten yang melanggar hukum</li>
			</ul>
		</div>
		<button class="accept-btn" onclick={onaccept}>
			Setuju & Lanjutkan
		</button>
	</div>
</div>

<style>
	.consent-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: flex-end;
		justify-content: center;
		z-index: 1000;
		padding: 0;
	}
	.consent-sheet {
		background: #fff;
		border-radius: 20px 20px 0 0;
		padding: 24px 20px;
		max-width: 480px;
		width: 100%;
		animation: slide-up 0.3s ease-out;
	}
	@keyframes slide-up {
		from { transform: translateY(100%); }
		to { transform: translateY(0); }
	}
	h2 {
		margin: 0 0 12px;
		font-size: 1.2rem;
	}
	.consent-body {
		font-size: 0.9rem;
		color: #4a5568;
		line-height: 1.5;
		margin-bottom: 20px;
	}
	.consent-body p {
		margin: 0 0 8px;
	}
	ul {
		margin: 0;
		padding-left: 20px;
	}
	li {
		margin-bottom: 4px;
	}
	.accept-btn {
		display: block;
		width: 100%;
		padding: 14px;
		font-size: 1.05rem;
		font-weight: 600;
		color: #fff;
		background: #e53e3e;
		border: none;
		border-radius: 12px;
		cursor: pointer;
	}
	.accept-btn:hover {
		opacity: 0.9;
	}
</style>
