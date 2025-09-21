<script lang="ts">
	import { BugIcon, GithubIcon, SettingsIcon } from '@lucide/svelte';
</script>

<footer class="w-full border-t-[1px] border-surface-500/20 px-4 py-4 xl:px-0">
	<div class="container mx-auto flex flex-col items-center justify-between gap-4 xl:flex-row">
		<div class="flex items-center gap-4">
			<p class="text-xs opacity-60">
				Minecraft manager built by
				<a class="underline" href="https://github.com/zanz1n">Izan Rodrigues</a>
			</p>
		</div>
		<nav class="flex items-center">
			<a
				class="btn-icon hover:preset-tonal"
				title="Github"
				href="https://github.com/zanz1n/mc-manager"
			>
				<GithubIcon />
			</a>
			<a
				class="btn-icon hover:preset-tonal"
				title="Report issue"
				href="https://github.com/zanz1n/mc-manager/issues"
			>
				<BugIcon />
			</a>
			<a class="btn-icon hover:preset-tonal" title="Report issue" href="/config">
				<SettingsIcon />
			</a>
		</nav>
	</div>
</footer>
