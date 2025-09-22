import { AuthService } from '@pb/auth_pb';
import { createClient, type Client } from '@connectrpc/connect';
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
			const setTokenHead = headers.get('set-token');
			if (setTokenHead && this.token) {
				this.token.jwt = setTokenHead;
			}
		};

		const res = await this.client.getSelf(
			{},
			{
				headers: {
					authorization: 'Bearer ' + this.token.jwt,
					'auth-refresh-token': this.token.refresh
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

	async fetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
		if (!this.token) {
			throw new UnauthorizedError();
		}

		if (isHeaders(init?.headers)) {
			init.headers.set('authorization', 'Bearer ' + this.token.jwt);
			init.headers.set('auth-refresh-token', this.token.refresh);
		}

		const res = await fetch(input, init);

		const newToken = res.headers.get('set-token');
		if (newToken) {
			this.token.jwt = newToken;
		}

		return res;
	}
}

function isHeaders(obj: unknown): obj is Headers {
	return (
		!!obj &&
		typeof obj == 'object' &&
		'append' in obj &&
		typeof obj['append'] == 'function' &&
		'set' in obj &&
		typeof obj['set'] == 'function'
	);
}
