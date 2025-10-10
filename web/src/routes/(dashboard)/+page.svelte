<script lang="ts">
	import type { InstanceGetManyResponse, Instance } from '@pb/instance_pb';
	import type { PageProps } from './$types';
	import { ProgressRing, Switch } from '@skeletonlabs/skeleton-svelte';
	import { toaster } from '@lib/svelte-toaster';
	import { onMount } from 'svelte';
	import { GlobeIcon, SearchIcon, UserIcon } from '@lucide/svelte';
	import InstanceComponent from './Instance.svelte';
	import Footer from '@components/Footer.svelte';
	import { instanceService } from '@lib/transport';

	const batchSize = 10;

	let { data }: PageProps = $props();

	let instances = $state<Instance[]>([]);
	let more = $state(true);
	let showAll = $state(false);

	async function load() {
		if (!more) return;
		// await new Promise((res) => setTimeout(res, 10 * 1000));

		let lastSeen: bigint;
		if (instances.length == 0) {
			// max int64
			lastSeen = 9223372036854775807n;
		} else {
			lastSeen = instances[instances.length - 1].id;
		}

		let promiseRes: Promise<InstanceGetManyResponse>;
		if (showAll) {
			promiseRes = instanceService.getMany({ lastSeen, limit: batchSize });
		} else {
			promiseRes = instanceService.getByUser({
				userId: data.user.id,
				pagination: {
					lastSeen,
					limit: batchSize
				}
			});
		}

		const [res, _] = await Promise.all([
			promiseRes,
			// fake delay to prevent api overload on rapid scrolling
			new Promise((res) => setTimeout(res, 500))
		]);

		instances.push(...res.instances);
		more = res.instances.length >= batchSize;
	}

	function onCheckedChange(event: { checked: boolean }) {
		instances = [];
		showAll = event.checked;
		more = true;
	}

	// Detect scroll end for loading more instances
	onMount(() => {
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting && more) {
					load().catch((e) => {
						if (e instanceof Error) {
							toaster.error({ title: e.message });
						} else {
							toaster.error({ title: 'Unknown error' });
						}
					});
				}
			},
			{ threshold: 0.5 }
		);

		const sentinel = document.getElementById('infinite-scroll-sentinel');
		if (sentinel) {
			observer.observe(sentinel);
		}

		return () => observer.disconnect();
	});
</script>

<svelte:head>
	<title>Instances</title>
</svelte:head>

<div class="container mx-auto my-8 scroll-mt-[70px] space-y-4 px-3 sm:px-4">
	<div class="flex flex-col items-center gap-4 lg:flex-row lg:justify-between">
		<div class="flex w-full flex-row items-center justify-between gap-8 lg:w-fit">
			<h1 class="h1">Instances</h1>
			<div class="lg:hidden">
				<Switch checked={showAll} {onCheckedChange} classes="lg:min-w-32">
					Show all
					{#snippet inactiveChild()}<UserIcon size="14" />{/snippet}
					{#snippet activeChild()}<GlobeIcon size="14" />{/snippet}
				</Switch>
			</div>
		</div>
		<div class="input-group w-full grid-cols-[auto_1fr_auto] lg:max-w-2xl">
			<div class="ig-cell preset-tonal">
				<SearchIcon size="16" />
			</div>
			<input class="ig-input" type="search" placeholder="Search..." />
			<button class="ig-btn preset-filled">Search</button>
		</div>
		<div class="not-lg:hidden"></div>
		<div class="not-lg:hidden">
			<Switch checked={showAll} {onCheckedChange} classes="lg:min-w-32">
				Show all
				{#snippet inactiveChild()}<UserIcon size="14" />{/snippet}
				{#snippet activeChild()}<GlobeIcon size="14" />{/snippet}
			</Switch>
		</div>
	</div>

	<main class="mt-8 flex flex-col items-center gap-2 sm:gap-4">
		{#each instances as instance (instance.id.toString())}
			<InstanceComponent {instance} />
		{/each}

		{#if more}
			<ProgressRing
				value={null}
				size="size-14"
				meterStroke="stroke-tertiary-600-400"
				trackStroke="stroke-tertiary-50-950"
			/>
		{/if}
	</main>

	<div id="infinite-scroll-sentinel" style="height: 1px;"></div>
</div>

<Footer />
