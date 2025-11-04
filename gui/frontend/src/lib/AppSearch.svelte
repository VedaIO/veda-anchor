<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import SearchResultItem from './SearchResultItem.svelte';
  import { showToast } from './toastStore';

  interface SearchResultItem {
    processName: string;
    exePath: string;
    commercialName: string;
    icon: string;
    otherInfo: string;
  }

  let q = '';
  let sinceDate = '';
  let sinceTime = '';
  let untilDate = '';
  let untilTime = '';
  let searchResults = writable<SearchResultItem[]>([]);
  let selectedApps: string[] = [];

  async function search(range?: {
    since: string;
    until: string;
  }): Promise<void> {
    let since = '';
    let until = '';

    if (range) {
      since = range.since;
      until = range.until;
    } else {
      if (sinceDate && sinceTime) {
        since = `${sinceDate}T${sinceTime}`;
      }
      if (untilDate && untilTime) {
        until = `${untilDate}T${untilTime}`;
      }
    }

    let url = '/api/search?q=' + encodeURIComponent(q);
    if (since) {
      url += '&since=' + encodeURIComponent(since);
    }
    if (until) {
      url += '&until=' + encodeURIComponent(until);
    }
    const res = await fetch(url, { cache: 'no-cache' });
    const data = await res.json();
    if (data && data.length > 0) {
      const items: SearchResultItem[] = await Promise.all(
        data.map(async (l: string[]) => {
          const processName = l[1];
          const exePath = l[4]; // exe_path is the 5th element
          let commercialName = '';
          let icon = '';

          if (exePath) {
            const appDetailsRes = await fetch(
              `/api/app-details?path=${encodeURIComponent(exePath)}`
            );
            if (appDetailsRes.ok) {
              const appDetails = await appDetailsRes.json();
              commercialName = appDetails.commercialName;
              icon = appDetails.icon;
            }
          }

          const otherInfo = l.filter((_, i) => i !== 1 && i !== 4).join(' | ');

          return {
            processName,
            exePath,
            commercialName,
            icon,
            otherInfo,
          };
        })
      );
      searchResults.set(items);
    } else {
      searchResults.set([]);
    }
  }

  async function block(): Promise<void> {
    if (selectedApps.length === 0) {
      showToast(
        'Vui lòng chọn một ứng dụng từ kết quả tìm kiếm để chặn.',
        'info'
      );
      return;
    }

    // Remove duplicates
    const uniqueApps = [...new Set(selectedApps)];

    const response = await fetch('/api/block', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ names: uniqueApps }),
    });

    if (!response.ok) {
      showToast(`Lỗi chặn ứng dụng: ${response.statusText}`, 'error');
      return;
    }

    showToast(`Đã chặn: ${uniqueApps.join(', ')}`, 'success');
    selectedApps = []; // Clear selection
  }

  onMount(() => {
    // Set default dates
    const now = new Date();
    const year = now.getFullYear();
    const month = (now.getMonth() + 1).toString().padStart(2, '0');
    const day = now.getDate().toString().padStart(2, '0');
    const today = `${year}-${month}-${day}`;
    sinceDate = today;
    untilDate = today;

    // Perform initial search
    search();
  });
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Tìm kiếm Log ứng dụng</h5>
    <div class="input-group mb-3">
      <input
        type="text"
        class="form-control"
        id="q"
        placeholder="Nhập tên ứng dụng..."
        bind:value={q}
      />
      <button class="btn btn-primary" type="button" on:click={search}>
        Tìm kiếm
      </button>
      <button class="btn btn-danger" type="button" on:click={block}>
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
            on:click={() => search({ since: '1 hour ago', until: 'now' })}
          >
            1 giờ qua
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() => search({ since: '24 hours ago', until: 'now' })}
          >
            24 giờ qua
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() => search({ since: '7 days ago', until: 'now' })}
          >
            7 ngày qua
          </button>
        </div>
      </div>
      <div class="card-body">
        <div class="row g-3 align-items-center">
          <div class="col-auto">
            <label class="col-form-label" for="since_date_input">Từ:</label>
          </div>
          <div class="col-auto">
            <input
              type="date"
              class="form-control"
              id="since_date_input"
              bind:value={sinceDate}
            />
          </div>
          <div class="col-auto">
            <input
              type="time"
              class="form-control"
              id="since_time_input"
              bind:value={sinceTime}
              step="60"
            />
          </div>
          <div class="col-auto">
            <label class="col-form-label" for="until_date_input">Đến:</label>
          </div>
          <div class="col-auto">
            <input
              type="date"
              class="form-control"
              id="until_date_input"
              bind:value={untilDate}
            />
          </div>
          <div class="col-auto">
            <input
              type="time"
              class="form-control"
              id="until_time_input"
              bind:value={untilTime}
              step="60"
            />
          </div>
          <div class="col-auto">
            <button class="btn btn-primary" on:click={search}>
              Xác nhận
            </button>
          </div>
        </div>
      </div>
    </div>

    <h5 class="mt-4">Kết quả tìm kiếm</h5>
    <div id="results" class="list-group">
      {#each $searchResults as item, i (item.processName + i)}
        <label class="list-group-item d-flex align-items-center">
          <input
            class="form-check-input me-2"
            type="checkbox"
            name="search-result-app"
            value={item.processName}
            bind:group={selectedApps}
          />
          <SearchResultItem {item} />
        </label>
      {:else}
        <div class="list-group-item">Không tìm thấy kết quả.</div>
      {/each}
    </div>
  </div>
</div>
