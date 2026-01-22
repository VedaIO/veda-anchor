<script lang="ts">
import { onMount } from 'svelte';
import { writable } from 'svelte/store';
import DateRangePicker from './DateRangePicker.svelte';
import { showToast } from './toastStore';

interface WebLogItem {
  timestamp: string;
  url: string;
  domain: string;
  title: string;
  iconUrl: string;
}

let q = '';
const webLogItems = writable<WebLogItem[]>([]);
let since: Date | null = null;
let until: Date | null = new Date();

function formatDateTime(date: Date | null): string {
  if (!date) return '';
  return date.toISOString();
}

async function loadWebLogs(
  query: string,
  sinceStr: string,
  untilStr: string,
): Promise<void> {
  console.log('loadWebLogs called with:', { query, sinceStr, untilStr });
  try {
    const data = await window.go.main.App.GetWebLogs(query, sinceStr, untilStr);
    console.log('GetWebLogs returned:', data);

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
            try {
              const webDetails = await window.go.main.App.GetWebDetails(domain);
              title = webDetails.title;
              iconUrl = webDetails.iconUrl;
            } catch (error) {
              console.error('Error fetching web details:', error);
            }
          }

          const timestamp = l[0];

          return {
            timestamp,
            url: urlString,
            domain,
            title,
            iconUrl,
          };
        }),
      );
      console.log('Processed items:', items);
      webLogItems.set(items);
    } else {
      console.log('No data returned');
      webLogItems.set([]);
    }
  } catch (error) {
    console.error('Error loading web logs:', error);
    webLogItems.set([]);
  }
}

function handleDateChange(
  event: CustomEvent<{ since: Date | null; until: Date | null }>,
) {
  since = event.detail.since;
  until = event.detail.until;
  loadWebLogs(q, formatDateTime(since), formatDateTime(until));
}

async function blockSelectedWebsites(): Promise<void> {
  const selectedDomains = Array.from(
    document.querySelectorAll('input[name="web-log-domain"]:checked'),
  ).map((cb) => (cb as HTMLInputElement).value);
  if (selectedDomains.length === 0) {
    showToast('Vui lòng chọn một trang web để chặn.', 'info');
    return;
  }

  const uniqueDomains = [...new Set(selectedDomains)];

  try {
    for (const domain of uniqueDomains) {
      await window.go.main.App.AddWebBlocklist(domain);
    }

    showToast(
      'Các trang web đã chọn đã được thêm vào danh sách chặn.',
      'success',
    );
    (
      document.querySelectorAll(
        'input[name="web-log-domain"]:checked',
      ) as NodeListOf<HTMLInputElement>
    ).forEach((cb) => {
      cb.checked = false;
    });
  } catch (error) {
    console.error('Error blocking websites:', error);
    showToast('Lỗi khi chặn trang web.', 'error');
  }
}

onMount(() => {
  const now = new Date();
  since = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000); // 7 days ago
  until = new Date();
  loadWebLogs(q, formatDateTime(since), formatDateTime(until));
});
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Lịch sử Truy cập Web</h5>
    <div class="input-group mb-3">
      <input
        type="text"
        class="form-control"
        id="q"
        placeholder="Nhập tên trang web..."
        bind:value={q}
        on:input={() =>
          loadWebLogs(q, formatDateTime(since), formatDateTime(until))}
      />
      <button class="btn btn-danger" on:click={blockSelectedWebsites}>
        Chặn mục đã chọn
      </button>
    </div>

    <div class="card mt-3">
      <div class="card-header d-flex align-items-center">
        <span class="me-3">Lọc theo thời gian định sẵn:</span>
        <div class="btn-group" role="group">
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() => loadWebLogs(q, '1 hour ago', 'now')}
          >
            1 giờ qua
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() => loadWebLogs(q, '24 hours ago', 'now')}
          >
            24 giờ qua
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() => loadWebLogs(q, '7 days ago', 'now')}
          >
            7 ngày qua
          </button>
        </div>
      </div>
      <div class="card-body">
        <span class="me-3">Lọc theo thời gian bất kỳ:</span>
        <DateRangePicker {since} {until} on:change={handleDateChange} />
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
