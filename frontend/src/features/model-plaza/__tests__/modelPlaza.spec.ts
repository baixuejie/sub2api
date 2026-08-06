import { describe, expect, it } from 'vitest'
import { BILLING_MODE_IMAGE, BILLING_MODE_TOKEN } from '@/constants/channel'
import type { ModelPlazaGroup, PlazaModel } from '../types/modelPlaza'
import {
  collectPlazaStats,
  effectiveGroupRate,
  effectiveModelRate,
  filterPlazaGroups,
  officialTokenPrice,
  paidRequestPrice,
  paidTokenPrice,
  tierLabel,
  USD_CREDIT_PER_CNY
} from '../utils/modelPlaza'

function model(name: string, mode = BILLING_MODE_TOKEN): PlazaModel {
  return {
    name,
    platform: 'openai',
    official_pricing: null,
    pricing: {
      billing_mode: mode,
      input_price: 0.000001,
      output_price: 0.000002,
      cache_write_price: null,
      cache_read_price: null,
      image_input_price: null,
      image_output_price: null,
      per_request_price: mode === BILLING_MODE_IMAGE ? 0.25 : null,
      intervals: []
    }
  }
}

function group(overrides: Partial<ModelPlazaGroup> = {}): ModelPlazaGroup {
  return {
    id: 1,
    name: 'OpenAI 标准组',
    description: '',
    platform: 'openai',
    subscription_type: 'standard',
    rate_multiplier: 0.8,
    peak_rate_enabled: false,
    peak_start: '',
    peak_end: '',
    peak_rate_multiplier: 1,
    is_exclusive: false,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    models: [model('gpt-5'), model('dall-e', BILLING_MODE_IMAGE)],
    ...overrides
  }
}

describe('model plaza extension view model', () => {
  it('uses the user-specific group rate for token prices', () => {
    const current = group({ user_rate_multiplier: 0.5 })

    expect(effectiveGroupRate(current)).toBe(0.5)
    expect(USD_CREDIT_PER_CNY).toBe(10)
    expect(paidTokenPrice(current, 0.000002)).toBe('¥0.10')
  })

  it('uses the independent image rate only for image billing', () => {
    const current = group({
      user_rate_multiplier: 0.5,
      image_rate_independent: true,
      image_rate_multiplier: 1.2
    })
    const imageModel = model('image-1', BILLING_MODE_IMAGE)
    const tokenModel = model('gpt-5')

    expect(effectiveModelRate(current, imageModel)).toBe(1.2)
    expect(effectiveModelRate(current, tokenModel)).toBe(0.5)
    expect(paidRequestPrice(current, imageModel, 0.25)).toBe('¥0.03')
  })

  it('keeps official token prices in USD without applying the group rate', () => {
    expect(officialTokenPrice(0.000002)).toBe('$2.00')
  })

  it('filters real groups and models without mutating the response', () => {
    const source = [
      group(),
      group({
        id: 2,
        name: 'Claude 组',
        platform: 'anthropic',
        models: [{ ...model('claude-sonnet'), platform: 'anthropic' }]
      })
    ]

    const result = filterPlazaGroups(source, {
      platform: 'openai',
      billing: BILLING_MODE_TOKEN,
      query: 'GPT'
    })

    expect(result).toHaveLength(1)
    expect(result[0].models.map((item) => item.name)).toEqual(['gpt-5'])
    expect(source[0].models).toHaveLength(2)
  })

  it('counts unique models across groups and preserves actual group totals', () => {
    const groups = [group(), group({ id: 2, name: 'OpenAI 专属组' })]

    expect(collectPlazaStats(groups)).toEqual({ groups: 2, models: 2, platforms: 1 })
  })

  it('formats configured and generated tier labels', () => {
    const base = {
      input_price: null,
      output_price: null,
      cache_write_price: null,
      cache_read_price: null,
      per_request_price: null
    }

    expect(tierLabel({ ...base, min_tokens: 0, max_tokens: 200_000 })).toBe('≤200K')
    expect(tierLabel({ ...base, min_tokens: 200_000, max_tokens: null })).toBe('>200K')
    expect(tierLabel({ ...base, min_tokens: 0, max_tokens: null, tier_label: '4K' })).toBe('4K')
  })
})
