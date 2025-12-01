<script lang="ts">
  import { navigate } from './router';
</script>

<div id="welcome-view">
  <div class="p-5 mb-4 bg-body-tertiary rounded-3">
    <div class="container-fluid py-5">
      <h1 class="display-5 fw-bold text-dark">Chào mừng đến với ProcGuard</h1>
      <p class="col-md-8 fs-4 text-dark">
        Đây là trung tâm điều khiển của bạn.<br />
        Từ đây, bạn có thể quản lý các ứng dụng và truy cập web được giám sát.
      </p>
    </div>
  </div>

  <div class="row align-items-md-stretch">
    <div class="col-md-6 mb-4">
      <div class="h-100 p-5 bg-body-tertiary border rounded-3">
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
      <div class="h-100 p-5 bg-body-tertiary border rounded-3">
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
</div>
