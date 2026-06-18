<template>
  <div class="upload-section">
    <div
      class="upload-zone"
      :class="{ 'upload-zone--active': isDragover, 'upload-zone--done': result }"
      @dragover.prevent="isDragover = true"
      @dragleave.prevent="isDragover = false"
      @drop.prevent="handleDrop"
      @click="triggerUpload"
    >
      <input
        ref="fileInput"
        type="file"
        accept=".html,.htm"
        class="upload-input"
        @change="handleFileChange"
      />

      <template v-if="!result">
        <div class="upload-icon-wrapper">
          <div class="upload-icon">
            <svg viewBox="0 0 40 40" width="40" height="40" fill="none">
              <rect x="4" y="8" width="32" height="24" rx="3" stroke="currentColor" stroke-width="2"/>
              <path d="M4 26l9-7 6 6 7-11 10 9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <circle cx="14" cy="17" r="2.5" stroke="currentColor" stroke-width="1.5"/>
            </svg>
          </div>
        </div>
        <p class="upload-title">拖拽或<span class="upload-link">点击上传</span>HTML 文件</p>
        <p class="upload-hint">支持 IMX 聊天记录导出的 .html 格式</p>
      </template>

      <template v-else>
        <div class="upload-result">
          <div class="result-icon">
            <svg viewBox="0 0 32 32" width="32" height="32" fill="none">
              <circle cx="16" cy="16" r="12" stroke="#10B981" stroke-width="2.5"/>
              <path d="M11 16l3.5 3.5 6.5-7" stroke="#10B981" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <div class="result-info">
            <p class="result-title">解析完成</p>
            <p class="result-detail">
              共 <strong>{{ result.totalCount }}</strong> 张图片
              <span class="result-sep">·</span>
              发送人 <strong>{{ result.sender }}</strong>
              <span class="result-sep">·</span>
              会话 <strong>{{ result.recipients }}</strong>
            </p>
          </div>
          <button class="btn-reset" @click.stop="$emit('reset')">
            <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M2 8a6 6 0 0111.47-2.5M14 8a6 6 0 01-11.47 2.5"/>
              <path d="M2 2v4h4M14 14v-4h-4"/>
            </svg>
            重新上传
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  loading: Boolean,
  result: Object
})

const emit = defineEmits(['upload', 'reset'])

const fileInput = ref(null)
const isDragover = ref(false)

function triggerUpload() {
  if (!props.result && !props.loading) {
    fileInput.value?.click()
  }
}

function handleFileChange(e) {
  const file = e.target.files?.[0]
  if (file) emit('upload', file)
  e.target.value = ''
}

function handleDrop(e) {
  isDragover.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) emit('upload', file)
}
</script>

<style scoped>
.upload-section {
  margin-bottom: 1rem;
}

.upload-zone {
  border: 1.5px dashed var(--color-border);
  border-radius: var(--radius-md);
  padding: 1.75rem 1.5rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: var(--color-bg-card);
  position: relative;
}

.upload-zone:hover {
  border-color: var(--color-primary-light);
  background: var(--color-primary-bg);
}

.upload-zone--active {
  border-color: var(--color-primary);
  background: var(--color-primary-bg);
  border-style: solid;
}

.upload-zone--done {
  border-style: solid;
  border-color: #A7F3D0;
  background: #F0FDF4;
  cursor: default;
  padding: 1rem 1.25rem;
}

.upload-input {
  display: none;
}

.upload-icon-wrapper {
  margin-bottom: 0.75rem;
}

.upload-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

.upload-zone--active .upload-icon {
  background: var(--color-primary);
  color: white;
}

.upload-title {
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--color-text);
  margin-bottom: 0.35rem;
}

.upload-link {
  color: var(--color-primary);
  font-weight: 600;
  text-decoration: underline;
  text-decoration-color: var(--color-primary-light);
  text-underline-offset: 2px;
}

.upload-hint {
  font-size: 0.78rem;
  color: var(--color-text-muted);
}

/* Result */
.upload-result {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  justify-content: center;
}

.result-icon {
  flex-shrink: 0;
}

.result-info {
  text-align: left;
  flex: 1;
  min-width: 200px;
}

.result-title {
  font-weight: 600;
  color: #065F46;
  font-size: 0.95rem;
}

.result-detail {
  font-size: 0.78rem;
  color: #047857;
  margin-top: 0.1rem;
}

.result-detail strong {
  color: #065F46;
  font-weight: 600;
}

.result-sep {
  color: #6EE7B7;
  margin: 0 0.15rem;
}

.btn-reset {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0.75rem;
  border-radius: var(--radius-sm);
  font-size: 0.78rem;
  font-weight: 500;
  border: 1px solid #A7F3D0;
  background: white;
  color: #047857;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.btn-reset:hover {
  background: #ECFDF5;
  border-color: #6EE7B7;
}

@media (max-width: 480px) {
  .upload-zone {
    padding: 1.25rem 1rem;
  }

  .upload-result {
    flex-direction: column;
    text-align: center;
  }

  .result-info {
    text-align: center;
  }

  .upload-title {
    font-size: 0.85rem;
  }
}
</style>
