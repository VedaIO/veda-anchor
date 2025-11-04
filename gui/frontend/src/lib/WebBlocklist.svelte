<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import { showToast } from './toastStore';

  interface WebBlocklistItem {
    domain: string;
    title: string;
    iconUrl: string;
  }

  let webBlocklistItems = writable<WebBlocklistItem[]>([]);
  let selectedWebsites: string[] = [];

  async function loadWebBlocklist(): Promise<void> {
    const res = await fetch('/api/web-blocklist', { cache: 'no-cache' });
    const data = await res.json();
    if (data && data.length > 0) {
      webBlocklistItems.set(data);
    } else {
      webBlocklistItems.set([]);
    }
  }

  async function removeWebBlocklist(domain: string): Promise<void> {
    if (confirm(`Bạn có chắc chắn muốn bỏ chặn ${domain} không?`)) {
      const response = await fetch('/api/web-blocklist/remove', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain: domain }),
      });
      if (response.ok) {
        showToast(`Đã bỏ chặn ${domain}`, 'success');
        loadWebBlocklist();
      } else {
        showToast(`Lỗi bỏ chặn ${domain}: ${response.statusText}`, 'error');
      }
    }
  }

  async function unblockSelectedWebsites(): Promise<void> {
    if (selectedWebsites.length === 0) {
      showToast('Vui lòng chọn các trang web để bỏ chặn.', 'info');
      return;
    }

    const removalPromises = selectedWebsites.map(async (domain) => {
      const response = await fetch('/api/web-blocklist/remove', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain }),
      });
      if (!response.ok) {
        showToast(`Lỗi bỏ chặn ${domain}: ${response.statusText}`, 'error');
        throw new Error(`Failed to unblock ${domain}`);
      }
    });

    try {
      await Promise.all(removalPromises);
      showToast(`Đã bỏ chặn: ${selectedWebsites.join(', ')}`, 'success');
    } catch {
      return;
    }

    loadWebBlocklist();
    selectedWebsites = [];
  }

  async function clearWebBlocklist(): Promise<void> {
    if (
      confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn web không?')
    ) {
      await fetch('/api/web-blocklist/clear', { method: 'POST' });
      showToast('Đã xóa toàn bộ danh sách chặn web.', 'success');
      loadWebBlocklist();
    }
  }

  async function saveWebBlocklist(): Promise<void> {
    const response = await fetch('/api/web-blocklist/save');
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = url;
    a.download = 'procguard_web_blocklist.json';
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
  }

  async function loadWebBlocklistFile(event: Event): Promise<void> {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (!file) {
      return;
    }
    const formData = new FormData();
    formData.append('file', file);

    await fetch('/api/web-blocklist/load', {
      method: 'POST',
      body: formData,
    });

    showToast('Đã tải lên và hợp nhất danh sách chặn web.', 'success');
    loadWebBlocklist();
  }

  onMount(() => {
    loadWebBlocklist();
  });
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Các trang web bị chặn</h5>
    <div class="btn-toolbar" role="toolbar">
      <div class="btn-group me-2" role="group">
        <button
          type="button"
          class="btn btn-primary"
          on:click={unblockSelectedWebsites}
        >
          Bỏ chặn mục đã chọn
        </button>
        <button
          type="button"
          class="btn btn-danger"
          on:click={clearWebBlocklist}
        >
          Xóa toàn bộ
        </button>
      </div>
      <div class="btn-group" role="group">
        <button
          type="button"
          class="btn btn-outline-secondary"
          on:click={saveWebBlocklist}
        >
          Lưu danh sách
        </button>
        <button
          type="button"
          class="btn btn-outline-secondary"
          on:click={() => document.getElementById('load-web-input')?.click()}
        >
          Tải lên danh sách
        </button>
      </div>
    </div>
    <input
      type="file"
      id="load-web-input"
      style="display: none"
      on:change={loadWebBlocklistFile}
    />
    <div id="web-blocklist-items" class="list-group mt-3">
      {#if $webBlocklistItems.length > 0}
        {#each $webBlocklistItems as item (item.domain)}
          <div
            class="list-group-item d-flex justify-content-between align-items-center"
          >
            <label class="flex-grow-1 mb-0 d-flex align-items-center">
              <input
                class="form-check-input me-2"
                type="checkbox"
                name="blocked-website"
                value={item.domain}
                bind:group={selectedWebsites}
              />
              {#if item.iconUrl}
                <img
                  src={item.iconUrl}
                  class="me-2"
                  style="width: 24px; height: 24px;"
                  alt="Website Icon"
                />
              {:else}
                <div class="me-2" style="width: 24px; height: 24px;"></div>
              {/if}
              <span class="fw-bold me-2">{item.title || item.domain}</span>
            </label>
            <button
              class="btn btn-sm btn-outline-danger"
              on:click={() => removeWebBlocklist(item.domain)}>&times;</button
            >
          </div>
        {/each}
      {:else}
        <div class="list-group-item">Hiện không có trang web nào bị chặn.</div>
      {/if}
    </div>
  </div>
</div>
