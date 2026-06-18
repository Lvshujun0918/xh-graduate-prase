<template>
  <div class="watermark-layer" aria-hidden="true">
    <div class="watermark-pattern">
      <span v-for="i in count" :key="i" class="watermark-text">{{ text }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  text: { type: String, default: '' }
})

const count = computed(() => 80)
</script>

<style scoped>
.watermark-layer {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 9999;
  overflow: hidden;
}

.watermark-pattern {
  position: absolute;
  inset: -50%;
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  gap: 80px 120px;
  transform: rotate(-20deg);
  opacity: 0.04;
}

.watermark-text {
  font-size: 14px;
  color: #000;
  white-space: nowrap;
  user-select: none;
  font-weight: 500;
  letter-spacing: 0.1em;
}
</style>
