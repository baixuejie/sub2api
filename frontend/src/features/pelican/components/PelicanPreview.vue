<template>
  <div ref="container" class="relative aspect-[4/3] overflow-hidden bg-white">
    <iframe
      v-if="documentHTML"
      :srcdoc="documentHTML"
      :title="title"
      sandbox=""
      referrerpolicy="no-referrer"
      tabindex="-1"
      class="pointer-events-none absolute left-0 top-0 border-0"
      :style="frameStyle"
    />
    <div
      v-else
      class="absolute inset-0 animate-pulse bg-gradient-to-br from-sky-50 via-white to-amber-50 dark:from-dark-800 dark:to-dark-900"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useElementSize } from '@vueuse/core'
import { pelicanAPI } from '../api'
import { buildPreviewHTML, PREVIEW_POLICY_VERSION } from '../previewPolicy'

const props = defineProps<{ runId: number; title: string }>()
const emit = defineEmits<{ unavailable: [] }>()
const container = ref<HTMLElement | null>(null)
const { width } = useElementSize(container)
const documentHTML = ref<string | null>(null)
const frameStyle = computed(() => ({
  width: '960px',
  height: '720px',
  transformOrigin: 'top left',
  transform: `scale(${width.value / 960})`
}))
let controller: AbortController | undefined
watch(
  () => props.runId,
  async (id) => {
    controller?.abort()
    const request = new AbortController()
    controller = request
    documentHTML.value = null
    try {
      const data = await pelicanAPI.artifact(id, request.signal)
      if (request.signal.aborted) return
      documentHTML.value =
        data.policy_version === PREVIEW_POLICY_VERSION ? buildPreviewHTML(data.html) : null
      if (!documentHTML.value) emit('unavailable')
    } catch {
      if (!request.signal.aborted) emit('unavailable')
    }
  },
  { immediate: true }
)
onBeforeUnmount(() => controller?.abort())
</script>
