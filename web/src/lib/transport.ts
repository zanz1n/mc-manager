import { createConnectTransport } from '@connectrpc/connect-web';
import { Auther } from './auth';
import { createClient } from '@connectrpc/connect';
import { InstanceService } from '@pb/instance_pb';

export const transport = createConnectTransport({
	baseUrl: '/api',
	interceptors: [Auther.getInstance().interceptor]
});

export const instanceService = createClient(InstanceService, transport);
