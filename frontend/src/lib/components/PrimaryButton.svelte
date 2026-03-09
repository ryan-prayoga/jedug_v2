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
		font-size: 1.1rem;
		font-weight: 600;
		color: #fff;
		background: #e53e3e;
		border: none;
		border-radius: 12px;
		cursor: pointer;
		transition: opacity 0.15s;
		text-align: center;
	}
	.primary-btn:hover:not(:disabled) {
		opacity: 0.9;
	}
	.primary-btn:active:not(:disabled) {
		transform: scale(0.98);
	}
	.primary-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
