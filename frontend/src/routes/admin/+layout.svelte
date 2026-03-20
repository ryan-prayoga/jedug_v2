<script lang="ts">
	import { goto, afterNavigate } from '$app/navigation';
	import { page } from '$app/stores';
	import { adminLogout, adminMe } from '$lib/api/admin';
	import { LogoutIcon, ShieldCheckIcon, UserIcon } from '$lib/icons';

	let { children } = $props();

	let ready = $state(false);
	let username = $state('');

	const pathname = $derived($page.url.pathname);
	const isLoginPage = $derived(pathname === '/admin/login');

	afterNavigate(({ from }) => {
		void syncAdminSession(from?.url?.pathname);
	});

	async function syncAdminSession(fromPathname?: string) {
		if (isLoginPage) {
			ready = true;
			return;
		}

		if (ready && fromPathname !== '/admin/login') return;

		ready = false;

		try {
			const res = await adminMe();
			username = res.data?.username ?? '';
			ready = true;
		} catch {
			goto('/admin/login');
		}
	}

	async function handleLogout() {
		try {
			await adminLogout();
		} catch {
			// Session may already be expired; navigation still needs to happen.
		}
		goto('/admin/login');
	}
</script>

<svelte:head>
	<title>JEDUG Admin</title>
</svelte:head>

{#if isLoginPage}
	{@render children()}
{:else if ready}
	<div class="admin-shell-bg">
		<header class="border-b border-white/80 bg-white/85 backdrop-blur-xl">
			<div class="admin-frame flex flex-col gap-4 py-4">
				<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
					<div class="flex items-center gap-3">
						<div class="flex size-12 items-center justify-center rounded-[20px] bg-brand-50 text-brand-600">
							<ShieldCheckIcon class="size-6" />
						</div>
						<div>
							<p class="text-[11px] font-bold uppercase tracking-[0.18em] text-brand-600">Moderation workspace</p>
							<a href="/admin" class="text-xl font-[800] tracking-[-0.04em] text-slate-950">JEDUG Admin</a>
						</div>
					</div>

					<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
						<nav class="flex items-center gap-2 rounded-[22px] border border-white/70 bg-white/75 p-1.5 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
							<a
								href="/admin/issues"
								class:bg-brand-500={pathname.startsWith('/admin/issues')}
								class:text-white={pathname.startsWith('/admin/issues')}
								class="rounded-[16px] px-4 py-2 text-sm font-semibold text-slate-600 transition hover:bg-slate-100"
							>
								Issues
							</a>
						</nav>

						<div class="flex items-center gap-3 rounded-[24px] border border-white/70 bg-white/75 px-4 py-2.5 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
							<div class="flex size-10 items-center justify-center rounded-2xl bg-slate-100 text-slate-700">
								<UserIcon class="size-5" />
							</div>
							<div class="min-w-0">
								<p class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">Signed in</p>
								<p class="truncate text-sm font-semibold text-slate-900">{username}</p>
							</div>
							<button class="btn-secondary min-h-10 px-4 py-2" onclick={handleLogout}>
								<LogoutIcon class="size-[18px]" />
								Logout
							</button>
						</div>
					</div>
				</div>
			</div>
		</header>
		<main class="admin-frame py-6">
			{@render children()}
		</main>
	</div>
{:else}
	<div class="admin-shell-bg">
		<div class="admin-frame flex min-h-dvh items-center justify-center">
			<div class="state-panel max-w-md">
				<div class="mx-auto size-11 animate-spin rounded-full border-[3px] border-slate-200 border-t-brand-500"></div>
				<p class="mt-4 text-sm font-semibold text-slate-700">Memuat sesi admin...</p>
			</div>
		</div>
	</div>
{/if}
