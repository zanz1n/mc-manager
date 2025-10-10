<script lang="ts">
	import type { LayoutProps } from './$types';
	import Header from '@components/Header.svelte';

	let { children, data }: LayoutProps = $props();
</script>

<div class="grid min-h-screen grid-rows-[auto_1fr]">
	<Header user={data.user} />

	{@render children?.()}
</div>
