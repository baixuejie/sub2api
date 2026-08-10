<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <header>
        <div class="flex items-center gap-2">
          <Icon name="book" size="lg" class="text-primary-500" aria-hidden="true" />
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('tutorials.title') }}</h1>
        </div>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('tutorials.description') }}</p>
      </header>

      <section class="card overflow-hidden p-0" aria-labelledby="tutorial-documents-title">
        <div class="border-b border-gray-100 px-5 py-4 dark:border-dark-700 sm:px-6">
          <h2 id="tutorial-documents-title" class="font-semibold text-gray-900 dark:text-white">{{ t('tutorials.documents.title') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('tutorials.documents.description') }}</p>
        </div>
        <iframe
          :src="TUTORIAL_DOCUMENT_URL"
          :title="t('tutorials.documents.title')"
          class="h-[min(70vh,680px)] w-full border-0 bg-white dark:bg-dark-900"
          loading="lazy"
          referrerpolicy="strict-origin-when-cross-origin"
        />
      </section>

      <section class="card p-5 sm:p-6" aria-labelledby="tutorial-videos-title">
        <div class="flex items-start justify-between gap-4 border-b border-gray-100 pb-4 dark:border-dark-700">
          <div>
            <h2 id="tutorial-videos-title" class="font-semibold text-gray-900 dark:text-white">{{ t('tutorials.videos.title') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('tutorials.videos.description') }}</p>
          </div>
          <Icon name="play" size="md" class="mt-0.5 flex-shrink-0 text-primary-500" aria-hidden="true" />
        </div>

        <div v-if="loading" class="flex min-h-40 items-center justify-center text-sm text-gray-500 dark:text-dark-400" role="status">
          {{ t('tutorials.videos.loading') }}
        </div>
        <div v-else-if="loadError" class="flex min-h-40 flex-col items-center justify-center gap-3 text-center" role="alert">
          <p class="text-sm text-red-600 dark:text-red-300">{{ t('tutorials.videos.error') }}</p>
          <button type="button" class="btn btn-secondary btn-sm" @click="loadSettings">
            <Icon name="refresh" size="sm" aria-hidden="true" />
            {{ t('common.refresh') }}
          </button>
        </div>
        <div v-else-if="videos.length === 0" class="flex min-h-40 flex-col items-center justify-center text-center">
          <Icon name="play" size="xl" class="mb-3 text-gray-300 dark:text-dark-600" aria-hidden="true" />
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('tutorials.videos.empty') }}</p>
        </div>
        <div v-else class="grid grid-cols-1 gap-4 pt-5 sm:grid-cols-2 xl:grid-cols-3">
          <a
            v-for="video in videos"
            :key="video.id"
            :href="safeVideoUrl(video.video_url)"
            target="_blank"
            rel="noopener noreferrer"
            class="group overflow-hidden rounded-xl border border-gray-200 bg-white transition-shadow hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-primary-500 dark:border-dark-700 dark:bg-dark-800"
            :aria-label="t('tutorials.videos.open', { title: video.title })"
          >
            <div class="relative aspect-video overflow-hidden bg-gray-100 dark:bg-dark-900">
              <img
                v-if="safeCoverUrl(video.cover_url)"
                :src="safeCoverUrl(video.cover_url)"
                :alt="video.title"
                class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                loading="lazy"
              />
              <div v-else class="flex h-full items-center justify-center text-gray-300 dark:text-dark-600">
                <Icon name="play" size="xl" aria-hidden="true" />
              </div>
              <span class="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover:bg-black/20">
                <span class="flex h-11 w-11 items-center justify-center rounded-full bg-white/95 text-primary-600 opacity-90 shadow-lg transition-transform group-hover:scale-110 dark:bg-dark-900/95 dark:text-primary-400">
                  <Icon name="play" size="md" aria-hidden="true" />
                </span>
              </span>
            </div>
            <div class="flex items-center justify-between gap-3 px-4 py-3">
              <h3 class="min-w-0 truncate text-sm font-medium text-gray-900 dark:text-white">{{ video.title }}</h3>
              <Icon name="externalLink" size="sm" class="flex-shrink-0 text-gray-400 dark:text-dark-400" aria-hidden="true" />
            </div>
          </a>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { sanitizeUrl } from '@/utils/url'
import { getPublicTutorialVideos, TUTORIAL_DOCUMENT_URL } from '../api/tutorials'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const loadError = ref(false)
const videos = computed(() =>
  getPublicTutorialVideos(appStore.cachedPublicSettings).filter((video) => Boolean(safeVideoUrl(video.video_url)))
)

function safeVideoUrl(url: string): string {
  return sanitizeUrl(url)
}

function safeCoverUrl(url: string): string {
  return sanitizeUrl(url, { allowDataUrl: true })
}

async function loadSettings(): Promise<void> {
  loading.value = true
  loadError.value = false
  try {
    const settings = await appStore.fetchPublicSettings()
    if (!settings) {
      loadError.value = true
    }
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadSettings()
})
</script>
