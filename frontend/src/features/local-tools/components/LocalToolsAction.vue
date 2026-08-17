<template>
  <div v-if="hasTools" ref="rootRef" class="relative inline-flex">
    <button
      type="button"
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-primary-50 hover:text-primary-600 disabled:cursor-not-allowed disabled:opacity-40 dark:hover:bg-primary-900/20 dark:hover:text-primary-400"
      :disabled="allToolsDisabled"
      :title="allToolsDisabled ? copy.unavailable : copy.menuTitle"
      aria-haspopup="menu"
      :aria-expanded="menuOpen"
      @click.stop="toggleMenu"
    >
      <span class="flex items-center gap-0.5">
        <Icon name="terminal" size="sm" />
        <Icon name="chevronDown" size="xs" class="opacity-70" />
      </span>
      <span class="text-xs">{{ copy.menuLabel }}</span>
    </button>

    <DeepSeekHarnessAction
      v-if="deepSeekHarnessEnabled"
      ref="harnessActionRef"
      :api-key-id="apiKey.id"
      :status="apiKey.status"
    >
      <template #trigger><span class="hidden" aria-hidden="true"></span></template>
    </DeepSeekHarnessAction>

    <Teleport to="body">
      <div
        v-if="menuOpen && menuPosition"
        ref="menuRef"
        class="fixed z-[100000030] w-64 max-w-[calc(100vw-16px)] overflow-hidden rounded-xl border border-gray-200 bg-white p-1.5 shadow-xl shadow-gray-900/10 dark:border-dark-600 dark:bg-dark-800 dark:shadow-black/30"
        :style="{ top: `${menuPosition.top}px`, left: `${menuPosition.left}px` }"
        role="menu"
        :aria-label="copy.menuTitle"
        @click.stop
      >
        <div class="px-2.5 pb-1.5 pt-1 text-[11px] font-semibold uppercase tracking-wide text-gray-400 dark:text-dark-500">
          {{ copy.selectTool }}
        </div>
        <button
          v-if="showCcSwitch"
          type="button"
          class="flex w-full items-center gap-3 rounded-lg px-2.5 py-2 text-left transition-colors hover:bg-blue-50 dark:hover:bg-blue-900/20"
          role="menuitem"
          @click="selectCcSwitch"
        >
          <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300">
            <Icon name="upload" size="sm" />
          </span>
          <span class="min-w-0">
            <span class="block text-sm font-medium text-gray-900 dark:text-white">{{ copy.ccSwitch }}</span>
            <span class="block truncate text-xs text-gray-500 dark:text-dark-400">{{ copy.ccSwitchDescription }}</span>
          </span>
        </button>

        <button
          v-if="deepSeekHarnessEnabled"
          type="button"
          class="flex w-full items-center gap-3 rounded-lg px-2.5 py-2 text-left transition-colors hover:bg-cyan-50 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-cyan-900/20"
          role="menuitem"
          :disabled="apiKey.status !== 'active'"
          :title="apiKey.status !== 'active' ? copy.unavailable : copy.deepSeekHarnessDescription"
          @click="selectDeepSeekHarness"
        >
          <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300">
            <Icon name="download" size="sm" />
          </span>
          <span class="min-w-0">
            <span class="block text-sm font-medium text-gray-900 dark:text-white">{{ copy.deepSeekHarness }}</span>
            <span class="block truncate text-xs text-gray-500 dark:text-dark-400">{{ copy.deepSeekHarnessDescription }}</span>
          </span>
        </button>
      </div>
    </Teleport>

    <BaseDialog
      :show="showCcsClientSelect"
      :title="t('keys.ccsClientSelect.title')"
      width="narrow"
      @close="closeCcsClientSelect"
    >
      <div class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ t('keys.ccsClientSelect.description') }}
        </p>
        <div class="grid grid-cols-2 gap-3">
          <button
            type="button"
            class="flex flex-col items-center gap-2 rounded-lg border-2 border-gray-200 p-4 transition-all hover:border-primary-500 hover:bg-primary-50 dark:border-dark-600 dark:hover:border-primary-500 dark:hover:bg-primary-900/20"
            @click="handleCcsClientSelect('claude')"
          >
            <Icon name="terminal" size="xl" class="text-gray-600 dark:text-gray-400" />
            <span class="font-medium text-gray-900 dark:text-white">{{ t('keys.ccsClientSelect.claudeCode') }}</span>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('keys.ccsClientSelect.claudeCodeDesc') }}</span>
          </button>
          <button
            type="button"
            class="flex flex-col items-center gap-2 rounded-lg border-2 border-gray-200 p-4 transition-all hover:border-primary-500 hover:bg-primary-50 dark:border-dark-600 dark:hover:border-primary-500 dark:hover:bg-primary-900/20"
            @click="handleCcsClientSelect('gemini')"
          >
            <Icon name="sparkles" size="xl" class="text-gray-600 dark:text-gray-400" />
            <span class="font-medium text-gray-900 dark:text-white">{{ t('keys.ccsClientSelect.geminiCli') }}</span>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('keys.ccsClientSelect.geminiCliDesc') }}</span>
          </button>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <button type="button" class="btn btn-secondary" @click="closeCcsClientSelect">
            {{ t('common.cancel') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ApiKey, PublicSettings } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import DeepSeekHarnessAction from '@/features/deepseek-harness/components/DeepSeekHarnessAction.vue'
import { useAppStore } from '@/stores/app'
import {
  buildCcSwitchImportDeeplink,
  type CcSwitchClientType
} from '@/utils/ccswitchImport'
import { localToolsCopy } from '../localToolsCopy'

interface Props {
  apiKey: ApiKey
  publicSettings: PublicSettings | null
  showCcSwitch?: boolean
  deepSeekHarnessEnabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showCcSwitch: true,
  deepSeekHarnessEnabled: false
})
const { locale, t } = useI18n()
const appStore = useAppStore()
const copy = computed(() => localToolsCopy[locale.value.toLowerCase().startsWith('zh') ? 'zh' : 'en'])
const rootRef = ref<HTMLElement | null>(null)
const menuRef = ref<HTMLElement | null>(null)
const harnessActionRef = ref<{ openDialog: () => void } | null>(null)
const menuOpen = ref(false)
const menuPosition = ref<{ top: number; left: number } | null>(null)
const hasTools = computed(() => props.showCcSwitch || props.deepSeekHarnessEnabled)
const allToolsDisabled = computed(() => !props.showCcSwitch && props.apiKey.status !== 'active')
const showCcsClientSelect = ref(false)
let protocolCheckTimer: ReturnType<typeof setTimeout> | null = null

function updateMenuPosition(): void {
  const trigger = rootRef.value?.querySelector('button')
  if (!trigger) return
  const rect = trigger.getBoundingClientRect()
  const width = Math.min(256, window.innerWidth - 16)
  const left = Math.max(8, Math.min(rect.right - width, window.innerWidth - width - 8))
  const estimatedHeight = props.showCcSwitch && props.deepSeekHarnessEnabled ? 150 : 105
  const top = window.innerHeight - rect.bottom >= estimatedHeight + 8
    ? rect.bottom + 6
    : Math.max(8, rect.top - estimatedHeight - 6)
  menuPosition.value = { top, left }
}

async function toggleMenu(): Promise<void> {
  if (allToolsDisabled.value) return
  menuOpen.value = !menuOpen.value
  if (menuOpen.value) {
    await nextTick()
    updateMenuPosition()
    window.addEventListener('resize', updateMenuPosition)
    window.addEventListener('scroll', updateMenuPosition, true)
    document.addEventListener('pointerdown', handleOutsidePointerDown, true)
  } else {
    removeGlobalListeners()
  }
}

function closeMenu(): void {
  menuOpen.value = false
  menuPosition.value = null
  removeGlobalListeners()
}

function selectCcSwitch(): void {
  closeMenu()
  const platform = props.apiKey.group?.platform || 'anthropic'
  if (platform === 'antigravity') {
    showCcsClientSelect.value = true
    return
  }
  executeCcsImport(platform === 'gemini' ? 'gemini' : 'claude')
}

function selectDeepSeekHarness(): void {
  closeMenu()
  harnessActionRef.value?.openDialog()
}

function executeCcsImport(clientType: CcSwitchClientType): void {
  const baseUrl = props.publicSettings?.api_base_url || window.location.origin
  const platform = props.apiKey.group?.platform || 'anthropic'
  const usageScript = `({
    request: {
      url: "{{baseUrl}}/v1/usage",
      method: "GET",
      headers: { "Authorization": "Bearer {{apiKey}}" }
    },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "USD";
      return {
        isValid: response?.is_active ?? response?.isValid ?? true,
        remaining,
        unit
      };
    }
  })`
  const providerName = (props.publicSettings?.site_name || 'sub2api').trim() || 'sub2api'
  const deeplink = buildCcSwitchImportDeeplink({
    baseUrl,
    platform,
    clientType,
    providerName,
    apiKey: props.apiKey.key,
    usageScript
  })

  try {
    window.open(deeplink, '_self')
    if (protocolCheckTimer) clearTimeout(protocolCheckTimer)
    protocolCheckTimer = setTimeout(() => {
      protocolCheckTimer = null
      if (document.hasFocus()) appStore.showError(t('keys.ccSwitchNotInstalled'))
    }, 100)
  } catch {
    appStore.showError(t('keys.ccSwitchNotInstalled'))
  }
}

function handleCcsClientSelect(clientType: CcSwitchClientType): void {
  showCcsClientSelect.value = false
  executeCcsImport(clientType)
}

function closeCcsClientSelect(): void {
  showCcsClientSelect.value = false
}

function handleOutsidePointerDown(event: PointerEvent): void {
  const target = event.target as Node | null
  if (target && (rootRef.value?.contains(target) || menuRef.value?.contains(target))) return
  closeMenu()
}

function removeGlobalListeners(): void {
  window.removeEventListener('resize', updateMenuPosition)
  window.removeEventListener('scroll', updateMenuPosition, true)
  document.removeEventListener('pointerdown', handleOutsidePointerDown, true)
}

onBeforeUnmount(() => {
  removeGlobalListeners()
  if (protocolCheckTimer) clearTimeout(protocolCheckTimer)
})
</script>
