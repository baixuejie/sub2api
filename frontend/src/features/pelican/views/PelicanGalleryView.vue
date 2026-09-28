<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div class="flex flex-wrap items-center gap-3">
            <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
              {{ t('pelican.title') }}
            </h1>
            <span
              v-if="status"
              class="rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300"
              >{{ t('pelican.intervalBadge', { minutes: status.interval_seconds / 60 }) }}</span
            >
          </div>
          <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
            {{ t('pelican.description') }}
          </p>
          <p class="mt-2 text-xs text-gray-400 dark:text-dark-500">
            gpt-6-astra · {{ t('pelican.high') }}
          </p>
        </div>
        <div class="flex items-center gap-2">
          <RouterLink v-if="auth.isAdmin" to="/admin/pelican" class="btn btn-secondary">{{
            t('pelican.settings')
          }}</RouterLink>
          <button
            class="btn btn-primary"
            :disabled="loading"
            :title="t('pelican.refreshHint')"
            @click="refresh"
          >
            {{ t('pelican.refresh') }}
          </button>
        </div>
      </header>

      <div class="card flex flex-wrap items-center justify-between gap-4 p-4">
        <div class="flex flex-wrap gap-3">
          <select
            v-model="topic"
            class="input w-auto min-w-36"
            :aria-label="t('pelican.filterTopics')"
          >
            <option value="">{{ t('pelican.allTopics') }}</option>
            <option v-for="id in topicIDs" :key="id" :value="id">
              {{ t(`pelican.topics.${id}`) }}
            </option>
          </select>
          <select
            v-model="groupID"
            class="input w-auto min-w-36"
            :aria-label="t('pelican.filterGroups')"
          >
            <option value="">{{ t('pelican.allGroups') }}</option>
            <option
              v-for="group in groups"
              :key="`${group.id}-${group.name}`"
              :value="String(group.id)"
            >
              {{ group.name }}
            </option>
          </select>
        </div>
        <div v-if="status" class="space-y-1 text-xs text-gray-500 dark:text-dark-400">
          <p>{{ t('pelican.lastSuccess') }} · {{ formatTime(status.last_success_at, locale) }}</p>
          <p>
            {{
              status.enabled
                ? `${t('pelican.nextRun')} · ${formatTime(status.next_run_at, locale)}`
                : t('pelican.paused')
            }}
          </p>
        </div>
      </div>

      <div
        v-if="items.length"
        class="grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 min-[1920px]:grid-cols-5"
      >
        <PelicanCard
          v-for="run in items"
          :key="run.id"
          :run="run"
          :active="activeIDs.has(run.id)"
          @open="selected = $event"
          @visibility="setVisibility"
          @unavailable="hideUnavailable"
        />
      </div>
      <div
        v-else
        class="rounded-2xl border border-dashed border-gray-200 py-24 text-center dark:border-dark-700"
      >
        <div class="mb-4 text-4xl" aria-hidden="true">❄</div>
        <p class="font-medium text-gray-700 dark:text-dark-200">{{ t('pelican.empty') }}</p>
        <p class="mt-2 text-sm text-gray-400">{{ t('pelican.emptyHint') }}</p>
      </div>
      <nav
        v-if="history.length || nextCursor"
        class="flex justify-center gap-3"
        aria-label="Pagination"
      >
        <button
          class="btn btn-secondary"
          :disabled="loading || !history.length"
          @click="previousPage"
        >
          {{ t('pelican.previous') }}
        </button>
        <button class="btn btn-secondary" :disabled="loading || !nextCursor" @click="nextPage">
          {{ t('pelican.next') }}
        </button>
      </nav>
    </div>
    <PelicanPreviewDialog :run="selected" @close="selected = null" @unavailable="hideUnavailable" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import AppLayout from '@/components/layout/AppLayout.vue'
import PelicanCard from '../components/PelicanCard.vue'
import PelicanPreviewDialog from '../components/PelicanPreviewDialog.vue'
import { pelicanAPI } from '../api'
import { formatTime, topicIDs, type GalleryStatus, type GroupTag, type PelicanRun } from '../types'

const { t, locale } = useI18n()
const auth = useAuthStore()
const items = ref<PelicanRun[]>([])
const groups = ref<GroupTag[]>([])
const status = ref<GalleryStatus | null>(null)
const selected = ref<PelicanRun | null>(null)
const topic = ref('')
const groupID = ref('')
const loading = ref(false)
const cursor = ref('')
const nextCursor = ref('')
const history = ref<string[]>([])
const visible = ref(new Set<number>())
const unavailable = new Set<number>()
const activeIDs = computed(
  () =>
    new Set(
      selected.value
        ? []
        : items.value
            .filter((item) => visible.value.has(item.id))
            .slice(0, 6)
            .map((item) => item.id)
    )
)
let request: AbortController | undefined
let timer: ReturnType<typeof setInterval> | undefined

function setVisibility(id: number, value: boolean) {
  const next = new Set(visible.value)
  if (value) next.add(id)
  else next.delete(id)
  visible.value = next
}
function hideUnavailable(id: number) {
  unavailable.add(id)
  items.value = items.value.filter((item) => item.id !== id)
  if (selected.value?.id === id) selected.value = null
}
async function refresh() {
  request?.abort()
  const current = new AbortController()
  request = current
  loading.value = true
  const results = await Promise.allSettled([
    pelicanAPI.list(
      {
        cursor: cursor.value || undefined,
        topic: topic.value || undefined,
        group_id: groupID.value ? Number(groupID.value) : undefined
      },
      false,
      current.signal
    ),
    pelicanAPI.status(current.signal),
    pelicanAPI.groups(current.signal)
  ])
  if (current.signal.aborted) return
  const [listResult, statusResult, groupsResult] = results
  if (listResult.status === 'fulfilled') {
    items.value = listResult.value.items.filter(
      (item) => item.status === 'succeeded' && !unavailable.has(item.id)
    )
    nextCursor.value = listResult.value.next_cursor
  }
  if (statusResult.status === 'fulfilled') status.value = statusResult.value
  if (groupsResult.status === 'fulfilled') groups.value = groupsResult.value
  // Failed reads retain the currently displayed creations without an error page.
  loading.value = false
}
function nextPage() {
  history.value.push(cursor.value)
  cursor.value = nextCursor.value
  void refresh()
}
function previousPage() {
  cursor.value = history.value.pop() ?? ''
  void refresh()
}
watch([topic, groupID], () => {
  cursor.value = ''
  history.value = []
  void refresh()
})
function onVisible() {
  if (!document.hidden && !loading.value) void refresh()
}
onMounted(() => {
  void refresh()
  timer = setInterval(onVisible, 60000)
  document.addEventListener('visibilitychange', onVisible)
})
onBeforeUnmount(() => {
  request?.abort()
  clearInterval(timer)
  document.removeEventListener('visibilitychange', onVisible)
})
</script>
