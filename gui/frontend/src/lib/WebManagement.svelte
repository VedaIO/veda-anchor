<script lang="ts">
  import { isExtensionInstalled } from './extensionStore';
  import WebLeaderboard from './WebLeaderboard.svelte';
  import WebLog from './WebLog.svelte';
  import WebBlocklist from './WebBlocklist.svelte';
  import { showToast } from './toastStore';

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
          <h5 class="card-title">Tiện ích mở rộng chưa được cài đặt</h5>
          <p>
            Để theo dõi và chặn các trang web, bạn phải tải tiện ích trình duyệt
            mở rộng của ProcGuard.
          </p>
          <button
            type="button"
            class="btn btn-primary"
            id="install-extension-btn-web"
            on:click={() =>
              showToast(
                'Vui lòng chờ tiện ích khả dụng trên cửa hàng.',
                'info'
              )}>Tải tiện ích</button
          >
          <button
            class="btn btn-secondary"
            id="reload-extension-check-btn"
            on:click={() => location.reload()}
          >
            Tôi đã tải nó, Tải lại
          </button>
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
