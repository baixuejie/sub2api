import { beforeEach, describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import RechargeNotice from '../components/RechargeNotice.vue'

const locale = vi.hoisted(() => ({ value: 'zh-CN' }))
const copyToClipboard = vi.hoisted(() => vi.fn().mockResolvedValue(true))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ locale }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard }),
}))

describe('充值公告 Extension', () => {
  beforeEach(() => {
    locale.value = 'zh-CN'
    copyToClipboard.mockClear()
  })

  it('完整展示充值比例、人工折扣条件和线上充值限制', () => {
    const wrapper = shallowMount(RechargeNotice)
    const text = wrapper.text()

    expect(text).toContain('充值优惠说明')
    expect(text).toContain('1 元人民币')
    expect(text).toContain('10 美元账户余额')
    expect(text).toContain('单笔满 200 元且未满 500 元')
    expect(text).toContain('95 折')
    expect(text).toContain('单笔满 500 元')
    expect(text).toContain('9 折')
    expect(text).toContain('包括大额充值')
    expect(text).toContain('Bml18303371673')
  })

  it('可以复制客服微信号并显示明确的成功提示', async () => {
    const wrapper = shallowMount(RechargeNotice)

    await wrapper.get('[data-test="copy-service-wechat"]').trigger('click')

    expect(copyToClipboard).toHaveBeenCalledWith('Bml18303371673', '客服微信号已复制')
  })

  it('英文界面使用对应英文公告', () => {
    locale.value = 'en'
    const wrapper = shallowMount(RechargeNotice)

    expect(wrapper.text()).toContain('Recharge offers')
    expect(wrapper.text()).toContain('all online payments made on this page')
  })

  it('倍率调整后与充值页的实际到账倍率保持一致', () => {
    const wrapper = shallowMount(RechargeNotice, {
      props: { balanceRechargeMultiplier: 0.14 },
    })

    expect(wrapper.text()).toContain('0.14 美元账户余额')
  })
})
