import type { Component } from 'vue'
import type { ApiKey, PublicSettings } from '@/types'
import DeepSeekHarnessAction from '@/features/deepseek-harness/components/DeepSeekHarnessAction.vue'
import { FeatureFlags, resolveFeatureFlag } from '@/utils/featureFlags'
import CcSwitchAction from './components/CcSwitchAction.vue'
import type { LocalToolsCopy } from './localToolsCopy'

type ToolCopyKey = keyof Pick<
  LocalToolsCopy,
  'ccSwitch' | 'ccSwitchDescription' | 'deepSeekHarness' | 'deepSeekHarnessDescription'
>

export interface LocalToolContext {
  apiKey: ApiKey
  publicSettings: PublicSettings | null
}

export interface LocalToolDefinition {
  id: string
  labelKey: ToolCopyKey
  descriptionKey: ToolCopyKey
  icon: 'upload' | 'download'
  iconClass: string
  hoverClass: string
  actionComponent: Component
  isVisible: (context: LocalToolContext) => boolean
  isDisabled: (context: LocalToolContext) => boolean
  actionProps: (context: LocalToolContext) => Record<string, unknown>
}

// This is the only frontend registration point for API-key local tools.
// Tool-specific dialogs and launch behavior stay inside each action component.
export const localToolRegistry: readonly LocalToolDefinition[] = [
  {
    id: 'cc-switch',
    labelKey: 'ccSwitch',
    descriptionKey: 'ccSwitchDescription',
    icon: 'upload',
    iconClass: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
    hoverClass: 'hover:bg-blue-50 dark:hover:bg-blue-900/20',
    actionComponent: CcSwitchAction,
    isVisible: ({ publicSettings }) => publicSettings?.hide_ccs_import_button !== true,
    isDisabled: () => false,
    actionProps: ({ apiKey, publicSettings }) => ({ apiKey, publicSettings })
  },
  {
    id: 'deepseek-harness',
    labelKey: 'deepSeekHarness',
    descriptionKey: 'deepSeekHarnessDescription',
    icon: 'download',
    iconClass: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300',
    hoverClass: 'hover:bg-cyan-50 dark:hover:bg-cyan-900/20',
    actionComponent: DeepSeekHarnessAction,
    isVisible: ({ publicSettings }) =>
      resolveFeatureFlag(FeatureFlags.deepSeekHarness, publicSettings),
    isDisabled: ({ apiKey }) => apiKey.status !== 'active',
    actionProps: ({ apiKey }) => ({ apiKeyId: apiKey.id, status: apiKey.status })
  }
]

export function resolveLocalTools(context: LocalToolContext): LocalToolDefinition[] {
  return localToolRegistry.filter((tool) => tool.isVisible(context))
}
