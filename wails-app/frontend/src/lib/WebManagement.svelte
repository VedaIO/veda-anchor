<script lang="ts">
  import { isExtensionInstalled } from './extensionStore';
  import WebLeaderboard from './WebLeaderboard.svelte';
  import WebLog from './WebLog.svelte';
  import WebBlocklist from './WebBlocklist.svelte';

  let activeTab: 'leaderboard' | 'log' | 'blocklist' = 'leaderboard';

  function showSubView(view: 'leaderboard' | 'log' | 'blocklist') {
    activeTab = view;
  }
</script>

<div id="web-management-view">
  {#if $isExtensionInstalled === false}
    <!-- "Not Installed" View -->
    <div id="web-extension-not-installed-view" class="text-center">
      <div class="card mt-3">
        <div class="card-body">
          <h5 class="card-title">Đã mất kết nối với tiện ích mở rộng</h5>
          <p>
            Để sử dụng tính năng quản lý web, vui lòng đảm bảo tiện ích
            ProcGuard đã được cài đặt và trình duyệt đang mở.
          </p>
          <p class="text-muted small">
            Nếu bạn vừa đóng trình duyệt, kết nối sẽ tự động được khôi phục khi
            bạn mở lại.
          </p>
          <button
            type="button"
            class="btn btn-primary"
            id="install-extension-btn-web"
            on:click={async () => {
              try {
                await window.go.main.App.OpenBrowser(
                  'https://chromewebstore.google.com/detail/procguard-web-monitor/hkanepohpflociaodcicmmfbdaohpceo'
                );
              } catch (err) {
                console.error('Failed to open browser:', err);
              }
            }}>Cài đặt tiện ích</button
          >
        </div>
      </div>
    </div>
  {:else if $isExtensionInstalled === true}
    <!-- Tab Navigation -->
    <ul class="nav nav-tabs" id="webManTabs" role="tablist">
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          class:active={activeTab === 'leaderboard'}
          id="web-leaderboard-tab"
          type="button"
          role="tab"
          on:click={() => showSubView('leaderboard')}
        >
          Bảng xếp hạng
        </button>
      </li>
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          class:active={activeTab === 'log'}
          id="web-log-tab"
          type="button"
          role="tab"
          on:click={() => showSubView('log')}
        >
          Lịch sử Web
        </button>
      </li>
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          class:active={activeTab === 'blocklist'}
          id="web-blocklist-tab"
          type="button"
          role="tab"
          on:click={() => showSubView('blocklist')}
        >
          Quản lý danh sách chặn
        </button>
      </li>
    </ul>

    <!-- Tab Content -->
    <div class="tab-content" id="webManTabsContent">
      {#if activeTab === 'leaderboard'}
        <div id="web-leaderboard-view" role="tabpanel">
          <WebLeaderboard />
        </div>
      {:else if activeTab === 'log'}
        <div id="web-log-view" role="tabpanel">
          <WebLog />
        </div>
      {:else if activeTab === 'blocklist'}
        <div id="web-blocklist-view" role="tabpanel">
          <WebBlocklist />
        </div>
      {/if}
    </div>
  {:else}
    <div class="text-center mt-5">
      <div class="spinner-border text-danger" role="status">
        <span class="visually-hidden">Đang tải...</span>
      </div>
      <p class="mt-2">Kiểm tra trạng thái tiện ích...</p>
    </div>
  {/if}
</div>
