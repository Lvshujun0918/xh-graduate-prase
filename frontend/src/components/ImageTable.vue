<template>
  <div class="table-section">
    <!-- 工具栏 -->
    <div class="table-toolbar">
      <div class="toolbar-left">
        <label class="checkbox-wrapper">
          <input
            type="checkbox"
            :checked="isAllSelected"
            :indeterminate="isIndeterminate"
            @change="$emit('selectAll', !isAllSelected)"
          />
          <span class="checkbox-label">
            已选 <strong>{{ selectedCount }}</strong> / {{ images.length }} 张
          </span>
        </label>
      </div>
      <div class="toolbar-right">
        <button
          class="btn btn-primary"
          :disabled="selectedCount === 0"
          @click="$emit('download')"
        >
          <svg viewBox="0 0 20 20" width="18" height="18" fill="currentColor">
            <path d="M10 3a1 1 0 011 1v7.586l2.293-2.293a1 1 0 011.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L9 11.586V4a1 1 0 011-1z"/>
            <path d="M3 14a1 1 0 011 1v1h12v-1a1 1 0 112 0v2a1 1 0 01-1 1H4a1 1 0 01-1-1v-2a1 1 0 011-1z"/>
          </svg>
          下载选中 ({{ selectedCount }})
        </button>
      </div>
    </div>

    <!-- 表格 - 桌面端 -->
    <div class="table-wrapper hidden-mobile">
      <table class="data-table">
        <thead>
          <tr>
            <th class="col-check">
              <input
                type="checkbox"
                :checked="isAllSelected"
                @change="$emit('selectAll', !isAllSelected)"
              />
            </th>
            <th class="col-preview">预览</th>
            <th class="col-time">发送时间</th>
            <th class="col-sender">发件人</th>
            <th class="col-recipient">收件人</th>
            <th class="col-size">尺寸</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(img, index) in pagedImages"
            :key="index"
            :class="{ 'row-selected': selectedIndices.has(actualIndex(index)) }"
          >
            <td class="col-check">
              <input
                type="checkbox"
                :checked="selectedIndices.has(actualIndex(index))"
                @change="$emit('select', actualIndex(index))"
              />
            </td>
            <td class="col-preview">
              <div class="img-thumb" @click="$emit('preview', img)">
                <img
                  :src="getProxyUrl(img.serverImgUrl)"
                  :alt="img.fileName"
                  loading="lazy"
                  @error="handleImgError"
                />
              </div>
            </td>
            <td class="col-time">{{ img.datetime }}</td>
            <td class="col-sender">{{ img.fromName }}</td>
            <td class="col-recipient">{{ img.sessionId }}</td>
            <td class="col-size">{{ img.imgWidth }} × {{ img.imgHeight }}</td>
          </tr>
          <tr v-if="pagedImages.length === 0">
            <td colspan="6" class="empty-row">暂无图片数据</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 卡片列表 - 移动端 -->
    <div class="card-list visible-mobile">
      <div
        v-for="(img, index) in pagedImages"
        :key="index"
        class="image-card"
        :class="{ 'card-selected': selectedIndices.has(actualIndex(index)) }"
      >
        <div class="card-header">
          <input
            type="checkbox"
            :checked="selectedIndices.has(actualIndex(index))"
            @change="$emit('select', actualIndex(index))"
          />
          <span class="card-time">{{ img.datetime }}</span>
        </div>
        <div class="card-preview" @click="$emit('preview', img)">
          <img
            :src="getProxyUrl(img.serverImgUrl)"
            :alt="img.fileName"
            loading="lazy"
            @error="handleImgError"
          />
        </div>
        <div class="card-info">
          <div class="card-row">
            <span class="card-label">发件人</span>
            <span class="card-value">{{ img.fromName }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">收件人</span>
            <span class="card-value">{{ img.sessionId }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">尺寸</span>
            <span class="card-value">{{ img.imgWidth }} × {{ img.imgHeight }}</span>
          </div>
        </div>
      </div>
      <div v-if="pagedImages.length === 0" class="empty-card">
        暂无图片数据
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination" v-if="totalPages > 1">
      <button
        class="page-btn"
        :disabled="currentPage === 1"
        @click="currentPage--"
      >
        ‹ 上一页
      </button>
      <span class="page-info">
        第 {{ currentPage }} / {{ totalPages }} 页
        <span class="page-total">（共 {{ images.length }} 条）</span>
      </span>
      <button
        class="page-btn"
        :disabled="currentPage === totalPages"
        @click="currentPage++"
      >
        下一页 ›
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { getProxyImageUrl } from '../api/index.js'

const props = defineProps({
  images: { type: Array, required: true },
  taskId: { type: String, required: true },
  selectedIndices: { type: Set, required: true }
})

defineEmits(['select', 'selectAll', 'preview', 'download'])

const currentPage = ref(1)
const pageSize = 20

const totalPages = computed(() => Math.ceil(props.images.length / pageSize))
const pagedImages = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return props.images.slice(start, start + pageSize)
})

const selectedCount = computed(() => props.selectedIndices.size)
const isAllSelected = computed(() =>
  props.images.length > 0 && selectedCount.value === props.images.length
)
const isIndeterminate = computed(() =>
  selectedCount.value > 0 && selectedCount.value < props.images.length
)

function actualIndex(pageIndex) {
  return (currentPage.value - 1) * pageSize + pageIndex
}

function getProxyUrl(url) {
  return getProxyImageUrl(url)
}

function handleImgError(e) {
  e.target.style.display = 'none'
  e.target.parentElement.classList.add('img-error')
}
</script>

<style scoped>
.table-section {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

/* ===== Toolbar ===== */
.table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border-light);
  flex-wrap: wrap;
  gap: 0.75rem;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.checkbox-wrapper {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.checkbox-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.checkbox-label strong {
  color: var(--color-primary-dark);
}

.toolbar-right {
  display: flex;
  gap: 0.5rem;
}

/* ===== Buttons ===== */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 500;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
  box-shadow: 0 1px 3px rgba(99, 102, 241, 0.3);
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-primary-dark);
  box-shadow: 0 2px 6px rgba(99, 102, 241, 0.4);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ===== Table (Desktop) ===== */
.table-wrapper {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.data-table thead {
  background: var(--color-bg);
  position: sticky;
  top: 0;
}

.data-table th {
  padding: 0.65rem 0.75rem;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  white-space: nowrap;
  border-bottom: 2px solid var(--color-border);
}

.data-table td {
  padding: 0.65rem 0.75rem;
  border-bottom: 1px solid var(--color-border-light);
  vertical-align: middle;
}

.data-table tbody tr {
  transition: background 0.15s;
}

.data-table tbody tr:hover {
  background: var(--color-bg-hover);
}

.data-table tbody tr.row-selected {
  background: #EEF2FF;
}

.col-check {
  width: 40px;
  text-align: center;
}

.col-preview {
  width: 70px;
}

.col-time {
  width: 150px;
  white-space: nowrap;
}

.col-sender {
  min-width: 80px;
}

.col-recipient {
  min-width: 80px;
}

.col-size {
  width: 100px;
  white-space: nowrap;
}

/* Thumbnail */
.img-thumb {
  width: 50px;
  height: 50px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  cursor: pointer;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  transition: transform 0.2s;
}

.img-thumb:hover {
  transform: scale(1.08);
  box-shadow: var(--shadow-md);
}

.img-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.img-thumb.img-error::after {
  content: '🖼';
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  font-size: 1.2rem;
}

.empty-row {
  text-align: center;
  color: var(--color-text-muted);
  padding: 2rem !important;
}

/* Checkbox styling */
input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--color-primary);
  cursor: pointer;
}

/* ===== Card List (Mobile) ===== */
.visible-mobile {
  display: none;
}

.card-list {
  padding: 0.75rem;
}

.image-card {
  background: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md);
  margin-bottom: 0.75rem;
  overflow: hidden;
  transition: all 0.2s;
}

.image-card.card-selected {
  border-color: var(--color-primary);
  background: #EEF2FF;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--color-border-light);
}

.card-time {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.card-preview {
  height: 180px;
  overflow: hidden;
  cursor: pointer;
  background: var(--color-bg);
}

.card-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-info {
  padding: 0.5rem 0.75rem;
}

.card-row {
  display: flex;
  justify-content: space-between;
  padding: 0.2rem 0;
  font-size: 0.8rem;
}

.card-label {
  color: var(--color-text-muted);
}

.card-value {
  font-weight: 500;
  color: var(--color-text);
}

.empty-card {
  text-align: center;
  padding: 2rem;
  color: var(--color-text-muted);
}

/* ===== Pagination ===== */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--color-border-light);
}

.page-btn {
  padding: 0.35rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-card);
  color: var(--color-text-secondary);
  font-size: 0.825rem;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.825rem;
  color: var(--color-text-secondary);
}

.page-total {
  color: var(--color-text-muted);
}

/* ===== Responsive ===== */
@media (max-width: 768px) {
  .hidden-mobile {
    display: none;
  }

  .visible-mobile {
    display: block;
  }

  .table-toolbar {
    padding: 0.5rem 0.75rem;
  }
}

@media (min-width: 769px) {
  .hidden-mobile {
    display: block;
  }

  .visible-mobile {
    display: none;
  }
}
</style>
