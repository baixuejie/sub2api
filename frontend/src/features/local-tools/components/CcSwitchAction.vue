<template>
  <BaseDialog
    :show="showClientSelect"
    :title="t('keys.ccsClientSelect.title')"
    width="narrow"
    @close="closeClientSelect"
  >
    <div class="space-y-4">
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ t('keys.ccsClientSelect.description') }}
      </p>
      <div class="grid grid-cols-2 gap-3">
        <button
          type="button"
          class="flex flex-col items-center gap-2 rounded-lg border-2 border-gray-200 p-4 transition-all hover:border-primary-500 hover:bg-primary-50 dark:border-dark-600 dark:hover:border-primary-500 dark:hover:bg-primary-900/20"
          @click="handleClientSelect('claude')"
        >
          <Icon name="terminal" size="xl" class="text-gray-600 dark:text-gray-400" />
          <span class="font-medium text-gray-900 dark:text-white">{{ t('keys.ccsClientSelect.claudeCode') }}</span>
          <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('keys.ccsClientSelect.claudeCodeDesc') }}</span>
        </button>
        <button
          type="button"
          class="flex flex-col items-center gap-2 rounded-lg border-2 border-gray-200 p-4 transition-all hover:border-primary-500 hover:bg-primary-50 dark:border-dark-600 dark:hover:border-primary-500 dark:hover:bg-primary-900/20"
          @click="handleClientSelect('gemini')"
        >
          <Icon name="sparkles" size="xl" class="text-gray-600 dark:text-gray-400" />
          <span class="font-medium text-gray-900 dark:text-white">{{ t('keys.ccsClientSelect.geminiCli') }}</span>
          <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('keys.ccsClientSelect.geminiCliDesc') }}</span>
        </button>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <button type="button" class="btn btn-secondary" @click="closeClientSelect">
          {{ t('common.cancel') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ApiKey, PublicSettings } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import {
  buildCcSwitchImportDeeplink,
  type CcSwitchClientType
} from '@/utils/ccswitchImport'

const props = defineProps<{
  apiKey: ApiKey
  publicSettings: PublicSettings | null
}>()

const { t } = useI18n()
const appStore = useAppStore()
const showClientSelect = ref(false)
let protocolCheckTimer: ReturnType<typeof setTimeout> | null = null

function openDialog(): void {
  const platform = props.apiKey.group?.platform || 'anthropic'
  if (platform === 'antigravity') {
    showClientSelect.value = true
    return
  }
  executeImport(platform === 'gemini' ? 'gemini' : 'claude')
}

function executeImport(clientType: CcSwitchClientType): void {
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

function handleClientSelect(clientType: CcSwitchClientType): void {
  showClientSelect.value = false
  executeImport(clientType)
}

function closeClientSelect(): void {
  showClientSelect.value = false
}

onBeforeUnmount(() => {
  if (protocolCheckTimer) clearTimeout(protocolCheckTimer)
})

defineExpose({ openDialog })
</script>
