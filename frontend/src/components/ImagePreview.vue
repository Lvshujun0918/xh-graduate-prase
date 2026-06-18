<template>
  <div class="preview-overlay" @click.self="$emit('close')">
    <div class="preview-modal">
      <div class="preview-header">
        <div class="preview-title">
          <h3>图片预览</h3>
          <span class="preview-name">{{ image.fileName }}</span>
        </div>
        <button class="preview-close" @click="$emit('close')">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      </div>
      <div class="preview-body">
        <img
          :src="proxyUrl"
          :alt="image.fileName"
          class="preview-image"
        />
      </div>
      <div class="preview-footer">
        <div class="preview-meta">
          <span>{{ image.datetime }}</span>
          <span class="meta-sep">·</span>
          <span>{{ image.fromName }}</span>
          <span class="meta-sep">·</span>
          <span>{{ image.imgWidth }} × {{ image.imgHeight }}</span>
        </div>
        <a :href="image.serverImgUrl" target="_blank" class="preview-original">
          查看原图
        </a>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { getProxyImageUrl } from '../api/index.js'

const props = defineProps({
  image: { type: Object, required: true }
})

defineEmits(['close'])

const proxyUrl = computed(() => getProxyImageUrl(props.image.serverImgUrl))
</script>

<style scoped>
.preview-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3000;
  padding: 1rem;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.preview-modal {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  max-width: 90vw;
  max-height: 90vh;
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.3);
  animation: scaleIn 0.25s ease;
}

@keyframes scaleIn {
  from { transform: scale(0.95); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-light);
}

.preview-title h3 {
  font-size: 1rem;
  font-weight: 600;
}

.preview-name {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  margin-top: 0.15rem;
  display: block;
}

.preview-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary);
  padding: 0.25rem;
  border-radius: var(--radius-sm);
  transition: all 0.2s;
  line-height: 0;
}

.preview-close:hover {
  background: var(--color-bg-hover);
  color: var(--color-text);
}

.preview-body {
  flex: 1;
  overflow: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: var(--color-bg);
}

.preview-image {
  max-width: 100%;
  max-height: 65vh;
  object-fit: contain;
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-md);
}

.preview-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--color-border-light);
  flex-wrap: wrap;
  gap: 0.5rem;
}

.preview-meta {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.meta-sep {
  color: var(--color-text-muted);
}

.preview-original {
  font-size: 0.8rem;
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 500;
}

.preview-original:hover {
  text-decoration: underline;
}

@media (max-width: 480px) {
  .preview-modal {
    max-width: 100vw;
    max-height: 100vh;
    border-radius: 0;
  }

  .preview-body {
    padding: 0.5rem;
  }
}
</style>
