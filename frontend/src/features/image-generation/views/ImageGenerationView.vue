<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <header class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div class="flex items-center gap-2">
            <Icon name="sparkles" size="lg" class="text-primary-500" aria-hidden="true" />
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('imageGeneration.title') }}</h1>
          </div>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('imageGeneration.description') }}</p>
        </div>
        <button
          type="button"
          class="btn btn-secondary btn-sm self-start sm:self-auto"
          :disabled="loadingOptions || generating"
          :title="t('common.refresh')"
          @click="loadOptions"
        >
          <Icon name="refresh" size="sm" :class="loadingOptions ? 'animate-spin' : ''" aria-hidden="true" />
          {{ t('common.refresh') }}
        </button>
      </header>

      <div v-if="pageError" class="flex items-start gap-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-300" role="alert">
        <Icon name="exclamationCircle" size="md" class="mt-0.5 flex-shrink-0" aria-hidden="true" />
        <span class="min-w-0 flex-1">{{ pageError }}</span>
        <button type="button" class="btn btn-ghost btn-sm -my-1 px-2 text-red-700 dark:text-red-300" @click="loadOptions">
          {{ t('imageGeneration.actions.retry') }}
        </button>
      </div>

      <div v-if="loadingOptions" class="card flex min-h-[360px] items-center justify-center p-8">
        <LoadingSpinner size="lg" />
      </div>

      <div v-else-if="groups.length === 0" class="card flex min-h-[360px] flex-col items-center justify-center p-8 text-center">
        <Icon name="sparkles" size="xl" class="mb-4 text-gray-300 dark:text-dark-600" aria-hidden="true" />
        <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-100">{{ t('imageGeneration.empty.title') }}</h2>
        <p class="mt-2 max-w-md text-sm text-gray-500 dark:text-dark-400">{{ t('imageGeneration.empty.description') }}</p>
        <button type="button" class="btn btn-secondary mt-5" @click="loadOptions">
          <Icon name="refresh" size="sm" aria-hidden="true" />
          {{ t('imageGeneration.actions.retry') }}
        </button>
      </div>

      <div v-else class="grid grid-cols-1 items-start gap-6 xl:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
        <!-- Prompt and model controls -->
        <section class="card p-5 sm:p-6">
          <div class="mb-5 flex items-center gap-2 border-b border-gray-100 pb-4 dark:border-dark-700">
            <Icon name="cog" size="md" class="text-primary-500" aria-hidden="true" />
            <h2 class="font-semibold text-gray-900 dark:text-white">{{ t('imageGeneration.controls.title') }}</h2>
          </div>

          <div class="space-y-4">
            <div>
              <label class="input-label" for="image-generation-group">{{ t('imageGeneration.controls.group') }}</label>
              <Select
                id="image-generation-group"
                v-model="selectedGroupId"
                :options="groupOptions"
                :disabled="generating"
                :aria-label="t('imageGeneration.controls.group')"
                searchable
              >
                <template #option="{ option }">
                  <span class="min-w-0 truncate">{{ option.label }}</span>
                </template>
              </Select>
              <p v-if="currentGroup?.description" class="input-hint">{{ currentGroup.description }}</p>
            </div>

            <div>
              <label class="input-label" for="image-generation-model">{{ t('imageGeneration.controls.model') }}</label>
              <Select
                id="image-generation-model"
                v-model="selectedModel"
                :options="modelOptions"
                :disabled="!currentGroup || generating"
                :aria-label="t('imageGeneration.controls.model')"
                searchable
              />
            </div>

            <div>
              <label class="input-label" for="image-generation-prompt">
                {{ t('imageGeneration.controls.prompt') }}
                <span class="text-red-500">*</span>
              </label>
              <textarea
                id="image-generation-prompt"
                v-model="prompt"
                class="input min-h-[156px] resize-y"
                :class="promptError ? 'input-error' : ''"
                :placeholder="t('imageGeneration.controls.promptPlaceholder')"
                :maxlength="MAX_PROMPT_LENGTH"
                :disabled="generating"
                rows="7"
                @input="promptError = ''"
              ></textarea>
              <div class="mt-1 flex items-start justify-between gap-2 text-xs">
                <span v-if="promptError" class="text-red-500">{{ promptError }}</span>
                <span v-else class="text-gray-500 dark:text-dark-400">{{ t('imageGeneration.controls.promptHint') }}</span>
                <span class="flex-shrink-0 text-gray-400 dark:text-dark-500">{{ prompt.length }}/{{ MAX_PROMPT_LENGTH }}</span>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label" for="image-generation-size">{{ t('imageGeneration.controls.size') }}</label>
                <Select id="image-generation-size" v-model="size" :options="sizeOptions" :disabled="!currentModel || generating" />
              </div>
              <div>
                <label class="input-label" for="image-generation-count">{{ t('imageGeneration.controls.count') }}</label>
                <input
                  id="image-generation-count"
                  v-model.number="count"
                  class="input"
                  type="number"
                  min="1"
                  :max="maxGenerationCount"
                  step="1"
                  :disabled="generating"
                  @blur="normalizeCount"
                />
              </div>
            </div>

            <div v-if="size === CUSTOM_SIZE_VALUE && currentModel?.custom_size" class="rounded-xl border border-gray-100 bg-gray-50/80 p-3 dark:border-dark-700 dark:bg-dark-900/50">
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="input-label" for="image-generation-custom-width">{{ t('imageGeneration.controls.customWidth') }}</label>
                  <input
                    id="image-generation-custom-width"
                    v-model.number="customWidth"
                    class="input"
                    type="number"
                    min="16"
                    :max="currentModel.custom_size.max_edge"
                    :step="currentModel.custom_size.edge_multiple"
                    :disabled="generating"
                  />
                </div>
                <div>
                  <label class="input-label" for="image-generation-custom-height">{{ t('imageGeneration.controls.customHeight') }}</label>
                  <input
                    id="image-generation-custom-height"
                    v-model.number="customHeight"
                    class="input"
                    type="number"
                    min="16"
                    :max="currentModel.custom_size.max_edge"
                    :step="currentModel.custom_size.edge_multiple"
                    :disabled="generating"
                  />
                </div>
              </div>
              <p class="input-hint mt-2">
                {{ t('imageGeneration.controls.customSizeHint', {
                  multiple: currentModel.custom_size.edge_multiple,
                  maxEdge: currentModel.custom_size.max_edge,
                  minPixels: currentModel.custom_size.min_pixels.toLocaleString(),
                  maxPixels: currentModel.custom_size.max_pixels.toLocaleString()
                }) }}
              </p>
              <p v-if="customSizeError" class="mt-1 text-xs text-red-500">
                {{ t(`imageGeneration.validation.customSize.${customSizeError}`) }}
              </p>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label" for="image-generation-quality">{{ t('imageGeneration.controls.quality') }}</label>
                <Select id="image-generation-quality" v-model="quality" :options="qualityOptions" :disabled="!currentModel || generating" />
              </div>
              <div>
                <label class="input-label" for="image-generation-format">{{ t('imageGeneration.controls.format') }}</label>
                <Select id="image-generation-format" v-model="outputFormat" :options="formatOptions" :disabled="!currentModel || generating" />
              </div>
            </div>

            <div v-if="isCompressionVisible" class="rounded-xl border border-gray-100 bg-gray-50/80 p-3 dark:border-dark-700 dark:bg-dark-900/50">
              <div class="mb-2 flex items-center justify-between gap-3">
                <label class="input-label mb-0" for="image-generation-compression">{{ t('imageGeneration.controls.compression') }}</label>
                <span class="text-sm font-medium tabular-nums text-gray-700 dark:text-gray-200">{{ compression }}%</span>
              </div>
              <input
                id="image-generation-compression"
                v-model.number="compression"
                class="w-full accent-primary-600"
                type="range"
                min="0"
                max="100"
                step="1"
                :disabled="generating"
              />
              <p class="input-hint">{{ t('imageGeneration.controls.compressionHint') }}</p>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label" for="image-generation-background">{{ t('imageGeneration.controls.background') }}</label>
                <Select id="image-generation-background" v-model="background" :options="backgroundOptions" :disabled="!currentModel || generating" />
              </div>
              <div>
                <label class="input-label" for="image-generation-moderation">{{ t('imageGeneration.controls.moderation') }}</label>
                <Select id="image-generation-moderation" v-model="moderation" :options="moderationOptions" :disabled="!currentModel || generating" />
              </div>
            </div>

            <button type="button" class="btn btn-primary w-full" :disabled="generating || !currentModel" @click="generate">
              <Icon v-if="!generating" name="sparkles" size="sm" aria-hidden="true" />
              <Icon v-else name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
              {{ generating ? t('imageGeneration.actions.generating') : t('imageGeneration.actions.generate') }}
            </button>
          </div>
        </section>

        <!-- Generated images -->
        <section class="card min-h-[520px] p-5 sm:p-6">
          <div class="mb-5 flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 pb-4 dark:border-dark-700">
            <div class="flex items-center gap-2">
              <Icon name="sparkles" size="md" class="text-primary-500" aria-hidden="true" />
              <h2 class="font-semibold text-gray-900 dark:text-white">{{ t('imageGeneration.results.title') }}</h2>
              <span v-if="images.length" class="badge badge-gray">{{ images.length }}</span>
            </div>
            <button v-if="images.length" type="button" class="btn btn-ghost btn-sm" :disabled="generating" @click="clearResults">
              <Icon name="trash" size="sm" aria-hidden="true" />
              {{ t('imageGeneration.actions.clear') }}
            </button>
          </div>

          <div v-if="generating" class="flex min-h-[400px] flex-col items-center justify-center text-center">
            <LoadingSpinner size="lg" />
            <p class="mt-4 text-sm text-gray-600 dark:text-gray-300">{{ t('imageGeneration.results.generating') }}</p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ t('imageGeneration.results.generatingHint') }}</p>
          </div>

          <div v-else-if="generationError" class="flex min-h-[400px] flex-col items-center justify-center text-center">
            <Icon name="exclamationTriangle" size="xl" class="mb-4 text-amber-500" aria-hidden="true" />
            <p class="max-w-md text-sm text-red-600 dark:text-red-300" role="alert">{{ generationError }}</p>
            <button type="button" class="btn btn-secondary mt-5" @click="generate">
              <Icon name="refresh" size="sm" aria-hidden="true" />
              {{ t('imageGeneration.actions.retry') }}
            </button>
          </div>

          <div v-else-if="images.length === 0" class="flex min-h-[400px] flex-col items-center justify-center text-center">
            <Icon name="sparkles" size="xl" class="mb-4 text-gray-300 dark:text-dark-600" aria-hidden="true" />
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('imageGeneration.results.empty') }}</p>
          </div>

          <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <figure v-for="(image, index) in images" :key="`${image.src}-${index}`" class="group overflow-hidden rounded-xl border border-gray-100 bg-gray-50 dark:border-dark-700 dark:bg-dark-900/50">
              <div class="relative flex aspect-square items-center justify-center overflow-hidden bg-gray-100 dark:bg-dark-900">
                <img
                  :src="image.src"
                  :alt="t('imageGeneration.results.imageAlt', { index: index + 1 })"
                  class="h-full w-full object-contain"
                  loading="lazy"
                />
                <button
                  type="button"
                  class="absolute right-3 top-3 inline-flex h-9 w-9 items-center justify-center rounded-lg bg-black/60 text-white opacity-0 shadow transition-opacity hover:bg-black/75 focus:opacity-100 focus:outline-none focus:ring-2 focus:ring-white/80 group-hover:opacity-100"
                  :title="t('imageGeneration.actions.download')"
                  @click="downloadImage(image)"
                >
                  <Icon name="download" size="sm" aria-hidden="true" />
                  <span class="sr-only">{{ t('imageGeneration.actions.download') }}</span>
                </button>
              </div>
              <figcaption v-if="image.revised_prompt" class="border-t border-gray-100 px-3 py-2 text-xs text-gray-500 dark:border-dark-700 dark:text-dark-400">
                {{ image.revised_prompt }}
              </figcaption>
            </figure>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import { sanitizeUrl } from '@/utils/url'
import { generateImages, loadImageGenerationOptions } from '../api/imageGeneration'
import {
  imageSource,
  validateCustomImageSize,
  type DisplayImage,
  type ImageGenerationBackground,
  type ImageGenerationFormat,
  type ImageGenerationGroupOption,
  type ImageGenerationModelOption,
  type ImageGenerationModeration,
  type ImageGenerationQuality
} from '../types/imageGeneration'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const MAX_PROMPT_LENGTH = 10_000
const CUSTOM_SIZE_VALUE = '__custom__'

const groups = ref<ImageGenerationGroupOption[]>([])
const loadingOptions = ref(false)
const pageError = ref('')
const generationError = ref('')
const generating = ref(false)
const promptError = ref('')

const selectedGroupId = ref<number | null>(null)
const selectedModel = ref('')
const prompt = ref('')
const count = ref(1)
const size = ref('1024x1024')
const quality = ref<ImageGenerationQuality>('auto')
const outputFormat = ref<ImageGenerationFormat>('png')
const compression = ref(80)
const background = ref<ImageGenerationBackground>('auto')
const moderation = ref<ImageGenerationModeration>('auto')
const customWidth = ref(2048)
const customHeight = ref(2048)
const images = ref<DisplayImage[]>([])

let optionsController: AbortController | null = null
let generationController: AbortController | null = null

const currentGroup = computed(() => groups.value.find((group) => group.id === selectedGroupId.value) ?? null)
const currentModel = computed<ImageGenerationModelOption | null>(() => currentGroup.value?.models.find((model) => model.name === selectedModel.value) ?? null)

const groupOptions = computed<SelectOption[]>(() => groups.value.map((group) => ({
  value: group.id,
  label: group.name,
  description: group.description || undefined
})))
const modelOptions = computed<SelectOption[]>(() => (currentGroup.value?.models ?? []).map((model) => ({
  value: model.name,
  label: model.name
})))
const sizeOptions = computed<SelectOption[]>(() => {
  const options = (currentModel.value?.sizes ?? []).map((value) => ({ value, label: value }))
  if (currentModel.value?.custom_size) {
    options.push({ value: CUSTOM_SIZE_VALUE, label: t('imageGeneration.values.size.custom') })
  }
  return options
})
const qualityOptions = computed<SelectOption[]>(() => (currentModel.value?.qualities ?? []).map((value) => ({ value, label: qualityLabel(value) })))
const formatOptions = computed<SelectOption[]>(() => (currentModel.value?.output_formats ?? []).map((value) => ({ value, label: formatLabel(value) })))
const backgroundOptions = computed<SelectOption[]>(() => (currentModel.value?.backgrounds ?? []).map((value) => ({ value, label: backgroundLabel(value) })))
const moderationOptions = computed<SelectOption[]>(() => (currentModel.value?.moderations ?? []).map((value) => ({ value, label: moderationLabel(value) })))
const maxGenerationCount = computed(() => Math.max(1, currentModel.value?.max_n ?? 10))
const isCompressionVisible = computed(() => currentModel.value?.supports_compression !== false && (outputFormat.value === 'jpeg' || outputFormat.value === 'webp'))
const customSizeError = computed(() => {
  const constraints = currentModel.value?.custom_size
  if (size.value !== CUSTOM_SIZE_VALUE || !constraints) return null
  return validateCustomImageSize(customWidth.value, customHeight.value, constraints)
})

function qualityLabel(value: string): string {
  return t(`imageGeneration.values.quality.${value}`)
}

function formatLabel(value: string): string {
  return t(`imageGeneration.values.format.${value}`)
}

function backgroundLabel(value: string): string {
  return t(`imageGeneration.values.background.${value}`)
}

function moderationLabel(value: string): string {
  return t(`imageGeneration.values.moderation.${value}`)
}

function updateModelDefaults(): void {
  const model = currentModel.value
  if (!model) {
    size.value = ''
    quality.value = 'auto'
    outputFormat.value = 'png'
    background.value = 'auto'
    moderation.value = 'auto'
    normalizeCount()
    return
  }

  const customSelected = size.value === CUSTOM_SIZE_VALUE && !!model.custom_size
  if (!customSelected && !model.sizes.includes(size.value)) size.value = model.sizes[0] || ''
  if (!model.qualities.includes(quality.value)) quality.value = model.qualities[0] || 'auto'
  if (!model.output_formats.includes(outputFormat.value)) outputFormat.value = model.output_formats[0] || 'png'
  if (!model.backgrounds.includes(background.value)) background.value = model.backgrounds[0] || 'auto'
  if (!model.moderations.includes(moderation.value)) moderation.value = model.moderations[0] || 'auto'
  normalizeCount()
}

watch(selectedGroupId, (groupId) => {
  const group = groups.value.find((item) => item.id === groupId)
  selectedModel.value = group?.models[0]?.name ?? ''
  updateModelDefaults()
})

watch(selectedModel, updateModelDefaults)

function normalizeCount(): void {
  const value = Number(count.value)
  count.value = Number.isFinite(value) ? Math.min(maxGenerationCount.value, Math.max(1, Math.round(value))) : 1
}

function errorMessage(cause: unknown, fallback: string): string {
  if (cause && typeof cause === 'object' && 'message' in cause) {
    const message = (cause as { message?: unknown }).message
    if (typeof message === 'string' && message.trim()) return message
  }
  return fallback
}

async function loadOptions(): Promise<void> {
  optionsController?.abort()
  const controller = new AbortController()
  optionsController = controller
  loadingOptions.value = true
  pageError.value = ''
  try {
    const response = await loadImageGenerationOptions({ signal: controller.signal })
    const useServerDefaults = groups.value.length === 0
    groups.value = response.groups
    if (useServerDefaults) {
      size.value = response.defaults.size
      quality.value = response.defaults.quality
      outputFormat.value = response.defaults.output_format
      background.value = response.defaults.background
      moderation.value = response.defaults.moderation
      count.value = response.defaults.n
    }
    if (!groups.value.some((group) => group.id === selectedGroupId.value)) {
      selectedGroupId.value = groups.value[0]?.id ?? null
    }
    if (!selectedModel.value || !currentGroup.value?.models.some((model) => model.name === selectedModel.value)) {
      selectedModel.value = currentGroup.value?.models[0]?.name ?? ''
    }
    updateModelDefaults()
  } catch (cause) {
    if (!controller.signal.aborted) {
      pageError.value = errorMessage(cause, t('imageGeneration.errors.options'))
      groups.value = []
    }
  } finally {
    if (optionsController === controller) {
      optionsController = null
      loadingOptions.value = false
    }
  }
}

function toDisplayImages(response: { data?: Array<{ b64_json?: string | null; url?: string | null; revised_prompt?: string | null; mime_type?: string | null }> }): DisplayImage[] {
  return (response.data ?? []).flatMap((image, index) => {
    const unsafeSource = imageSource(image, outputFormat.value)
    const src = sanitizeUrl(unsafeSource, { allowDataUrl: true, allowRelative: true })
    if (!src) return []
    const extension = outputFormat.value === 'jpeg' ? 'jpg' : outputFormat.value
    return [{ ...image, src, downloadName: `sub2api-image-${Date.now()}-${index + 1}.${extension}` }]
  })
}

async function generate(): Promise<void> {
  normalizeCount()
  generationError.value = ''
  const trimmedPrompt = prompt.value.trim()
  if (!trimmedPrompt) {
    promptError.value = t('imageGeneration.validation.promptRequired')
    return
  }
  if (trimmedPrompt.length > MAX_PROMPT_LENGTH) {
    promptError.value = t('imageGeneration.validation.promptTooLong')
    return
  }
  promptError.value = ''
  if (!selectedGroupId.value || !currentModel.value) {
    generationError.value = t('imageGeneration.errors.noModel')
    return
  }
  if (customSizeError.value) {
    generationError.value = t(`imageGeneration.validation.customSize.${customSizeError.value}`)
    return
  }

  generationController?.abort()
  const controller = new AbortController()
  generationController = controller
  generating.value = true
  try {
    const payload = {
      group_id: selectedGroupId.value,
      model: currentModel.value.name,
      prompt: trimmedPrompt,
      n: count.value,
      size: size.value === CUSTOM_SIZE_VALUE ? `${customWidth.value}x${customHeight.value}` : size.value,
      quality: quality.value,
      output_format: outputFormat.value,
      ...(isCompressionVisible.value ? { output_compression: compression.value } : {}),
      background: background.value,
      moderation: moderation.value
    }
    const response = await generateImages(payload, { signal: controller.signal })
    const displayImages = toDisplayImages(response)
    if (displayImages.length === 0) {
      generationError.value = t('imageGeneration.errors.emptyResponse')
    } else {
      images.value = displayImages
    }
  } catch (cause) {
    if (!controller.signal.aborted) {
      generationError.value = errorMessage(cause, t('imageGeneration.errors.generate'))
    }
  } finally {
    if (generationController === controller) {
      generationController = null
      generating.value = false
    }
  }
}

function clearResults(): void {
  images.value = []
  generationError.value = ''
}

async function downloadImage(image: DisplayImage): Promise<void> {
  try {
    const response = await fetch(image.src, { credentials: 'omit' })
    if (!response.ok) throw new Error('download failed')
    const blob = await response.blob()
    const objectUrl = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = objectUrl
    anchor.download = image.downloadName
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(objectUrl)
  } catch {
    // Remote URLs may disallow CORS; opening the protected URL is the least
    // surprising fallback and does not expose any credential to the page.
    window.open(image.src, '_blank', 'noopener,noreferrer')
  }
}

onMounted(() => {
  void loadOptions()
})

onUnmounted(() => {
  optionsController?.abort()
  generationController?.abort()
})
</script>
