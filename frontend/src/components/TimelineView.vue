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

    <!-- 日期筛选栏 -->
    <div class="date-filter-bar" v-if="allDateGroups.length > 0">
      <button
        v-if="filterStart"
        class="filter-clear-btn"
        @click="clearFilter"
      >
        <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M4 4l8 8M12 4l-8 8"/>
        </svg>
        清除筛选
      </button>
      <div class="date-filter-scroll" ref="filterScrollRef">
        <button
          v-for="dg in allDateGroups"
          :key="dg.date"
          class="date-filter-chip"
          :class="{
            'chip--start': filterStart === dg.date && filterEnd,
            'chip--end': filterEnd === dg.date,
            'chip--in-range': isDateInFilterRange(dg.date) && filterStart !== dg.date && filterEnd !== dg.date,
            'chip--single': filterStart === dg.date && !filterEnd
          }"
          @click="toggleDateFilter(dg.date)"
        >
          <span class="chip-date">{{ formatDateShort(dg.date) }}</span>
          <span class="chip-count">{{ dg.count }}</span>
        </button>
      </div>
      <div class="filter-summary" v-if="filterStart">
        <span class="filter-label">{{ filterEnd ? `${formatDateShort(filterStart)} — ${formatDateShort(filterEnd)}` : formatDateShort(filterStart) }}</span>
        <span class="filter-total">{{ filteredImages.length }} 张</span>
      </div>
    </div>
    <transition name="slide">
      <div v-if="selectionMode && isMobile" class="selection-banner">
        <span>选择模式 · 点击图片进行多选</span>
        <button @click="exitSelectionMode">完成</button>
      </div>
    </transition>

    <!-- 时间线主体 + 快速跳转 -->
    <div class="timeline-body">
      <!-- 时间线 -->
      <div class="timeline-scroll" ref="scrollContainer" @scroll="handleScroll">
        <!-- 空状态 -->
        <div v-if="filteredImages.length === 0 && images.length > 0" class="timeline-empty">
          <div class="empty-icon">
            <svg viewBox="0 0 48 48" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="24" cy="24" r="12"/>
              <path d="M16 16l16 16M32 16l-16 16"/>
            </svg>
          </div>
          <p>该日期范围内无图片</p>
          <button class="empty-reset-btn" @click="clearFilter">清除筛选</button>
        </div>
        <div v-else-if="images.length === 0" class="timeline-empty">
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
          :ref="el => setGroupRef(group.date, el)"
        >
          <!-- 日期分隔线 -->
          <div class="timeline-date-separator" :id="`date-${group.date}`">
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
      <div v-if="!hasMore && filteredImages.length > 0" class="timeline-end">
        <span>— 已展示全部 {{ filteredImages.length }} 张图片 —</span>
      </div>
    </div>

    <!-- 快速跳转指示器（层级钻取） -->
    <nav class="quick-jump" v-if="yearTree.length > 0 && !isMobile">
      <div class="quick-jump-title">快速跳转</div>
      <div class="jump-tree">
        <div
          v-for="year in yearTree"
          :key="'y-'+year.year"
          class="jump-node jump-year"
          :class="{ 'node--expanded': year.expanded, 'node--active': year.active }"
        >
          <button class="jump-node-btn" @click="toggleYear(year)">
            <span class="jump-caret">
              <svg viewBox="0 0 16 16" width="10" height="10" fill="currentColor">
                <path d="M6 4l4 4-4 4"/>
              </svg>
            </span>
            <span class="jump-node-label">{{ year.year }}年</span>
            <span class="jump-node-count">{{ year.totalCount }}</span>
          </button>
          <div v-if="year.expanded" class="jump-children">
            <div
              v-for="month in year.months"
              :key="'m-'+month.key"
              class="jump-node jump-month"
              :class="{ 'node--expanded': month.expanded, 'node--active': month.active }"
            >
              <button class="jump-node-btn" @click="toggleMonth(month)">
                <span class="jump-caret">
                  <svg viewBox="0 0 16 16" width="10" height="10" fill="currentColor">
                    <path d="M6 4l4 4-4 4"/>
                  </svg>
                </span>
                <span class="jump-node-label">{{ month.label }}</span>
                <span class="jump-node-count">{{ month.totalCount }}</span>
              </button>
              <div v-if="month.expanded" class="jump-children">
                <button
                  v-for="day in month.days"
                  :key="'d-'+day.date"
                  class="jump-node jump-day"
                  :class="{ 'node--active': day.date === activeDate }"
                  @click="jumpToDate(day.date)"
                >
                  <span class="jump-day-dot"></span>
                  <span class="jump-node-label">{{ formatDateShort(day.date) }}</span>
                  <span class="jump-node-count">{{ day.count }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </nav>
  </div>
</div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick, reactive } from 'vue'
import { getProxyImageUrl } from '../api/index.js'

const props = defineProps({
  result: { type: Object, required: true },
  images: { type: Array, required: true },
  taskId: { type: String, required: true },
  selectedIndices: { type: Set, required: true }
})

const emit = defineEmits(['select', 'selectRange', 'preview', 'download', 'clearSelection', 'selectAll'])

// ---- 所有日期分组（用于筛选栏和快速跳转）----
const allDateGroups = computed(() => {
  const groups = []
  let currentDate = ''
  let currentGroup = null

  for (const img of props.images) {
    const date = img.datetime.split(' ')[0]
    if (date !== currentDate) {
      currentGroup = { date, count: 0 }
      groups.push(currentGroup)
      currentDate = date
    }
    currentGroup.count++
  }
  return groups
})

// ---- 层级日期树（年→月→日）----
const yearTree = computed(() => {
  const yearMap = new Map()
  for (const dg of allDateGroups.value) {
    const parts = dg.date.split('-')
    const year = parts[0]
    const month = parts[1]
    const day = parts[2]

    if (!yearMap.has(year)) {
      yearMap.set(year, { months: new Map(), totalCount: 0 })
    }
    const y = yearMap.get(year)
    y.totalCount += dg.count

    const monthKey = `${year}-${month}`
    if (!y.months.has(monthKey)) {
      y.months.set(monthKey, {
        key: monthKey,
        label: `${parseInt(month)}月`,
        days: [],
        totalCount: 0,
        expanded: false,
        active: false
      })
    }
    const m = y.months.get(monthKey)
    m.totalCount += dg.count
    m.days.push({ date: dg.date, count: dg.count })
  }

  // 转为数组并计算 active 状态
  const years = []
  const ad = activeDate.value
  const adParts = ad.split('-')
  const adYear = adParts[0]
  const adMonth = adParts.length >= 2 ? `${adParts[0]}-${adParts[1]}` : ''

  for (const [year, ydata] of yearMap) {
    const months = []
    let yearActive = false
    for (const [mkey, mdata] of ydata.months) {
      const monthActive = mdata.days.some(d => d.date === ad)
      if (monthActive) yearActive = true
      // 保留之前的展开状态
      const existing = expandedState.value.get(mkey)
      months.push({
        ...mdata,
        active: monthActive,
        expanded: existing !== undefined ? existing : false
      })
    }
    const existing = expandedState.value.get(`y-${year}`)
    years.push({
      year,
      months,
      totalCount: ydata.totalCount,
      active: yearActive,
      expanded: existing !== undefined ? existing : false
    })
  }

  return years
})

// 展开状态（Map<nodeKey, boolean>）
const expandedState = ref(new Map())

function toggleYear(year) {
  const key = `y-${year.year}`
  const newMap = new Map(expandedState.value)
  newMap.set(key, !newMap.get(key))
  expandedState.value = newMap
}

function toggleMonth(month) {
  const newMap = new Map(expandedState.value)
  newMap.set(month.key, !newMap.get(month.key))
  expandedState.value = newMap
}

// activeDate 变化时自动展开祖先
watch(activeDate, (newDate) => {
  if (!newDate) return
  const parts = newDate.split('-')
  if (parts.length < 2) return
  const year = parts[0]
  const monthKey = `${parts[0]}-${parts[1]}`
  const newMap = new Map(expandedState.value)
  newMap.set(`y-${year}`, true)
  newMap.set(monthKey, true)
  expandedState.value = newMap
})
const filterStart = ref('')
const filterEnd = ref('')
const filterScrollRef = ref(null)

const filteredImages = computed(() => {
  if (!filterStart.value) return props.images
  const start = filterStart.value
  const end = filterEnd.value || filterStart.value
  return props.images.filter(img => {
    const d = img.datetime.split(' ')[0]
    return d >= start && d <= end
  })
})

function toggleDateFilter(date) {
  if (!filterStart.value) {
    // 第一次点击：设为起始日期
    filterStart.value = date
    filterEnd.value = ''
  } else if (filterEnd.value) {
    // 已有完整范围：重新开始
    filterStart.value = date
    filterEnd.value = ''
  } else if (date === filterStart.value) {
    // 再次点击同一天：取消筛选
    filterStart.value = ''
  } else if (date < filterStart.value) {
    // 点击更早日期：交换起始
    filterEnd.value = filterStart.value
    filterStart.value = date
  } else {
    // 点击更晚日期：设为终止
    filterEnd.value = date
  }
}

function clearFilter() {
  filterStart.value = ''
  filterEnd.value = ''
}

function isDateInFilterRange(date) {
  if (!filterStart.value) return false
  const start = filterStart.value
  const end = filterEnd.value || filterStart.value
  return date >= start && date <= end
}

// ---- 无极滚动 ----
const BATCH_SIZE = 40
const displayCount = ref(BATCH_SIZE)
const scrollContainer = ref(null)
const loadTrigger = ref(null)

const hasMore = computed(() => displayCount.value < filteredImages.value.length)

// 按日期分组可见图片
const visibleGroups = computed(() => {
  const visible = filteredImages.value.slice(0, displayCount.value)
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
    displayCount.value = Math.min(displayCount.value + BATCH_SIZE, filteredImages.value.length)
  }
}

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

function handleScroll() {
  if (!scrollContainer.value || !hasMore.value) return
  const { scrollTop, scrollHeight, clientHeight } = scrollContainer.value
  if (scrollHeight - scrollTop - clientHeight < 400) {
    loadMore()
  }
}

watch(displayCount, () => { nextTick(() => setupObserver()) })

// ---- 快速跳转 ----
const groupRefs = reactive({})
const activeDate = ref('')

function setGroupRef(date, el) {
  if (el) groupRefs[date] = el
}

function jumpToDate(date) {
  const el = groupRefs[date]
  if (el && scrollContainer.value) {
    const offsetTop = el.offsetTop - 80
    scrollContainer.value.scrollTo({ top: offsetTop, behavior: 'smooth' })
  }
}

// 检测当前可见日期
function updateActiveDate() {
  if (!scrollContainer.value || allDateGroups.value.length === 0) return
  const container = scrollContainer.value
  const scrollMid = container.scrollTop + 120

  let closest = allDateGroups.value[0].date
  let closestDist = Infinity

  for (const dg of allDateGroups.value) {
    const el = groupRefs[dg.date]
    if (!el) continue
    const dist = Math.abs(el.offsetTop - scrollMid)
    if (dist < closestDist) {
      closestDist = dist
      closest = dg.date
    }
  }
  activeDate.value = closest
}

// ---- 选择逻辑 ----
const isMobile = ref(false)
const selectionMode = ref(false)
const lastClickedIndex = ref(-1)
let longPressTimer = null

function detectMobile() { isMobile.value = window.innerWidth < 768 }

function handleCardClick(index, event) {
  if (selectionMode.value && isMobile.value) {
    emit('select', index)
    return
  }
  if (event.shiftKey && lastClickedIndex.value >= 0) {
    const start = Math.min(lastClickedIndex.value, index)
    const end = Math.max(lastClickedIndex.value, index)
    const rangeIndices = []
    for (let i = start; i <= end; i++) rangeIndices.push(i)
    emit('selectRange', rangeIndices)
    return
  }
  emit('select', index)
  lastClickedIndex.value = index
}

function handleContextMenu(index) {
  if (!selectionMode.value && isMobile.value) {
    enterSelectionMode()
    emit('select', index)
  }
}

function enterSelectionMode() { selectionMode.value = true }
function exitSelectionMode() { selectionMode.value = false }

const isAllSelected = computed(() =>
  props.images.length > 0 && props.selectedIndices.size === props.images.length
)

function toggleSelectAll() {
  emit('selectAll', !isAllSelected.value)
}

// ---- 工具函数 ----
function getProxyUrl(url) { return getProxyImageUrl(url) }

function formatTime(datetime) {
  const parts = datetime.split(' ')
  return parts.length >= 2 ? parts[1] : datetime
}

function formatDateShort(dateStr) {
  const parts = dateStr.split('-')
  return parts.length === 3 ? `${parts[1]}-${parts[2]}` : dateStr
}

function onImgError(e) { e.target.style.display = 'none' }

// ---- 生命周期 ----
onMounted(() => {
  detectMobile()
  window.addEventListener('resize', detectMobile)
  nextTick(() => {
    setupObserver()
    updateActiveDate()
  })
})

onUnmounted(() => {
  window.removeEventListener('resize', detectMobile)
  if (observer) observer.disconnect()
})

// 监听滚动更新 activeDate
watch(scrollContainer, (el) => {
  if (el) {
    el.addEventListener('scroll', updateActiveDate, { passive: true })
  }
})

// 筛选变化时重置滚动
watch([filterStart, filterEnd], () => {
  displayCount.value = BATCH_SIZE
  if (scrollContainer.value) scrollContainer.value.scrollTop = 0
  nextTick(() => { setupObserver(); updateActiveDate() })
})

// 任务切换重置
watch(() => props.result?.taskId, () => {
  displayCount.value = BATCH_SIZE
  lastClickedIndex.value = -1
  selectionMode.value = false
  filterStart.value = ''
  filterEnd.value = ''
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

.timeline-body {
  flex: 1;
  display: flex;
  min-height: 0;
  position: relative;
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

/* ========== 日期筛选栏 ========== */
.date-filter-bar {
  display: flex;
  align-items: center;
  gap: 0;
  padding: 0.5rem 1rem;
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border-light);
  overflow: hidden;
}

.filter-clear-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.6rem;
  border-radius: 6px;
  font-size: 0.7rem;
  font-weight: 500;
  border: 1px solid #FECACA;
  background: #FEF2F2;
  color: #DC2626;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  margin-right: 0.5rem;
  font-family: inherit;
  transition: all 0.2s;
}

.filter-clear-btn:hover {
  background: #FEE2E2;
}

.date-filter-scroll {
  flex: 1;
  display: flex;
  gap: 0.35rem;
  overflow-x: auto;
  padding: 0.15rem 0;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.date-filter-scroll::-webkit-scrollbar {
  display: none;
}

.date-filter-chip {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.1rem;
  min-width: 44px;
  padding: 0.3rem 0.5rem;
  border-radius: 8px;
  border: 1.5px solid var(--color-border-light);
  background: var(--color-bg);
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  white-space: nowrap;
  flex-shrink: 0;
}

.date-filter-chip:hover {
  border-color: var(--color-primary-light);
  background: var(--color-primary-bg);
}

/* 单日选中 */
.date-filter-chip.chip--single {
  border-color: var(--color-primary);
  background: var(--color-primary);
  color: white;
}

/* 范围起始 */
.date-filter-chip.chip--start {
  border-color: var(--color-primary);
  background: var(--color-primary);
  color: white;
}

/* 范围终止 */
.date-filter-chip.chip--end {
  border-color: var(--color-primary);
  background: var(--color-primary);
  color: white;
}

/* 范围内（非起止） */
.date-filter-chip.chip--in-range {
  border-color: #A5B4FC;
  background: #EEF2FF;
  color: var(--color-primary-dark);
}

.chip--single .chip-count,
.chip--start .chip-count,
.chip--end .chip-count {
  opacity: 0.85;
}

.chip--in-range .chip-count {
  opacity: 0.7;
}

.filter-summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
  margin-left: 0.75rem;
  padding-left: 0.75rem;
  border-left: 1px solid var(--color-border);
}

.filter-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-primary-dark);
  white-space: nowrap;
}

.filter-total {
  font-size: 0.7rem;
  color: var(--color-text-muted);
  background: var(--color-bg);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-weight: 500;
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
  min-width: 0;
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

/* ========== 快速跳转（层级钻取） ========== */
.quick-jump {
  width: 125px;
  flex-shrink: 0;
  border-left: 1px solid var(--color-border-light);
  background: var(--color-bg-card);
  padding: 0.6rem 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  user-select: none;
}

.quick-jump-title {
  font-size: 0.6rem;
  font-weight: 700;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0 0.5rem 0.5rem;
  margin: 0 0.5rem 0.35rem;
  border-bottom: 1px solid var(--color-border-light);
}

.jump-tree {
  display: flex;
  flex-direction: column;
  padding: 0 0.25rem;
}

/* 树节点 */
.jump-node {
  position: relative;
}

.jump-node-btn {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  width: 100%;
  padding: 0.25rem 0.3rem;
  border-radius: 5px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-family: inherit;
  font-size: 0.68rem;
  color: var(--color-text-secondary);
  transition: all 0.12s ease;
  text-align: left;
}

.jump-node-btn:hover {
  background: var(--color-bg-hover);
}

/* 展开/折叠箭头 */
.jump-caret {
  width: 14px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--color-text-muted);
  transition: transform 0.2s ease;
}

.node--expanded > .jump-node-btn .jump-caret {
  transform: rotate(90deg);
}

/* 节点标签和计数 */
.jump-node-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.jump-node-count {
  font-size: 0.58rem;
  color: var(--color-text-muted);
  opacity: 0.6;
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
}

/* 当前活跃祖先节点 */
.node--active > .jump-node-btn {
  color: var(--color-primary-dark);
  font-weight: 600;
  background: var(--color-primary-bg);
}

/* 子节点容器 */
.jump-children {
  padding-left: 0.8rem;
  border-left: 1.5px solid var(--color-border-light);
  margin-left: 0.5rem;
}

/* 年份节点 */
.jump-year > .jump-node-btn {
  font-weight: 600;
  color: var(--color-text);
}

/* 月份节点 */
.jump-month > .jump-node-btn {
  font-size: 0.65rem;
}

/* 日期节点（叶子） */
.jump-day {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  width: 100%;
  padding: 0.2rem 0.3rem;
  border-radius: 5px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-family: inherit;
  font-size: 0.65rem;
  color: var(--color-text-secondary);
  transition: all 0.12s ease;
  text-align: left;
}

.jump-day:hover {
  background: var(--color-bg-hover);
}

.jump-day.node--active {
  background: var(--color-primary);
  color: white;
  font-weight: 600;
}

.jump-day.node--active .jump-node-count {
  color: rgba(255,255,255,0.7);
}

.jump-day-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--color-border);
  flex-shrink: 0;
  transition: all 0.2s;
}

.jump-day.node--active .jump-day-dot {
  background: white;
  box-shadow: 0 0 0 2px rgba(255,255,255,0.3);
}

/* 空状态重置按钮 */
.empty-reset-btn {
  margin-top: 0.75rem;
  padding: 0.35rem 0.85rem;
  border-radius: 8px;
  border: 1px solid var(--color-primary-light);
  background: var(--color-primary-bg);
  color: var(--color-primary-dark);
  font-size: 0.8rem;
  cursor: pointer;
  font-family: inherit;
  font-weight: 500;
  transition: all 0.2s;
}

.empty-reset-btn:hover {
  background: var(--color-primary);
  color: white;
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

  .quick-jump {
    width: 90px;
  }
}

@media (max-width: 768px) {
  .timeline-container {
    height: auto;
    min-height: 400px;
  }

  .timeline-body {
    flex-direction: column;
  }

  .quick-jump {
    display: none;
  }

  .timeline-toolbar {
    padding: 0.65rem 0.75rem;
    position: sticky;
    top: 0;
  }

  .date-filter-bar {
    padding: 0.4rem 0.5rem;
    overflow-x: auto;
  }

  .date-filter-scroll {
    gap: 0.25rem;
  }

  .date-filter-chip {
    min-width: 38px;
    padding: 0.25rem 0.4rem;
  }

  .chip-date {
    font-size: 0.62rem;
  }

  .chip-count {
    font-size: 0.55rem;
  }

  .filter-summary {
    display: none;
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

  .date-filter-chip {
    min-width: 34px;
  }
}
</style>
