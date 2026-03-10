<div class="landing">
	<div class="hero">
		<h1 class="title">JEDUG</h1>
		<p class="tagline">Lapor, pantau, viralkan.<br />Biar nggak ada yang kejedug lagi.</p>
	</div>

	<div class="description">
		<p>
			Jalan rusak di sekitarmu? Jangan cuma ngomel — laporkan. JEDUG adalah platform publik
			untuk mendokumentasi dan memantau kondisi jalan rusak di Indonesia. Anonim, cepat, langsung dari HP.
		</p>
		<p>
			Setiap laporan masuk jadi data terbuka yang bisa dilihat siapa saja.
			Makin banyak laporan, makin sulit diabaikan.
		</p>
	</div>

	<div class="actions">
		<a href="/lapor" class="btn btn-primary">
			Laporkan Jalan Rusak
		</a>
		<a href="/issues" class="btn btn-secondary">
			Lihat Laporan Publik
		</a>
	</div>
</div>

<style>
	.landing {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding-top: 48px;
		text-align: center;
	}
	.hero {
		margin-bottom: 24px;
	}
	.title {
		font-size: 48px;
		font-weight: 900;
		color: #E5484D;
		letter-spacing: -1.5px;
		margin-bottom: 8px;
		line-height: 1;
	}
	.tagline {
		font-size: 16px;
		color: #64748B;
		line-height: 1.5;
	}
	.description {
		max-width: 400px;
		margin-bottom: 32px;
	}
	.description p {
		font-size: 14px;
		color: #64748B;
		line-height: 1.6;
		margin-bottom: 12px;
	}
	.actions {
		display: flex;
		flex-direction: column;
		gap: 12px;
		width: 100%;
		max-width: 320px;
	}
	.btn {
		display: flex;
		align-items: center;
		justify-content: center;
		text-align: center;
		text-decoration: none;
		padding: 14px 24px;
		font-size: 16px;
		font-weight: 600;
		border-radius: 12px;
		min-height: 48px;
		transition: opacity 0.15s, transform 0.1s;
	}
	.btn:active {
		transform: scale(0.97);
	}
	.btn-primary {
		background: #E5484D;
		color: #fff;
		box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04);
	}
	.btn-primary:hover {
		opacity: 0.88;
	}
	.btn-secondary {
		background: #fff;
		color: #0F172A;
		border: 1px solid #E2E8F0;
	}
	.btn-secondary:hover {
		background: #F8FAFC;
	}
</style>
