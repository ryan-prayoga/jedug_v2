<script lang="ts">
	let { message = 'Memuat...' }: { message?: string } = $props();
</script>

<div class="loading-state">
	<div class="spinner"></div>
	<p>{message}</p>
</div>

<style>
	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		color: #666;
	}
	.spinner {
		width: 32px;
		height: 32px;
		border: 3px solid #e2e2e2;
		border-top-color: #e53e3e;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	p {
		margin-top: 1rem;
		font-size: 0.9rem;
	}
</style>
