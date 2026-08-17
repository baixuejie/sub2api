import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import type { PublicSettings } from '@/types'
import { FeatureFlags, resolveFeatureFlag } from '@/utils/featureFlags'

const here = dirname(fileURLToPath(import.meta.url))
const read = (path: string) => readFileSync(resolve(here, path), 'utf8')

describe('DeepSeek Harness extension integration surface', () => {
  it('keeps the API key view limited to mounting the feature action', () => {
    const keysView = read('../../../views/user/KeysView.vue')
    const localToolsPosition = keysView.indexOf('<LocalToolsAction')
    const togglePosition = keysView.indexOf('toggleKeyStatus(row)')

    expect(keysView).toContain(
      "import LocalToolsAction from '@/features/local-tools/components/LocalToolsAction.vue'"
    )
    expect(keysView).toContain(
      '<LocalToolsAction :api-key="row" :public-settings="publicSettings" />'
    )
    expect(keysView).not.toContain('deepSeekHarnessEnabled')
    expect(keysView).not.toContain('deep-seek-harness-enabled')
    expect(keysView).not.toContain('show-cc-switch')
    expect(keysView).not.toContain('importToCcswitch')
    expect(keysView).not.toContain('buildCcSwitchImportDeeplink')
    expect(keysView).toContain('class="flex flex-wrap items-center gap-1"')
    expect(localToolsPosition).toBeGreaterThan(-1)
    expect(togglePosition).toBeGreaterThan(localToolsPosition)

    const localTools = read('../../local-tools/components/LocalToolsAction.vue')
    expect(localTools).toContain("from '../localToolRegistry'")
    expect(localTools).not.toContain('DeepSeekHarnessAction')
    expect(localTools).not.toContain('buildCcSwitchImportDeeplink')

    const registry = read('../../local-tools/localToolRegistry.ts')
    expect(registry).toContain("id: 'cc-switch'")
    expect(registry).toContain("id: 'deepseek-harness'")
    expect(registry).toContain('actionComponent: DeepSeekHarnessAction')
    expect(registry).toContain('resolveFeatureFlag(FeatureFlags.deepSeekHarness, publicSettings)')
  })

  it('registers the public setting as a fail-closed opt-in feature flag', () => {
    const registry = read('../../../utils/featureFlags.ts')

    expect(registry).toContain('deepSeekHarness: defineFlag({')
    expect(registry).toContain("key: 'deepseek_harness_enabled'")
    expect(registry).toContain("mode: 'opt-in'")
    expect(resolveFeatureFlag(FeatureFlags.deepSeekHarness, undefined)).toBe(false)
    expect(
      resolveFeatureFlag(FeatureFlags.deepSeekHarness, {
        deepseek_harness_enabled: true
      } as PublicSettings)
    ).toBe(true)
  })

  it('never accepts a raw API key as a component prop or session request field', () => {
    const action = read('../components/DeepSeekHarnessAction.vue')
    const api = read('../api/deepseekHarness.ts')
    const types = read('../types/deepseekHarness.ts')

    expect(action).toContain('apiKeyId: number')
    expect(action).not.toContain('apiKey: string')
    expect(api).toContain('api_key_id: apiKeyId')
    expect(types).not.toMatch(/api_key:\s*string/)
  })
})
