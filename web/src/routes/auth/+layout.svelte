<script lang="ts">
	import Footer from '@components/Footer.svelte';

	let { children } = $props();
</script>

<div class="flex h-screen flex-col items-center justify-between sm:gap-4">
	<div></div>

	<div
		class="bg-noise flex w-full flex-col items-center justify-center gap-2 card p-8 sm:w-[450px] sm:border-[1px] sm:border-surface-200-800 sm:bg-surface-50-950"
	>
		{@render children?.()}
	</div>

	<Footer />
</div>
