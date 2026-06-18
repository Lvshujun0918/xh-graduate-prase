<template>
  <footer class="app-footer">
    <div class="footer-content">
      <div class="footer-left">
        <span class="footer-author">© {{ currentYear }} lsj</span>
      </div>
      <div class="footer-right">
        <span class="footer-version" v-if="version">
          <span class="version-label">版本</span>
          <span class="version-value">{{ version }}</span>
        </span>
      </div>
    </div>
  </footer>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getVersion } from '../api/index.js'

const currentYear = computed(() => new Date().getFullYear())
const version = ref('')

onMounted(async () => {
  try {
    const res = await getVersion()
    // 只展示 commitId 的前 7 位
    const data = res.data || {}
    if (data.commitId && data.commitId !== 'unknown') {
      version.value = data.commitId.substring(0, 7)
    } else if (data.version && data.version !== 'dev') {
      version.value = data.version
    }
  } catch {
    version.value = ''
  }
})
</script>

<style scoped>
.app-footer {
  background: var(--color-bg-card);
  border-top: 1px solid var(--color-border-light);
  padding: 1rem 1.5rem;
  margin-top: auto;
}

.footer-content {
  max-width: 1100px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.footer-author {
  font-weight: 500;
  color: var(--color-text-secondary);
}

.footer-version {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.version-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.version-value {
  font-family: 'SF Mono', 'Fira Code', 'Fira Mono', 'Roboto Mono', monospace;
  background: var(--color-bg);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

@media (max-width: 480px) {
  .app-footer {
    padding: 0.75rem 1rem;
  }

  .footer-content {
    flex-direction: column;
    gap: 0.25rem;
    text-align: center;
  }
}
</style>
