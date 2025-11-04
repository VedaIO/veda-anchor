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

  onMount(async () => {
    checkExtension();
    const res = await fetch('/api/is-authenticated');
    const data = await res.json();
    isAuthenticated.set(data.authenticated);

    if (!$isAuthenticated && window.location.pathname !== '/login') {
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
            <a class="nav-link" href="/logout">Đăng xuất</a>
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
