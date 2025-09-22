import { createConnectTransport } from '@connectrpc/connect-web';
import { Auther } from './auth';

export const transport = createConnectTransport({
	baseUrl: '/api',
	fetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
		return Auther.getInstance().fetch(input, init);
	}
});
