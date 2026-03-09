<nav class="app-header">
	<a href="/" class="logo">JEDUG</a>
	<div class="nav-links">
		<a href="/lapor">Lapor</a>
		<a href="/issues">Daftar</a>
	</div>
</nav>

<style>
	.app-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		background: #fff;
		border-bottom: 1px solid #e2e8f0;
		position: sticky;
		top: 0;
		z-index: 100;
	}
	.logo {
		font-size: 1.3rem;
		font-weight: 800;
		color: #e53e3e;
		text-decoration: none;
		letter-spacing: -0.5px;
	}
	.nav-links {
		display: flex;
		gap: 16px;
	}
	.nav-links a {
		font-size: 0.9rem;
		color: #4a5568;
		text-decoration: none;
		font-weight: 500;
	}
	.nav-links a:hover {
		color: #e53e3e;
	}
</style>
