<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import BrowserPushCard from '$lib/components/BrowserPushCard.svelte';
	import NearbyAlertsPanel from '$lib/components/NearbyAlertsPanel.svelte';
	import NotificationPreferencesPanel from '$lib/components/NotificationPreferencesPanel.svelte';
	import {
		BellIcon,
		ChartIcon,
		DocumentIcon,
		MapIcon,
		NotificationIcon,
		TrashIcon
	} from '$lib/icons';
	import { requestIssueDetailRefresh } from '$lib/utils/issue-detail-refresh';
	import { notificationsState, unreadNotificationCount } from '$lib/stores/notifications';

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

	function handlePanelToggle() {
		openNotif = !openNotif;
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

<header class="sticky top-0 z-[100] border-b border-white/70 bg-white/80 backdrop-blur-xl">
	<div class="mx-auto flex w-full max-w-[1200px] flex-col gap-3 px-4 py-3 sm:px-6 lg:px-8">
		<div class="flex items-center justify-between gap-4">
			<a
				href="/"
				class="flex min-w-0 items-center gap-3 rounded-[24px] border border-white/60 bg-white/75 px-3 py-2 shadow-[0_12px_28px_rgba(15,23,42,0.06)] backdrop-blur"
			>
				<div class="flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
					<MapIcon class="size-6" />
				</div>
				<div class="min-w-0">
					<p class="truncate text-[11px] font-bold uppercase tracking-[0.18em] text-brand-600">
						Civic-Tech
					</p>
					<div class="flex flex-wrap items-center gap-x-2 gap-y-1">
						<strong class="text-lg font-[800] tracking-[-0.04em] text-slate-950">JEDUG</strong>
						<span class="text-xs text-slate-500 sm:text-sm">Pantau Jalan Rusak</span>
					</div>
				</div>
			</a>

			<div class="relative shrink-0">
				<button
					bind:this={notifButtonEl}
					type="button"
					class="btn-icon relative border-brand-100/70 bg-white/90 text-slate-700"
					onclick={handlePanelToggle}
					aria-haspopup="dialog"
					aria-expanded={openNotif}
					aria-label="Lihat notifikasi"
				>
					<BellIcon class="size-5" />
					{#if unreadCount > 0}
						<span class="absolute -right-1 -top-1 inline-flex min-h-5 min-w-5 items-center justify-center rounded-full bg-brand-500 px-1.5 text-[10px] font-bold text-white shadow-[0_10px_22px_rgba(229,72,77,0.28)]">
							{unreadCount > 99 ? '99+' : unreadCount}
						</span>
					{/if}
				</button>

				{#if openNotif}
					<div
						bind:this={notifPanelEl}
						class="absolute right-0 top-[calc(100%+0.85rem)] z-[120] flex w-[min(95vw,26rem)] flex-col overflow-hidden rounded-[28px] border border-white/80 bg-white/95 shadow-[0_24px_60px_rgba(15,23,42,0.18)] backdrop-blur-xl"
					>
						<div class="border-b border-slate-100 px-4 py-4">
							<div class="flex items-center gap-3">
								<div class="flex size-10 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
									<NotificationIcon class="size-5" />
								</div>
								<div class="min-w-0 flex-1">
									<h2 class="text-sm font-bold text-slate-950">Pusat Notifikasi</h2>
									<p class="text-xs leading-5 text-slate-500">
										Update issue, preferensi, dan nearby alerts terkumpul di satu panel.
									</p>
								</div>
								<span class="badge-tint">{unreadCount} unread</span>
							</div>
						</div>

						<div class="max-h-[75dvh] overflow-y-auto px-3 py-3">
							<BrowserPushCard
								variant="compact"
								lead="Aktifkan notifikasi browser agar perubahan issue tetap masuk walau tab JEDUG tidak sedang dibuka."
							/>
							<NotificationPreferencesPanel />
							<NearbyAlertsPanel />

							<div class="mt-3 rounded-[24px] border border-slate-100 bg-slate-50/75 p-2">
								<div class="flex items-center justify-between px-2 py-2">
									<h3 class="text-sm font-bold text-slate-900">Update terbaru</h3>
									<span class="text-xs font-semibold text-slate-400">
										{notifState.items.length} item
									</span>
								</div>

								{#if notifState.loading}
									<div class="state-panel border-0 bg-white/80 px-4 py-6">
										<div class="mx-auto size-9 animate-spin rounded-full border-[3px] border-slate-200 border-t-brand-500"></div>
										<p class="mt-3 text-sm text-slate-500">Memuat notifikasi...</p>
									</div>
								{:else if notifState.error}
									<div class="error-panel m-2">{notifState.error}</div>
								{:else if notifState.items.length === 0}
									<div class="state-panel border-0 bg-white/80 px-4 py-6">
										<div class="mx-auto flex size-12 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
											<NotificationIcon class="size-6" />
										</div>
										<p class="mt-3 text-sm font-semibold text-slate-700">Belum ada notifikasi</p>
										<p class="mt-1 text-xs leading-5 text-slate-500">
											Ikuti issue atau aktifkan pantauan area agar update penting masuk ke sini.
										</p>
									</div>
								{:else}
									<ul class="space-y-2">
										{#each notifState.items as item (item.id)}
											<li class="group flex items-stretch gap-2">
												<button
													type="button"
													class={`min-w-0 flex-1 rounded-[22px] border px-4 py-3 text-left shadow-[0_8px_22px_rgba(15,23,42,0.05)] transition hover:-translate-y-0.5 hover:border-slate-300 ${!item.read_at ? 'border-brand-100 bg-brand-50/80' : 'border-slate-200 bg-white'}`}
													onclick={() => handleNotificationClick(item.id, item.issue_id)}
												>
													<div class="flex items-start gap-3">
														<div
															class:border-brand-100={!item.read_at}
															class:bg-brand-50={!item.read_at}
															class="mt-0.5 flex size-10 shrink-0 items-center justify-center rounded-2xl border border-slate-200 bg-slate-50 text-slate-600"
														>
															<BellIcon class="size-[18px]" />
														</div>
														<div class="min-w-0 flex-1">
															<div class="flex items-start justify-between gap-3">
																<p class="text-sm font-bold leading-5 text-slate-900">
																	{item.title}
																</p>
																{#if !item.read_at}
																	<span class="mt-0.5 size-2 shrink-0 rounded-full bg-brand-500"></span>
																{/if}
															</div>
															<p class="mt-1 text-xs leading-5 text-slate-600">
																{item.message}
															</p>
															<p class="mt-2 text-[11px] font-semibold uppercase tracking-[0.14em] text-slate-400">
																{formatNotifTime(item.created_at)}
															</p>
														</div>
													</div>
												</button>
												<button
													type="button"
													class="btn-icon size-11 self-center rounded-2xl border-slate-200 bg-white text-slate-500"
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

		<nav
			class="grid grid-cols-3 gap-2 rounded-[24px] border border-white/70 bg-white/75 p-1.5 shadow-[0_12px_28px_rgba(15,23,42,0.05)] backdrop-blur"
			aria-label="Navigasi utama"
		>
			{#each navItems as item}
				{@const NavIcon = item.icon}
				<a
					href={item.href}
					class:bg-brand-500={isActiveNav(item.href, pathname)}
					class:text-white={isActiveNav(item.href, pathname)}
					class:shadow-[0_14px_30px_rgba(229,72,77,0.2)]={isActiveNav(item.href, pathname)}
					class="flex min-h-12 items-center justify-center gap-2 rounded-[18px] px-3 py-2 text-sm font-semibold text-slate-600 transition hover:bg-slate-100 hover:text-slate-900"
					aria-current={isActiveNav(item.href, pathname) ? 'page' : undefined}
				>
					<NavIcon class="size-[18px]" />
					<span>{item.label}</span>
				</a>
			{/each}
		</nav>
	</div>
</header>
