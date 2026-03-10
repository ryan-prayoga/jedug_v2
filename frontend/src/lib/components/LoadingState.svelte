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
		padding: 48px 16px;
		color: #64748B;
	}
	.spinner {
		width: 32px;
		height: 32px;
		border: 3px solid #E2E8F0;
		border-top-color: #E5484D;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	p {
		margin-top: 16px;
		font-size: 14px;
	}
</style>
