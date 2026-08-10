<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl">
      <header class="mb-5 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div class="lg:hidden">
          <div class="flex items-center gap-2">
            <Icon name="book" size="lg" class="text-primary-500" aria-hidden="true" />
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('tutorials.title') }}</h1>
          </div>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('tutorials.description') }}</p>
        </div>

        <div
          class="tabs w-full sm:ml-auto sm:w-auto"
          role="tablist"
          :aria-label="t('tutorials.title')"
        >
          <button
            id="tutorial-documents-tab"
            type="button"
            role="tab"
            aria-controls="tutorial-documents-panel"
            :aria-selected="activePanel === 'documents'"
            :tabindex="activePanel === 'documents' ? 0 : -1"
            class="tab inline-flex flex-1 items-center justify-center gap-2 sm:flex-none"
            :class="{ 'tab-active': activePanel === 'documents' }"
            @click="activePanel = 'documents'"
            @keydown.right.prevent="focusPanel('videos')"
            @keydown.left.prevent="focusPanel('videos')"
          >
            <Icon name="book" size="sm" aria-hidden="true" />
            {{ t('tutorials.documents.title') }}
          </button>
          <button
            id="tutorial-videos-tab"
            type="button"
            role="tab"
            aria-controls="tutorial-videos-panel"
            :aria-selected="activePanel === 'videos'"
            :tabindex="activePanel === 'videos' ? 0 : -1"
            class="tab inline-flex flex-1 items-center justify-center gap-2 sm:flex-none"
            :class="{ 'tab-active': activePanel === 'videos' }"
            @click="activePanel = 'videos'"
            @keydown.right.prevent="focusPanel('documents')"
            @keydown.left.prevent="focusPanel('documents')"
          >
            <Icon name="play" size="sm" aria-hidden="true" />
            {{ t('tutorials.videos.title') }}
            <span
              v-if="videos.length > 0"
              class="ml-0.5 inline-flex min-w-5 items-center justify-center rounded-full bg-primary-100 px-1.5 py-0.5 text-[11px] font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300"
            >
              {{ videos.length }}
            </span>
          </button>
        </div>
      </header>

      <section
        class="overflow-hidden rounded-lg border border-gray-200/80 bg-white/90 shadow-sm dark:border-dark-700 dark:bg-dark-800/80"
        :aria-label="t('tutorials.title')"
      >
        <div class="flex min-h-[76px] items-center justify-between gap-4 border-b border-gray-100 px-5 py-4 dark:border-dark-700 sm:px-6">
          <div class="min-w-0">
            <h2 class="font-semibold text-gray-900 dark:text-white">
              {{ activePanel === 'documents' ? t('tutorials.documents.title') : t('tutorials.videos.title') }}
            </h2>
            <p class="mt-1 truncate text-sm text-gray-500 dark:text-dark-400">
              {{ activePanel === 'documents' ? t('tutorials.documents.description') : t('tutorials.videos.description') }}
            </p>
          </div>
          <Icon
            :name="activePanel === 'documents' ? 'book' : 'play'"
            size="md"
            class="flex-shrink-0 text-primary-500"
            aria-hidden="true"
          />
        </div>

        <Transition name="tutorial-panel" mode="out-in">
          <div
            v-if="activePanel === 'documents'"
            id="tutorial-documents-panel"
            role="tabpanel"
            aria-labelledby="tutorial-documents-tab"
            class="h-[min(68vh,720px)] min-h-[420px] bg-white dark:bg-dark-900"
          >
            <iframe
              :src="TUTORIAL_DOCUMENT_URL"
              :title="t('tutorials.documents.title')"
              class="h-full w-full border-0"
              loading="lazy"
              referrerpolicy="strict-origin-when-cross-origin"
            />
          </div>

          <div
            v-else
            id="tutorial-videos-panel"
            role="tabpanel"
            aria-labelledby="tutorial-videos-tab"
            class="min-h-[min(68vh,720px)] p-5 sm:p-6"
          >
            <div v-if="loading" class="flex min-h-[min(58vh,620px)] items-center justify-center text-sm text-gray-500 dark:text-dark-400" role="status">
              {{ t('tutorials.videos.loading') }}
            </div>
            <div v-else-if="loadError" class="flex min-h-[min(58vh,620px)] flex-col items-center justify-center gap-3 text-center" role="alert">
              <p class="text-sm text-red-600 dark:text-red-300">{{ t('tutorials.videos.error') }}</p>
              <button type="button" class="btn btn-secondary btn-sm" @click="loadSettings">
                <Icon name="refresh" size="sm" aria-hidden="true" />
                {{ t('common.refresh') }}
              </button>
            </div>
            <div v-else-if="videos.length === 0" class="flex min-h-[min(58vh,620px)] flex-col items-center justify-center text-center">
              <Icon name="play" size="xl" class="mb-3 text-gray-300 dark:text-dark-600" aria-hidden="true" />
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('tutorials.videos.empty') }}</p>
            </div>
            <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
              <a
                v-for="video in videos"
                :key="video.id"
                :href="safeVideoUrl(video.video_url)"
                target="_blank"
                rel="noopener noreferrer"
                class="group overflow-hidden rounded-lg border border-gray-200 bg-white transition-shadow hover:shadow-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:border-dark-700 dark:bg-dark-800"
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
          </div>
        </Transition>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { sanitizeUrl } from '@/utils/url'
import { getPublicTutorialVideos, TUTORIAL_DOCUMENT_URL } from '../api/tutorials'

const { t } = useI18n()
const appStore = useAppStore()
const activePanel = ref<'documents' | 'videos'>('documents')
const loading = ref(true)
const loadError = ref(false)
const videos = computed(() =>
  getPublicTutorialVideos(appStore.cachedPublicSettings).filter((video) => Boolean(safeVideoUrl(video.video_url)))
)

function safeVideoUrl(url: string): string {
  return sanitizeUrl(url)
}

function safeCoverUrl(url: string): string {
  return sanitizeUrl(url, { allowRelative: true, allowDataUrl: true })
}

async function focusPanel(panel: 'documents' | 'videos'): Promise<void> {
  activePanel.value = panel
  await nextTick()
  document.getElementById(`tutorial-${panel}-tab`)?.focus()
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

<style scoped>
.tutorial-panel-enter-active,
.tutorial-panel-leave-active {
  transition: opacity 160ms ease, transform 160ms ease;
}

.tutorial-panel-enter-from,
.tutorial-panel-leave-to {
  opacity: 0;
  transform: translateY(4px);
}

@media (prefers-reduced-motion: reduce) {
  .tutorial-panel-enter-active,
  .tutorial-panel-leave-active {
    transition: none;
  }
}
</style>
