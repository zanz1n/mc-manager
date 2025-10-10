<script lang="ts">
	import { resolve } from '$app/paths';
	import { HashIcon, JoystickIcon, ServerIcon } from '@lucide/svelte';
	import { Distribution } from '@pb/distribution_pb';
	import { InstanceState, type Instance } from '@pb/instance_pb';

	type Props = { instance: Instance };

	let { instance }: Props = $props();

	function distroName(distro: Distribution): string {
		switch (distro) {
			case Distribution.PAPER:
				return 'Paper';
			case Distribution.VANILLA:
				return 'Vanilla';
		}
	}

	function stateName(state: InstanceState): string {
		switch (state) {
			case InstanceState.STATE_OFFLINE:
				return 'Offline';
			case InstanceState.STATE_RUNNING:
				return 'Running';
			case InstanceState.STATE_SHUTTING_DOWN:
				return 'Shutting Down';
			case InstanceState.STATE_STARTING:
				return 'Starting';
		}
	}

	function stateColor(state: InstanceState): string {
		switch (state) {
			case InstanceState.STATE_OFFLINE:
				return 'preset-filled-error-500';
			case InstanceState.STATE_RUNNING:
				return 'preset-filled-success-500';
			case InstanceState.STATE_SHUTTING_DOWN:
				return 'preset-filled-warning-500';
			case InstanceState.STATE_STARTING:
				return 'preset-filled-warning-500';
		}
	}
</script>

<a href={resolve(`/instance/${instance.id.toString()}`)} class="w-full">
	<section
		class="flex w-full flex-row items-center justify-between card border-1 border-surface-200-800 bg-surface-50-950 p-4 shadow"
	>
		<div>
			<h3 class="h4">{instance.name}</h3>
			<code class="hidden code sm:inline">ID: {instance.id}</code>
		</div>

		<div class="flex flex-row items-center gap-4">
			<div class="btn preset-outlined-secondary-500 not-lg:hidden">
				<HashIcon size={18} />
				<span>{distroName(instance.versionDistro) + ' ' + instance.version}</span>
			</div>
			<div class="btn preset-outlined-tertiary-500 not-lg:hidden">
				<JoystickIcon size={18} />
				<span>{instance.players}/{instance.limits?.maxPlayers} players</span>
			</div>
			<div class="btn {stateColor(instance.state)}">
				<ServerIcon size={18} />
				<span>{stateName(instance.state)}</span>
			</div>
		</div>
	</section>
</a>
