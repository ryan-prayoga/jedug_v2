<div class="landing">
	<section class="hero" aria-labelledby="home-title">
		<div class="hero-kicker">Pantau Jalan Rusak</div>
		<h1 id="home-title" class="title">JEDUG</h1>
		<p class="tagline">Platform publik untuk melaporkan dan memantau jalan rusak secara cepat.</p>
		<p class="trust">Data terbuka, anonim, dan bisa dipantau siapa saja.</p>
	</section>

	<section class="actions" aria-label="Aksi utama">
		<a href="/lapor" class="btn btn-primary">Laporkan Jalan Rusak</a>
		<a href="/issues" class="btn btn-secondary">Lihat Peta Laporan</a>
	</section>

	<section class="intro" aria-label="Ringkasan JEDUG">
		<p>
			JEDUG membantu warga mengubah keluhan jalan rusak menjadi bukti yang rapi dan bisa ditindak.
			Setiap laporan memperkuat visibilitas titik kerusakan di peta publik.
		</p>
		<div class="feature-points">
			<div class="point">Lapor cepat dari HP</div>
			<div class="point">Titik kerusakan terlihat di peta</div>
			<div class="point">Moderasi untuk menjaga kualitas data</div>
		</div>
	</section>
</div>

<style>
	.landing {
		display: flex;
		flex-direction: column;
		gap: 20px;
		padding: 20px 0 28px;
	}

	.hero {
		text-align: left;
		padding: 20px;
		border: 1px solid #E2E8F0;
		border-radius: 16px;
		background: linear-gradient(180deg, #FFFFFF 0%, #F8FAFC 100%);
		box-shadow: 0 4px 16px rgba(15, 23, 42, 0.05);
	}

	.hero-kicker {
		display: inline-flex;
		align-items: center;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.8px;
		text-transform: uppercase;
		color: #E5484D;
		background: #FEF2F2;
		border: 1px solid #FECACA;
		border-radius: 999px;
		padding: 5px 10px;
		margin-bottom: 12px;
	}

	.title {
		font-size: clamp(40px, 13vw, 56px);
		font-weight: 900;
		line-height: 0.95;
		letter-spacing: -1.5px;
		color: #E5484D;
		margin: 0;
	}

	.tagline {
		font-size: 16px;
		font-weight: 600;
		line-height: 1.45;
		color: #0F172A;
		margin: 12px 0 0;
	}

	.trust {
		font-size: 13px;
		line-height: 1.5;
		color: #64748B;
		margin: 10px 0 0;
	}

	.actions {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.btn {
		display: flex;
		align-items: center;
		justify-content: center;
		text-align: center;
		text-decoration: none;
		padding: 14px 24px;
		font-size: 15px;
		font-weight: 700;
		border-radius: 12px;
		min-height: 52px;
		transition: opacity 0.15s, transform 0.1s, box-shadow 0.15s;
		letter-spacing: 0.1px;
	}

	.btn:active {
		transform: scale(0.97);
	}

	.btn-primary {
		color: #fff;
		background: linear-gradient(180deg, #EB5960 0%, #E5484D 100%);
		border: 1px solid rgba(173, 40, 45, 0.35);
		box-shadow: 0 6px 16px rgba(229, 72, 77, 0.22), 0 1px 3px rgba(0,0,0,0.08);
	}

	.btn-primary:hover {
		opacity: 0.94;
		box-shadow: 0 8px 20px rgba(229, 72, 77, 0.28), 0 1px 3px rgba(0,0,0,0.1);
	}

	.btn-secondary {
		color: #0F172A;
		background: #fff;
		border: 1px solid #E2E8F0;
		box-shadow: 0 1px 3px rgba(0,0,0,0.05);
	}

	.btn-secondary:hover {
		background: #F8FAFC;
	}

	.intro {
		padding: 18px;
		border: 1px solid #E2E8F0;
		border-radius: 16px;
		background: #fff;
		box-shadow: 0 1px 3px rgba(0,0,0,0.04);
	}

	.intro p {
		font-size: 14px;
		line-height: 1.65;
		color: #475569;
		margin: 0;
	}

	.feature-points {
		display: grid;
		grid-template-columns: 1fr;
		gap: 8px;
		margin-top: 14px;
	}

	.point {
		font-size: 13px;
		font-weight: 600;
		color: #334155;
		background: #F8FAFC;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		padding: 10px 12px;
	}

	@media (min-width: 640px) {
		.feature-points {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
	}
</style>
