<script lang="ts">
import { onMount } from 'svelte';
import { showToast } from './toastStore';

interface AppLeaderboardItem {
  rank: number;
  name: string;
  processName: string;
  icon: string;
  count: number;
}

let leaderboardData: AppLeaderboardItem[] = [];

async function loadAppLeaderboard(since = '', until = ''): Promise<void> {
  try {
    console.log('Fetching app leaderboard...', { since, until });
    const data = await window.go.main.App.GetAppLeaderboard(since, until);
    console.log('App leaderboard data received:', data);
    if (data && data.length > 0) {
      leaderboardData = data;
    } else {
      leaderboardData = [];
    }
  } catch (error) {
    console.error('Error loading app leaderboard:', error);
    leaderboardData = [];
  }
}

async function blockApp(
  processName: string,
  displayName: string,
): Promise<void> {
  try {
    await window.go.main.App.BlockApps([processName]);
    showToast(`Đã chặn ${displayName}`, 'success');
    loadAppLeaderboard(); // Refresh
  } catch (error) {
    console.error('Error blocking app:', error);
    showToast('Lỗi khi chặn ứng dụng.', 'error');
  }
}

onMount(() => {
  loadAppLeaderboard();
  const pollingTimer = setInterval(() => {
    loadAppLeaderboard();
  }, 5000); // 5 seconds

  return () => clearInterval(pollingTimer);
});
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Bảng xếp hạng ứng dụng</h5>
    <div id="app-leaderboard-table-container" class="table-responsive">
      {#if leaderboardData.length > 0}
        <table class="table table-hover">
          <thead>
            <tr>
              <th scope="col">Xếp hạng</th>
              <th scope="col">Ứng dụng</th>
              <th scope="col">Số lần dùng</th>
              <th scope="col">Hành động</th>
            </tr>
          </thead>
          <tbody>
            {#each leaderboardData as item (item.name)}
              <tr>
                <th scope="row"
                  ><span class="badge bg-primary">{item.rank}</span></th
                >
                <td>
                  {#if item.icon}
                    <img
                      src="data:image/png;base64,{item.icon}"
                      class="me-2"
                      style="width: 24px; height: 24px;"
                      alt="App Icon"
                    />
                  {:else}
                    <div class="me-2" style="width: 24px; height: 24px;"></div>
                  {/if}
                  <span class="fw-bold">{item.name}</span>
                </td>
                <td>{item.count}</td>
                <td>
                  <button
                    type="button"
                    class="btn btn-sm btn-danger"
                    on:click={() => blockApp(item.processName, item.name)}
                  >
                    Chặn
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {:else}
        <div class="list-group-item">No data for leaderboard.</div>
      {/if}
    </div>
  </div>
</div>
