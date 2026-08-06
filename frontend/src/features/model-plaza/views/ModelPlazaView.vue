<template>
  <AppLayout v-if="isEmbedded">
    <PlazaPageBody :response="response" :loading="loading" :error="error" embedded @retry="load" />
  </AppLayout>

  <div v-else class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <ExtensionNavBar />
    <main class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <PlazaPageBody :response="response" :loading="loading" :error="error" @retry="load" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import ExtensionNavBar from '../components/ExtensionNavBar.vue'
import PlazaPageBody from '../components/PlazaPageBody.vue'
import { loadModelPlaza } from '../api/modelPlaza'
import type { ModelPlazaResponse } from '../types/modelPlaza'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const response = ref<ModelPlazaResponse | null>(null)
const loading = ref(true)
const error = ref(false)
let controller: AbortController | null = null

const isEmbedded = computed(() => route.query.embedded === '1' && authStore.isAuthenticated)

async function load(): Promise<void> {
  controller?.abort()
  const requestController = new AbortController()
  controller = requestController
  loading.value = true
  error.value = false
  try {
    void appStore.fetchPublicSettings()
    response.value = await loadModelPlaza({ signal: requestController.signal })
  } catch (cause) {
    if ((cause as { name?: string }).name !== 'AbortError') error.value = true
  } finally {
    if (controller === requestController && !requestController.signal.aborted) loading.value = false
  }
}

onMounted(() => {
  void load()
})

onUnmounted(() => {
  controller?.abort()
})
</script>
