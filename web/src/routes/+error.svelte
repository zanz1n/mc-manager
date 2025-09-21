<script lang="ts">
	import { page } from '$app/state';
</script>

<div class="flex h-full w-full items-center justify-center">
	<div class="text-center">
		<h1 class="text-8xl">{page.status}</h1>
		<h2 class="text-2xl">{page.error?.message}</h2>
	</div>
</div>
