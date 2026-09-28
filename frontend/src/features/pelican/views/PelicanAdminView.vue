<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <header class="flex flex-wrap justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('pelican.adminTitle') }}
          </h1>
          <p class="mt-2 text-sm text-gray-500">{{ t('pelican.adminDescription') }}</p>
        </div>
        <RouterLink to="/pelican" class="btn btn-secondary self-start">{{
          t('pelican.gallery')
        }}</RouterLink>
      </header>
      <p
        v-if="error"
        role="alert"
        class="rounded-xl bg-amber-50 p-3 text-sm text-amber-800 dark:bg-amber-950/40 dark:text-amber-200"
      >
        {{ error }}
      </p>
      <p
        v-if="notice"
        role="status"
        class="rounded-xl bg-emerald-50 p-3 text-sm text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-200"
      >
        {{ notice }}
      </p>
      <button v-if="!config" class="btn btn-secondary" :disabled="loading" @click="load">
        {{ loading ? t('pelican.loading') : t('pelican.reload') }}
      </button>

      <form v-if="config && form" class="card space-y-6 p-5 sm:p-6" @submit.prevent="save">
        <div
          class="flex flex-wrap items-start justify-between gap-4 border-b border-gray-100 pb-5 dark:border-dark-700"
        >
          <div class="space-y-2">
            <label class="flex items-center gap-3 font-medium"
              ><input
                v-model="form.enabled"
                type="checkbox"
                :disabled="!config.encryption_ready"
                class="h-4 w-4 rounded"
              />{{ t('pelican.enabled') }}</label
            >
            <p class="max-w-lg text-xs leading-5 text-gray-500">{{ t('pelican.scheduleHint') }}</p>
          </div>
          <div class="rounded-xl bg-gray-50 px-4 py-3 text-sm dark:bg-dark-800">
            <span class="font-mono">gpt-6-astra</span
            ><span class="ml-3 text-gray-500"
              >{{ t('pelican.high') }} · {{ t('pelican.hourly') }}</span
            >
          </div>
        </div>
        <p v-if="!config.encryption_ready" class="text-sm text-amber-700 dark:text-amber-300">
          {{ t('pelican.encryptionRequired') }}
        </p>
        <div>
          <label for="pelican-key" class="mb-2 block text-sm font-medium">{{
            t('pelican.apiKey')
          }}</label>
          <div class="flex gap-3">
            <input
              id="pelican-key"
              v-model="apiKey"
              type="password"
              autocomplete="new-password"
              maxlength="1024"
              class="input min-w-0 flex-1"
              :disabled="!config.encryption_ready"
              :placeholder="
                config.key_configured ? t('pelican.keySaved') : t('pelican.keyPlaceholder')
              "
            /><button
              v-if="config.key_configured"
              type="button"
              class="btn btn-secondary shrink-0"
              :disabled="saving"
              @click="confirmClear = true"
            >
              {{ t('pelican.clearKey') }}
            </button>
          </div>
          <p v-if="config.key_unavailable" class="mt-2 text-xs text-amber-700 dark:text-amber-300">
            {{ t('pelican.keyUnavailable') }}
          </p>
        </div>
        <fieldset>
          <legend class="text-sm font-medium">{{ t('pelican.groups') }}</legend>
          <p class="mt-1 text-xs leading-5 text-gray-500">{{ t('pelican.groupsHint') }}</p>
          <div
            class="mt-3 grid max-h-64 grid-cols-1 gap-2 overflow-y-auto rounded-xl border border-gray-200 p-3 sm:grid-cols-2 lg:grid-cols-3 dark:border-dark-700"
          >
            <label
              v-for="group in groups"
              :key="group.id"
              class="flex cursor-pointer items-center gap-2 rounded-lg p-2 text-sm hover:bg-gray-50 dark:hover:bg-dark-800"
              ><input
                v-model="form.selected_group_ids"
                type="checkbox"
                :value="group.id"
                class="h-4 w-4 rounded"
              /><span class="break-all">{{ group.name }}</span></label
            >
            <p v-if="!groups.length" class="p-2 text-sm text-gray-400">
              {{ t('pelican.noGroups') }}
            </p>
          </div>
          <p class="mt-2 text-xs text-gray-500">
            {{ t('pelican.selectedGroups', { count: form.selected_group_ids.length }) }}
          </p>
        </fieldset>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label for="pelican-mode" class="mb-2 block text-sm font-medium">{{
              t('pelican.topicMode')
            }}</label
            ><select id="pelican-mode" v-model="form.topic_mode" class="input">
              <option value="rotate">{{ t('pelican.rotate') }}</option>
              <option value="fixed">{{ t('pelican.fixed') }}</option>
            </select>
          </div>
          <div v-if="form.topic_mode === 'fixed'">
            <label for="pelican-topic" class="mb-2 block text-sm font-medium">{{
              t('pelican.fixedTopic')
            }}</label
            ><select id="pelican-topic" v-model="form.fixed_topic_id" class="input">
              <option v-for="id in topicIDs" :key="id" :value="id">
                {{ t(`pelican.topics.${id}`) }}
              </option>
            </select>
          </div>
        </div>
        <div class="grid gap-4 sm:grid-cols-3">
          <div>
            <label for="pelican-tokens" class="mb-2 block text-sm font-medium">{{
              t('pelican.maxTokens')
            }}</label
            ><input
              id="pelican-tokens"
              v-model.number="form.max_output_tokens"
              required
              type="number"
              min="4096"
              max="32768"
              class="input"
            />
          </div>
          <div>
            <label for="pelican-timeout" class="mb-2 block text-sm font-medium">{{
              t('pelican.timeout')
            }}</label
            ><input
              id="pelican-timeout"
              v-model.number="form.timeout_seconds"
              required
              type="number"
              min="60"
              max="600"
              class="input"
            />
          </div>
          <div>
            <label for="pelican-retention" class="mb-2 block text-sm font-medium">{{
              t('pelican.retention')
            }}</label
            ><input
              id="pelican-retention"
              v-model.number="form.retention_days"
              required
              type="number"
              min="7"
              max="90"
              class="input"
            />
          </div>
        </div>
        <p
          class="rounded-xl bg-sky-50 p-3 text-xs leading-5 text-sky-800 dark:bg-sky-950/30 dark:text-sky-200"
        >
          {{ t('pelican.generationHint') }}
        </p>
        <div class="flex flex-wrap items-center justify-between gap-4">
          <p class="text-xs text-gray-500">
            {{ t('pelican.nextRun') }} · {{ formatTime(config.next_run_at, locale) }}
          </p>
          <div class="flex gap-2">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="loading || saving"
              @click="load"
            >
              {{ t('pelican.reload') }}</button
            ><button type="submit" class="btn btn-primary" :disabled="saving">
              {{ saving ? t('pelican.saving') : t('pelican.save') }}
            </button>
          </div>
        </div>
      </form>

      <section class="card overflow-hidden">
        <div class="flex items-center justify-between gap-4 p-5">
          <div>
            <h2 class="font-semibold">{{ t('pelican.records') }}</h2>
            <p class="mt-1 text-xs text-gray-500">{{ t('pelican.recordsHint') }}</p>
          </div>
          <button class="btn btn-secondary" :disabled="runsLoading" @click="loadRuns">
            {{ t('pelican.reload') }}
          </button>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-800">
              <tr>
                <th class="px-5 py-3">{{ t('pelican.time') }}</th>
                <th class="px-5 py-3">{{ t('pelican.topic') }}</th>
                <th class="px-5 py-3">{{ t('pelican.status') }}</th>
                <th class="px-5 py-3">{{ t('pelican.tokens') }}</th>
                <th class="px-5 py-3">
                  <span class="sr-only">{{ t('pelican.detail') }}</span>
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="run in runs" :key="run.id">
                <td class="whitespace-nowrap px-5 py-3">
                  {{ formatTime(run.scheduled_for, locale) }}
                </td>
                <td class="px-5 py-3">{{ t(`pelican.topics.${run.topic_id}`) }}</td>
                <td class="px-5 py-3">
                  <span
                    :class="
                      run.status === 'succeeded'
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-gray-500'
                    "
                    >{{ t(`pelican.statuses.${run.status}`) }}</span
                  >
                </td>
                <td class="px-5 py-3">{{ run.total_tokens?.toLocaleString() ?? '—' }}</td>
                <td class="px-5 py-3">
                  <button class="text-primary-600 hover:underline" @click="openDetail(run)">
                    {{ t('pelican.detail') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-if="!runs.length" class="p-8 text-center text-sm text-gray-400">
          {{ t('pelican.noRecords') }}
        </p>
        <div v-if="runHistory.length || nextRunCursor" class="flex justify-center gap-3 p-4">
          <button
            class="btn btn-secondary"
            :disabled="runsLoading || !runHistory.length"
            @click="previousRuns"
          >
            {{ t('pelican.previous') }}</button
          ><button
            class="btn btn-secondary"
            :disabled="runsLoading || !nextRunCursor"
            @click="nextRuns"
          >
            {{ t('pelican.next') }}
          </button>
        </div>
      </section>
    </div>
    <BaseDialog
      :show="confirmClear"
      :title="t('pelican.clearKey')"
      width="narrow"
      @close="confirmClear = false"
      ><p class="text-sm">{{ t('pelican.clearConfirm') }}</p>
      <template #footer
        ><button class="btn btn-secondary" @click="confirmClear = false">
          {{ t('pelican.cancel') }}</button
        ><button class="btn btn-primary" :disabled="saving" @click="clearKey">
          {{ t('pelican.confirm') }}
        </button></template
      ></BaseDialog
    >
    <BaseDialog :show="!!detail" :title="t('pelican.detail')" width="wide" @close="detail = null">
      <div v-if="detail" class="space-y-4 text-sm">
        <p>
          {{ t('pelican.status') }}: {{ t(`pelican.statuses.${detail.run.status}`) }} ·
          {{ formatTime(detail.run.scheduled_for, locale) }}
        </p>
        <p v-if="detail.run.error_code">
          {{ t('pelican.diagnostic') }}: <code>{{ detail.run.error_code }}</code>
        </p>
        <p v-if="detail.skipped_hours">
          {{ t('pelican.skippedHours') }}: {{ detail.skipped_hours }}
        </p>
        <div>
          <h3 class="mb-2 font-medium">{{ t('pelican.prompt') }}</h3>
          <pre class="whitespace-pre-wrap break-words rounded-xl bg-gray-50 p-3 dark:bg-dark-800">{{
            detail.prompt
          }}</pre>
        </div>
        <details>
          <summary class="cursor-pointer font-medium">{{ t('pelican.request') }}</summary>
          <pre class="mt-2 overflow-x-auto rounded-xl bg-gray-50 p-3 dark:bg-dark-800">{{
            JSON.stringify(detail.request, null, 2)
          }}</pre>
        </details>
        <details v-if="source">
          <summary class="cursor-pointer font-medium">{{ t('pelican.source') }}</summary>
          <pre
            class="mt-2 max-h-96 overflow-auto whitespace-pre-wrap break-words rounded-xl bg-gray-50 p-3 text-xs dark:bg-dark-800"
            >{{ source }}</pre
          >
        </details>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { getAll } from '@/api/admin/groups'
import { pelicanAPI } from '../api'
import {
  formatTime,
  topicIDs,
  type GroupTag,
  type PelicanConfig,
  type PelicanRun,
  type RunDetail,
  type SaveConfig
} from '../types'

const { t, locale } = useI18n()
const config = ref<PelicanConfig | null>(null)
const form = ref<SaveConfig | null>(null)
const groups = ref<GroupTag[]>([])
const apiKey = ref('')
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const notice = ref('')
const confirmClear = ref(false)
const runs = ref<PelicanRun[]>([])
const runsLoading = ref(false)
const runCursor = ref('')
const nextRunCursor = ref('')
const runHistory = ref<string[]>([])
const detail = ref<RunDetail | null>(null)
const source = ref('')
let disposed = false
let detailRequest = 0
let runsRequest = 0
let timer: ReturnType<typeof setInterval> | undefined

function applyConfig(value: PelicanConfig) {
  config.value = value
  const {
    revision,
    enabled,
    selected_group_ids,
    topic_mode,
    fixed_topic_id,
    max_output_tokens,
    timeout_seconds,
    retention_days
  } = value
  form.value = {
    revision,
    enabled,
    selected_group_ids: [...selected_group_ids],
    topic_mode,
    fixed_topic_id,
    max_output_tokens,
    timeout_seconds,
    retention_days
  }
  apiKey.value = ''
}
async function load() {
  loading.value = true
  error.value = ''
  try {
    const [value, availableGroups] = await Promise.all([pelicanAPI.config(), getAll()])
    if (disposed) return
    groups.value = availableGroups.map(({ id, name }) => ({ id, name }))
    applyConfig(value)
    // Removed or disabled groups no longer participate in future label snapshots.
    if (form.value)
      form.value.selected_group_ids = form.value.selected_group_ids.filter((id) =>
        groups.value.some((group) => group.id === id)
      )
  } catch {
    if (!disposed) error.value = t('pelican.loadFailed')
  } finally {
    loading.value = false
  }
}
function saveError(err: unknown) {
  const status =
    typeof err === 'object' && err !== null && 'status' in err
      ? (err as { status: number }).status
      : 0
  error.value = t(status === 409 ? 'pelican.conflict' : 'pelican.saveFailed')
}
async function save() {
  if (!form.value || saving.value) return
  saving.value = true
  error.value = ''
  notice.value = ''
  try {
    const payload: SaveConfig = {
      ...form.value,
      selected_group_ids: [...form.value.selected_group_ids]
    }
    if (apiKey.value.trim()) payload.api_key = apiKey.value.trim()
    const value = await pelicanAPI.save(payload)
    if (!disposed) {
      applyConfig(value)
      notice.value = t('pelican.saved')
    }
  } catch (err) {
    if (!disposed) saveError(err)
  } finally {
    saving.value = false
  }
}
async function clearKey() {
  if (!config.value || saving.value) return
  saving.value = true
  error.value = ''
  notice.value = ''
  try {
    const value = await pelicanAPI.clearKey(config.value.revision)
    if (!disposed) {
      applyConfig(value)
      confirmClear.value = false
      notice.value = t('pelican.cleared')
    }
  } catch (err) {
    if (!disposed) saveError(err)
  } finally {
    saving.value = false
  }
}
async function loadRuns() {
  const sequence = ++runsRequest
  runsLoading.value = true
  try {
    const page = await pelicanAPI.list({ cursor: runCursor.value || undefined }, true)
    if (disposed || sequence !== runsRequest) return
    runs.value = page.items
    nextRunCursor.value = page.next_cursor
  } catch {
    /* Keep the existing log table if polling is temporarily unavailable. */
  } finally {
    if (sequence === runsRequest) runsLoading.value = false
  }
}
function nextRuns() {
  runHistory.value.push(runCursor.value)
  runCursor.value = nextRunCursor.value
  void loadRuns()
}
function previousRuns() {
  runCursor.value = runHistory.value.pop() ?? ''
  void loadRuns()
}
async function openDetail(run: PelicanRun) {
  const sequence = ++detailRequest
  source.value = ''
  try {
    const value = await pelicanAPI.detail(run.id)
    if (disposed || sequence !== detailRequest) return
    detail.value = value
    if (run.status === 'succeeded') {
      const text = await pelicanAPI.source(run.id)
      if (!disposed && sequence === detailRequest) source.value = text.source
    }
  } catch {
    if (!disposed && sequence === detailRequest) error.value = t('pelican.detailFailed')
  }
}
onMounted(() => {
  void load()
  void loadRuns()
  timer = setInterval(() => {
    if (!document.hidden && !runsLoading.value) void loadRuns()
  }, 60000)
})
onBeforeUnmount(() => {
  disposed = true
  apiKey.value = ''
  clearInterval(timer)
})
</script>
