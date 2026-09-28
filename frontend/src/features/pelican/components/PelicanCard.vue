<template>
  <article
    ref="container"
    class="group overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm transition-shadow hover:shadow-lg dark:border-dark-700 dark:bg-dark-900"
  >
    <button
      type="button"
      class="block w-full text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500"
      :aria-label="`${t('pelican.preview')} · ${title}`"
      @click="emit('open', run)"
    >
      <PelicanPreview
        v-if="active"
        :run-id="run.id"
        :title="title"
        @unavailable="emit('unavailable', run.id)"
      />
      <div
        v-else
        class="aspect-[4/3] bg-gradient-to-br from-sky-50 via-white to-amber-50 dark:from-dark-800 dark:to-dark-900"
      />
      <div class="space-y-3 p-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="font-semibold text-gray-900 dark:text-gray-100">{{ title }}</h2>
          <span
            aria-hidden="true"
            class="text-gray-400 transition-transform group-hover:translate-x-1"
            >↗</span
          >
        </div>
        <time
          :datetime="run.scheduled_for"
          class="block text-xs text-gray-500 dark:text-dark-400"
          >{{ formatTime(run.scheduled_for, locale) }}</time
        >
        <div class="flex min-h-6 flex-wrap gap-1.5">
          <span
            v-for="label in run.groups"
            :key="label.id"
            class="rounded-md bg-sky-50 px-2 py-1 text-xs font-medium text-sky-700 dark:bg-sky-950/50 dark:text-sky-300"
            >{{ label.name }}</span
          >
        </div>
      </div>
    </button>
  </article>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIntersectionObserver } from '@vueuse/core'
import PelicanPreview from './PelicanPreview.vue'
import { formatTime, type PelicanRun } from '../types'

const props = defineProps<{ run: PelicanRun; active: boolean }>()
const emit = defineEmits<{
  open: [run: PelicanRun]
  visibility: [id: number, visible: boolean]
  unavailable: [id: number]
}>()
const { t, locale } = useI18n()
const title = computed(() => t(`pelican.topics.${props.run.topic_id}`))
const container = ref<HTMLElement | null>(null)
useIntersectionObserver(
  container,
  ([entry]) => emit('visibility', props.run.id, entry?.isIntersecting ?? false),
  { threshold: 0.05 }
)
onBeforeUnmount(() => emit('visibility', props.run.id, false))
</script>
