<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		goto('/admin/issues', { replaceState: true });
	});
</script>

<p>Mengarahkan ke daftar issue...</p>
