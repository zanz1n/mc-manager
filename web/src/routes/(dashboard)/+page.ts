import { InstanceService } from '@pb/instance_pb';
import { createClient } from '@connectrpc/connect';
import { transport } from '@lib/transport';
import type { PageLoad } from './$types';

const instanceSrv = createClient(InstanceService, transport);

export const load: PageLoad = async ({ parent }) => {
	const data = await parent();

	const res = await instanceSrv.getByUser({
		userId: data.user.id,
		pagination: {
			lastSeen: 0xffffffffffffffffn,
			limit: 15
		}
	});

	return {
		instances: res.instances,
		next: res.instances.length >= 15
	};
};
