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
		padding: 48px 16px;
		color: #DC2626;
	}
	.icon {
		font-size: 48px;
		margin-bottom: 12px;
		line-height: 1;
	}
	p {
		font-size: 14px;
		text-align: center;
		margin-bottom: 16px;
		max-width: 280px;
		line-height: 1.5;
	}
	.retry-btn {
		padding: 10px 20px;
		font-size: 14px;
		font-weight: 600;
		color: #DC2626;
		background: #FEF2F2;
		border: 1px solid #FECACA;
		border-radius: 12px;
		cursor: pointer;
		transition: background 0.15s;
	}
	.retry-btn:hover {
		background: #FEE2E2;
	}
</style>
