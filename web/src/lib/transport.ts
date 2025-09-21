import { createConnectTransport } from '@connectrpc/connect-web';
import { Auther } from './auth';

export const transport = createConnectTransport({
	baseUrl: '/api',
	fetch: Auther.getInstance().fetch
});
