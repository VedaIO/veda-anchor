<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import { showToast } from './toastStore';

  interface WebLogItem {
    timestamp: string;
    url: string;
    domain: string;
    title: string;
    iconUrl: string;
  }

  let webLogItems = writable<WebLogItem[]>([]);
  let webSinceDate = '';
  let webSinceTime = '';
  let webUntilDate = '';
  let webUntilTime = '';

  async function searchWebLogs(range?: {
    since: string;
    until: string;
  }): Promise<void> {
    let since = '';
    let until = '';

    if (range) {
      since = range.since;
      until = range.until;
    } else {
      if (webSinceDate && webSinceTime) {
        since = `${webSinceDate}T${webSinceTime}`;
      }
      if (webUntilDate && webUntilTime) {
        until = `${webUntilDate}T${webUntilTime}`;
      }
    }
    await loadWebLogs(since, until);
  }

  async function loadWebLogs(since = '', until = ''): Promise<void> {
    let url = '/api/web-logs';
    // eslint-disable-next-line svelte/prefer-svelte-reactivity
    const params = new URLSearchParams();
    if (since) {
      params.append('since', since);
    }
    if (until) {
      params.append('until', until);
    }
    const queryString = params.toString();
    if (queryString) {
      url += `?${queryString}`;
    }

    const res = await fetch(url);
    const data = await res.json();
    if (data && data.length > 0) {
      const items: WebLogItem[] = await Promise.all(
        data.map(async (l: string[]) => {
          const urlString = l[1];
          let domain = '';
          try {
            const url = new URL(urlString);
            domain = url.hostname;
          } catch {
            // Ignore invalid URLs
          }

          let title = '';
          let iconUrl = '';
          if (domain) {
            const webDetailsRes = await fetch(
              `/api/web-details?domain=${encodeURIComponent(domain)}`
            );
            if (webDetailsRes.ok) {
              const webDetails = await webDetailsRes.json();
              title = webDetails.title;
              iconUrl = webDetails.iconUrl;
            }
          }

          const timestamp = l[0]; // Just the timestamp

          return {
            timestamp,
            url: urlString,
            domain,
            title,
            iconUrl,
          };
        })
      );
      webLogItems.set(items);
    } else {
      webLogItems.set([]);
    }
  }

  async function blockSelectedWebsites(): Promise<void> {
    const selectedDomains = Array.from(
      document.querySelectorAll('input[name="web-log-domain"]:checked')
    ).map((cb) => (cb as HTMLInputElement).value);
    if (selectedDomains.length === 0) {
      showToast('Vui lòng chọn một trang web để chặn.', 'info');
      return;
    }

    const uniqueDomains = [...new Set(selectedDomains)];

    for (const domain of uniqueDomains) {
      await fetch('/api/web-blocklist/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain: domain }),
      });
    }

    showToast(
      'Các trang web đã chọn đã được thêm vào danh sách chặn.',
      'success'
    );
    // Uncheck all boxes
    (
      document.querySelectorAll(
        'input[name="web-log-domain"]:checked'
      ) as NodeListOf<HTMLInputElement>
    ).forEach((cb) => (cb.checked = false));
  }

  onMount(() => {
    // Set default dates
    const now = new Date();
    const year = now.getFullYear();
    const month = (now.getMonth() + 1).toString().padStart(2, '0');
    const day = now.getDate().toString().padStart(2, '0');
    const today = `${year}-${month}-${day}`;
    webSinceDate = today;
    webUntilDate = today;

    loadWebLogs();
  });
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Lịch sử Truy cập Web</h5>
    <div class="mb-3">
      <button class="btn btn-danger" on:click={blockSelectedWebsites}>
        Chặn mục đã chọn
      </button>
    </div>

    <div class="card mt-3">
      <div class="card-header d-flex align-items-center">
        <span class="me-3">Lọc theo thời gian:</span>
        <div class="btn-group" role="group">
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() =>
              searchWebLogs({ since: '1 hour ago', until: 'now' })}
          >
            1 giờ qua
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() =>
              searchWebLogs({ since: '24 hours ago', until: 'now' })}
          >
            24 giờ qua
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() =>
              searchWebLogs({ since: '7 days ago', until: 'now' })}
          >
            7 ngày qua
          </button>
        </div>
      </div>
      <div class="card-body">
        <div class="row g-3 align-items-center">
          <div class="col-auto">
            <label class="col-form-label" for="web_since_date_input">Từ:</label>
          </div>
          <div class="col-auto">
            <input
              type="date"
              class="form-control"
              id="web_since_date_input"
              bind:value={webSinceDate}
            />
          </div>
          <div class="col-auto">
            <input
              type="time"
              class="form-control"
              id="web_since_time_input"
              bind:value={webSinceTime}
              step="60"
            />
          </div>
          <div class="col-auto">
            <label class="col-form-label" for="web_until_date_input">Đến:</label
            >
          </div>
          <div class="col-auto">
            <input
              type="date"
              class="form-control"
              id="web_until_date_input"
              bind:value={webUntilDate}
            />
          </div>
          <div class="col-auto">
            <input
              type="time"
              class="form-control"
              id="web_until_time_input"
              bind:value={webUntilTime}
              step="60"
            />
          </div>
          <div class="col-auto">
            <button class="btn btn-primary" on:click={searchWebLogs}>
              Xác nhận
            </button>
          </div>
        </div>
      </div>
    </div>

    <h5 class="mt-4">Lịch sử truy cập</h5>
    <div id="web-log-items" class="list-group">
      {#if $webLogItems.length > 0}
        {#each $webLogItems as item (item.timestamp + item.url)}
          <label class="list-group-item d-flex align-items-center">
            <input
              class="form-check-input me-2"
              type="checkbox"
              name="web-log-domain"
              value={item.domain}
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
            <span class="text-muted ms-auto">{item.timestamp}</span>
          </label>
        {/each}
      {:else}
        <div class="list-group-item">Chưa có lịch sử truy cập web.</div>
      {/if}
    </div>
  </div>
</div>
