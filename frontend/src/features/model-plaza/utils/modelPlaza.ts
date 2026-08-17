import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_VIDEO,
  BILLING_MODE_TOKEN,
  type BillingMode
} from '@/constants/channel'
import { formatScaled } from '@/utils/pricing'
import type {
  ModelPlazaGroup,
  ModelPricingInterval,
  PlazaModel
} from '../types/modelPlaza'

export type BillingFilter = 'all' | BillingMode

export interface PlazaFilters {
  platform: string
  billing: BillingFilter
  query: string
}

export interface PlazaStats {
  groups: number
  models: number
  platforms: number
}

const PER_MILLION = 1_000_000
const MIN_DECIMALS = 2

export function effectiveGroupRate(group: ModelPlazaGroup): number {
  return group.user_rate_multiplier ?? group.rate_multiplier
}

export function modelBillingMode(model: PlazaModel): BillingMode {
  return model.pricing?.billing_mode ?? BILLING_MODE_TOKEN
}

export function effectiveModelRate(group: ModelPlazaGroup, model: PlazaModel): number {
  if (modelBillingMode(model) === BILLING_MODE_IMAGE && group.image_rate_independent) {
    return group.image_rate_multiplier
  }
  return effectiveGroupRate(group)
}

export function paidTokenPrice(
  group: ModelPlazaGroup,
  value: number | null | undefined,
  rechargeMultiplier = 10
): string {
  if (value == null) return '-'
  return formatCny(value * effectiveGroupRate(group), PER_MILLION, rechargeMultiplier)
}

export function paidRequestPrice(
  group: ModelPlazaGroup,
  model: PlazaModel,
  value: number | null | undefined,
  rechargeMultiplier = 10
): string {
  if (value == null) return '-'
  return formatCny(value * effectiveModelRate(group, model), 1, rechargeMultiplier)
}

export function officialTokenPrice(value: number | null | undefined): string {
  if (value == null) return '-'
  return formatScaled(value, PER_MILLION, MIN_DECIMALS)
}

export function tokenIntervals(model: PlazaModel): ModelPricingInterval[] {
  if (modelBillingMode(model) !== BILLING_MODE_TOKEN) return []
  return model.pricing?.intervals ?? []
}

export function requestIntervals(model: PlazaModel): ModelPricingInterval[] {
  if (modelBillingMode(model) === BILLING_MODE_TOKEN) return []
  return (model.pricing?.intervals ?? []).filter((interval) => interval.per_request_price != null)
}

export function hasConfiguredPricing(model: PlazaModel): boolean {
  const pricing = model.pricing
  if (!pricing) return false
  if (modelBillingMode(model) === BILLING_MODE_TOKEN) {
    return Boolean(
      tokenIntervals(model).length ||
        pricing.input_price != null ||
        pricing.output_price != null ||
        pricing.cache_write_price != null ||
        pricing.cache_read_price != null
    )
  }
  return Boolean(requestIntervals(model).length || pricing.per_request_price != null)
}

export function tierLabel(interval: ModelPricingInterval): string {
  if (interval.tier_label) return interval.tier_label
  if (interval.max_tokens == null) return `>${formatTokenCount(interval.min_tokens)}`
  if (interval.min_tokens === 0) return `≤${formatTokenCount(interval.max_tokens)}`
  return `${formatTokenCount(interval.min_tokens)}–${formatTokenCount(interval.max_tokens)}`
}

export function filterPlazaGroups(
  groups: ModelPlazaGroup[],
  filters: PlazaFilters
): ModelPlazaGroup[] {
  const query = filters.query.trim().toLocaleLowerCase()
  return groups
    .filter((group) => filters.platform === 'all' || group.platform === filters.platform)
    .map((group) => ({
      ...group,
      models: group.models.filter((model) => {
        if (filters.billing !== 'all' && modelBillingMode(model) !== filters.billing) return false
        return !query || model.name.toLocaleLowerCase().includes(query)
      })
    }))
    .filter((group) => group.models.length > 0)
    .sort(
      (left, right) =>
        effectiveGroupRate(left) - effectiveGroupRate(right) ||
        left.name.localeCompare(right.name)
    )
}

export function sortModels(models: PlazaModel[]): PlazaModel[] {
  const modeOrder: Record<BillingMode, number> = {
    [BILLING_MODE_TOKEN]: 0,
    [BILLING_MODE_PER_REQUEST]: 1,
    [BILLING_MODE_IMAGE]: 2,
    [BILLING_MODE_VIDEO]: 3
  }
  return [...models].sort((left, right) => {
    const modeDifference = modeOrder[modelBillingMode(left)] - modeOrder[modelBillingMode(right)]
    return modeDifference || left.name.localeCompare(right.name)
  })
}

export function collectPlazaStats(groups: ModelPlazaGroup[]): PlazaStats {
  const models = new Set<string>()
  const platforms = new Set<string>()
  for (const group of groups) {
    if (group.platform) platforms.add(group.platform)
    for (const model of group.models) models.add(`${model.platform}:${model.name}`)
  }
  return { groups: groups.length, models: models.size, platforms: platforms.size }
}

export function modelResultCount(groups: ModelPlazaGroup[]): number {
  return groups.reduce((count, group) => count + group.models.length, 0)
}

function formatTokenCount(value: number): string {
  if (value >= 1_000_000) return `${trimNumber(value / 1_000_000)}M`
  if (value >= 1_000) return `${trimNumber(value / 1_000)}K`
  return String(value)
}

function trimNumber(value: number): string {
  return String(Math.round(value * 100) / 100)
}

/** 将 USD 额度价格换算为用户实际支付的人民币价格。 */
function formatCny(value: number, scale: number, rechargeMultiplier: number): string {
  const multiplier = Number.isFinite(rechargeMultiplier) && rechargeMultiplier > 0 ? rechargeMultiplier : 10
  return formatScaled(value / multiplier, scale, MIN_DECIMALS).replace(/^\$/, '¥')
}
