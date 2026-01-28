<script lang="ts">
import { onMount } from 'svelte';
import { navigate } from './router';

interface ScreenTimeItem {
  name: string;
  executablePath: string;
  icon: string;
  durationSeconds: number;
}

let screenTimeData: ScreenTimeItem[] = [];
let totalScreenTime = 0;

// Format seconds to "Xh Xm" or "Xm Xs"
function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}p`;
  } else if (minutes > 0) {
    return `${minutes}p ${secs}s`;
  }
  return `${secs}s`;
}

async function loadScreenTime(): Promise<void> {
  try {
    const data = await window.go.main.App.GetScreenTime();
    screenTimeData = data || [];

    const total = await window.go.main.App.GetTotalScreenTime();
    totalScreenTime = total || 0;
  } catch (error) {
    console.error('Error loading screen time:', error);
    screenTimeData = [];
    totalScreenTime = 0;
  }
}

onMount(() => {
  loadScreenTime();
  // Refresh every 10 seconds
  const timer = setInterval(loadScreenTime, 10000);
  return () => clearInterval(timer);
});
</script>

<div id="welcome-view">
  <!-- Welcome Jumbotron -->
  <div class="p-3 p-md-5 mb-4 bg-body-tertiary rounded-3">
    <div class="container-fluid py-3 py-md-5">
      <h1 class="display-5 fw-bold text-dark">Chào mừng đến với ProcGuard</h1>
      <p class="col-md-8 fs-4 text-dark">
        Đây là trung tâm điều khiển của bạn.<br />
        Từ đây, bạn có thể quản lý các ứng dụng và truy cập web được giám sát.
      </p>
    </div>
  </div>

  <div class="row align-items-md-stretch">
    <div class="col-md-6 mb-4">
      <div class="h-100 p-3 p-md-5 bg-body-tertiary border rounded-3">
        <h2 class="text-dark">Quản lý Ứng dụng</h2>
        <p class="text-dark">
          Xem lại lịch sử sử dụng ứng dụng, chặn hoặc bỏ chặn các chương trình.
        </p>
        <button
          class="btn btn-outline-light"
          type="button"
          on:click={() => navigate('/apps')}
        >
          Đi tới Quản lý Ứng dụng
        </button>
      </div>
    </div>
    <div class="col-md-6 mb-4">
      <div class="h-100 p-3 p-md-5 bg-body-tertiary border rounded-3">
        <h2 class="text-dark">Quản lý Web</h2>
        <p class="text-dark">
          Xem lại lịch sử truy cập web và quản lý danh sách các trang web bị
          chặn.
        </p>
        <button
          class="btn btn-outline-light"
          type="button"
          on:click={() => navigate('/web')}
        >
          Đi tới Quản lý Web
        </button>
      </div>
    </div>
  </div>

  <!-- Screen Time Card -->
  <div class="card mb-4">
    <div class="card-body">
      <div class="d-flex justify-content-between align-items-center mb-3">
        <h5 class="card-title mb-0">
          <i class="bi bi-clock-history me-2"></i>Thời gian sử dụng ứng dụng hôm
          nay
        </h5>
        <span class="badge fs-6 text-black" style="background-color: #f76f53;"
          >{formatDuration(totalScreenTime)}</span
        >
      </div>

      {#if screenTimeData.length > 0}
        <div
          class="screen-time-list"
          style="max-height: 400px; overflow-y: auto;"
        >
          {#each screenTimeData as item (item.executablePath)}
            <div
              class="screen-time-item d-flex align-items-center py-2 border-bottom"
            >
              {#if item.icon}
                <img
                  src="data:image/png;base64,{item.icon}"
                  class="me-3"
                  style="width: 32px; height: 32px;"
                  alt="App Icon"
                />
              {:else}
                <div
                  class="me-3 bg-secondary rounded d-flex align-items-center justify-content-center"
                  style="width: 32px; height: 32px;"
                >
                  <i class="bi bi-app text-white"></i>
                </div>
              {/if}
              <div class="flex-grow-1">
                <div class="fw-semibold">{item.name}</div>
              </div>
              <div class="text-muted">
                {formatDuration(item.durationSeconds)}
              </div>
            </div>
          {/each}
        </div>
      {:else}
        <p class="text-muted mb-0">Chưa có dữ liệu thời gian sử dụng.</p>
      {/if}
    </div>
  </div>
</div>
