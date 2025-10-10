<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Auther } from '@lib/auth';
	import { toaster } from '@lib/svelte-toaster';

	let error = $state<string>();

	let email = $state('');
	let username = $state('');
	let firstName = $state('');
	let lastName = $state('');
	let minecraftUser = $state('');
	let password = $state('');
</script>

<svelte:head>
	<title>Singup</title>
</svelte:head>

<h1 class="text-3xl">Signup</h1>

<form
	class="mx-auto flex w-full max-w-md flex-col items-center justify-center gap-4"
	onsubmit={async (e) => {
		e.preventDefault();

		try {
			await Auther.getInstance().signup({
				email,
				username,
				firstName,
				lastName,
				minecraftUser,
				password
			});
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
		<span class="label-text">Username</span>
		<input
			class="input"
			name="Username"
			type="text"
			placeholder="Enter Username"
			required
			bind:value={username}
		/>
	</label>

	<label class="label">
		<span class="label-text">First name</span>
		<input
			class="input"
			name="First name"
			type="text"
			placeholder="Enter first name"
			required
			bind:value={firstName}
		/>
	</label>

	<label class="label">
		<span class="label-text">Last name</span>
		<input
			class="input"
			name="Last name"
			type="text"
			placeholder="Enter last name"
			required
			bind:value={lastName}
		/>
	</label>

	<label class="label">
		<span class="label-text">Minecraft user</span>
		<input
			class="input"
			name="Minecraft user"
			type="text"
			placeholder="Enter minecraft user"
			bind:value={minecraftUser}
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

	<button type="submit" class="mt-4 btn w-full preset-filled">Signup</button>

	<p class="text-base">
		Or <a class="anchor" href={resolve('/auth/login')}>login to an existing account</a>
	</p>
</form>
