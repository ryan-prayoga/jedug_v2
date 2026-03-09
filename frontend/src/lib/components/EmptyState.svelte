<script lang="ts">
	let { message = 'Belum ada data', icon = '📭' }: { message?: string; icon?: string } = $props();
</script>

<div class="empty-state">
	<span class="icon">{icon}</span>
	<p>{message}</p>
</div>

<style>
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		color: #888;
	}
	.icon {
		font-size: 2.5rem;
		margin-bottom: 0.75rem;
	}
	p {
		font-size: 0.95rem;
		text-align: center;
	}
</style>
