<script lang="ts">
	import { createClient } from '@connectrpc/connect';
	import { transport } from '@lib/transport';
	import { InstanceService, type PartialInstance } from '@pb/instance_pb';
	import type { PageProps } from './$types';
	import { ProgressRing } from '@skeletonlabs/skeleton-svelte';
	import { toaster } from '@lib/svelte-toaster';
	import InfiniteScroll from '@components/InfiniteScroll.svelte';

	const instanceSrv = createClient(InstanceService, transport);
	const batchSize = 10;

	let { data }: PageProps = $props();

	let instances = $state<PartialInstance[]>(data.instances);
	let isFinished = $state(!data.next);

	const load = async () => {
		if (isFinished || instances.length == 0) {
			return;
		}

		const lastInstance = instances[instances.length - 1];

		const res = await instanceSrv.getByUser({
			userId: data.user.id,
			pagination: {
				lastSeen: lastInstance.id,
				limit: batchSize
			}
		});

		instances.push(...res.instances);
		isFinished = res.instances.length < batchSize;
	};
</script>

<div>
	{#each instances as instance (instance.id.toString())}
		<p>{instance.name}</p>
	{/each}

	{#if !isFinished}
		<ProgressRing
			value={null}
			size="size-14"
			meterStroke="stroke-tertiary-600-400"
			trackStroke="stroke-tertiary-50-950"
		/>
	{/if}

	<InfiniteScroll
		hasMore={!isFinished}
		horizontal={false}
		threshold={100}
		loadMore={() => {
			load().catch((e) => {
				if (e instanceof Error) {
					toaster.error({ title: e.message });
				} else {
					toaster.error({ title: 'Unknown error' });
				}
			});
		}}
	/>
</div>
