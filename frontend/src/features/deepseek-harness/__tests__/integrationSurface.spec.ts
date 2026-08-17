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
    const ccsPosition = keysView.indexOf('importToCcswitch(row)')
    const harnessPosition = keysView.indexOf('<DeepSeekHarnessAction')
    const togglePosition = keysView.indexOf('toggleKeyStatus(row)')

    expect(keysView).toContain(
      "import DeepSeekHarnessAction from '@/features/deepseek-harness/components/DeepSeekHarnessAction.vue'"
    )
    expect(keysView).toContain('v-if="deepSeekHarnessEnabled"')
    expect(keysView).toContain('resolveFeatureFlag(FeatureFlags.deepSeekHarness, publicSettings.value)')
    expect(keysView).toContain('class="flex flex-wrap items-center gap-1"')
    expect(ccsPosition).toBeGreaterThan(-1)
    expect(harnessPosition).toBeGreaterThan(ccsPosition)
    expect(togglePosition).toBeGreaterThan(harnessPosition)
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
