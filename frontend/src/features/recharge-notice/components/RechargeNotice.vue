<template>
  <section
    data-test="recharge-notice"
    aria-labelledby="recharge-notice-title"
    class="overflow-hidden rounded-lg border border-amber-300 bg-white shadow-sm dark:border-amber-700/70 dark:bg-dark-800"
  >
    <div class="flex items-start gap-3 border-b border-amber-200 bg-amber-50 px-4 py-3 dark:border-amber-800/70 dark:bg-amber-950/30 sm:px-5">
      <div class="mt-0.5 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-amber-100 text-amber-700 dark:bg-amber-900/60 dark:text-amber-300">
        <Icon name="exclamationTriangle" size="sm" aria-hidden="true" />
      </div>
      <div class="min-w-0">
        <h2 id="recharge-notice-title" class="text-base font-semibold text-amber-950 dark:text-amber-100">
          {{ localText('充值优惠说明', 'Recharge offers') }}
        </h2>
        <p class="mt-0.5 text-xs leading-5 text-amber-800 dark:text-amber-300">
          {{ localText('大额优惠需提前联系客服，并由管理员人工办理。', 'Contact support before paying to receive a manual large-recharge discount.') }}
        </p>
      </div>
    </div>

    <div class="px-4 py-4 sm:px-5">
      <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <span class="text-sm font-medium text-gray-600 dark:text-gray-300">
          {{ localText('充值比例', 'Recharge rate') }}
        </span>
        <p class="text-lg font-bold text-gray-950 dark:text-white">
          <span class="text-primary-600 dark:text-primary-400">1 {{ localText('元人民币', 'CNY') }}</span>
          <span class="mx-1.5 text-gray-400">=</span>
          <span class="text-emerald-600 dark:text-emerald-400">{{ rechargeMultiplierLabel }} {{ localText('美元账户余额', 'USD account balance') }}</span>
        </p>
      </div>

      <div class="mt-4 grid border-y border-gray-200 dark:border-dark-600 dark:divide-dark-600 sm:grid-cols-2 sm:divide-x sm:divide-gray-200">
        <div class="py-3 sm:pr-5">
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ localText('单笔满 200 元且未满 500 元', 'CNY 200 to under CNY 500') }}
          </p>
          <p class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
            {{ localText('人工充值', 'Manual recharge') }}
            <span class="ml-1 text-emerald-600 dark:text-emerald-400">{{ localText('95 折', '5% off') }}</span>
          </p>
        </div>
        <div class="border-t border-gray-200 py-3 dark:border-dark-600 sm:border-t-0 sm:pl-5">
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ localText('单笔满 500 元', 'CNY 500 or more') }}
          </p>
          <p class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
            {{ localText('人工充值', 'Manual recharge') }}
            <span class="ml-1 text-emerald-600 dark:text-emerald-400">{{ localText('9 折', '10% off') }}</span>
          </p>
        </div>
      </div>

      <div class="mt-3 flex items-start gap-2 text-xs leading-5 text-gray-600 dark:text-gray-300">
        <Icon name="infoCircle" size="sm" class="mt-0.5 flex-shrink-0 text-amber-600 dark:text-amber-400" aria-hidden="true" />
        <p>
          {{ localText('未满 200 元，或直接在本页面在线充值（包括大额充值），均不享受折扣优惠。', 'Recharges under CNY 200 and all online payments made on this page, including large payments, are not eligible for discounts.') }}
        </p>
      </div>

      <div class="mt-4 flex flex-col gap-3 rounded-lg bg-gray-50 px-3 py-3 dark:bg-dark-700 sm:flex-row sm:items-center sm:justify-between">
        <div class="min-w-0">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('客服微信', 'Support WeChat') }}</p>
          <p class="mt-0.5 break-all font-mono text-sm font-semibold text-gray-900 dark:text-white">{{ SERVICE_WECHAT }}</p>
        </div>
        <button
          type="button"
          data-test="copy-service-wechat"
          class="btn btn-secondary btn-sm flex w-full flex-shrink-0 items-center justify-center gap-1.5 sm:w-auto"
          :aria-label="localText('复制客服微信号', 'Copy support WeChat ID')"
          @click="copyServiceWechat"
        >
          <Icon name="copy" size="sm" aria-hidden="true" />
          {{ localText('复制微信号', 'Copy WeChat ID') }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'

const SERVICE_WECHAT = 'Bml18303371673'

const props = withDefaults(defineProps<{
  balanceRechargeMultiplier?: number
}>(), {
  balanceRechargeMultiplier: 10,
})

const { locale } = useI18n()
const { copyToClipboard } = useClipboard()
const isZhLocale = computed(() => locale.value.startsWith('zh'))
const rechargeMultiplierLabel = computed(() => {
  const multiplier = Number.isFinite(props.balanceRechargeMultiplier) && props.balanceRechargeMultiplier > 0
    ? props.balanceRechargeMultiplier
    : 10
  return multiplier.toFixed(2).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
})

function localText(zh: string, en: string): string {
  return isZhLocale.value ? zh : en
}

function copyServiceWechat(): void {
  void copyToClipboard(
    SERVICE_WECHAT,
    localText('客服微信号已复制', 'Support WeChat ID copied'),
  )
}
</script>
