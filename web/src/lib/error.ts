import type { HttpError } from '@sveltejs/kit';

export class AppError extends Error implements HttpError {
	body: App.Error;

	constructor(
		public status: number,
		message: string
	) {
		super(message);
		this.body = { message };
	}

	action() {}
}
