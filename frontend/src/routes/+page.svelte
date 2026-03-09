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
			📸 Laporkan Jalan Rusak
		</a>
		<a href="/issues" class="btn btn-secondary">
			📋 Lihat Laporan Publik
		</a>
	</div>
</div>

<style>
	.landing {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding-top: 2.5rem;
		text-align: center;
	}
	.hero {
		margin-bottom: 1.5rem;
	}
	.title {
		font-size: 3rem;
		font-weight: 900;
		color: #e53e3e;
		letter-spacing: -1px;
		margin-bottom: 8px;
	}
	.tagline {
		font-size: 1.1rem;
		color: #4a5568;
		line-height: 1.5;
	}
	.description {
		max-width: 400px;
		margin-bottom: 2rem;
	}
	.description p {
		font-size: 0.95rem;
		color: #718096;
		line-height: 1.6;
		margin-bottom: 0.75rem;
	}
	.actions {
		display: flex;
		flex-direction: column;
		gap: 12px;
		width: 100%;
		max-width: 320px;
	}
	.btn {
		display: block;
		text-align: center;
		text-decoration: none;
		padding: 16px 24px;
		font-size: 1.1rem;
		font-weight: 600;
		border-radius: 12px;
		transition: opacity 0.15s;
	}
	.btn:hover {
		opacity: 0.9;
	}
	.btn-primary {
		background: #e53e3e;
		color: #fff;
	}
	.btn-secondary {
		background: #fff;
		color: #2d3748;
		border: 1px solid #e2e8f0;
	}
	.btn-secondary:hover {
		background: #f7fafc;
	}
</style>
