<template>
  <div class="timeline-container">
    <!-- 顶部操作栏 -->
    <div class="timeline-toolbar">
      <div class="toolbar-stats">
        <div class="stat-item">
          <span class="stat-value">{{ images.length }}</span>
          <span class="stat-label">图片</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-value">{{ result.sender }}</span>
          <span class="stat-label">发件人</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-value">{{ result.recipients }}</span>
          <span class="stat-label">会话</span>
        </div>
      </div>
      <div class="toolbar-actions">
        <button
          v-if="selectedCount > 0"
          class="action-btn action-btn--clear"
          @click="$emit('clearSelection')"
        >
          取消选择
        </button>
        <button
          class="action-btn action-btn--select"
          @click="toggleSelectAll"
        >
          {{ isAllSelected ? '取消全选' : '全选' }}
        </button>
        <button
          class="action-btn action-btn--primary"
          :disabled="selectedCount === 0"
          @click="$emit('download')"
        >
          <svg viewBox="0 0 20 20" width="16" height="16" fill="currentColor">
            <path d="M10 3a1 1 0 011 1v7.586l2.293-2.293a1 1 0 011.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L9 11.586V4a1 1 0 011-1z"/>
            <path d="M3 14a1 1 0 011 1v1h12v-1a1 1 0 112 0v2a1 1 0 01-1 1H4a1 1 0 01-1-1v-2a1 1 0 011-1z"/>
          </svg>
          下载{{ selectedCount > 0 ? ` (${selectedCount})` : '' }}
        </button>
      </div>
    </div>

    <!-- 选中提示条 -->
    <transition name="slide">
      <div v-if="selectionMode && isMobile" class="selection-banner">
        <span>选择模式 · 点击图片进行多选</span>
        <button @click="exitSelectionMode">完成</button>
      </div>
    </transition>

    <!-- 时间线 -->
    <div class="timeline-scroll" ref="scrollContainer" @scroll="handleScroll">
      <!-- 空状态 -->
      <div v-if="images.length === 0" class="timeline-empty">
        <div class="empty-icon">
          <svg viewBox="0 0 48 48" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="8" y="12" width="32" height="24" rx="3"/>
            <circle cx="16" cy="20" r="3"/>
            <path d="M8 30l10-8 6 5 8-9 8 7"/>
          </svg>
        </div>
        <p>暂无图片数据</p>
      </div>

      <!-- 时间线条目 -->
      <div
        v-for="(group, gIdx) in visibleGroups"
        :key="gIdx"
        class="timeline-group"
      >
        <!-- 日期分隔线 -->
        <div class="timeline-date-separator">
          <span class="date-badge">{{ group.date }}</span>
          <span class="date-count">{{ group.items.length }} 张</span>
        </div>

        <!-- 图片网格 -->
        <div class="timeline-grid">
          <div
            v-for="item in group.items"
            :key="item.index"
            class="timeline-card"
            :class="{
              'card--selected': selectedIndices.has(item.index),
              'card--selection-mode': selectionMode && isMobile
            }"
            @click="handleCardClick(item.index, $event)"
            @contextmenu.prevent="handleContextMenu(item.index)"
          >
            <!-- 选择指示器 -->
            <div class="card-check" :class="{ 'check--visible': selectedIndices.has(item.index) || selectionMode }">
              <div class="check-circle" :class="{ 'check--active': selectedIndices.has(item.index) }">
                <svg v-if="selectedIndices.has(item.index)" viewBox="0 0 24 24" width="16" height="16" fill="white">
                  <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                </svg>
              </div>
            </div>

            <!-- 缩略图 -->
            <div class="card-image" @click.stop="$emit('preview', item)">
              <img
                :src="getProxyUrl(item.serverImgUrl)"
                :alt="item.fileName"
                loading="lazy"
                @error="onImgError"
              />
              <div class="image-overlay">
                <svg viewBox="0 0 24 24" width="24" height="24" fill="white">
                  <circle cx="12" cy="12" r="10" stroke="white" stroke-width="1.5" fill="none"/>
                  <path d="M12 8v8M8 12h8" stroke="white" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
              </div>
            </div>

            <!-- 信息区 -->
            <div class="card-info">
              <div class="card-time">{{ formatTime(item.datetime) }}</div>
              <div class="card-meta">
                <span class="meta-tag">{{ item.fromName }}</span>
                <span class="meta-dim">{{ item.imgWidth }}×{{ item.imgHeight }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 加载更多触发器 -->
      <div
        v-if="hasMore"
        ref="loadTrigger"
        class="timeline-loader"
      >
        <div class="loader-dot"></div>
        <span>加载更多...</span>
      </div>

      <!-- 全部加载完毕 -->
      <div v-if="!hasMore && images.length > 0" class="timeline-end">
        <span>— 已展示全部 {{ images.length }} 张图片 —</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { getProxyImageUrl } from '../api/index.js'

const props = defineProps({
  result: { type: Object, required: true },
  images: { type: Array, required: true },
  taskId: { type: String, required: true },
  selectedIndices: { type: Set, required: true }
})

const emit = defineEmits(['select', 'selectRange', 'preview', 'download', 'clearSelection', 'selectAll'])

// ---- 无极滚动 ----
const BATCH_SIZE = 40
const displayCount = ref(BATCH_SIZE)
const scrollContainer = ref(null)
const loadTrigger = ref(null)

const hasMore = computed(() => displayCount.value < props.images.length)

// 按日期分组可见图片
const visibleGroups = computed(() => {
  const visible = props.images.slice(0, displayCount.value)
  const groups = []
  let currentDate = ''
  let currentGroup = null

  for (const img of visible) {
    const date = img.datetime.split(' ')[0]
    if (date !== currentDate) {
      currentGroup = { date, items: [] }
      groups.push(currentGroup)
      currentDate = date
    }
    currentGroup.items.push(img)
  }
  return groups
})

function loadMore() {
  if (hasMore.value) {
    displayCount.value = Math.min(displayCount.value + BATCH_SIZE, props.images.length)
  }
}

// Intersection Observer 自动加载
let observer = null
function setupObserver() {
  if (observer) observer.disconnect()
  if (!loadTrigger.value) return

  observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting && hasMore.value) {
      loadMore()
    }
  }, { root: scrollContainer.value, rootMargin: '200px' })

  observer.observe(loadTrigger.value)
}

// 手动滚动兜底
function handleScroll() {
  if (!scrollContainer.value || !hasMore.value) return
  const { scrollTop, scrollHeight, clientHeight } = scrollContainer.value
  if (scrollHeight - scrollTop - clientHeight < 400) {
    loadMore()
  }
}

watch(displayCount, () => {
  nextTick(() => setupObserver())
})

onMounted(() => {
  nextTick(() => setupObserver())
})

onUnmounted(() => {
  if (observer) observer.disconnect()
})

// ---- 选择逻辑 ----
const isMobile = ref(false)
const selectionMode = ref(false)
const lastClickedIndex = ref(-1)
let longPressTimer = null

function detectMobile() {
  isMobile.value = window.innerWidth < 768
}

function handleCardClick(index, event) {
  if (selectionMode.value && isMobile.value) {
    // 移动端选择模式：点击切换选中
    emit('select', index)
    return
  }

  if (event.shiftKey && lastClickedIndex.value >= 0) {
    // Shift+点击：区间选择
    const start = Math.min(lastClickedIndex.value, index)
    const end = Math.max(lastClickedIndex.value, index)
    const rangeIndices = []
    for (let i = start; i <= end; i++) {
      rangeIndices.push(i)
    }
    emit('selectRange', rangeIndices)
    return
  }

  // 普通点击：切换单个
  emit('select', index)
  lastClickedIndex.value = index
}

function handleContextMenu(index) {
  if (!selectionMode.value && isMobile.value) {
    enterSelectionMode()
    emit('select', index)
  }
}

function enterSelectionMode() {
  selectionMode.value = true
}

function exitSelectionMode() {
  selectionMode.value = false
}

// 长按进入选择模式（移动端）
function onTouchStart(index, event) {
  if (isMobile.value && !selectionMode.value) {
    longPressTimer = setTimeout(() => {
      enterSelectionMode()
      emit('select', index)
    }, 500)
  }
}

function onTouchEnd() {
  clearTimeout(longPressTimer)
}

// ---- 全选 ----
const isAllSelected = computed(() =>
  props.images.length > 0 && props.selectedIndices.size === props.images.length
)

function toggleSelectAll() {
  emit('selectAll', !isAllSelected.value)
}

// ---- 工具函数 ----
function getProxyUrl(url) {
  return getProxyImageUrl(url)
}

function formatTime(datetime) {
  const parts = datetime.split(' ')
  if (parts.length >= 2) {
    return parts[1] // 只显示时分秒
  }
  return datetime
}

function onImgError(e) {
  e.target.style.display = 'none'
}

// 监听窗口大小
onMounted(() => {
  detectMobile()
  window.addEventListener('resize', detectMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', detectMobile)
})

// 监听 result 变化时重置
watch(() => props.result?.taskId, () => {
  displayCount.value = BATCH_SIZE
  lastClickedIndex.value = -1
  selectionMode.value = false
})
</script>

<style scoped>
/* ========== 容器 ========== */
.timeline-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 180px);
  min-height: 500px;
  background: var(--color-bg);
}

/* ========== 工具栏 ========== */
.timeline-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 100;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.toolbar-stats {
  display: flex;
  align-items: center;
  gap: 0;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 0.75rem;
}

.stat-value {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text);
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}

.stat-label {
  font-size: 0.675rem;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-top: 0.1rem;
}

.stat-divider {
  width: 1px;
  height: 28px;
  background: var(--color-border);
}

.toolbar-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.45rem 0.85rem;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 500;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  white-space: nowrap;
}

.action-btn--select {
  background: var(--color-bg);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}

.action-btn--select:hover {
  background: var(--color-bg-hover);
  border-color: var(--color-text-muted);
}

.action-btn--clear {
  background: #FEF2F2;
  color: #DC2626;
  border: 1px solid #FECACA;
}

.action-btn--clear:hover {
  background: #FEE2E2;
}

.action-btn--primary {
  background: linear-gradient(135deg, #6366F1, #4F46E5);
  color: white;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
}

.action-btn--primary:hover:not(:disabled) {
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.4);
  transform: translateY(-1px);
}

.action-btn--primary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  box-shadow: none;
}

/* ========== 选择模式横幅（移动端） ========== */
.selection-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  background: #EEF2FF;
  border-bottom: 1px solid #C7D2FE;
  font-size: 0.8rem;
  color: var(--color-primary-dark);
}

.selection-banner button {
  background: var(--color-primary);
  color: white;
  border: none;
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
  font-size: 0.75rem;
  cursor: pointer;
  font-family: inherit;
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-100%);
}

/* ========== 滚动区域 ========== */
.timeline-scroll {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 1rem 1.25rem 2rem;
  scroll-behavior: smooth;
}

.timeline-scroll::-webkit-scrollbar {
  width: 5px;
}
.timeline-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.timeline-scroll::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: 10px;
}

/* ========== 空状态 ========== */
.timeline-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 1rem;
  color: var(--color-text-muted);
}

.empty-icon {
  margin-bottom: 1rem;
  opacity: 0.4;
}

.timeline-empty p {
  font-size: 0.95rem;
}

/* ========== 日期分组 ========== */
.timeline-group {
  margin-bottom: 1.5rem;
}

.timeline-date-separator {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  padding: 0.5rem 0;
  margin-bottom: 0.75rem;
  position: sticky;
  top: -16px;
  background: var(--color-bg);
  z-index: 10;
}

.date-badge {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text);
  background: var(--color-bg-card);
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-border-light);
}

.date-count {
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

/* ========== 图片网格 ========== */
.timeline-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
  gap: 0.75rem;
}

/* ========== 图片卡片 ========== */
.timeline-card {
  background: var(--color-bg-card);
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  box-shadow: var(--shadow-sm);
}

.timeline-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  border-color: var(--color-border);
}

.timeline-card.card--selected {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
  background: #F5F3FF;
}

.timeline-card.card--selection-mode {
  border-color: var(--color-border);
}

.timeline-card.card--selection-mode:hover {
  border-color: var(--color-primary-light);
}

/* 选择复选框 */
.card-check {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 5;
  opacity: 0;
  transform: scale(0.8);
  transition: all 0.2s ease;
}

.timeline-card:hover .card-check,
.check--visible {
  opacity: 1;
  transform: scale(1);
}

.check-circle {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.8);
  background: rgba(0, 0, 0, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  backdrop-filter: blur(4px);
}

.check-circle.check--active {
  background: var(--color-primary);
  border-color: var(--color-primary);
}

/* 缩略图 */
.card-image {
  position: relative;
  aspect-ratio: 4/3;
  overflow: hidden;
  background: #F1F5F9;
}

.card-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.timeline-card:hover .card-image img {
  transform: scale(1.05);
}

.image-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0);
  transition: background 0.2s ease;
  opacity: 0;
}

.timeline-card:hover .image-overlay {
  background: rgba(0, 0, 0, 0.15);
  opacity: 1;
}

.card-image img[style*="display: none"] + .image-overlay,
.card-image:has(img[style*="display: none"]) .image-overlay {
  opacity: 0;
}

/* 信息区 */
.card-info {
  padding: 0.5rem 0.65rem;
}

.card-time {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.25rem;
  font-variant-numeric: tabular-nums;
}

.card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.meta-tag {
  font-size: 0.7rem;
  color: var(--color-text-secondary);
  background: var(--color-bg);
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta-dim {
  font-size: 0.7rem;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

/* ========== 加载 & 结束 ========== */
.timeline-loader {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1.5rem;
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.loader-dot {
  width: 8px;
  height: 8px;
  background: var(--color-primary);
  border-radius: 50%;
  animation: pulse 1s ease infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.8); }
}

.timeline-end {
  text-align: center;
  padding: 2rem 0;
  color: var(--color-text-muted);
  font-size: 0.78rem;
}

/* ========== Responsive ========== */
@media (max-width: 1024px) {
  .timeline-grid {
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  }
}

@media (max-width: 768px) {
  .timeline-container {
    height: auto;
    min-height: 400px;
  }

  .timeline-toolbar {
    padding: 0.65rem 0.75rem;
    position: sticky;
    top: 0;
  }

  .toolbar-stats {
    gap: 0;
  }

  .stat-item {
    padding: 0 0.4rem;
  }

  .stat-value {
    font-size: 0.85rem;
  }

  .stat-label {
    font-size: 0.6rem;
  }

  .toolbar-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .action-btn {
    font-size: 0.75rem;
    padding: 0.35rem 0.65rem;
  }

  .timeline-scroll {
    padding: 0.5rem;
  }

  .timeline-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
  }

  .timeline-date-separator {
    top: 100px;
    padding: 0.35rem 0;
    margin-bottom: 0.5rem;
  }

  .card-info {
    padding: 0.4rem 0.5rem;
  }

  .card-time {
    font-size: 0.7rem;
  }

  .card-check {
    opacity: 0.6;
    transform: scale(0.9);
    top: 4px;
    right: 4px;
  }

  .check-circle {
    width: 20px;
    height: 20px;
  }
}

@media (max-width: 480px) {
  .timeline-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.4rem;
  }

  .timeline-card {
    border-radius: 8px;
  }

  .stat-divider {
    display: none;
  }
}
</style>
