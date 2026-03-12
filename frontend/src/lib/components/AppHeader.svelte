<nav class="app-header">
	<a href="/" class="logo-group">
		<span class="logo-text">JEDUG</span>
		<span class="logo-sub">Pantau Jalan Rusak</span>
	</a>
	<div class="nav-links">
		<a href="/lapor" class="nav-link">Lapor</a>
		<a href="/issues" class="nav-link">Peta</a>
		<a href="/stats" class="nav-link">Statistik</a>
	</div>
</nav>

<style>
	.app-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		background: #fff;
		border-bottom: 1px solid #E2E8F0;
		position: sticky;
		top: 0;
		z-index: 100;
	}
	.logo-group {
		display: flex;
		align-items: baseline;
		gap: 8px;
		text-decoration: none;
	}
	.logo-text {
		font-size: 1.25rem;
		font-weight: 800;
		color: #E5484D;
		letter-spacing: -0.5px;
		line-height: 1;
	}
	.logo-sub {
		font-size: 11px;
		color: #64748B;
		font-weight: 400;
		letter-spacing: 0;
		display: none;
	}
	@media (min-width: 360px) {
		.logo-sub {
			display: inline;
		}
	}
	.nav-links {
		display: flex;
		gap: 4px;
	}
	.nav-link {
		font-size: 14px;
		color: #64748B;
		text-decoration: none;
		font-weight: 500;
		padding: 6px 12px;
		border-radius: 8px;
		transition: background 0.15s, color 0.15s;
	}
	.nav-link:hover {
		color: #E5484D;
		background: #FEF2F2;
	}

	@media (max-width: 420px) {
		.nav-links {
			gap: 2px;
		}
		.nav-link {
			padding: 6px 8px;
			font-size: 13px;
		}
	}
</style>
