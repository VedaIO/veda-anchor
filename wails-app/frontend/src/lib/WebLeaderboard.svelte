<script lang="ts">
import { onMount } from 'svelte';
import { showToast } from './toastStore';

interface WebLeaderboardItem {
  rank: number;
  domain: string;
  title: string;
  icon: string;
  count: number;
}

let leaderboardData: WebLeaderboardItem[] = [];

async function loadWebLeaderboard(since = '', until = ''): Promise<void> {
  try {
    const data = await window.go.main.App.GetWebLeaderboard(since, until);
    if (data && data.length > 0) {
      leaderboardData = data;
    } else {
      leaderboardData = [];
    }
  } catch (error) {
    console.error('Error loading web leaderboard:', error);
    leaderboardData = [];
  }
}

async function blockDomain(domain: string): Promise<void> {
  try {
    await window.go.main.App.AddWebBlocklist(domain);
    showToast(`Đã chặn ${domain}`, 'success');
    loadWebLeaderboard(); // Refresh
  } catch (error) {
    console.error('Error blocking domain:', error);
    showToast('Lỗi khi chặn trang web.', 'error');
  }
}

onMount(() => {
  loadWebLeaderboard();
  const pollingTimer = setInterval(() => {
    loadWebLeaderboard();
  }, 5000); // 5 seconds

  return () => clearInterval(pollingTimer);
});
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Bảng xếp hạng Web</h5>
    <div id="web-leaderboard-table-container" class="table-responsive">
      {#if leaderboardData.length > 0}
        <table class="table table-hover">
          <thead>
            <tr>
              <th scope="col">Xếp hạng</th>
              <th scope="col">Trang Web</th>
              <th scope="col">Số lần mở</th>
              <th scope="col">Hành động</th>
            </tr>
          </thead>
          <tbody>
            {#each leaderboardData as item (item.domain)}
              <tr>
                <th scope="row"
                  ><span class="badge bg-primary">{item.rank}</span></th
                >
                <td>
                  {#if item.icon}
                    <img
                      src={item.icon}
                      class="me-2"
                      style="width: 24px; height: 24px;"
                      alt="Website Icon"
                    />
                  {:else}
                    <div class="me-2" style="width: 24px; height: 24px;"></div>
                  {/if}
                  <span class="fw-bold">{item.title || item.domain}</span>
                </td>
                <td>{item.count}</td>
                <td>
                  <button
                    type="button"
                    class="btn btn-sm btn-danger"
                    on:click={() => blockDomain(item.domain)}
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
