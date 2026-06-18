<template>
  <div class="app-container">
    <!-- 水印 -->
    <Watermark text="图片批量下载工具" />

    <!-- 顶部导航 -->
    <header class="app-header">
      <div class="header-content">
        <div class="logo">
          <svg class="logo-icon" viewBox="0 0 32 32" width="32" height="32">
            <rect width="32" height="32" rx="6" fill="url(#logo-grad)"/>
            <defs>
              <linearGradient id="logo-grad" x1="0" y1="0" x2="32" y2="32">
                <stop offset="0%" stop-color="#6366F1"/>
                <stop offset="100%" stop-color="#8B5CF6"/>
              </linearGradient>
            </defs>
            <path d="M8 10l4 6-4 6h3l4-6-4-6H8zm7 0l4 6-4 6h3l4-6-4-6h-3z" fill="white"/>
          </svg>
          <h1 class="app-title">图片批量下载工具</h1>
        </div>
        <p class="app-subtitle">上传聊天记录 HTML 文件，轻松批量下载图片</p>
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

      <!-- 图片表格 -->
      <ImageTable
        v-if="result && result.images && result.images.length > 0"
        :images="result.images"
        :taskId="result.taskId"
        :selectedIndices="selectedIndices"
        @select="handleSelect"
        @selectAll="handleSelectAll"
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
    <div v-if="errorMsg" class="toast toast-error" @click="errorMsg = ''">
      {{ errorMsg }}
    </div>

    <!-- 成功提示 -->
    <div v-if="successMsg" class="toast toast-success" @click="successMsg = ''">
      {{ successMsg }}
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { uploadHTML, downloadImages } from './api/index.js'
import FileUpload from './components/FileUpload.vue'
import ImageTable from './components/ImageTable.vue'
import ImagePreview from './components/ImagePreview.vue'
import Watermark from './components/Watermark.vue'
import AppFooter from './components/AppFooter.vue'

const loading = ref(false)
const result = ref(null)
const selectedIndices = ref(new Set())
const previewImage = ref(null)
const errorMsg = ref('')
const successMsg = ref('')

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

  try {
    const res = await downloadImages(result.value.taskId, indices)
    const blob = res.data
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url

    // 从响应头获取文件名
    const disposition = res.headers['content-disposition']
    if (disposition) {
      const match = disposition.match(/filename="?(.+?)"?$/i)
      if (match) a.download = match[1]
    } else if (indices.length === 1) {
      a.download = result.value.images[indices[0]].fileName
    } else {
      a.download = 'images.zip'
    }

    a.click()
    URL.revokeObjectURL(url)
    successMsg.value = `成功下载 ${indices.length} 张图片`
  } catch (err) {
    errorMsg.value = '下载失败，请重试'
  }
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
  --color-success: #10B981;
  --color-warning: #F59E0B;
  --color-danger: #EF4444;
  --color-bg: #F8FAFC;
  --color-bg-card: #FFFFFF;
  --color-bg-hover: #F1F5F9;
  --color-text: #1E293B;
  --color-text-secondary: #64748B;
  --color-text-muted: #94A3B8;
  --color-border: #E2E8F0;
  --color-border-light: #F1F5F9;
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1);
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 16px;
  --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC',
    'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial,
    sans-serif;
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
  background: linear-gradient(135deg, var(--color-primary-dark), var(--color-primary));
  color: white;
  padding: 2rem 1.5rem;
  text-align: center;
  box-shadow: var(--shadow-md);
}

.header-content {
  max-width: 900px;
  margin: 0 auto;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.logo-icon {
  flex-shrink: 0;
}

.app-title {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.app-subtitle {
  font-size: 0.9rem;
  opacity: 0.85;
  font-weight: 400;
}

/* ===== Main ===== */
.app-main {
  flex: 1;
  max-width: 1100px;
  width: 100%;
  margin: 0 auto;
  padding: 1.5rem 1rem 3rem;
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
    padding: 1.5rem 1rem;
  }

  .app-title {
    font-size: 1.25rem;
  }

  .app-subtitle {
    font-size: 0.8rem;
  }

  .app-main {
    padding: 1rem 0.75rem 2rem;
  }
}

@media (max-width: 480px) {
  .app-header {
    padding: 1.25rem 0.75rem;
  }

  .logo {
    gap: 0.5rem;
  }

  .logo-icon {
    width: 26px;
    height: 26px;
  }

  .app-title {
    font-size: 1.1rem;
  }
}
</style>
