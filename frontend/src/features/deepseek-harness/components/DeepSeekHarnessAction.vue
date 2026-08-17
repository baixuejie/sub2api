<template>
  <slot name="trigger" :open="openDialog" :disabled="disabled">
    <button
      type="button"
      :disabled="disabled"
      :title="disabled ? copy.unavailableKey : copy.action"
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-cyan-50 hover:text-cyan-700 disabled:cursor-not-allowed disabled:opacity-40 dark:hover:bg-cyan-900/20 dark:hover:text-cyan-300"
      @click="openDialog"
    >
      <Icon name="download" size="sm" />
      <span class="text-xs">{{ copy.action }}</span>
    </button>
  </slot>

  <BaseDialog
    :show="showDialog"
    :title="copy.title"
    width="normal"
    :close-on-escape="!startingSession"
    @close="closeDialog"
  >
    <div class="min-h-64 space-y-5">
      <div v-if="loadingProfile" role="status" class="flex min-h-52 items-center justify-center gap-3 text-sm text-gray-500 dark:text-dark-400">
        <span class="h-5 w-5 animate-spin rounded-full border-2 border-gray-200 border-t-cyan-600 dark:border-dark-600 dark:border-t-cyan-400" aria-hidden="true"></span>
        <span>{{ copy.loading }}</span>
      </div>

      <template v-else-if="profileResponse">
        <div class="grid gap-x-5 gap-y-3 text-sm sm:grid-cols-[8rem_minmax(0,1fr)]">
          <span class="text-gray-500 dark:text-dark-400">{{ copy.key }}</span>
          <span class="min-w-0 break-all font-medium text-gray-900 dark:text-white">
            {{ profileResponse.profile.api_key_name }} · {{ profileResponse.profile.key_hint }}
          </span>
          <span class="text-gray-500 dark:text-dark-400">{{ copy.group }}</span>
          <span class="min-w-0 break-words font-medium text-gray-900 dark:text-white">
            {{ profileResponse.profile.group_name }} · {{ profileResponse.profile.platform }}
          </span>
          <span class="text-gray-500 dark:text-dark-400">{{ copy.endpoint }}</span>
          <span class="min-w-0 break-all font-mono text-xs text-gray-900 dark:text-white">
            {{ profileResponse.profile.base_url }}
          </span>
          <span class="text-gray-500 dark:text-dark-400">{{ copy.protocol }}</span>
          <span class="min-w-0 break-all font-mono text-xs text-gray-900 dark:text-white">
            {{ profileResponse.profile.protocol }}
          </span>
        </div>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{ copy.model }}</span>
          <select
            v-model="selectedModel"
            class="input w-full"
            :disabled="installInProgress"
          >
            <option
              v-for="model in profileResponse.profile.available_models"
              :key="model.id"
              :value="model.id"
            >
              {{ model.name }} ({{ model.id }})
            </option>
          </select>
        </label>

        <div class="border-y border-gray-200 py-3 text-sm dark:border-dark-700">
          <div class="mb-2 font-medium text-gray-700 dark:text-dark-200">{{ copy.environment }}</div>
          <div class="flex flex-wrap gap-x-6 gap-y-2 text-xs text-gray-500 dark:text-dark-400">
            <span>{{ copy.nodeRequirement }} {{ profileResponse.required_node }}</span>
            <span>{{ copy.helperVersion }} &gt;= {{ profileResponse.minimum_helper_version }}</span>
            <span>{{ copy.harnessVersion }} {{ profileResponse.dsh_version }}</span>
          </div>
        </div>

        <div v-if="session" class="space-y-4" aria-live="polite">
          <div class="flex items-center justify-between gap-3 text-sm">
            <span class="font-medium text-gray-900 dark:text-white">{{ stageLabel }}</span>
            <span class="font-mono text-xs text-gray-500 dark:text-dark-400">{{ displayProgress }}%</span>
          </div>
          <div
            class="h-2 overflow-hidden rounded bg-gray-100 dark:bg-dark-700"
            role="progressbar"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-valuenow="displayProgress"
          >
            <div
              class="h-full bg-cyan-600 transition-[width] duration-300 dark:bg-cyan-400"
              :class="session.status === 'failed' ? '!bg-red-500' : ''"
              :style="{ width: `${displayProgress}%` }"
            ></div>
          </div>
          <p v-if="session.message" class="break-words text-sm text-gray-500 dark:text-dark-400">
            {{ session.message }}
          </p>

          <div
            v-if="helperMayBeMissing && session.status === 'awaiting_helper'"
            class="border-l-2 border-amber-500 pl-3"
          >
            <div class="text-sm font-medium text-gray-900 dark:text-white">{{ copy.helperMissing }}</div>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ copy.helperMissingDetail }}</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <a
                v-if="helperDownloadURL"
                :href="helperDownloadURL"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-secondary btn-sm"
              >
                <Icon name="download" size="sm" />
                {{ copy.installHelper }}
              </a>
              <a
                v-else-if="safeReleasesPageURL"
                :href="safeReleasesPageURL"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-secondary btn-sm"
              >
                <Icon name="download" size="sm" />
                {{ copy.installHelper }}
              </a>
              <button type="button" class="btn btn-secondary btn-sm" @click="relaunchHelper">
                <Icon name="terminal" size="sm" />
                {{ copy.relaunch }}
              </button>
            </div>
          </div>

          <div v-if="session.status === 'completed'" class="border-l-2 border-emerald-500 pl-3">
            <div class="text-sm font-medium text-emerald-700 dark:text-emerald-300">{{ copy.completed }}</div>
            <a
              v-if="safeRunningHarnessURL"
              :href="safeRunningHarnessURL"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-primary btn-sm mt-3 inline-flex"
            >
              <Icon name="terminal" size="sm" />
              {{ copy.openHarness }}
            </a>
          </div>

          <div v-if="session.status === 'failed' || session.status === 'expired'" class="border-l-2 border-red-500 pl-3">
            <div class="text-sm font-medium text-red-700 dark:text-red-300">
              {{ session.status === 'expired' ? copy.expired : copy.failed }}
            </div>
            <p v-if="session.error_code" class="mt-1 font-mono text-xs text-gray-500 dark:text-dark-400">
              {{ session.error_code }}
            </p>
            <a
              v-if="helperUpdateRequired && helperDownloadURL"
              :href="helperDownloadURL"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary btn-sm mt-3 inline-flex"
            >
              <Icon name="download" size="sm" />
              {{ copy.updateHelper }}
            </a>
            <a
              v-else-if="helperUpdateRequired && safeReleasesPageURL"
              :href="safeReleasesPageURL"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary btn-sm mt-3 inline-flex"
            >
              <Icon name="download" size="sm" />
              {{ copy.updateHelper }}
            </a>
          </div>
        </div>

        <div v-if="errorMessage && !session" role="alert" class="border-l-2 border-red-500 pl-3 text-sm text-red-700 dark:text-red-300">
          {{ errorMessage }}
        </div>
      </template>
    </div>

    <template #footer>
      <div class="flex w-full flex-wrap justify-end gap-2">
        <button type="button" class="btn btn-secondary" @click="closeDialog">
          {{ session?.status === 'completed' ? copy.close : copy.cancel }}
        </button>
        <button
          v-if="!session || session.status === 'failed' || session.status === 'expired'"
          type="button"
          class="btn btn-primary"
          :disabled="loadingProfile || startingSession || !profileResponse"
          @click="session ? retryInstall() : startInstall()"
        >
          <span
            v-if="startingSession"
            class="h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"
            aria-hidden="true"
          ></span>
          {{ session ? copy.retry : copy.installAndStart }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import type { ApiKey } from '@/types'
import {
  createHarnessSession,
  getHarnessProfile,
  getHarnessSession
} from '../api/deepseekHarness'
import { deepSeekHarnessCopy } from '../locales'
import type {
  HarnessInstallSession,
  HarnessProfileResponse
} from '../types/deepseekHarness'
import {
  safeDownloadURL,
  safeHarnessURL,
  safeLaunchURI,
  selectHelperDownload
} from '../utils/urlPolicy'

interface Props {
  apiKeyId: number
  status: ApiKey['status']
}

const props = defineProps<Props>()
const { locale } = useI18n()
const appStore = useAppStore()
const copy = computed(() => deepSeekHarnessCopy[locale.value.toLowerCase().startsWith('zh') ? 'zh' : 'en'])
const disabled = computed(() => props.status !== 'active')
const showDialog = ref(false)
const loadingProfile = ref(false)
const startingSession = ref(false)
const helperMayBeMissing = ref(false)
const errorMessage = ref('')
const profileResponse = ref<HarnessProfileResponse | null>(null)
const session = ref<HarnessInstallSession | null>(null)
const selectedModel = ref('')
const installInProgress = computed(() =>
  Boolean(session.value && !['completed', 'failed', 'expired'].includes(session.value.status))
)
const stageLabel = computed(() => {
  const key = session.value?.stage || session.value?.status || ''
  return copy.value.stages[key] || copy.value.stages[session.value?.status || ''] || key
})
const displayProgress = computed(() => {
  const progress = Number(session.value?.progress ?? 0)
  return Number.isFinite(progress) ? Math.min(100, Math.max(0, Math.round(progress))) : 0
})
const helperDownloadURL = computed(() => {
  if (!profileResponse.value || typeof navigator === 'undefined') return ''
  const browser = navigator as Navigator & { userAgentData?: { architecture?: string } }
  return selectHelperDownload(
    profileResponse.value.helper_downloads,
    navigator.platform || '',
    navigator.userAgent || '',
    browser.userAgentData?.architecture || ''
  )
})
const safeReleasesPageURL = computed(() =>
  safeDownloadURL(profileResponse.value?.helper_downloads.releases_page)
)
const safeRunningHarnessURL = computed(() => safeHarnessURL(session.value?.harness_url))
const helperUpdateRequired = computed(() => session.value?.error_code === 'helper_update_required')
let profileController: AbortController | null = null
let createController: AbortController | null = null
let pollController: AbortController | null = null
let pollTimer: ReturnType<typeof setTimeout> | null = null
let helperTimer: ReturnType<typeof setTimeout> | null = null
let profileGeneration = 0
let createGeneration = 0
let pollFailures = 0
let notifiedSessionId = ''

async function openDialog(): Promise<void> {
  if (disabled.value) return
  showDialog.value = true
  errorMessage.value = ''
  if (session.value && installInProgress.value) {
    schedulePoll(0)
    return
  }
  if (!profileResponse.value) {
    await loadProfile()
  }
}

async function loadProfile(): Promise<void> {
  profileController?.abort()
  const controller = new AbortController()
  const generation = ++profileGeneration
  profileController = controller
  loadingProfile.value = true
  errorMessage.value = ''
  try {
    const result = await getHarnessProfile(props.apiKeyId, controller.signal)
    if (controller.signal.aborted || generation !== profileGeneration || !showDialog.value) return
    profileResponse.value = result
    selectedModel.value = result.profile.selected_model || result.profile.default_model
  } catch (error) {
    if (!isAbortError(error) && generation === profileGeneration && showDialog.value) {
      errorMessage.value = errorText(error)
    }
  } finally {
    if (generation === profileGeneration) {
      loadingProfile.value = false
      profileController = null
    }
  }
}

async function startInstall(): Promise<void> {
  if (!profileResponse.value || startingSession.value || !showDialog.value) return
  createController?.abort()
  const controller = new AbortController()
  const generation = ++createGeneration
  createController = controller
  startingSession.value = true
  helperMayBeMissing.value = false
  errorMessage.value = ''
  pollFailures = 0
  try {
    const created = await createHarnessSession(
      {
        api_key_id: props.apiKeyId,
        model: selectedModel.value || undefined
      },
      controller.signal
    )
    if (controller.signal.aborted || generation !== createGeneration || !showDialog.value) return
    session.value = created
    notifiedSessionId = ''
    invokeHelper(created.launch_uri)
    schedulePoll(600)
  } catch (error) {
    if (!isAbortError(error) && generation === createGeneration && showDialog.value) {
      errorMessage.value = errorText(error)
      appStore.showError(errorMessage.value)
    }
  } finally {
    if (generation === createGeneration) {
      startingSession.value = false
      createController = null
    }
  }
}

function invokeHelper(launchURI?: string): void {
  const validatedURI = safeLaunchURI(launchURI)
  if (!validatedURI || !showDialog.value) {
    helperMayBeMissing.value = true
    return
  }
  try {
    window.location.assign(validatedURI)
  } catch {
    helperMayBeMissing.value = true
  }
  clearHelperTimer()
  helperTimer = setTimeout(() => {
    if (session.value?.status === 'awaiting_helper') {
      helperMayBeMissing.value = true
    }
  }, 3500)
}

async function relaunchHelper(): Promise<void> {
  if (!session.value?.launch_uri || ticketExpired(session.value.ticket_expires_at)) {
    await retryInstall()
    return
  }
  helperMayBeMissing.value = false
  invokeHelper(session.value.launch_uri)
  schedulePoll(500)
}

function schedulePoll(delay = 1500): void {
  clearPollTimer()
  if (!showDialog.value || !session.value || !installInProgress.value) return
  if (sessionExpired(session.value.expires_at)) {
    session.value = { ...session.value, status: 'expired', stage: 'expired' }
    return
  }
  pollTimer = setTimeout(pollSession, delay)
}

async function pollSession(): Promise<void> {
  if (!session.value || !showDialog.value) return
  const expectedSessionId = session.value.id
  pollController?.abort()
  const controller = new AbortController()
  pollController = controller
  try {
    const latest = await getHarnessSession(expectedSessionId, controller.signal)
    if (
      controller.signal.aborted ||
      !showDialog.value ||
      pollController !== controller ||
      session.value?.id !== expectedSessionId
    ) {
      return
    }
    pollFailures = 0
    errorMessage.value = ''
    session.value = { ...session.value, ...latest }
    if (latest.status !== 'awaiting_helper') {
      helperMayBeMissing.value = false
      clearHelperTimer()
    }
    if (latest.status === 'completed') {
      if (notifiedSessionId !== latest.id) {
        notifiedSessionId = latest.id
        appStore.showSuccess(copy.value.completed)
      }
      return
    }
    if (latest.status === 'failed' || latest.status === 'expired') return
    schedulePoll()
  } catch (error) {
    if (isAbortError(error) || !showDialog.value || session.value?.id !== expectedSessionId) return
    pollFailures += 1
    errorMessage.value = errorText(error)
    const status = errorStatus(error)
    if ([401, 403, 404].includes(status) || pollFailures >= 5) {
      session.value = {
        ...session.value,
        status: status === 404 ? 'expired' : 'failed',
        stage: status === 404 ? 'expired' : 'failed',
        error_code: 'status_poll_failed',
        message: errorMessage.value
      }
      return
    }
    schedulePoll(Math.min(15000, 1500 * 2 ** pollFailures))
  } finally {
    if (pollController === controller) pollController = null
  }
}

async function retryInstall(): Promise<void> {
  clearTimers()
  session.value = null
  profileResponse.value = null
  selectedModel.value = ''
  errorMessage.value = ''
  helperMayBeMissing.value = false
  await loadProfile()
  if (profileResponse.value && showDialog.value) await startInstall()
}

function closeDialog(): void {
  showDialog.value = false
  clearTimers()
}

function clearPollTimer(): void {
  if (pollTimer) {
    clearTimeout(pollTimer)
    pollTimer = null
  }
  pollController?.abort()
  pollController = null
}

function clearHelperTimer(): void {
  if (helperTimer) {
    clearTimeout(helperTimer)
    helperTimer = null
  }
}

function clearTimers(): void {
  clearPollTimer()
  clearHelperTimer()
  profileGeneration += 1
  createGeneration += 1
  profileController?.abort()
  createController?.abort()
  profileController = null
  createController = null
  loadingProfile.value = false
  startingSession.value = false
}

function ticketExpired(value?: string): boolean {
  if (!value) return true
  const expiresAt = new Date(value).getTime()
  return !Number.isFinite(expiresAt) || expiresAt <= Date.now()
}

function sessionExpired(value: string): boolean {
  return ticketExpired(value)
}

function errorStatus(error: unknown): number {
  if (typeof error !== 'object' || error === null || !('status' in error)) return 0
  const status = Number((error as { status?: unknown }).status)
  return Number.isFinite(status) ? status : 0
}

function errorText(error: unknown): string {
  if (typeof error === 'object' && error !== null) {
    const value = error as { message?: unknown; response?: { data?: { message?: unknown } } }
    const responseMessage = value.response?.data?.message
    if (typeof responseMessage === 'string' && responseMessage.trim()) return responseMessage
    if (typeof value.message === 'string' && value.message.trim()) return value.message
  }
  return copy.value.failed
}

function isAbortError(error: unknown): boolean {
  return Boolean(
    typeof error === 'object' &&
      error !== null &&
      ('code' in error || 'name' in error) &&
      ((error as { code?: string }).code === 'ERR_CANCELED' ||
        (error as { name?: string }).name === 'AbortError')
  )
}

defineExpose({ openDialog })

onUnmounted(clearTimers)
</script>
