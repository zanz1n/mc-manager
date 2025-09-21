import { Auther } from '@lib/auth';
import type { LayoutLoad } from './$types';
import { goto } from '$app/navigation';
import type { User } from '@pb/user_pb';
import { resolve } from '$app/paths';

export const load: LayoutLoad = async () => {
	try {
		const user = await Auther.getInstance().getUser();
		return { user };
	} catch (_error) {
		await goto(resolve('/auth/login'));
		// component unreachable due to redirection
		// just for types
		return null as unknown as { user: User };
	}
};
