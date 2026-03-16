<script lang="ts">
  import { getLocation } from "$lib/utils/geolocation";
  import { nearbyAlertsState } from "$lib/stores/nearby-alerts";

  const nearbyState = $derived($nearbyAlertsState);

  let open = $state(false);
  let locating = $state(false);
  let formLabel = $state("");
  let formLatitude = $state("");
  let formLongitude = $state("");
  let formRadius = $state("500");
  let formError = $state<string | null>(null);
  let drafts = $state<Record<string, { label: string; radius: string }>>({});

  const limitReached = $derived(nearbyState.items.length >= 10);

  $effect(() => {
    if (!open) return;
    void nearbyAlertsState.init();
  });

  $effect(() => {
    const activeIDs = new Set(nearbyState.items.map((item) => item.id));
    for (const item of nearbyState.items) {
      if (!drafts[item.id]) {
        drafts[item.id] = {
          label: item.label ?? "",
          radius: String(item.radius_m),
        };
      }
    }
    for (const id of Object.keys(drafts)) {
      if (!activeIDs.has(id)) {
        delete drafts[id];
      }
    }
  });

  function isSaving(key: string): boolean {
    return nearbyState.savingKeys.includes(key);
  }

  function updateDraftLabel(id: string, value: string) {
    drafts[id] = {
      ...(drafts[id] ?? { label: "", radius: "500" }),
      label: value,
    };
  }

  function updateDraftRadius(id: string, value: string) {
    drafts[id] = {
      ...(drafts[id] ?? { label: "", radius: "500" }),
      radius: value,
    };
  }

  async function useCurrentLocation() {
    locating = true;
    formError = null;
    try {
      const location = await getLocation({ forceFresh: true });
      formLatitude = location.latitude.toFixed(6);
      formLongitude = location.longitude.toFixed(6);
    } catch (error) {
      formError = error instanceof Error ? error.message : "Belum bisa mengambil lokasi browser.";
    } finally {
      locating = false;
    }
  }

  async function createAlert() {
    formError = null;
    const latitude = Number.parseFloat(formLatitude);
    const longitude = Number.parseFloat(formLongitude);
    const radius = Number.parseInt(formRadius, 10);

    if (Number.isNaN(latitude) || Number.isNaN(longitude)) {
      formError = "Isi latitude dan longitude dulu, atau pakai lokasi browser.";
      return;
    }
    if (latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) {
      formError = "Koordinat belum valid. Pastikan latitude dan longitude benar.";
      return;
    }
    if (!Number.isFinite(radius) || radius < 100 || radius > 5000) {
      formError = "Radius harus antara 100 sampai 5000 meter.";
      return;
    }

    const created = await nearbyAlertsState.create({
      label: formLabel.trim() || undefined,
      latitude,
      longitude,
      radius_m: radius,
      enabled: true,
    });
    if (!created) return;

    formLabel = "";
    formLatitude = "";
    formLongitude = "";
    formRadius = "500";
  }

  async function saveItem(id: string) {
    const draft = drafts[id];
    if (!draft) return;

    const radius = Number.parseInt(draft.radius, 10);
    if (!Number.isFinite(radius) || radius < 100 || radius > 5000) {
      formError = "Radius harus antara 100 sampai 5000 meter.";
      return;
    }

    formError = null;
    await nearbyAlertsState.update(
      id,
      {
        label: draft.label.trim(),
        radius_m: radius,
      },
      `save:${id}`,
    );
  }

  async function toggleItem(id: string, enabled: boolean) {
    await nearbyAlertsState.update(id, { enabled }, `toggle:${id}`);
  }

  async function deleteItem(id: string) {
    formError = null;
    await nearbyAlertsState.remove(id);
  }

  function formatCoordinates(latitude: number, longitude: number): string {
    return `${latitude.toFixed(5)}, ${longitude.toFixed(5)}`;
  }
</script>

<section class="nearby-card">
  <button
    type="button"
    class="nearby-toggle"
    aria-expanded={open}
    onclick={() => (open = !open)}
  >
    <div>
      <div class="nearby-title">Pantau area sekitar rumah</div>
      <p>Beri tahu saya jika ada laporan baru di sekitar lokasi ini.</p>
    </div>
    <span>{open ? "Tutup" : "Atur"}</span>
  </button>

  {#if open}
    <div class="nearby-body">
      <div class="nearby-intro">
        Simpan beberapa lokasi seperti rumah, kantor, atau area yang ingin kamu pantau. Nearby Alerts tetap mengikuti preferensi notifikasi dan channel yang kamu aktifkan.
      </div>

      <div class="nearby-form">
        <div class="nearby-form-title">Tambah lokasi pantauan</div>
        <div class="nearby-grid">
          <label>
            <span>Label lokasi</span>
            <input
              type="text"
              placeholder="Mis. Rumah / Kantor"
              bind:value={formLabel}
              maxlength="80"
            />
          </label>
          <label>
            <span>Radius (meter)</span>
            <input
              type="number"
              min="100"
              max="5000"
              step="50"
              bind:value={formRadius}
            />
          </label>
          <label>
            <span>Latitude</span>
            <input
              type="number"
              step="0.000001"
              placeholder="-6.200000"
              bind:value={formLatitude}
            />
          </label>
          <label>
            <span>Longitude</span>
            <input
              type="number"
              step="0.000001"
              placeholder="106.816666"
              bind:value={formLongitude}
            />
          </label>
        </div>

        <div class="nearby-actions">
          <button
            type="button"
            class="secondary-button"
            disabled={locating || nearbyState.creating}
            onclick={useCurrentLocation}
          >
            {locating ? "Mengambil lokasi..." : "Gunakan lokasi saya"}
          </button>
          <button
            type="button"
            class="primary-button"
            disabled={limitReached || nearbyState.creating}
            onclick={createAlert}
          >
            {nearbyState.creating ? "Menyimpan..." : "Tambah lokasi"}
          </button>
        </div>

        {#if limitReached}
          <div class="nearby-note nearby-warning">
            Maksimum 10 lokasi pantauan per browser agar notifikasi tetap ringan dan tidak spammy.
          </div>
        {/if}
      </div>

      {#if formError}
        <div class="nearby-note nearby-error">{formError}</div>
      {/if}

      {#if nearbyState.loading}
        <div class="nearby-note">Memuat lokasi pantauan...</div>
      {:else if nearbyState.unavailableMessage}
        <div class="nearby-note">{nearbyState.unavailableMessage}</div>
      {:else if nearbyState.error}
        <div class="nearby-note nearby-error">{nearbyState.error}</div>
      {/if}

      {#if nearbyState.items.length > 0}
        <div class="nearby-list">
          {#each nearbyState.items as item (item.id)}
            <article class="nearby-item" aria-busy={isSaving(`save:${item.id}`) || isSaving(`toggle:${item.id}`)}>
              <div class="nearby-item-head">
                <div>
                  <div class="nearby-item-title">{item.label || "Area pantauan"}</div>
                  <p>{formatCoordinates(item.latitude, item.longitude)}</p>
                </div>
                <label class="nearby-switch">
                  <span>{item.enabled ? "Aktif" : "Nonaktif"}</span>
                  <input
                    type="checkbox"
                    checked={item.enabled}
                    disabled={isSaving(`toggle:${item.id}`)}
                    onchange={(event) => toggleItem(item.id, (event.currentTarget as HTMLInputElement).checked)}
                  />
                </label>
              </div>

              <div class="nearby-grid">
                <label>
                  <span>Label</span>
                  <input
                    type="text"
                    value={drafts[item.id]?.label ?? item.label ?? ""}
                    maxlength="80"
                    oninput={(event) => updateDraftLabel(item.id, (event.currentTarget as HTMLInputElement).value)}
                  />
                </label>
                <label>
                  <span>Radius (meter)</span>
                  <input
                    type="number"
                    min="100"
                    max="5000"
                    step="50"
                    value={drafts[item.id]?.radius ?? String(item.radius_m)}
                    oninput={(event) => updateDraftRadius(item.id, (event.currentTarget as HTMLInputElement).value)}
                  />
                </label>
              </div>

              <div class="nearby-actions">
                <button
                  type="button"
                  class="secondary-button"
                  disabled={nearbyState.deletingIDs.includes(item.id)}
                  onclick={() => deleteItem(item.id)}
                >
                  {nearbyState.deletingIDs.includes(item.id) ? "Menghapus..." : "Hapus"}
                </button>
                <button
                  type="button"
                  class="primary-button"
                  disabled={isSaving(`save:${item.id}`)}
                  onclick={() => saveItem(item.id)}
                >
                  {isSaving(`save:${item.id}`) ? "Menyimpan..." : "Simpan perubahan"}
                </button>
              </div>
            </article>
          {/each}
        </div>
      {:else if !nearbyState.loading && !nearbyState.error && !nearbyState.unavailableMessage}
        <div class="nearby-note">
          Belum ada area pantauan. Tambahkan lokasi yang penting buatmu supaya laporan baru di sekitar area itu bisa langsung masuk ke notifikasi.
        </div>
      {/if}
    </div>
  {/if}
</section>

<style>
  .nearby-card {
    margin: 4px 8px 12px;
    border: 1px solid #E2E8F0;
    border-radius: 12px;
    background: #FFFFFF;
    overflow: hidden;
  }

  .nearby-toggle {
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

  .nearby-title {
    font-size: 14px;
    font-weight: 700;
    color: #0F172A;
  }

  .nearby-toggle p {
    margin-top: 4px;
    font-size: 12px;
    color: #64748B;
    line-height: 1.5;
  }

  .nearby-toggle span {
    flex-shrink: 0;
    font-size: 12px;
    font-weight: 700;
    color: #B42318;
  }

  .nearby-body {
    display: grid;
    gap: 12px;
    padding: 0 12px 12px;
    border-top: 1px solid #F1F5F9;
  }

  .nearby-intro,
  .nearby-note {
    padding: 12px;
    border-radius: 12px;
    background: #F8FAFC;
    font-size: 12px;
    color: #64748B;
    line-height: 1.55;
  }

  .nearby-warning {
    background: #FFF7ED;
    color: #9A3412;
  }

  .nearby-error {
    background: #FFF1F2;
    color: #B42318;
  }

  .nearby-form,
  .nearby-item {
    display: grid;
    gap: 10px;
    padding: 12px;
    border-radius: 12px;
    background: #F8FAFC;
  }

  .nearby-form-title,
  .nearby-item-title {
    font-size: 13px;
    font-weight: 700;
    color: #0F172A;
  }

  .nearby-list {
    display: grid;
    gap: 10px;
  }

  .nearby-item-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .nearby-item-head p {
    margin-top: 4px;
    font-size: 12px;
    color: #64748B;
  }

  .nearby-switch {
    display: grid;
    gap: 6px;
    justify-items: end;
    font-size: 12px;
    font-weight: 600;
    color: #64748B;
  }

  .nearby-switch input {
    width: 18px;
    height: 18px;
    accent-color: #E5484D;
  }

  .nearby-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .nearby-grid label {
    display: grid;
    gap: 6px;
  }

  .nearby-grid span {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: #94A3B8;
  }

  .nearby-grid input {
    width: 100%;
    min-height: 40px;
    padding: 10px 12px;
    border: 1px solid #D8E1EC;
    border-radius: 10px;
    background: #FFFFFF;
    font-size: 13px;
    color: #0F172A;
  }

  .nearby-grid input:focus {
    outline: none;
    border-color: #E5484D;
    box-shadow: 0 0 0 3px rgba(229, 72, 77, 0.12);
  }

  .nearby-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .primary-button,
  .secondary-button {
    min-height: 44px;
    padding: 0 14px;
    border-radius: 10px;
    border: 1px solid transparent;
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
  }

  .primary-button {
    background: #E5484D;
    color: #FFFFFF;
  }

  .secondary-button {
    background: #FFFFFF;
    color: #0F172A;
    border-color: #E2E8F0;
  }

  .primary-button:disabled,
  .secondary-button:disabled {
    opacity: 0.55;
    cursor: default;
  }

  @media (max-width: 480px) {
    .nearby-grid {
      grid-template-columns: 1fr;
    }

    .nearby-actions {
      flex-direction: column;
    }

    .primary-button,
    .secondary-button {
      width: 100%;
    }
  }
</style>