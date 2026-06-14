<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import {
		BellIcon,
		ChartIcon,
		DocumentIcon,
		MapIcon,
		TrashIcon,
		MoonIcon,
		SunIcon
	} from '$lib/icons';
	import { requestIssueDetailRefresh } from '$lib/utils/issue-detail-refresh';
	import { notificationsState, unreadNotificationCount } from '$lib/stores/notifications';
	import { theme, setTheme, getInitialTheme, type ThemeMode } from '$lib/stores/theme';

	const pathname = $derived($page.url.pathname);
	const unreadCount = $derived($unreadNotificationCount);
	const notifState = $derived($notificationsState);
	const navItems = [
		{ href: '/lapor', label: 'Lapor', icon: DocumentIcon },
		{ href: '/issues', label: 'Peta', icon: MapIcon },
		{ href: '/stats', label: 'Statistik', icon: ChartIcon }
	];

	let openNotif = $state(false);
	let notifButtonEl = $state<HTMLButtonElement | null>(null);
	let notifPanelEl = $state<HTMLDivElement | null>(null);
	let BrowserPushCardComponent =
		$state<typeof import('$lib/components/BrowserPushCard.svelte').default | null>(null);
	let NearbyAlertsPanelComponent =
		$state<typeof import('$lib/components/NearbyAlertsPanel.svelte').default | null>(null);
	let NotificationPreferencesPanelComponent =
		$state<typeof import('$lib/components/NotificationPreferencesPanel.svelte').default | null>(null);
	let notifPanelsLoading = $state(false);
	let notifPanelsError = $state<string | null>(null);

	// Theme state
	let currentTheme = $state<ThemeMode>('system');
	
	onMount(() => {
		currentTheme = getInitialTheme();
	});

	const ThemeIcon = $derived(currentTheme === 'dark' || (currentTheme === 'system' && typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches) ? SunIcon : MoonIcon);

	function toggleTheme() {
		const newTheme: ThemeMode = currentTheme === 'dark' ? 'light' : 'dark';
		currentTheme = newTheme;
		setTheme(newTheme);
	}

	function isLaporActive(path: string): boolean {
		return path === '/lapor' || path.startsWith('/lapor/');
	}

	function isPetaActive(path: string): boolean {
		return path === '/issues' || path.startsWith('/issues/');
	}

	function isStatsActive(path: string): boolean {
		return path === '/stats' || path.startsWith('/stats/');
	}

	function isActiveNav(href: string, path: string): boolean {
		if (href === '/lapor') return isLaporActive(path);
		if (href === '/issues') return isPetaActive(path);
		if (href === '/stats') return isStatsActive(path);
		return false;
	}

	function formatNotifTime(input: string): string {
		const date = new Date(input);
		if (Number.isNaN(date.getTime())) return '';
		return date.toLocaleString('id-ID', {
			day: '2-digit',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	async function handleNotificationClick(id: string, issueID: string) {
		await notificationsState.markRead(id);
		openNotif = false;

		const targetPath = `/issues/${issueID}`;
		if (pathname === targetPath) {
			requestIssueDetailRefresh({ issueID, source: 'notification' });
			return;
		}

		await goto(targetPath);
	}

	async function handleNotificationDelete(event: MouseEvent, id: string) {
		event.preventDefault();
		event.stopPropagation();
		await notificationsState.delete(id);
	}

	async function ensureNotifPanelsLoaded() {
		if (
			notifPanelsLoading ||
			(BrowserPushCardComponent &&
				NearbyAlertsPanelComponent &&
				NotificationPreferencesPanelComponent)
		) {
			return;
		}

		notifPanelsLoading = true;
		notifPanelsError = null;

		try {
			const [browserPushCardModule, nearbyAlertsPanelModule, notificationPreferencesPanelModule] =
				await Promise.all([
					import('$lib/components/BrowserPushCard.svelte'),
					import('$lib/components/NearbyAlertsPanel.svelte'),
					import('$lib/components/NotificationPreferencesPanel.svelte')
				]);
			BrowserPushCardComponent = browserPushCardModule.default;
			NearbyAlertsPanelComponent = nearbyAlertsPanelModule.default;
			NotificationPreferencesPanelComponent = notificationPreferencesPanelModule.default;
		} catch {
			notifPanelsError = 'Panel notifikasi tambahan belum bisa dimuat. Coba buka lagi.';
		} finally {
			notifPanelsLoading = false;
		}
	}

	function handlePanelToggle() {
		openNotif = !openNotif;
		if (openNotif) {
			void ensureNotifPanelsLoaded();
		}
	}

	onMount(() => {
		function handlePointerDown(event: PointerEvent) {
			if (!openNotif) return;
			const target = event.target as Node | null;
			if (notifPanelEl?.contains(target) || notifButtonEl?.contains(target)) return;
			openNotif = false;
		}

		function handleKeyDown(event: KeyboardEvent) {
			if (event.key === 'Escape') {
				openNotif = false;
			}
		}

		window.addEventListener('pointerdown', handlePointerDown);
		window.addEventListener('keydown', handleKeyDown);

		return () => {
			window.removeEventListener('pointerdown', handlePointerDown);
			window.removeEventListener('keydown', handleKeyDown);
		};
	});

	$effect(() => {
		pathname;
		openNotif = false;
	});
</script>

<header class="sticky top-0 z-[100] border-b border-hairline bg-paper">
	<div class="mx-auto flex w-full max-w-[1120px] flex-col gap-4 px-5 py-4 sm:px-6 lg:px-8">
		<div class="flex items-center justify-between gap-4">
			<a href="/" class="flex min-w-0 items-baseline gap-3">
				<strong class="font-serif text-xl font-semibold tracking-[-0.02em] text-ink">JEDUG</strong>
				<span class="hidden text-xs text-muted sm:inline">Pantau Jalan Rusak</span>
			</a>

			<div class="flex items-center gap-2">
				<!-- Theme Toggle -->
				<button
					type="button"
					class="btn-icon"
					onclick={toggleTheme}
					aria-label={currentTheme === 'dark' ? 'Aktifkan mode terang' : 'Aktifkan mode gelap'}
					title={currentTheme === 'dark' ? 'Mode Terang' : 'Mode Gelap'}
				>
					<ThemeIcon class="size-5" />
				</button>

				<div class="relative shrink-0">
				<button
					bind:this={notifButtonEl}
					type="button"
					class="btn-icon relative"
					onclick={handlePanelToggle}
					aria-haspopup="dialog"
					aria-expanded={openNotif}
					aria-label="Lihat notifikasi"
				>
					<BellIcon class="size-5" />
					{#if unreadCount > 0}
						<span class="absolute -right-1 -top-1 inline-flex min-h-5 min-w-5 items-center justify-center rounded-full bg-brand px-1.5 text-[10px] font-semibold text-paper">
							{unreadCount > 99 ? '99+' : unreadCount}
						</span>
					{/if}
				</button>

				{#if openNotif}
					<div
						bind:this={notifPanelEl}
						class="absolute right-0 top-[calc(100%+0.75rem)] z-[120] flex w-[min(95vw,26rem)] flex-col overflow-hidden rounded-[12px] border border-hairline bg-surface"
					>
						<div class="flex items-center gap-3 border-b border-hairline px-4 py-4">
							<div class="min-w-0 flex-1">
								<p class="kicker mb-1">Notifikasi</p>
								<h2 class="font-serif text-base font-semibold text-ink">Pusat Notifikasi</h2>
								<p class="mt-1 text-xs leading-5 text-muted">
									Update issue, preferensi, dan nearby alerts dalam satu panel.
								</p>
							</div>
							<span class="badge-muted">{unreadCount} unread</span>
						</div>

						<div class="max-h-[75dvh] overflow-y-auto px-3 py-3">
							{#if notifPanelsLoading && !BrowserPushCardComponent}
								<div class="state-panel px-4 py-6">
									<div class="mx-auto size-9 animate-spin rounded-full border-[3px] border-hairline border-t-brand"></div>
									<p class="mt-3 text-sm text-muted">Menyiapkan panel notifikasi...</p>
								</div>
							{:else}
								{#if notifPanelsError}
									<div class="error-panel mx-1 mb-3">{notifPanelsError}</div>
								{/if}
								{#if BrowserPushCardComponent}
									<BrowserPushCardComponent
										variant="compact"
										lead="Aktifkan notifikasi browser agar perubahan issue tetap masuk walau tab JEDUG tidak sedang dibuka."
									/>
								{/if}
								{#if NotificationPreferencesPanelComponent}
									<NotificationPreferencesPanelComponent />
								{/if}
								{#if NearbyAlertsPanelComponent}
									<NearbyAlertsPanelComponent />
								{/if}
							{/if}

							<div class="mt-3 rounded-[8px] border border-hairline bg-sunken p-2">
								<div class="flex items-center justify-between px-2 py-2">
									<h3 class="text-sm font-semibold text-ink">Update terbaru</h3>
									<span class="text-xs font-medium text-subtle">
										{notifState.items.length} item
									</span>
								</div>

								{#if notifState.loading}
									<div class="state-panel px-4 py-6">
										<div class="mx-auto size-9 animate-spin rounded-full border-[3px] border-hairline border-t-brand"></div>
										<p class="mt-3 text-sm text-muted">Memuat notifikasi...</p>
									</div>
								{:else if notifState.error}
									<div class="error-panel m-2">{notifState.error}</div>
								{:else if notifState.items.length === 0}
									<div class="state-panel px-4 py-6">
										<p class="text-sm font-semibold text-ink">Belum ada notifikasi</p>
										<p class="mt-1 text-xs leading-5 text-muted">
											Ikuti issue atau aktifkan pantauan area agar update penting masuk ke sini.
										</p>
									</div>
								{:else}
									<ul class="space-y-2">
										{#each notifState.items as item (item.id)}
											<li class="group flex items-stretch gap-2">
												<button
													type="button"
													class={`min-w-0 flex-1 rounded-[8px] border px-4 py-3 text-left transition-colors hover:border-ink ${!item.read_at ? 'border-brand/30 bg-brand-tint' : 'border-hairline bg-surface'}`}
													onclick={() => handleNotificationClick(item.id, item.issue_id)}
												>
													<div class="min-w-0 flex-1">
														<div class="flex items-start justify-between gap-3">
															<p class="text-sm font-semibold leading-5 text-ink">
																{item.title}
															</p>
															{#if !item.read_at}
																<span class="mt-1.5 size-2 shrink-0 rounded-full bg-brand"></span>
															{/if}
														</div>
														<p class="mt-1 text-xs leading-5 text-muted">
															{item.message}
														</p>
														<p class="mt-2 text-[11px] font-semibold uppercase tracking-[0.14em] text-subtle">
															{formatNotifTime(item.created_at)}
														</p>
													</div>
												</button>
												<button
													type="button"
													class="btn-icon size-11 self-center text-subtle"
													aria-label="Hapus notifikasi"
													disabled={notifState.deletingIDs.includes(item.id)}
													onclick={(event) => handleNotificationDelete(event, item.id)}
												>
													<TrashIcon class="size-[18px]" />
												</button>
											</li>
										{/each}
									</ul>
								{/if}
							</div>
						</div>
					</div>
				{/if}
			</div>
			</div>
		</div>

		<nav class="-mb-px flex items-center gap-6" aria-label="Navigasi utama">
			{#each navItems as item}
				{@const NavIcon = item.icon}
				{@const active = isActiveNav(item.href, pathname)}
				<a
					href={item.href}
					class:text-ink={active}
					class:border-ink={active}
					class="flex items-center gap-2 border-b-2 border-transparent pb-2.5 text-sm font-semibold text-muted transition-colors hover:text-ink"
					aria-current={active ? 'page' : undefined}
				>
					<NavIcon class="size-[18px]" />
					<span>{item.label}</span>
				</a>
			{/each}
		</nav>
	</div>
</header>
