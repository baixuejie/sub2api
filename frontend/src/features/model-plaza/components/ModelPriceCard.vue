<template>
  <article
    class="model-card overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-card transition-shadow hover:shadow-card-hover dark:border-dark-700/60 dark:bg-dark-800/60"
    :style="accentStyle"
  >
    <div class="h-1" :style="{ backgroundColor: platformAccentColor(group.platform) }" aria-hidden="true"></div>

    <header class="flex items-start justify-between gap-3 px-4 pb-3 pt-4 sm:px-5">
      <div class="flex min-w-0 items-start gap-3">
        <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gray-50 dark:bg-dark-900/70" :style="{ color: platformAccentColor(group.platform) }">
          <PlatformIcon :platform="group.platform as GroupPlatform" size="md" />
        </span>
        <div class="min-w-0">
          <h3 class="truncate font-mono text-[15px] font-semibold text-gray-900 dark:text-white" :title="model.name">
            {{ model.name }}
          </h3>
          <div class="mt-1 flex flex-wrap items-center gap-1.5">
            <span class="billing-badge">{{ billingLabel }}</span>
            <span v-if="!hasConfiguredPricing(model)" class="billing-badge billing-badge-muted">{{ copy.noPricing }}</span>
          </div>
        </div>
      </div>
      <button
        type="button"
        class="btn-icon shrink-0 p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-700 dark:hover:text-white"
        :aria-label="copied ? copy.copied : copy.copyModel"
        :title="copied ? copy.copied : copy.copyModel"
        @click="copyModelName"
      >
        <Icon :name="copied ? 'check' : 'copy'" size="sm" :class="copied ? 'text-emerald-500' : undefined" aria-hidden="true" />
      </button>
    </header>

    <div v-if="hasConfiguredPricing(model)" class="px-4 pb-4 sm:px-5">
      <template v-if="isTokenBilling">
        <div v-if="tokenRows.length" class="divide-y divide-gray-100 rounded-xl border border-gray-100 dark:divide-dark-700/70 dark:border-dark-700">
          <div v-for="(interval, index) in tokenRows" :key="index" class="price-row px-3 py-3">
            <div class="mb-2 flex items-center justify-between gap-2">
              <span class="text-[11px] font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">
                {{ interval ? tierLabel(interval) : copy.actualPrice }}
              </span>
              <span class="text-[11px] text-gray-400 dark:text-dark-500">{{ copy.perMillion }}</span>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <PriceCell :label="copy.input" :value="tokenPrice(interval, 'input_price')" />
              <PriceCell :label="copy.output" :value="tokenPrice(interval, 'output_price')" />
            </div>
            <div v-if="hasCache(interval)" class="mt-2 grid grid-cols-2 gap-2 border-t border-gray-100 pt-2 dark:border-dark-700/70">
              <PriceCell compact :label="copy.cacheWrite" :value="tokenPrice(interval, 'cache_write_price')" />
              <PriceCell compact :label="copy.cacheRead" :value="tokenPrice(interval, 'cache_read_price')" />
            </div>
          </div>
        </div>
      </template>

      <template v-else>
        <div class="rounded-xl border border-gray-100 px-3 py-3 dark:border-dark-700">
          <div class="mb-2 flex items-center justify-between gap-2">
            <span class="text-[11px] font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">{{ copy.actualPrice }}</span>
            <span class="text-[11px] text-gray-400 dark:text-dark-500">{{ billingLabel === copy.imageBilling ? copy.perImage : copy.perRequest }}</span>
          </div>
          <div v-if="requestRows.length > 1" class="space-y-2">
            <div v-for="(interval, index) in requestRows" :key="index" class="flex items-center justify-between gap-3 border-b border-gray-100 pb-2 last:border-0 last:pb-0 dark:border-dark-700/70">
              <span class="text-xs text-gray-500 dark:text-dark-400">{{ interval ? tierLabel(interval) : copy.actualPrice }}</span>
              <span class="font-mono text-base font-semibold text-gray-900 dark:text-white">{{ requestPrice(interval) }}<span class="ml-1 text-xs font-normal text-gray-400">{{ billingLabel === copy.imageBilling ? copy.perImage : copy.perRequest }}</span></span>
            </div>
          </div>
          <div v-else class="flex items-baseline justify-between gap-3">
            <span class="text-xs text-gray-500 dark:text-dark-400">{{ requestRows[0] ? tierLabel(requestRows[0]) : copy.actualPrice }}</span>
            <span class="font-mono text-lg font-semibold text-gray-900 dark:text-white">{{ requestPrice(requestRows[0]) }}<span class="ml-1 text-xs font-normal text-gray-400">{{ billingLabel === copy.imageBilling ? copy.perImage : copy.perRequest }}</span></span>
          </div>
        </div>
      </template>
    </div>

    <div v-else class="mx-4 mb-4 rounded-xl border border-dashed border-gray-200 px-3 py-5 text-center text-sm text-gray-400 dark:mx-5 dark:border-dark-700 dark:text-dark-500">
      {{ copy.noPricing }}
    </div>

    <footer class="border-t border-gray-100 px-4 py-3 dark:border-dark-700/70 sm:px-5">
      <div class="flex flex-wrap items-center justify-between gap-2 text-xs">
        <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
          <Icon name="calculator" size="xs" aria-hidden="true" />
          {{ copy.rate }}
          <span class="font-mono font-semibold text-gray-800 dark:text-dark-200">{{ formatMultiplier(effectiveRate) }}x</span>
        </span>
        <span v-if="usesIndependentImageRate" class="text-gray-400 dark:text-dark-500">{{ copy.imageRate }} {{ formatMultiplier(group.image_rate_multiplier) }}x</span>
        <span v-else-if="hasCustomRate" class="text-primary-600 dark:text-primary-400">{{ copy.personalRate }}</span>
      </div>
      <details v-if="hasOfficialPricing" class="mt-2 border-t border-gray-100 pt-2 text-xs dark:border-dark-700/70">
        <summary class="cursor-pointer select-none text-gray-400 transition-colors hover:text-gray-700 dark:text-dark-500 dark:hover:text-dark-200">{{ copy.officialReference }}</summary>
        <div class="mt-2 grid grid-cols-2 gap-x-3 gap-y-1 font-mono text-gray-400 dark:text-dark-500">
          <span>{{ copy.input }} {{ officialTokenPrice(model.official_pricing?.input_price) }}</span>
          <span>{{ copy.output }} {{ officialTokenPrice(model.official_pricing?.output_price) }}</span>
          <span v-if="model.official_pricing?.cache_write_price != null">{{ copy.cacheWrite }} {{ officialTokenPrice(model.official_pricing.cache_write_price) }}</span>
          <span v-if="model.official_pricing?.cache_read_price != null">{{ copy.cacheRead }} {{ officialTokenPrice(model.official_pricing.cache_read_price) }}</span>
        </div>
      </details>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import PriceCell from './PriceCell.vue'
import type { GroupPlatform } from '@/types'
import { platformAccentColor } from '@/utils/platformColors'
import { BILLING_MODE_IMAGE, BILLING_MODE_TOKEN } from '@/constants/channel'
import { useModelPlazaLocale } from '../locales'
import type { ModelPlazaGroup, ModelPricingInterval, PlazaModel } from '../types/modelPlaza'
import {
  effectiveModelRate,
  hasConfiguredPricing,
  modelBillingMode,
  officialTokenPrice,
  paidRequestPrice,
  paidTokenPrice,
  requestIntervals,
  tierLabel,
  tokenIntervals
} from '../utils/modelPlaza'

const props = defineProps<{
  group: ModelPlazaGroup
  model: PlazaModel
}>()

const copy = useModelPlazaLocale()
const copied = ref(false)
let copiedTimer: ReturnType<typeof setTimeout> | undefined

const accentStyle = computed(() => ({ '--plaza-accent': platformAccentColor(props.group.platform) }))
const isTokenBilling = computed(() => modelBillingMode(props.model) === BILLING_MODE_TOKEN)
const billingLabel = computed(() => {
  if (modelBillingMode(props.model) === BILLING_MODE_IMAGE) return copy.value.imageBilling
  if (isTokenBilling.value) return copy.value.tokenBilling
  return copy.value.requestBilling
})
const effectiveRate = computed(() => effectiveModelRate(props.group, props.model))
const hasCustomRate = computed(
  () => props.group.user_rate_multiplier != null && props.group.user_rate_multiplier !== props.group.rate_multiplier
)
const usesIndependentImageRate = computed(
  () => modelBillingMode(props.model) === BILLING_MODE_IMAGE && props.group.image_rate_independent
)
const tokenRows = computed<(ModelPricingInterval | null)[]>(() => {
  const intervals = tokenIntervals(props.model)
  return intervals.length ? intervals : [null]
})
const requestRows = computed<(ModelPricingInterval | null)[]>(() => {
  const intervals = requestIntervals(props.model)
  return intervals.length ? intervals : [null]
})
const hasOfficialPricing = computed(() => {
  const pricing = props.model.official_pricing
  return Boolean(
    pricing &&
      (pricing.input_price != null ||
        pricing.output_price != null ||
        pricing.cache_write_price != null ||
        pricing.cache_read_price != null)
  )
})

function tokenPrice(
  interval: ModelPricingInterval | null,
  field: 'input_price' | 'output_price' | 'cache_write_price' | 'cache_read_price'
): string {
  const value = interval ? interval[field] : props.model.pricing?.[field]
  return paidTokenPrice(props.group, value)
}

function requestPrice(interval: ModelPricingInterval | null): string {
  const value = interval?.per_request_price ?? props.model.pricing?.per_request_price
  return paidRequestPrice(props.group, props.model, value)
}

function hasCache(interval: ModelPricingInterval | null): boolean {
  const pricing = interval ?? props.model.pricing
  return Boolean(pricing?.cache_write_price != null || pricing?.cache_read_price != null)
}

function formatMultiplier(value: number): string {
  return String(Math.round(value * 1000) / 1000)
}

async function copyModelName(): Promise<void> {
  let success = false
  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(props.model.name)
      success = true
    } else if (typeof document !== 'undefined') {
      const textarea = document.createElement('textarea')
      textarea.value = props.model.name
      textarea.setAttribute('readonly', 'true')
      textarea.style.cssText = 'position:fixed;left:-9999px;top:0;opacity:0'
      document.body.appendChild(textarea)
      textarea.select()
      success = document.execCommand('copy')
      document.body.removeChild(textarea)
    }
  } catch {
    success = false
  }
  if (!success) return
  copied.value = true
  if (copiedTimer) clearTimeout(copiedTimer)
  copiedTimer = setTimeout(() => {
    copied.value = false
  }, 1800)
}

onUnmounted(() => {
  if (copiedTimer) clearTimeout(copiedTimer)
})
</script>

<style scoped>
.billing-badge {
  @apply inline-flex items-center rounded-md bg-primary-50 px-1.5 py-0.5 text-[10px] font-medium text-primary-700 dark:bg-primary-900/25 dark:text-primary-300;
}

.billing-badge-muted {
  @apply bg-gray-100 text-gray-500 dark:bg-dark-700/70 dark:text-dark-400;
}

.price-row {
  background: color-mix(in srgb, var(--plaza-accent) 3%, transparent);
}

.price-row + .price-row {
  background: color-mix(in srgb, var(--plaza-accent) 1.5%, transparent);
}
</style>
