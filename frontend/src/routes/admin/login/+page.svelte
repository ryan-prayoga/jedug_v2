<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { adminLogin, adminMe } from '$lib/api/admin';
	import {
		CheckCircleIcon,
		EyeClosedIcon,
		EyeIcon,
		LoginIcon,
		ShieldCheckIcon,
		UserIcon
	} from '$lib/icons';
	import {
		clearRememberedAdminUsername,
		getRememberedAdminUsername,
		setRememberedAdminUsername
	} from '$lib/utils/storage';

	let username = $state('');
	let password = $state('');
	let rememberMe = $state(false);
	let showPassword = $state(false);
	let error = $state('');
	let loading = $state(false);

	onMount(() => {
		const remembered = getRememberedAdminUsername();
		if (remembered) {
			username = remembered;
			rememberMe = true;
		}
		void redirectIfSessionExists();
	});

	async function redirectIfSessionExists() {
		try {
			await adminMe();
			goto('/admin');
		} catch {
			// No active admin session; stay on login page.
		}
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			await adminLogin(username, password);
			if (rememberMe) {
				setRememberedAdminUsername(username);
			} else {
				clearRememberedAdminUsername();
			}
			goto('/admin');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login gagal';
		} finally {
			loading = false;
		}
	}
</script>

<div class="admin-shell-bg">
	<div class="admin-frame flex min-h-dvh items-center justify-center py-10">
		<div class="grid w-full max-w-[1080px] overflow-hidden rounded-[4px] border border-hairline bg-surface lg:grid-cols-[minmax(0,0.95fr)_minmax(380px,0.85fr)]">
			<section class="hidden border-r border-hairline bg-sunken p-8 lg:flex lg:flex-col lg:justify-between">
				<div class="space-y-6">
					<span class="section-kicker">Admin workspace</span>
					<div class="space-y-3">
						<h1 class="section-title text-[clamp(2.2rem,4vw,3.2rem)]">
							Moderasi issue publik dengan UI yang lebih rapi dan operasional-friendly.
						</h1>
						<p class="section-copy max-w-[46ch]">
							Area admin dipoles untuk membantu pemindaian issue, tindakan moderasi, dan review bukti tanpa mengubah kontrak backend yang sudah berjalan.
						</p>
					</div>
				</div>

				<div class="space-y-3">
					<div class="jedug-card-soft flex gap-3 px-4 py-4">
						<div class="flex size-10 shrink-0 items-center justify-center rounded-[8px] bg-sunken text-muted">
							<ShieldCheckIcon class="size-5" />
						</div>
						<div>
							<p class="text-sm font-bold text-ink">Cookie HttpOnly tetap dipakai</p>
							<p class="mt-1 text-xs leading-5 text-muted">Sesi admin tetap mengikuti kebijakan server, bukan disimpan di localStorage.</p>
						</div>
					</div>
					<div class="jedug-card-soft flex gap-3 px-4 py-4">
						<div class="flex size-10 shrink-0 items-center justify-center rounded-[8px] bg-sunken text-muted">
							<UserIcon class="size-5" />
						</div>
						<div>
							<p class="text-sm font-bold text-ink">Ingat saya tetap aman</p>
							<p class="mt-1 text-xs leading-5 text-muted">Yang disimpan hanya username untuk memudahkan login berikutnya, bukan password.</p>
						</div>
					</div>
				</div>
			</section>

			<section class="p-5 sm:p-8">
				<div class="mx-auto flex w-full max-w-[420px] flex-col gap-6 py-4">
					<div class="space-y-4">
						<div class="flex size-14 items-center justify-center rounded-[4px] bg-sunken text-muted">
							<LoginIcon class="size-7" />
						</div>
						<div class="space-y-2">
							<p class="text-[11px] font-bold uppercase tracking-[0.18em] text-brand">JEDUG Admin</p>
							<h2 class="font-serif text-3xl font-semibold tracking-[-0.02em] text-ink">Masuk untuk moderasi</h2>
							<p class="text-sm leading-6 text-muted">
								Gunakan kredensial admin yang valid. Sesi tetap diamankan oleh cookie server-side dengan TTL backend saat ini.
							</p>
						</div>
					</div>

					<form class="space-y-4" onsubmit={handleSubmit}>
						{#if error}
							<div class="error-panel">{error}</div>
						{/if}

						<label class="input-shell">
							<span class="input-label">Username</span>
							<div class="relative">
								<span class="pointer-events-none absolute inset-y-0 left-4 flex items-center text-subtle">
									<UserIcon class="size-5" />
								</span>
								<input
									type="text"
									class="input-field w-full pl-12"
									bind:value={username}
									required
									autocomplete="username"
									disabled={loading}
								/>
							</div>
						</label>

						<label class="input-shell">
							<span class="input-label">Password</span>
							<div class="relative">
								<span class="pointer-events-none absolute inset-y-0 left-4 flex items-center text-subtle">
									<ShieldCheckIcon class="size-5" />
								</span>
								<input
									type={showPassword ? 'text' : 'password'}
									class="input-field w-full pl-12 pr-12"
									bind:value={password}
									required
									autocomplete="current-password"
									disabled={loading}
								/>
								<button
									type="button"
									class="absolute inset-y-0 right-3 my-auto flex size-9 items-center justify-center rounded-[8px] text-muted transition hover:bg-sunken hover:text-ink"
									aria-label={showPassword ? 'Sembunyikan password' : 'Tampilkan password'}
									aria-pressed={showPassword}
									onclick={() => (showPassword = !showPassword)}
								>
									{#if showPassword}
										<EyeClosedIcon class="size-5" />
									{:else}
										<EyeIcon class="size-5" />
									{/if}
								</button>
							</div>
						</label>

						<label class="flex items-start gap-3 rounded-[4px] border border-hairline bg-sunken px-4 py-3">
							<input type="checkbox" class="mt-1 h-4 w-4 accent-[#e5484d]" bind:checked={rememberMe} />
							<span>
								<span class="block text-sm font-semibold text-ink">Ingat saya di browser ini</span>
								<span class="mt-1 block text-xs leading-5 text-muted">
									Menyimpan username saja. Password tidak disimpan, dan durasi sesi tetap mengikuti kebijakan server.
								</span>
							</span>
						</label>

						<button type="submit" class="btn-primary w-full" disabled={loading}>
							{#if loading}
								<div class="size-4 animate-spin rounded-full border-2 border-hairline border-t-white"></div>
								Memproses...
							{:else}
								<LoginIcon class="size-[18px]" />
								Masuk ke Dashboard
							{/if}
						</button>
					</form>

					<div class="rounded-[4px] border border-hairline bg-sunken px-4 py-4 text-sm text-verify-community">
						<div class="flex items-start gap-3">
							<CheckCircleIcon class="mt-0.5 size-5 shrink-0" />
							<p>
								Session login admin sekarang memakai cookie `HttpOnly` + `SameSite=Strict` dengan TTL backend 12 jam. Karena backend belum mendukung varian remember-me, opsi di atas sengaja dibatasi ke penyimpanan username saja.
							</p>
						</div>
					</div>
				</div>
			</section>
		</div>
	</div>
</div>
