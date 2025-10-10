import { AuthService } from '@pb/auth_pb';
import { createClient, type Client, type Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import type { User } from '@pb/user_pb';
import { AppError } from './error';
import { goto } from '$app/navigation';
import { resolve } from '$app/paths';

export class UnauthorizedError extends AppError {
	constructor() {
		super(400, 'Unauthorized');
	}

	action(): void {
		goto(resolve('/auth/login'));
	}
}

export type SignupData = {
	email: string;
	username: string;
	firstName: string;
	lastName: string;
	minecraftUser: string;
	password: string;
};

export class Auther {
	private static __instance?: Auther;

	static getInstance(): Auther {
		if (!this.__instance) {
			this.__instance = new Auther();
		}
		return this.__instance;
	}

	private token?: { jwt: string; refresh: string };
	private client: Client<typeof AuthService>;

	constructor(client?: Client<typeof AuthService>) {
		const refresh = localStorage.getItem('refresh-token');
		if (refresh) {
			this.token = { refresh: refresh, jwt: 'replace' };
		}

		if (!client) {
			const transport = createConnectTransport({ baseUrl: '/api' });
			client = createClient(AuthService, transport);
		}
		this.client = client;
	}

	async getUser(): Promise<User> {
		if (!this.token) {
			throw new UnauthorizedError();
		}

		const onHeader = (headers: Headers) => {
			const setTokenHead = headers.get('Set-Token');
			if (setTokenHead && this.token) {
				this.token.jwt = setTokenHead;
			}
		};

		const res = await this.client.getSelf(
			{},
			{
				headers: {
					Authorization: 'Bearer ' + this.token.jwt,
					'Auth-Refresh-Token': this.token.refresh
				},
				onHeader
			}
		);

		return res;
	}

	async login(email: string, password: string) {
		const res = await this.client.login({ email, password });
		this.token = { jwt: res.token, refresh: res.refreshToken };

		localStorage.setItem('refresh-token', res.refreshToken);
	}

	async signup(data: SignupData) {
		const res = await this.client.signup(data);
		this.token = { jwt: res.token, refresh: res.refreshToken };

		localStorage.setItem('refresh-token', res.refreshToken);
	}

	logout() {
		this.token = undefined;
	}

	readonly interceptor: Interceptor = (fn) => {
		return async (req) => {
			if (!this.token) {
				throw new UnauthorizedError();
			}

			req.header.set('Authorization', 'Bearer ' + this.token.jwt);
			req.header.set('Auth-Refresh-Token', this.token.refresh);

			const res = await fn(req);
			const newToken = res.header.get('Set-Token');
			if (newToken) {
				this.token.jwt = newToken;
			}

			return res;
		};
	};
}
