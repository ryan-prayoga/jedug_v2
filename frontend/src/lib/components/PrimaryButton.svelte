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
		font-weight: 600;
		color: #fff;
		background: #E5484D;
		border: none;
		border-radius: 12px;
		cursor: pointer;
		transition: opacity 0.15s, transform 0.1s;
		text-align: center;
		min-height: 48px;
	}
	.primary-btn:hover:not(:disabled) {
		opacity: 0.88;
	}
	.primary-btn:active:not(:disabled) {
		transform: scale(0.97);
	}
	.primary-btn:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}
</style>
