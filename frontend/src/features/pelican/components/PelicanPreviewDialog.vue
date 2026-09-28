<template>
  <BaseDialog
    :show="!!run"
    :title="run ? t(`pelican.topics.${run.topic_id}`) : ''"
    width="extra-wide"
    @close="emit('close')"
  >
    <template v-if="run">
      <PelicanPreview
        :run-id="run.id"
        :title="t(`pelican.topics.${run.topic_id}`)"
        class="rounded-xl"
        @unavailable="emit('unavailable', run.id)"
      />
      <div class="mt-4 flex flex-wrap gap-2">
        <span
          v-for="group in run.groups"
          :key="group.id"
          class="rounded-md bg-sky-50 px-2 py-1 text-xs text-sky-700 dark:bg-sky-950 dark:text-sky-300"
          >{{ group.name }}</span
        >
      </div>
      <dl class="mt-4 grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
        <div>
          <dt class="text-gray-500">{{ t('pelican.model') }}</dt>
          <dd class="mt-1 font-mono">{{ run.model }}</dd>
        </div>
        <div>
          <dt class="text-gray-500">{{ t('pelican.effort') }}</dt>
          <dd class="mt-1">{{ t('pelican.high') }}</dd>
        </div>
        <div>
          <dt class="text-gray-500">{{ t('pelican.tokens') }}</dt>
          <dd class="mt-1">{{ run.total_tokens?.toLocaleString() ?? t('pelican.unknown') }}</dd>
        </div>
        <div>
          <dt class="text-gray-500">{{ t('pelican.duration') }}</dt>
          <dd class="mt-1">
            {{ run.latency_ms === null ? '—' : `${(run.latency_ms / 1000).toFixed(1)} s` }}
          </dd>
        </div>
      </dl>
      <p class="mt-4 text-xs text-gray-500">
        {{ formatTime(run.finished_at, locale) }} · {{ t('pelican.inputTokens') }}
        {{ run.input_tokens ?? '—' }} / {{ t('pelican.outputTokens') }}
        {{ run.output_tokens ?? '—' }}
      </p>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PelicanPreview from './PelicanPreview.vue'
import { formatTime, type PelicanRun } from '../types'
defineProps<{ run: PelicanRun | null }>()
const emit = defineEmits<{ close: []; unavailable: [id: number] }>()
const { t, locale } = useI18n()
</script>
