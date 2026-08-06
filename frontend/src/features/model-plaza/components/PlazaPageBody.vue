<template>
  <div class="space-y-6">
    <div v-if="loading" class="space-y-5" aria-busy="true">
      <div class="h-20 animate-pulse rounded-xl bg-gray-100 dark:bg-dark-800"></div>
      <div class="h-14 animate-pulse rounded-xl bg-gray-100 dark:bg-dark-800"></div>
      <div class="grid gap-4 lg:grid-cols-2">
        <div v-for="index in 4" :key="index" class="h-52 animate-pulse rounded-2xl bg-gray-100 dark:bg-dark-800"></div>
      </div>
    </div>

    <div v-else-if="error" class="empty-state rounded-xl border border-red-200 bg-red-50/70 dark:border-red-500/30 dark:bg-red-950/20">
      <Icon name="exclamationCircle" size="xl" class="text-red-400" aria-hidden="true" />
      <p class="empty-state-title">{{ t('modelPlaza.loadFailed') }}</p>
      <button type="button" class="btn btn-secondary btn-sm mt-3" @click="$emit('retry')">
        <Icon name="refresh" size="sm" aria-hidden="true" />
        {{ copy.retry }}
      </button>
    </div>

    <PlazaExplorer v-else :response="response" :embedded="embedded" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlazaExplorer from './PlazaExplorer.vue'
import { useModelPlazaLocale } from '../locales'
import type { ModelPlazaResponse } from '../types/modelPlaza'

defineProps<{
  response: ModelPlazaResponse | null
  loading: boolean
  error: boolean
  embedded?: boolean
}>()

defineEmits<{
  retry: []
}>()

const { t } = useI18n()
const copy = useModelPlazaLocale()
</script>
