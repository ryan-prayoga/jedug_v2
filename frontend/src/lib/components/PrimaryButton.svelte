<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';

	let { children, ...rest }: HTMLButtonAttributes & { children: Snippet } = $props();
</script>

<button class="primary-btn" {...rest}>
	{@render children()}
</button>

<style>
	.primary-btn {
		display: block;
		width: 100%;
		padding: 14px 24px;
		font-size: 16px;
		font-weight: 700;
		color: #fff;
		background: linear-gradient(180deg, #EB5960 0%, #E5484D 100%);
		border: 1px solid rgba(173, 40, 45, 0.35);
		border-radius: 12px;
		cursor: pointer;
		transition: opacity 0.15s, transform 0.1s, box-shadow 0.15s;
		text-align: center;
		min-height: 52px;
		box-shadow: 0 6px 16px rgba(229, 72, 77, 0.22), 0 1px 3px rgba(0,0,0,0.08);
		letter-spacing: 0.1px;
	}
	.primary-btn:hover:not(:disabled) {
		opacity: 0.94;
		box-shadow: 0 8px 20px rgba(229, 72, 77, 0.28), 0 1px 3px rgba(0,0,0,0.1);
	}
	.primary-btn:active:not(:disabled) {
		transform: scale(0.97);
		box-shadow: 0 4px 12px rgba(229, 72, 77, 0.2);
	}
	.primary-btn:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}
</style>
