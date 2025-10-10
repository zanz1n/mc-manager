<script lang="ts">
	import { MoonIcon, SunIcon } from '@lucide/svelte';
	import { Switch } from '@skeletonlabs/skeleton-svelte';

	let checked = $state(false);

	function setMode(mode: 'light' | 'dark') {
		const other = mode == 'light' ? 'dark' : 'light';

		document.documentElement.classList.remove(other);
		if (!document.documentElement.classList.contains(mode)) {
			document.documentElement.classList.add(mode);
		}

		localStorage.setItem('theme-mode', mode);
	}

	$effect(() => {
		const storedMode = localStorage.getItem('theme-mode');
		let isDark = true;
		if (!storedMode) {
			isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		} else {
			isDark = storedMode == 'dark';
		}
		checked = !isDark;
	});

	function onCheckedChange(event: { checked: boolean }) {
		setMode(event.checked ? 'light' : 'dark');
		checked = event.checked;
	}
</script>

<Switch {checked} {onCheckedChange} controlActive="bg-surface-200">
	{#snippet inactiveChild()}<MoonIcon size="14" />{/snippet}
	{#snippet activeChild()}<SunIcon size="14" />{/snippet}
</Switch>
