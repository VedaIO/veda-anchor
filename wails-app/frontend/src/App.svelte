<script lang="ts">
  import { currentPath, navigate } from './lib/router';
  import Welcome from './lib/Welcome.svelte';
  import AppManagement from './lib/AppManagement.svelte';
  import WebManagement from './lib/WebManagement.svelte';
  import Settings from './lib/Settings.svelte';
  import Login from './lib/Login.svelte';
  import { onMount } from 'svelte';
  import { isAuthenticated } from './lib/authStore';
  import Toast from './lib/Toast.svelte';
  import {
    isUninstallModalOpen,
    uninstallPassword,
    uninstallError,
    handleUninstallSubmit,
  } from './lib/modalStore';

  const routes: { [key: string]: any } = {
    '/': Welcome,
    '/apps': AppManagement,
    '/web': WebManagement,
    '/settings': Settings,
    '/login': Login,
  };

  import { checkExtension } from './lib/extensionStore';

  /**
   * Handle stopping the ProcGuard daemon completely
   * This is different from just closing the window - it stops background monitoring
   */
  async function handleStop() {
    if (confirm('Bạn có chắc chắn muốn dừng ProcGuard không?')) {
      try {
        await window.go.main.App.Stop();
        alert('ProcGuard đã được dừng.');
      } catch (error) {
        console.error('Lỗi khi dừng ProcGuard:', error);
        alert('Đã có lỗi xảy ra khi cố gắng dừng ProcGuard.');
      }
    }
  }

  /**
   * Handle user logout
   * Calls backend Logout method then navigates to login page
   * CRITICAL: Must call backend first to clear session, then update frontend state
   */
  async function handleLogout() {
    try {
      // Call backend to clear authentication session
      await window.go.main.App.Logout();
      // Update frontend state
      isAuthenticated.set(false);
      // Navigate to login page using hash routing
      navigate('/login');
    } catch (error) {
      console.error('Lỗi khi đăng xuất:', error);
      alert('Đã có lỗi xảy ra khi đăng xuất.');
    }
  }

  /**
   * Check authentication status and extension on mount
   * Extension polling starts automatically
   */
  onMount(async () => {
    // Check extension status (starts polling automatically)
    checkExtension();

    // Check if user is authenticated
    const authenticated = await window.go.main.App.GetIsAuthenticated();
    isAuthenticated.set(authenticated);

    // Redirect to login if not authenticated
    if (!$isAuthenticated && $currentPath !== '/login') {
      navigate('/login');
    }
  });
</script>

{#if $isAuthenticated}
  <nav class="navbar navbar-expand-lg navbar-light">
    <div class="container-fluid">
      <a
        class="navbar-brand"
        href="/"
        on:click|preventDefault={() => navigate('/')}>ProcGuard</a
      >
      <button
        class="navbar-toggler"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#navbarNav"
        aria-controls="navbarNav"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarNav">
        <ul class="navbar-nav me-auto">
          <li class="nav-item">
            <a
              class="nav-link"
              href="/"
              on:click|preventDefault={() => navigate('/')}>Trang chủ</a
            >
          </li>
          <li class="nav-item">
            <a
              class="nav-link"
              href="/settings"
              on:click|preventDefault={() => navigate('/settings')}>Cài đặt</a
            >
          </li>
        </ul>
        <ul class="navbar-nav">
          <li class="nav-item">
            <button class="nav-link btn" on:click={handleStop}
              >Dừng ProcGuard</button
            >
          </li>
          <li class="nav-item">
            <!-- Logout button - uses handleLogout() instead of href to properly clear session -->
            <button class="nav-link btn" on:click={handleLogout}
              >Đăng xuất</button
            >
          </li>
        </ul>
      </div>
    </div>
  </nav>

  <main class="container mt-4">
    <svelte:component this={routes[$currentPath]} />
  </main>

  <Toast />
{:else}
  <svelte:component this={routes['/login']} />
{/if}

<!-- Uninstall Modal -->
{#if $isUninstallModalOpen}
  <div class="modal-backdrop fade show"></div>
  <div
    class="modal fade show"
    style="display: block;"
    id="uninstall-modal"
    tabindex="-1"
    aria-labelledby="uninstallModalLabel"
    aria-hidden="false"
  >
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="uninstallModalLabel">
            Xác nhận gỡ cài đặt
          </h5>
          <button
            type="button"
            class="btn-close"
            on:click={() => isUninstallModalOpen.set(false)}
            aria-label="Close"
          ></button>
        </div>
        <div class="modal-body">
          <p>Vui lòng nhập mật khẩu của bạn để tiếp tục.</p>
          <form
            on:submit|preventDefault={() => {
              handleUninstallSubmit();
            }}
          >
            <div class="mb-3">
              <input
                type="password"
                class="form-control"
                id="uninstall-password"
                placeholder="Mật khẩu"
                required
                bind:value={$uninstallPassword}
              />
            </div>
            {#if $uninstallError}
              <p class="text-danger" style="display: block">
                {$uninstallError}
              </p>
            {/if}
            <button type="submit" class="btn btn-danger w-100">
              Xác nhận
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
{/if}
