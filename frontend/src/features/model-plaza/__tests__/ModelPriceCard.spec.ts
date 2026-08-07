import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { BILLING_MODE_PER_REQUEST, BILLING_MODE_TOKEN } from '@/constants/channel'
import ModelPriceCard from '../components/ModelPriceCard.vue'
import type { ModelPlazaGroup, PlazaModel } from '../types/modelPlaza'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh-CN' }
    })
  }
})

function tokenModel(overrides: Partial<PlazaModel> = {}): PlazaModel {
  return {
    name: 'claude-sonnet',
    platform: 'anthropic',
    pricing: {
      billing_mode: BILLING_MODE_TOKEN,
      input_price: 3e-6,
      output_price: 15e-6,
      cache_write_price: null,
      cache_read_price: null,
      image_input_price: null,
      image_output_price: null,
      per_request_price: null,
      intervals: []
    },
    official_pricing: {
      input_price: 3e-6,
      output_price: 15e-6,
      cache_write_price: null,
      cache_read_price: null
    },
    ...overrides
  }
}

function group(models: PlazaModel[]): ModelPlazaGroup {
  return {
    id: 1,
    name: 'Anthropic 标准组',
    description: '',
    platform: 'anthropic',
    subscription_type: 'standard',
    rate_multiplier: 0.5,
    peak_rate_enabled: false,
    peak_start: '',
    peak_end: '',
    peak_rate_multiplier: 1,
    is_exclusive: false,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    models
  }
}

function mountCard(model: PlazaModel, showOfficialPrice = false, rechargeMultiplier = 10) {
  return mount(ModelPriceCard, {
    props: {
      group: group([model]),
      model,
      rechargeMultiplier,
      showOfficialPrice
    },
    global: {
      stubs: {
        Icon: { template: '<i />' },
        PlatformIcon: { template: '<i />' }
      }
    }
  })
}

describe('ModelPriceCard price modes', () => {
  it('shows the actual CNY price using the configured recharge ratio', () => {
    const wrapper = mountCard(tokenModel(), false, 10)

    expect(wrapper.text()).toContain('实际价格（人民币）')
    expect(wrapper.text()).toContain('¥0.15')
    expect(wrapper.text()).toContain('¥0.75')
    expect(wrapper.text()).toContain('¥ / 1M token')
    expect(wrapper.text()).not.toContain('$3.00')
  })

  it('updates actual prices when the recharge ratio changes', () => {
    const wrapper = mountCard(tokenModel(), false, 5)
    expect(wrapper.text()).toContain('¥0.30')
    expect(wrapper.text()).toContain('¥1.50')
  })

  it('switches token models to the official USD reference price', () => {
    const wrapper = mountCard(tokenModel(), true)

    expect(wrapper.text()).toContain('官方参考价')
    expect(wrapper.text()).toContain('$3.00')
    expect(wrapper.text()).toContain('$15.00')
    expect(wrapper.text()).toContain('$ / 1M token')
    expect(wrapper.text()).toContain('官方参考价不计入分组倍率')
    expect(wrapper.text()).not.toContain('¥0.15')
  })

  it('shows the official one-hour cache write price when it is available', () => {
    const wrapper = mountCard(tokenModel({
      official_pricing: {
        input_price: null,
        output_price: null,
        cache_write_price: null,
        cache_write_1h_price: 6e-6,
        cache_read_price: null
      }
    }), true)

    expect(wrapper.text()).toContain('缓存写入（1 小时）')
    expect(wrapper.text()).toContain('$6.00')
  })

  it('reacts when the price mode changes on the same card instance', async () => {
    const wrapper = mountCard(tokenModel())

    expect(wrapper.text()).toContain('¥0.15')
    await wrapper.setProps({ showOfficialPrice: true })
    expect(wrapper.text()).toContain('$3.00')
    expect(wrapper.text()).not.toContain('¥0.15')
  })

  it('does not present token reference prices as official per-request prices', () => {
    const requestModel = tokenModel({
      name: 'search-tool',
      pricing: {
        billing_mode: BILLING_MODE_PER_REQUEST,
        input_price: null,
        output_price: null,
        cache_write_price: null,
        cache_read_price: null,
        image_input_price: null,
        image_output_price: null,
        per_request_price: 0.2,
        intervals: []
      }
    })
    const wrapper = mountCard(requestModel, true)

    expect(wrapper.text()).toContain('暂无该计费方式的官方参考价格')
    expect(wrapper.text()).not.toContain('$3.00')
    expect(wrapper.text()).not.toContain('¥0.01')
  })
})
