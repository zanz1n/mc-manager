<script lang="ts">
	import Footer from '@components/Footer.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<div class="grid h-screen grid-rows-[auto_1fr_auto]">
	<header
		class="sticky top-0 z-50 flex h-[70px] w-full items-center border border-surface-100-900/30 bg-surface-50-950/75 backdrop-blur-lg"
	>
		<div
			class="container mx-auto grid max-w-screen-2xl grid-cols-[auto_1fr_auto] items-center gap-4 px-4 xl:grid-cols-[1fr_auto_1fr] xl:px-10"
		>
			<div class="flex items-stretch justify-start gap-4"></div>
			<div class="flex items-center gap-2"></div>
			<div class="flex items-stretch justify-end gap-2"></div>
		</div>
	</header>

	<div class="space-y-4 p-4">
		{@render children?.()}
	</div>

	<Footer />
</div>
