<template>
  <div class="app-container">
    <!-- 水印 -->
    <Watermark text="图片批量下载工具" />

    <!-- 顶部导航 -->
    <header class="app-header">
      <div class="header-content">
        <div class="header-left">
          <div class="logo">
            <svg viewBox="0 0 36 36" width="36" height="36">
              <defs>
                <linearGradient id="logo-grad-h" x1="0" y1="0" x2="36" y2="36">
                  <stop offset="0%" stop-color="#6366F1"/>
                  <stop offset="100%" stop-color="#8B5CF6"/>
                </linearGradient>
              </defs>
              <rect width="36" height="36" rx="8" fill="url(#logo-grad-h)"/>
              <path d="M10 12l4 6-4 6h3l4-6-4-6h-3zm8 0l4 6-4 6h3l4-6-4-6h-3z" fill="white"/>
            </svg>
          </div>
          <div class="header-text">
            <h1 class="app-title">学海聊天记录图片批量下载工具</h1>
            <p class="app-subtitle">上传聊天记录 · 智能解析 · 一键下载</p>
          </div>
        </div>
        <div class="header-right hidden-mobile">
          <div class="header-badge">
            <span class="badge-dot"></span>
            LSJ
          </div>
        </div>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="app-main">
      <!-- 上传区 -->
      <FileUpload
        :loading="loading"
        :result="result"
        @upload="handleUpload"
        @reset="handleReset"
      />

      <!-- 图片时间线 -->
      <TimelineView
        v-if="result && result.images && result.images.length > 0"
        :result="result"
        :images="result.images"
        :taskId="result.taskId"
        :selectedIndices="selectedIndices"
        @select="handleSelect"
        @selectRange="handleSelectRange"
        @selectAll="handleSelectAll"
        @clearSelection="handleClearSelection"
        @preview="handlePreview"
        @download="handleDownload"
      />
    </main>

    <!-- 图片预览弹窗 -->
    <ImagePreview
      v-if="previewImage"
      :image="previewImage"
      @close="previewImage = null"
    />

    <!-- 页脚 -->
    <AppFooter />

    <!-- 加载遮罩 -->
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner">
        <div class="spinner"></div>
        <p>正在解析文件...</p>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMsg" class="toast toast-error" @click="errorMsg = ''; clearTimeout(toastTimer)">
      {{ errorMsg }}
    </div>

    <!-- 成功提示 -->
    <div v-if="successMsg" class="toast toast-success" @click="successMsg = ''; clearTimeout(toastTimer)">
      {{ successMsg }}
    </div>
    <!-- 下载进度弹窗 -->
    <DownloadProgress
      v-if="downloadImages_list"
      ref="downloadProgressRef"
      :images="downloadImages_list"
      @close="downloadImages_list = null"
    />
  </div>
</template>

<script setup>
import { ref, nextTick, watch } from 'vue'
import { uploadHTML } from './api/index.js'
import FileUpload from './components/FileUpload.vue'
import TimelineView from './components/TimelineView.vue'
import ImagePreview from './components/ImagePreview.vue'
import DownloadProgress from './components/DownloadProgress.vue'
import Watermark from './components/Watermark.vue'
import AppFooter from './components/AppFooter.vue'

const loading = ref(false)
const result = ref(null)
const selectedIndices = ref(new Set())
const previewImage = ref(null)
const errorMsg = ref('')
const successMsg = ref('')
const downloadImages_list = ref(null)
const downloadProgressRef = ref(null)

// Toast 自动消失
let toastTimer = null
watch(errorMsg, (val) => {
  clearTimeout(toastTimer)
  if (val) toastTimer = setTimeout(() => { errorMsg.value = '' }, 4000)
})
watch(successMsg, (val) => {
  clearTimeout(toastTimer)
  if (val) toastTimer = setTimeout(() => { successMsg.value = '' }, 2500)
})

async function handleUpload(file) {
  loading.value = true
  errorMsg.value = ''
  result.value = null
  selectedIndices.value = new Set()

  try {
    const res = await uploadHTML(file)
    if (res.data.error) {
      errorMsg.value = res.data.error
      return
    }
    result.value = res.data
    successMsg.value = `成功解析 ${res.data.totalCount} 张图片`
  } catch (err) {
    errorMsg.value = err.response?.data?.error || '上传失败，请重试'
  } finally {
    loading.value = false
  }
}

function handleReset() {
  result.value = null
  selectedIndices.value = new Set()
  errorMsg.value = ''
}

function handleSelect(index) {
  const newSet = new Set(selectedIndices.value)
  if (newSet.has(index)) {
    newSet.delete(index)
  } else {
    newSet.add(index)
  }
  selectedIndices.value = newSet
}

function handleSelectRange(indices) {
  const newSet = new Set(selectedIndices.value)
  for (const i of indices) {
    newSet.add(i)
  }
  selectedIndices.value = newSet
}

function handleClearSelection() {
  selectedIndices.value = new Set()
}

function handleSelectAll(selectAll) {
  if (selectAll && result.value) {
    selectedIndices.value = new Set(result.value.images.map((_, i) => i))
  } else {
    selectedIndices.value = new Set()
  }
}

function handlePreview(image) {
  previewImage.value = image
}

async function handleDownload() {
  if (!result.value) return

  const indices = Array.from(selectedIndices.value)
  if (indices.length === 0) {
    errorMsg.value = '请先选择要下载的图片'
    return
  }

  const selectedImages = indices.map(i => result.value.images[i])

  // 单张图片：直接代理下载
  if (selectedImages.length === 1) {
    const img = selectedImages[0]
    const proxyUrl = `/api/proxy-image?url=${encodeURIComponent(img.serverImgUrl)}`
    const a = document.createElement('a')
    a.href = proxyUrl
    a.download = img.fileName
    a.click()
    successMsg.value = '开始下载 1 张图片'
    return
  }

  // 多张图片：打开进度弹窗，前端并行下载 + JSZip 打包
  downloadImages_list.value = selectedImages
  await nextTick()
  downloadProgressRef.value?.start()
}
</script>

<style>
/* ===== CSS Reset & Variables ===== */
*,
*::before,
*::after {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  --color-primary: #6366F1;
  --color-primary-dark: #4F46E5;
  --color-primary-light: #A5B4FC;
  --color-primary-bg: #EEF2FF;
  --color-success: #10B981;
  --color-warning: #F59E0B;
  --color-danger: #EF4444;
  --color-bg: #F1F5F9;
  --color-bg-card: #FFFFFF;
  --color-bg-hover: #F8FAFC;
  --color-text: #0F172A;
  --color-text-secondary: #475569;
  --color-text-muted: #94A3B8;
  --color-border: #E2E8F0;
  --color-border-light: #F1F5F9;
  --shadow-xs: 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.08), 0 4px 6px -4px rgba(0, 0, 0, 0.04);
  --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -6px rgba(0, 0, 0, 0.04);
  --radius-xs: 4px;
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 16px;
  --radius-xl: 20px;
  --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC',
    'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial,
    sans-serif;
  --font-mono: 'SF Mono', 'Fira Code', 'Fira Mono', 'Roboto Mono', monospace;
}

html {
  font-size: 16px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

body {
  font-family: var(--font-sans);
  background: var(--color-bg);
  color: var(--color-text);
  line-height: 1.6;
  min-height: 100vh;
}

/* ===== App Layout ===== */
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
}

/* ===== Header ===== */
.app-header {
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border);
  padding: 1rem 1.5rem;
  box-shadow: var(--shadow-xs);
  position: sticky;
  top: 0;
  z-index: 500;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.header-text {
  display: flex;
  flex-direction: column;
}

.app-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--color-text);
  letter-spacing: -0.02em;
  line-height: 1.3;
}

.app-subtitle {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  font-weight: 400;
  letter-spacing: 0.02em;
}

.header-right {
  display: flex;
  align-items: center;
}

.header-badge {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-bg);
  padding: 0.3rem 0.7rem;
  border-radius: 20px;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.badge-dot {
  width: 6px;
  height: 6px;
  background: var(--color-primary);
  border-radius: 50%;
  animation: dotPulse 2s ease infinite;
}

@keyframes dotPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* ===== Main ===== */
.app-main {
  flex: 1;
  max-width: 1300px;
  width: 100%;
  margin: 0 auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
}

/* ===== Loading Overlay ===== */
.loading-overlay {
  position: fixed;
  inset: 0;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.loading-spinner {
  text-align: center;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-spinner p {
  color: var(--color-text-secondary);
  font-size: 0.95rem;
}

/* ===== Toast ===== */
.toast {
  position: fixed;
  top: 1rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.75rem 1.5rem;
  border-radius: var(--radius-md);
  font-size: 0.9rem;
  font-weight: 500;
  z-index: 2000;
  cursor: pointer;
  animation: slideDown 0.3s ease;
  box-shadow: var(--shadow-lg);
  max-width: 90vw;
  text-align: center;
}

.toast-error {
  background: #FEF2F2;
  color: var(--color-danger);
  border: 1px solid #FECACA;
}

.toast-success {
  background: #ECFDF5;
  color: var(--color-success);
  border: 1px solid #A7F3D0;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateX(-50%) translateY(-20px); }
  to { opacity: 1; transform: translateX(-50%) translateY(0); }
}

/* ===== Responsive ===== */
@media (max-width: 768px) {
  .app-header {
    padding: 0.75rem 1rem;
  }

  .app-title {
    font-size: 1rem;
  }

  .app-subtitle {
    font-size: 0.7rem;
  }

  .header-badge {
    font-size: 0.65rem;
    padding: 0.2rem 0.5rem;
  }

  .app-main {
    padding: 0.5rem;
  }

  .hidden-mobile {
    display: none;
  }
}

@media (max-width: 480px) {
  .app-header {
    padding: 0.65rem 0.75rem;
  }

  .logo svg {
    width: 28px;
    height: 28px;
  }

  .app-title {
    font-size: 0.9rem;
  }
}
</style>
