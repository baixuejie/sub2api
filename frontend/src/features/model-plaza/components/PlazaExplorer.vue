<template>
  <div class="space-y-6">
    <div v-if="!embedded" class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <p class="mb-2 text-xs font-semibold uppercase tracking-[0.18em] text-primary-600 dark:text-primary-400">Sub2API</p>
        <h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white sm:text-3xl">{{ t('modelPlaza.title') }}</h1>
        <p class="mt-1.5 text-sm text-gray-500 dark:text-dark-400">{{ t('modelPlaza.description') }}</p>
      </div>
      <div class="text-right text-xs text-gray-400 dark:text-dark-500">
        <span class="font-mono text-gray-700 dark:text-dark-200">{{ stats.models }}</span> {{ copy.modelUnit }}
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-x-6 gap-y-3 border-y border-gray-200/80 py-3.5 text-sm dark:border-dark-700/70">
      <span class="mr-1 text-xs font-semibold uppercase tracking-[0.16em] text-gray-400 dark:text-dark-500">{{ copy.overview }}</span>
      <span class="inline-flex items-baseline gap-1.5"><strong class="font-mono text-base text-gray-900 dark:text-white">{{ stats.groups }}</strong><span class="text-gray-500 dark:text-dark-400">{{ copy.groupUnit }}</span></span>
      <span class="inline-flex items-baseline gap-1.5"><strong class="font-mono text-base text-gray-900 dark:text-white">{{ stats.models }}</strong><span class="text-gray-500 dark:text-dark-400">{{ copy.modelUnit }}</span></span>
      <span class="inline-flex items-baseline gap-1.5"><strong class="font-mono text-base text-gray-900 dark:text-white">{{ stats.platforms }}</strong><span class="text-gray-500 dark:text-dark-400">{{ copy.platformUnit }}</span></span>
      <div class="ml-auto flex flex-wrap items-center gap-x-4 gap-y-2">
        <span class="inline-flex items-center gap-1.5 text-xs text-gray-500 dark:text-dark-400">
          <Icon name="calculator" size="xs" class="text-primary-500" aria-hidden="true" />
          {{ copy.priceBasis }}
        </span>
        <label class="inline-flex cursor-pointer select-none items-center gap-2 text-xs font-medium text-gray-700 dark:text-dark-200">
          <input
            v-model="showOfficialPrice"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-800"
            data-test="official-price-toggle"
          />
          <span>{{ copy.showOfficialPrice }}</span>
        </label>
      </div>
    </div>

    <div v-if="descriptionHtml" class="plaza-description rounded-xl border border-primary-100 bg-primary-50/60 px-4 py-3 text-sm dark:border-primary-900/40 dark:bg-primary-950/20" v-html="descriptionHtml"></div>

    <p v-if="!isAuthenticated" class="flex items-center gap-1.5 text-xs text-gray-400 dark:text-dark-500">
      <Icon name="infoCircle" size="xs" aria-hidden="true" />
      {{ t('modelPlaza.anonymousHint') }}
    </p>

    <div class="space-y-3 border-b border-gray-200/80 pb-5 dark:border-dark-700/70">
      <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
        <label class="relative block min-w-0 xl:max-w-sm xl:flex-1">
          <span class="sr-only">{{ t('modelPlaza.filters.searchPlaceholder') }}</span>
          <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-dark-500" aria-hidden="true" />
          <input v-model="query" type="search" class="input pl-10 pr-10" :placeholder="t('modelPlaza.filters.searchPlaceholder')" />
          <button v-if="query" type="button" class="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-white" :aria-label="copy.clearSearch" :title="copy.clearSearch" @click="query = ''">
            <Icon name="x" size="xs" aria-hidden="true" />
          </button>
        </label>

        <div class="flex max-w-full items-center gap-1 overflow-x-auto rounded-xl bg-gray-100 p-1 dark:bg-dark-800" role="group" :aria-label="copy.allBilling">
          <button v-for="option in billingOptions" :key="option.value" type="button" class="whitespace-nowrap rounded-lg px-3 py-1.5 text-xs font-medium transition-colors" :class="billing === option.value ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'" @click="billing = option.value">
            {{ option.label }}
          </button>
        </div>
      </div>

      <div class="flex items-start gap-2">
        <span class="w-10 shrink-0 pt-1.5 text-[10px] font-semibold uppercase tracking-[0.16em] text-gray-400 dark:text-dark-500">{{ t('modelPlaza.filters.platformLabel') }}</span>
        <div class="flex min-w-0 flex-1 gap-1.5 overflow-x-auto pb-1">
          <button v-for="option in platformOptions" :key="option.value" type="button" class="inline-flex shrink-0 items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium transition-colors" :class="platform === option.value ? 'bg-primary-600 text-white shadow-sm shadow-primary-500/20' : 'bg-white text-gray-600 ring-1 ring-inset ring-gray-200 hover:bg-gray-50 dark:bg-dark-800 dark:text-dark-300 dark:ring-dark-700 dark:hover:bg-dark-700'" @click="platform = option.value">
            <PlatformIcon v-if="option.value !== 'all'" :platform="option.value as GroupPlatform" size="xs" />
            {{ option.label }}
          </button>
        </div>
      </div>

      <div class="flex items-center justify-between gap-3 text-xs text-gray-400 dark:text-dark-500">
        <span v-if="hasFilters">{{ copy.searchSummary.replace('{count}', String(resultCount)) }}</span>
        <span v-else>{{ copy.basePriceHint }}</span>
        <span v-if="hasFilters" class="font-mono">{{ resultCount }}</span>
      </div>
    </div>

    <div v-if="filteredGroups.length" class="grid gap-7 lg:grid-cols-[15rem_minmax(0,1fr)] lg:items-start">
      <GroupNavigator :groups="filteredGroups" :selected-id="activeGroup?.id ?? null" @select="selectGroup" />

      <div class="min-w-0 space-y-5">
        <label class="block lg:hidden">
          <span class="input-label">{{ copy.selectGroup }}</span>
          <select class="input" :value="activeGroup?.id ?? ''" @change="selectGroup(Number(($event.target as HTMLSelectElement).value))">
            <option v-for="group in filteredGroups" :key="group.id" :value="group.id">{{ group.name }} · {{ group.models.length }} {{ copy.modelCount }}</option>
          </select>
        </label>

        <section v-if="activeGroup" :key="activeGroup.id" class="space-y-4">
          <header class="border-l-4 pl-4" :style="{ borderColor: platformAccentColor(activeGroup.platform) }">
            <div class="flex flex-wrap items-center gap-2">
              <PlatformIcon :platform="activeGroup.platform as GroupPlatform" size="md" :style="{ color: platformAccentColor(activeGroup.platform) }" />
              <h2 class="text-xl font-semibold text-gray-950 dark:text-white">{{ activeGroup.name }}</h2>
              <span class="badge badge-gray">{{ activeGroup.models.length }} {{ copy.modelCount }}</span>
              <span v-if="activeGroup.is_exclusive" class="badge badge-purple"><Icon name="shield" size="xs" aria-hidden="true" />{{ copy.exclusive }}</span>
              <span v-if="activeGroup.subscription_type === 'subscription'" class="badge badge-primary">{{ copy.subscription }}</span>
            </div>
            <p v-if="activeGroup.description" class="mt-2 max-w-3xl text-sm leading-6 text-gray-500 dark:text-dark-400">{{ activeGroup.description }}</p>
            <div class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-dark-400">
              <span class="inline-flex items-center gap-1.5"><Icon name="calculator" size="xs" aria-hidden="true" />{{ copy.rate }} <strong class="font-mono text-gray-800 dark:text-dark-200">{{ formatMultiplier(effectiveGroupRate(activeGroup)) }}x</strong></span>
              <span v-if="activeGroup.user_rate_multiplier != null && activeGroup.user_rate_multiplier !== activeGroup.rate_multiplier" class="text-primary-600 dark:text-primary-400">{{ copy.personalRate }}</span>
              <span v-if="hasPeakRate(activeGroup)" class="inline-flex items-center gap-1.5 text-amber-600 dark:text-amber-400"><Icon name="clock" size="xs" aria-hidden="true" />{{ peakWindow(activeGroup) }}</span>
            </div>
          </header>

          <div class="grid gap-4 xl:grid-cols-2">
            <ModelPriceCard
              v-for="model in sortModels(activeGroup.models)"
              :key="`${activeGroup.id}-${model.name}`"
              :group="activeGroup"
              :model="model"
              :recharge-multiplier="props.response?.balance_recharge_multiplier ?? 1"
              :show-official-price="showOfficialPrice"
            />
          </div>
        </section>
      </div>
    </div>

    <div v-else class="empty-state rounded-xl border border-dashed border-gray-300 dark:border-dark-700">
      <Icon name="search" size="xl" class="text-gray-300 dark:text-dark-600" aria-hidden="true" />
      <p class="empty-state-title">{{ query.trim() ? t('modelPlaza.noSearchResult') : t('modelPlaza.empty') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { GroupPlatform } from '@/types'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { formatPeakRateWindow, hasPeakRate, serverTimezoneLabel } from '@/utils/peak-rate'
import { platformAccentColor, platformLabel } from '@/utils/platformColors'
import GroupNavigator from './GroupNavigator.vue'
import ModelPriceCard from './ModelPriceCard.vue'
import { useModelPlazaLocale } from '../locales'
import type { ModelPlazaGroup, ModelPlazaResponse } from '../types/modelPlaza'
import {
  collectPlazaStats,
  effectiveGroupRate,
  filterPlazaGroups,
  modelResultCount,
  type BillingFilter,
  sortModels
} from '../utils/modelPlaza'

const props = defineProps<{
  response: ModelPlazaResponse | null
  embedded?: boolean
}>()

const { t } = useI18n()
const copy = useModelPlazaLocale()
const authStore = useAuthStore()
const appStore = useAppStore()
const query = ref('')
const platform = ref('all')
const billing = ref<BillingFilter>('all')
const showOfficialPrice = ref(false)
const selectedGroupId = ref<number | null>(null)

const isAuthenticated = computed(() => authStore.isAuthenticated)
const groups = computed(() => props.response?.groups ?? [])
const stats = computed(() => collectPlazaStats(groups.value))
const descriptionHtml = computed(() => {
  const description = props.response?.description?.trim()
  return description ? DOMPurify.sanitize(marked.parse(description) as string) : ''
})
const platformOptions = computed(() => [
  { value: 'all', label: t('modelPlaza.filters.all') },
  ...[...new Set(groups.value.map((group) => group.platform).filter(Boolean))]
    .sort()
    .map((value) => ({ value, label: platformLabel(value) }))
])
const billingOptions = computed(() => [
  { value: 'all' as const, label: copy.value.allBilling },
  { value: 'token' as const, label: copy.value.tokenBilling },
  { value: 'per_request' as const, label: copy.value.requestBilling },
  { value: 'image' as const, label: copy.value.imageBilling }
])
const filteredGroups = computed(() =>
  filterPlazaGroups(groups.value, { platform: platform.value, billing: billing.value, query: query.value })
)
const activeGroup = computed(() =>
  filteredGroups.value.find((group) => group.id === selectedGroupId.value) ?? filteredGroups.value[0] ?? null
)
const resultCount = computed(() => modelResultCount(filteredGroups.value))
const hasFilters = computed(() => platform.value !== 'all' || billing.value !== 'all' || query.value.trim() !== '')

watch(
  filteredGroups,
  (next) => {
    if (!next.some((group) => group.id === selectedGroupId.value)) {
      selectedGroupId.value = next[0]?.id ?? null
    }
  },
  { immediate: true }
)

function selectGroup(id: number): void {
  if (filteredGroups.value.some((group) => group.id === id)) selectedGroupId.value = id
}

function formatMultiplier(value: number): string {
  return String(Math.round(value * 1000) / 1000)
}

function peakWindow(group: ModelPlazaGroup): string {
  return formatPeakRateWindow(group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}
</script>

<style scoped>
.plaza-description {
  line-height: 1.7;
  overflow-wrap: anywhere;
}

.plaza-description :deep(p) {
  @apply text-gray-700 dark:text-dark-200;
}

.plaza-description :deep(a) {
  @apply text-primary-700 underline underline-offset-4 dark:text-primary-300;
}

.plaza-description :deep(ul) {
  @apply ml-5 list-disc;
}
</style>
