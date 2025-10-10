<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Auther } from '@lib/auth';
	import { toaster } from '@lib/svelte-toaster';

	let error = $state<string>();

	let email = $state('');
	let password = $state('');
</script>

<svelte:head>
	<title>Login</title>
</svelte:head>

<h1 class="text-3xl">Login</h1>

<form
	class="mx-auto flex w-full max-w-md flex-col items-center justify-center gap-4"
	onsubmit={async (e) => {
		e.preventDefault();

		try {
			await Auther.getInstance().login(email, password);
			toaster.success({ title: 'Logged in' });
			await goto(resolve('/'));
		} catch (err) {
			if (err instanceof Error) {
				error = err.message;
			}
		}
	}}
>
	<p class="text-center text-base text-error-500">{error}</p>

	<label class="label">
		<span class="label-text">Email</span>
		<input
			class="input"
			name="Email"
			type="email"
			placeholder="Enter Email"
			required
			bind:value={email}
		/>
	</label>

	<label class="label">
		<span class="label-text">Password</span>
		<input
			class="input"
			name="Password"
			type="password"
			placeholder="Enter Password"
			required
			bind:value={password}
		/>
	</label>

	<button type="submit" class="mt-4 btn w-full preset-filled">Login</button>

	<p class="text-base">
		Or <a class="anchor" href={resolve('/auth/signup')}>create an account</a>
	</p>
</form>
