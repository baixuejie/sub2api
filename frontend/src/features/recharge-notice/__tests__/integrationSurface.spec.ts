import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const here = dirname(fileURLToPath(import.meta.url))
const read = (path: string) => readFileSync(resolve(here, path), 'utf8')

describe('充值公告 Extension 集成边界', () => {
  it('公告实现保留在独立 Extension 中，充值页仅负责挂载', () => {
    const paymentView = read('../../../views/user/PaymentView.vue')
    const notice = read('../components/RechargeNotice.vue')

    expect(paymentView).toContain("import RechargeNotice from '@/features/recharge-notice/components/RechargeNotice.vue'")
    expect(paymentView).toContain('<RechargeNotice :balance-recharge-multiplier="balanceRechargeMultiplier" />')
    expect(notice).toContain("const SERVICE_WECHAT = 'Bml18303371673'")
    expect(notice).toContain('useClipboard()')
  })
})
