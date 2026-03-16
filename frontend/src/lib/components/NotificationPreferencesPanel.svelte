<script lang="ts">
  import { browserPushState } from "$lib/stores/browser-push";
  import { notificationPreferencesState } from "$lib/stores/notification-preferences";

  const prefsState = $derived($notificationPreferencesState);
  const pushState = $derived($browserPushState);

  let open = $state(false);

  const preferences = $derived(prefsState.preferences);
  const channelsDisabled = $derived(
    !preferences || !preferences.notifications_enabled,
  );
  const pushAvailable = $derived(pushState.status === "subscribed");
  const pushToggleDisabled = $derived(
    channelsDisabled || pushState.busy || (!pushAvailable && !preferences?.push_enabled),
  );
  const pushHint = $derived.by(() => {
    if (!preferences) return "";
    if (pushAvailable) {
      return "Update akan dikirim ke browser ini saat langganan push masih aktif.";
    }
    if (pushState.status === "denied") {
      return "Izin browser ditolak. Aktifkan lagi dari pengaturan browser jika perlu.";
    }
    if (pushState.status === "unsupported") {
      return "Browser ini belum mendukung push notification.";
    }
    return "Aktifkan notifikasi browser di kartu atas dulu untuk memakai channel ini.";
  });

  function isSaving(key: string): boolean {
    return prefsState.savingKeys.includes(key);
  }

  async function toggle(key: string, value: boolean) {
    await notificationPreferencesState.update({ [key]: value }, key);
  }
</script>

<section class="prefs-card">
  <button
    type="button"
    class="prefs-toggle"
    aria-expanded={open}
    onclick={() => (open = !open)}
  >
    <div>
      <div class="prefs-title">Preferensi Notifikasi</div>
      <p>Atur update mana yang masih ingin kamu terima.</p>
    </div>
    <span>{open ? "Tutup" : "Atur"}</span>
  </button>

  {#if open}
    <div class="prefs-body">
      {#if prefsState.loading}
        <div class="prefs-empty">Memuat preferensi...</div>
      {:else if prefsState.unavailableMessage}
        <div class="prefs-empty">{prefsState.unavailableMessage}</div>
      {:else if prefsState.error}
        <div class="prefs-empty prefs-error">{prefsState.error}</div>
      {:else if preferences}
        <div class="prefs-section">
          <div class="prefs-section-title">Umum</div>
          <label class="prefs-row">
            <div class="prefs-copy">
              <div>Semua notifikasi</div>
              <p>Matikan seluruh update untuk laporan yang kamu ikuti di browser ini.</p>
            </div>
            <input
              type="checkbox"
              checked={preferences.notifications_enabled}
              disabled={isSaving("notifications_enabled")}
              onchange={(event) =>
                toggle(
                  "notifications_enabled",
                  (event.currentTarget as HTMLInputElement).checked,
                )}
            />
          </label>
        </div>

        <div class="prefs-section">
          <div class="prefs-section-title">Channel</div>
          <label class="prefs-row">
            <div class="prefs-copy">
              <div>Notifikasi di dalam aplikasi</div>
              <p>Update baru muncul di panel lonceng saat kamu membuka JEDUG.</p>
            </div>
            <input
              type="checkbox"
              checked={preferences.in_app_enabled}
              disabled={channelsDisabled || isSaving("in_app_enabled")}
              onchange={(event) =>
                toggle(
                  "in_app_enabled",
                  (event.currentTarget as HTMLInputElement).checked,
                )}
            />
          </label>

          <label class="prefs-row">
            <div class="prefs-copy">
              <div>Notifikasi browser</div>
              <p>{pushHint}</p>
            </div>
            <input
              type="checkbox"
              checked={preferences.push_enabled}
              disabled={pushToggleDisabled || isSaving("push_enabled")}
              onchange={(event) =>
                toggle(
                  "push_enabled",
                  (event.currentTarget as HTMLInputElement).checked,
                )}
            />
          </label>
        </div>

        <div class="prefs-section">
          <div class="prefs-section-title">Jenis update</div>
          <label class="prefs-row">
            <div class="prefs-copy">
              <div>Foto baru pada laporan yang kamu ikuti</div>
              <p>Dipakai saat ada bukti foto tambahan pada issue yang sama.</p>
            </div>
            <input
              type="checkbox"
              checked={preferences.notify_on_photo_added}
              disabled={channelsDisabled || isSaving("notify_on_photo_added")}
              onchange={(event) =>
                toggle(
                  "notify_on_photo_added",
                  (event.currentTarget as HTMLInputElement).checked,
                )}
            />
          </label>

          <label class="prefs-row">
            <div class="prefs-copy">
              <div>Perubahan status laporan</div>
              <p>Misalnya saat issue ditandai selesai, ditolak, atau diarsipkan.</p>
            </div>
            <input
              type="checkbox"
              checked={preferences.notify_on_status_updated}
              disabled={channelsDisabled || isSaving("notify_on_status_updated")}
              onchange={(event) =>
                toggle(
                  "notify_on_status_updated",
                  (event.currentTarget as HTMLInputElement).checked,
                )}
            />
          </label>

          <label class="prefs-row">
            <div class="prefs-copy">
              <div>Perubahan tingkat keparahan</div>
              <p>Dipakai saat tingkat kerusakan issue dinaikkan oleh laporan baru.</p>
            </div>
            <input
              type="checkbox"
              checked={preferences.notify_on_severity_changed}
              disabled={channelsDisabled || isSaving("notify_on_severity_changed")}
              onchange={(event) =>
                toggle(
                  "notify_on_severity_changed",
                  (event.currentTarget as HTMLInputElement).checked,
                )}
            />
          </label>

          <label class="prefs-row">
            <div class="prefs-copy">
              <div>Laporan korban baru</div>
              <p>Dipakai saat ada laporan korban baru atau jumlah korban meningkat.</p>
            </div>
            <input
              type="checkbox"
              checked={preferences.notify_on_casualty_reported}
              disabled={channelsDisabled || isSaving("notify_on_casualty_reported")}
              onchange={(event) =>
                toggle(
                  "notify_on_casualty_reported",
                  (event.currentTarget as HTMLInputElement).checked,
                )}
            />
          </label>

          <label class="prefs-row">
            <div class="prefs-copy">
              <div>Laporan baru di area pantauan</div>
              <p>Dipakai saat ada issue baru yang masuk ke radius lokasi Nearby Alerts milikmu.</p>
            </div>
            <input
              type="checkbox"
              checked={preferences.notify_on_nearby_issue_created}
              disabled={channelsDisabled || isSaving("notify_on_nearby_issue_created")}
              onchange={(event) =>
                toggle(
                  "notify_on_nearby_issue_created",
                  (event.currentTarget as HTMLInputElement).checked,
                )}
            />
          </label>
        </div>
      {/if}
    </div>
  {/if}
</section>

<style>
  .prefs-card {
    margin: 4px 8px 12px;
    border: 1px solid #E2E8F0;
    border-radius: 12px;
    background: #FFFFFF;
    overflow: hidden;
  }

  .prefs-toggle {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px;
    border: none;
    background: transparent;
    text-align: left;
    cursor: pointer;
  }

  .prefs-title {
    font-size: 14px;
    font-weight: 700;
    color: #0F172A;
  }

  .prefs-toggle p {
    margin-top: 4px;
    font-size: 12px;
    color: #64748B;
    line-height: 1.5;
  }

  .prefs-toggle span {
    flex-shrink: 0;
    font-size: 12px;
    font-weight: 700;
    color: #B42318;
  }

  .prefs-body {
    display: grid;
    gap: 14px;
    padding: 0 12px 12px;
    border-top: 1px solid #F1F5F9;
  }

  .prefs-section {
    display: grid;
    gap: 10px;
  }

  .prefs-section-title {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #94A3B8;
  }

  .prefs-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    padding: 12px;
    border-radius: 12px;
    background: #F8FAFC;
  }

  .prefs-copy {
    display: grid;
    gap: 4px;
  }

  .prefs-copy div {
    font-size: 13px;
    font-weight: 600;
    color: #0F172A;
    line-height: 1.4;
  }

  .prefs-copy p {
    font-size: 12px;
    color: #64748B;
    line-height: 1.5;
  }

  .prefs-row input {
    width: 18px;
    height: 18px;
    margin-top: 2px;
    accent-color: #E5484D;
  }

  .prefs-empty {
    padding: 12px;
    border-radius: 12px;
    background: #F8FAFC;
    font-size: 13px;
    color: #64748B;
    line-height: 1.5;
  }

  .prefs-error {
    color: #B42318;
    background: #FFF1F2;
  }
</style>
