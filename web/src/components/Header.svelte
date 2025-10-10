<script lang="ts">
	import { resolve } from '$app/paths';
	import type { User } from '@pb/user_pb';
	import LightSwitch from './LightSwitch.svelte';
	import { Avatar, Popover } from '@skeletonlabs/skeleton-svelte';
	import {
		BugIcon,
		GithubIcon,
		LayoutDashboardIcon,
		LogOutIcon,
		PencilIcon,
		SettingsIcon,
		UserIcon,
		XIcon
	} from '@lucide/svelte';
	import { Auther } from '@lib/auth';
	import { goto } from '$app/navigation';

	type Props = { user: User };

	let { user }: Props = $props();

	let fullName = user.firstName + ' ' + user.lastName;

	let openState = $state(false);

	function popoverClose() {
		openState = false;
	}

	function signout() {
		Auther.getInstance().logout();
		goto(resolve('/auth/login'));
	}
</script>

<header
	class="sticky top-0 z-50 flex h-[70px] w-full items-center border border-surface-100-900/30 bg-surface-50-950/75 backdrop-blur-lg"
>
	<div
		class="container mx-auto grid max-w-screen-2xl grid-cols-[auto_1fr_auto] items-center gap-4 px-4 xl:grid-cols-[1fr_auto_1fr] xl:px-10"
	>
		<div class="flex items-stretch justify-start gap-4">
			<a href={resolve('/')}>
				<h4 class="h4">MC Manager</h4>
			</a>
		</div>
		<div class="flex items-center gap-2"></div>
		<div class="flex items-stretch justify-end gap-4">
			<LightSwitch />

			<Popover
				zIndex="100"
				open={openState}
				onOpenChange={(e) => (openState = e.open)}
				positioning={{ placement: 'bottom' }}
				contentBase="card border border-surface-200-800 bg-surface-50-950 p-4 space-y-4 max-w-[320px]"
			>
				{#snippet trigger()}
					<Avatar
						classes="border-2 border-surface-300-600-token hover:!border-primary-500 w-11 h-11"
						name={fullName}
						initials={[0, 1]}
					/>
				{/snippet}

				{#snippet content()}
					<div class="flex justify-between gap-4">
						<div class="flex flex-row items-center gap-2">
							<Avatar
								classes="border-2 border-surface-300-600-token w-12 h-12"
								name={fullName}
								initials={[0, 1]}
							/>
							<div>
								<p class="mb-0 text-lg font-bold">{fullName}</p>
								<p class="mt-0 text-base opacity-70">{user.username}</p>
							</div>
						</div>
						<button
							class="btn-icon hover:preset-tonal"
							onclick={popoverClose}
							title="Close"
							aria-label="Close"
						>
							<XIcon />
						</button>
					</div>

					<button class="btn w-full justify-start preset-filled px-2" onclick={signout}>
						<LogOutIcon size={18} />
						<span>Sign Out</span>
					</button>

					<hr class="hr" />

					<div class="flex w-full flex-col gap-2">
						{#if user.admin}
							<a
								class="btn w-full justify-start px-2 hover:preset-tonal"
								title="Dashboard"
								href="/admin"
							>
								<LayoutDashboardIcon size={18} />
								<span>Admin Dashboard</span>
							</a>
						{/if}
						<a
							class="btn w-full justify-start px-2 hover:preset-tonal"
							title="Account"
							href="/account"
						>
							<UserIcon size={18} />
							<span>Acount</span>
						</a>
						<a
							class="btn w-full justify-start px-2 hover:preset-tonal"
							title="Config"
							href="/config"
						>
							<SettingsIcon size={18} />
							<span>Configuration</span>
						</a>
						<a
							class="btn w-full justify-start px-2 hover:preset-tonal"
							title="appearance"
							href="/config/appearance"
						>
							<PencilIcon size={18} />
							<span>Appearance</span>
						</a>
					</div>

					<hr class="hr" />

					<div class="flex w-full flex-col gap-2">
						<a
							class="btn w-full justify-start px-2 hover:preset-tonal"
							title="GitHub"
							href="https://github.com/zanz1n/mc-manager"
							target="_blank"
						>
							<GithubIcon size={18} />
							<span>GitHub</span>
						</a>
						<a
							class="btn w-full justify-start px-2 hover:preset-tonal"
							title="Report issue"
							href="https://github.com/zanz1n/mc-manager/issues"
							target="_blank"
						>
							<BugIcon size={18} />
							<span>Report Bugs</span>
						</a>
					</div>
				{/snippet}
			</Popover>
		</div>
	</div>
</header>
