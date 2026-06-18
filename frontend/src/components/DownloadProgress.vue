<template>
  <div class="download-overlay" @click.self="!downloading && $emit('close')">
    <div class="download-modal">
      <!-- 头部 -->
      <div class="download-header">
        <h3>{{ downloading ? '正在下载' : completed ? '下载完成' : '准备下载' }}</h3>
        <button v-if="!downloading" class="dl-close-btn" @click="$emit('close')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <!-- 进度条 -->
      <div class="download-progress-section">
        <div class="progress-bar-wrapper">
          <div
            class="progress-bar-fill"
            :style="{ width: progressPercent + '%' }"
            :class="{ 'bar--complete': completed, 'bar--error': hasError }"
          ></div>
        </div>
        <div class="progress-stats">
          <span class="stat-count">
            <template v-if="hasError">下载失败</template>
            <template v-else>{{ downloaded }} / {{ total }} 张</template>
          </span>
          <span class="stat-percent">{{ progressPercent }}%</span>
        </div>
      </div>

      <!-- 速度与时间 -->
      <div class="download-meta" v-if="downloading && !hasError">
        <span>{{ formatSpeed(speed) }}</span>
        <span class="meta-sep">·</span>
        <span>剩余 {{ formatTime(eta) }}</span>
      </div>

      <!-- 错误信息 -->
      <div v-if="hasError" class="download-error">
        <p>{{ errorMsg }}</p>
        <div class="error-actions">
          <button class="dl-btn dl-btn--retry" @click="retryFailed">
            重试失败项 ({{ failedList.length }})
          </button>
          <button class="dl-btn dl-btn--skip" @click="skipAndFinish">
            跳过并打包已完成项
          </button>
        </div>
      </div>

      <!-- 完成状态 -->
      <div v-if="completed" class="download-complete">
        <div class="complete-icon">
          <svg viewBox="0 0 32 32" width="32" height="32" fill="none">
            <circle cx="16" cy="16" r="12" stroke="#10B981" stroke-width="2.5"/>
            <path d="M11 16l3.5 3.5 6.5-7" stroke="#10B981" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <p>成功下载 {{ successCount }} 张图片</p>
        <p class="complete-size">总大小 {{ formatSize(totalBytes) }}</p>
      </div>

      <!-- 操作按钮 -->
      <div class="download-footer" v-if="completed && !hasError">
        <button class="dl-btn dl-btn--primary" @click="saveZip">
          保存 ZIP 文件
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onUnmounted } from 'vue'
import JSZip from 'jszip'

const props = defineProps({
  images: { type: Array, required: true }
})

const emit = defineEmits(['close'])

const downloading = ref(false)
const completed = ref(false)
const hasError = ref(false)
const errorMsg = ref('')
const downloaded = ref(0)
const total = ref(0)
const speed = ref(0)
const eta = ref(0)
const totalBytes = ref(0)
const successCount = ref(0)
const failedList = ref([])
const zipBlob = ref(null)

let abortController = null
let startTime = 0
let lastDownloaded = 0
let lastTime = 0

const progressPercent = computed(() => {
  if (total.value === 0) return 0
  return Math.round((downloaded.value / total.value) * 100)
})

const PARALLEL = 4

// 开始下载
async function start() {
  downloading.value = true
  completed.value = false
  hasError.value = false
  errorMsg.value = ''
  downloaded.value = 0
  total.value = props.images.length
  successCount.value = 0
  failedList.value = []
  totalBytes.value = 0
  zipBlob.value = null
  startTime = Date.now()
  lastDownloaded = 0
  lastTime = Date.now()
  abortController = new AbortController()

  const zip = new JSZip()
  const queue = [...props.images]

  // 并行下载 worker
  async function worker() {
    while (queue.length > 0 && !abortController.signal.aborted) {
      const img = queue.shift()
      if (!img) break
      try {
        const blob = await fetchOne(img, abortController.signal)
        zip.file(img.fileName, blob)
        totalBytes.value += blob.size
        successCount.value++
      } catch (err) {
        if (err.name !== 'AbortError') {
          failedList.value.push(img)
        }
      }
      downloaded.value++
      updateSpeed()
    }
  }

  // 启动并行 workers
  const workers = []
  for (let i = 0; i < PARALLEL; i++) {
    workers.push(worker())
  }

  try {
    await Promise.all(workers)

    if (failedList.value.length > 0 && successCount.value === 0) {
      hasError.value = true
      errorMsg.value = `全部 ${failedList.value.length} 张图片下载失败`
    } else if (failedList.value.length > 0) {
      hasError.value = true
      errorMsg.value = `${failedList.value.length} 张图片下载失败，${successCount.value} 张成功`
    } else {
      // 全部成功：生成 ZIP
      const zipData = await zip.generateAsync({
        type: 'blob',
        compression: 'DEFLATE',
        compressionOptions: { level: 1 }
      }, (meta) => {
        // zip 压缩进度（不额外展示，速度很快）
      })
      zipBlob.value = zipData
      completed.value = true
    }
  } catch (err) {
    hasError.value = true
    errorMsg.value = '下载过程出错: ' + err.message
  }

  downloading.value = false
}

async function fetchOne(img, signal) {
  const proxyUrl = `/api/proxy-image?url=${encodeURIComponent(img.serverImgUrl)}`
  const resp = await fetch(proxyUrl, { signal })
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
  return resp.blob()
}

function updateSpeed() {
  const now = Date.now()
  const elapsed = (now - lastTime) / 1000
  if (elapsed >= 1) {
    const delta = downloaded.value - lastDownloaded
    speed.value = delta / elapsed
    lastDownloaded = downloaded.value
    lastTime = now
    if (speed.value > 0) {
      eta.value = Math.round((total.value - downloaded.value) / speed.value)
    }
  }
}

function retryFailed() {
  // 将失败项重新加入队列
  hasError.value = false
  errorMsg.value = ''
  total.value = failedList.value.length
  // 简单重启（用新的 images）
  const failed = [...failedList.value]
  failedList.value = []
  const saved = props.images
  // 替换 images 引用，重新开始
  props.images.splice(0, props.images.length, ...failed)
  start()
}

function skipAndFinish() {
  hasError.value = false
  completed.value = true
}

async function saveZip() {
  if (!zipBlob.value) return
  const url = URL.createObjectURL(zipBlob.value)
  const a = document.createElement('a')
  a.href = url
  a.download = `images_${new Date().toISOString().slice(0, 10)}.zip`
  a.click()
  URL.revokeObjectURL(url)
}

function formatSpeed(bytesPerSec) {
  if (bytesPerSec < 1024) return `${bytesPerSec.toFixed(0)} B/s`
  if (bytesPerSec < 1048576) return `${(bytesPerSec / 1024).toFixed(1)} KB/s`
  return `${(bytesPerSec / 1048576).toFixed(1)} MB/s`
}

function formatTime(seconds) {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

function formatSize(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1048576).toFixed(1)} MB`
}

onUnmounted(() => {
  if (abortController) abortController.abort()
})

defineExpose({ start })
</script>

<style scoped>
.download-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 4000;
  padding: 1rem;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.download-modal {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 420px;
  box-shadow: var(--shadow-xl);
  overflow: hidden;
  animation: scaleIn 0.25s ease;
}

@keyframes scaleIn {
  from { transform: scale(0.95); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

.download-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-light);
}

.download-header h3 {
  font-size: 0.95rem;
  font-weight: 600;
}

.dl-close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 0.25rem;
  border-radius: 6px;
  transition: all 0.2s;
  line-height: 0;
}

.dl-close-btn:hover {
  background: var(--color-bg-hover);
  color: var(--color-text);
}

/* 进度条 */
.download-progress-section {
  padding: 1.25rem 1.25rem 0.75rem;
}

.progress-bar-wrapper {
  height: 8px;
  background: var(--color-bg);
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 0.5rem;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #6366F1, #8B5CF6);
  border-radius: 10px;
  transition: width 0.3s ease;
  min-width: 0;
}

.progress-bar-fill.bar--complete {
  background: #10B981;
}

.progress-bar-fill.bar--error {
  background: #EF4444;
}

.progress-stats {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.stat-percent {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

/* 速度 */
.download-meta {
  padding: 0 1.25rem 0.75rem;
  font-size: 0.72rem;
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.meta-sep {
  opacity: 0.4;
}

/* 错误 */
.download-error {
  padding: 0.75rem 1.25rem;
  text-align: center;
}

.download-error p {
  font-size: 0.82rem;
  color: #DC2626;
  margin-bottom: 0.75rem;
}

.error-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
}

/* 完成 */
.download-complete {
  padding: 1.25rem;
  text-align: center;
}

.complete-icon {
  margin-bottom: 0.5rem;
}

.download-complete p {
  font-size: 0.85rem;
  color: var(--color-text);
  font-weight: 500;
}

.complete-size {
  font-size: 0.72rem !important;
  color: var(--color-text-muted) !important;
  font-weight: 400 !important;
  margin-top: 0.25rem;
}

/* 按钮 */
.download-footer {
  padding: 0.75rem 1.25rem 1.25rem;
  text-align: center;
}

.dl-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 500;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.dl-btn--primary {
  background: var(--color-primary);
  color: white;
  width: 100%;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
}

.dl-btn--primary:hover {
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.4);
}

.dl-btn--retry {
  background: #FEF2F2;
  color: #DC2626;
  border: 1px solid #FECACA;
}

.dl-btn--retry:hover {
  background: #FEE2E2;
}

.dl-btn--skip {
  background: var(--color-bg);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}
</style>
