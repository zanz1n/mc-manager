<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { Toaster } from '@skeletonlabs/skeleton-svelte';
	import { toaster } from '@lib/svelte-toaster';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<Toaster {toaster} />

{@render children?.()}
