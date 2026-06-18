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
        <div class="upload-icon">
          <svg viewBox="0 0 48 48" width="48" height="48" fill="none">
            <rect x="6" y="10" width="36" height="28" rx="4" stroke="currentColor" stroke-width="2.5"/>
            <path d="M6 30l10-8 6 5 8-10 12 10" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <circle cx="16" cy="20" r="3" stroke="currentColor" stroke-width="2"/>
          </svg>
        </div>
        <p class="upload-title">点击或拖拽上传 HTML 文件</p>
        <p class="upload-hint">支持聊天记录导出的 .html 文件</p>
      </template>

      <template v-else>
        <div class="upload-result">
          <div class="result-check">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="none">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
              <path d="M8 12l3 3 5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <div class="result-info">
            <p class="result-title">解析完成</p>
            <p class="result-detail">
              共 <strong>{{ result.totalCount }}</strong> 张图片 ·
              发送人：<strong>{{ result.sender }}</strong> ·
              会话：<strong>{{ result.recipients }}</strong>
            </p>
          </div>
          <button class="btn btn-outline btn-sm" @click.stop="$emit('reset')">
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
  margin-bottom: 1.5rem;
}

.upload-zone {
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-lg);
  padding: 2.5rem 2rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.25s ease;
  background: var(--color-bg-card);
  position: relative;
}

.upload-zone:hover {
  border-color: var(--color-primary-light);
  background: #EEF2FF;
}

.upload-zone--active {
  border-color: var(--color-primary);
  background: #EEF2FF;
  transform: scale(1.01);
}

.upload-zone--done {
  border-style: solid;
  border-color: var(--color-success);
  background: #ECFDF5;
  cursor: default;
}

.upload-input {
  display: none;
}

.upload-icon {
  color: var(--color-primary-light);
  margin-bottom: 1rem;
}

.upload-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.5rem;
}

.upload-hint {
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

/* Result */
.upload-result {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
  justify-content: center;
}

.result-check {
  color: var(--color-success);
  flex-shrink: 0;
}

.result-info {
  text-align: left;
}

.result-title {
  font-weight: 600;
  color: var(--color-text);
  font-size: 1.05rem;
}

.result-detail {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  margin-top: 0.15rem;
}

.result-detail strong {
  color: var(--color-primary-dark);
}

/* Button */
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

.btn-outline {
  background: transparent;
  border: 1.5px solid var(--color-border);
  color: var(--color-text-secondary);
}

.btn-outline:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.btn-sm {
  padding: 0.35rem 0.75rem;
  font-size: 0.8rem;
}

@media (max-width: 480px) {
  .upload-zone {
    padding: 1.5rem 1rem;
  }

  .upload-result {
    flex-direction: column;
    text-align: center;
  }

  .result-info {
    text-align: center;
  }
}
</style>
