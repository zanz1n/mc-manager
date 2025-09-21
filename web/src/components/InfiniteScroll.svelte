<script lang="ts">
	import { onDestroy } from 'svelte';

	type Props = {
		threshold: number;
		horizontal: boolean;
		hasMore: boolean;
		elementScroll?: Element;
		loadMore: () => void;
	};

	let {
		threshold = 0,
		horizontal = false,
		hasMore = true,
		elementScroll,
		loadMore
	}: Props = $props();

	let isLoadMore = $state(false);
	let component = $state<Element>();

	$effect(() => {
		if (component || elementScroll) {
			const element = elementScroll ? elementScroll : component?.parentNode;

			if (element) {
				element.addEventListener('scroll', onScroll);
				element.addEventListener('resize', onScroll);
			}
		}
	});

	const onScroll = (e: Event) => {
		if (!e.target) {
			return;
		}
		const element = e.target as Element;

		let offset: number;
		if (horizontal) {
			offset = element.scrollWidth - element.clientWidth - element.scrollLeft;
		} else {
			offset = element.scrollHeight - element.clientHeight - element.scrollTop;
		}

		if (offset <= threshold) {
			if (!isLoadMore && hasMore) {
				loadMore();
			}
			isLoadMore = true;
		} else {
			isLoadMore = false;
		}
	};

	onDestroy(() => {
		if (component || elementScroll) {
			const element = elementScroll ? elementScroll : component?.parentNode;

			if (element) {
				element.removeEventListener('scroll', null);
				element.removeEventListener('resize', null);
			}
		}
	});
</script>

<div bind:this={component} style="width:0px"></div>
