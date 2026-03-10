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
		font-size: 18px;
		font-weight: 700;
		color: #0F172A;
	}
	.consent-body {
		font-size: 14px;
		color: #64748B;
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
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		padding: 14px;
		font-size: 16px;
		font-weight: 600;
		color: #fff;
		background: #E5484D;
		border: none;
		border-radius: 12px;
		min-height: 48px;
		cursor: pointer;
		transition: opacity 0.15s, transform 0.1s;
	}
	.accept-btn:hover {
		opacity: 0.88;
	}
	.accept-btn:active {
		transform: scale(0.97);
	}
</style>
