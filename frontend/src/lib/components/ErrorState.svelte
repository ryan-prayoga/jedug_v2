<script lang="ts">
	let { message = 'Terjadi kesalahan', onretry }: { message?: string; onretry?: () => void } = $props();
</script>

<div class="error-state">
	<span class="icon">⚠️</span>
	<p>{message}</p>
	{#if onretry}
		<button class="retry-btn" onclick={onretry}>Coba Lagi</button>
	{/if}
</div>

<style>
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		color: #c53030;
	}
	.icon {
		font-size: 2.5rem;
		margin-bottom: 0.75rem;
	}
	p {
		font-size: 0.95rem;
		text-align: center;
		margin-bottom: 1rem;
	}
	.retry-btn {
		padding: 8px 20px;
		font-size: 0.9rem;
		color: #c53030;
		background: #fff5f5;
		border: 1px solid #feb2b2;
		border-radius: 8px;
		cursor: pointer;
	}
	.retry-btn:hover {
		background: #fed7d7;
	}
</style>
