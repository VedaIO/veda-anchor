<script lang="ts">
  // Window runtime functions
  function minimize() {
    window.runtime.WindowMinimise();
  }

  function maximize() {
    window.runtime.WindowToggleMaximise();
  }

  function close() {
    window.runtime.Quit();
  }
</script>

<div class="title-bar" style="--wails-draggable: drag">
  <div class="title-drag-region">
    <!-- Optional: App Icon or Title could go here -->
    <!-- <span class="app-title">ProcGuard</span> -->
  </div>
  <div class="window-controls">
    <button
      class="control-btn minimize"
      on:click={minimize}
      title="Minimize"
      style="--wails-draggable: no-drag"
    >
      <svg width="10" height="1" viewBox="0 0 10 1">
        <path d="M0 0h10v1H0z" fill="currentColor" />
      </svg>
    </button>
    <button
      class="control-btn maximize"
      on:click={maximize}
      title="Maximize"
      style="--wails-draggable: no-drag"
    >
      <svg width="10" height="10" viewBox="0 0 10 10">
        <path d="M1 1h8v8H1V1zm1 1v6h6V2H2z" fill="currentColor" />
      </svg>
    </button>
    <button
      class="control-btn close"
      on:click={close}
      title="Close"
      style="--wails-draggable: no-drag"
    >
      <svg width="10" height="10" viewBox="0 0 10 10">
        <path
          d="M1.1 0L0 1.1 3.9 5 0 8.9 1.1 10 5 6.1 8.9 10 10 8.9 6.1 5 10 1.1 8.9 0 5 3.9z"
          fill="currentColor"
        />
      </svg>
    </button>
  </div>
</div>

<style>
  .title-bar {
    height: 32px; /* Compact height for title bar */
    width: 100%;
    background-color: #f2f0e3;
    display: flex;
    justify-content: flex-end; /* Controls on the right */
    align-items: center;
    user-select: none;
    -webkit-user-select: none;
    z-index: 9999; /* Always on top */
  }

  .title-drag-region {
    flex: 1; /* Takes up all remaining space */
    height: 100%;
  }

  .window-controls {
    display: flex;
    height: 100%;
  }

  .control-btn {
    width: 46px;
    height: 100%;
    border: none;
    background: transparent;
    color: #000;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: default;
    transition:
      background-color 0.1s,
      color 0.1s;
  }

  .control-btn:hover {
    background-color: rgba(255, 255, 255, 0.1);
    color: #000;
  }

  .control-btn.close:hover {
    background-color: #f76f53;
    color: #000;
  }
  .control-btn.maximize:hover {
    background-color: #000;
    color: #ffffff;
  }
  .control-btn.minimize:hover {
    background-color: #000;
    color: #ffffff;
  }
  .control-btn:active {
    background-color: #000;
    color: #ffffff;
  }

  .control-btn.close:active {
    color: #000;
  }
</style>
