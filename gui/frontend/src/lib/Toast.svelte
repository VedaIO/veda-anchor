<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from './toastStore';
  import { Toast } from 'bootstrap';

  let toastElement: HTMLElement;
  let toastInstance: Toast;

  onMount(() => {
    toastInstance = new Toast(toastElement, { autohide: true, delay: 3000 });
  });

  toast.subscribe((msg) => {
    if (msg && toastInstance) {
      toastInstance.show();
    }
  });
</script>

<div class="toast-container position-fixed top-0 end-0 p-3">
  <div
    bind:this={toastElement}
    class="toast"
    role="alert"
    aria-live="assertive"
    aria-atomic="true"
  >
    {#if $toast}
      <div
        class="toast-header text-black"
        class:bg-dark={$toast.type === 'info'}
        class:bg-danger={$toast.type === 'error'}
        style={$toast.type === 'success' ? 'background-color: #f76f53;' : ''}
      >
        <strong class="me-auto">Thông báo</strong>
        <button
          type="button"
          class="btn-close btn-close-white"
          data-bs-dismiss="toast"
          aria-label="Close"
        ></button>
      </div>
      <div class="toast-body">
        {$toast.message}
      </div>
    {/if}
  </div>
</div>
