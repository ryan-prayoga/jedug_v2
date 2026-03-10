<script lang="ts">
	let {
		message = 'Belum ada data',
		icon = '📭',
		ctaHref = '',
		ctaLabel = ''
	}: {
		message?: string;
		icon?: string;
		ctaHref?: string;
		ctaLabel?: string;
	} = $props();
</script>

<div class="empty-state">
	<span class="icon">{icon}</span>
	<p>{message}</p>
	{#if ctaHref && ctaLabel}
		<a href={ctaHref} class="empty-cta">{ctaLabel}</a>
	{/if}
</div>

<style>
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 16px;
		color: #64748B;
	}
	.icon {
		font-size: 48px;
		margin-bottom: 12px;
		line-height: 1;
	}
	p {
		font-size: 14px;
		text-align: center;
		max-width: 280px;
		line-height: 1.5;
	}
	.empty-cta {
		margin-top: 16px;
		display: inline-block;
		padding: 10px 20px;
		font-size: 14px;
		font-weight: 600;
		color: #fff;
		background: #E5484D;
		border-radius: 12px;
		text-decoration: none;
		transition: opacity 0.15s;
	}
	.empty-cta:hover {
		opacity: 0.88;
	}
</style>
