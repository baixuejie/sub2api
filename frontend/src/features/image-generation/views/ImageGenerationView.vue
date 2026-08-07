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
        <div class="flex flex-wrap gap-2 self-start sm:self-auto">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="loadingOptions || generating"
            :title="t('common.refresh')"
            @click="loadOptions"
          >
            <Icon name="refresh" size="sm" :class="loadingOptions ? 'animate-spin' : ''" aria-hidden="true" />
            {{ t('common.refresh') }}
          </button>
          <router-link
            to="/image-generation/settings"
            class="btn btn-secondary btn-sm"
            :title="t('imageGeneration.actions.settings')"
          >
            <Icon name="cog" size="sm" aria-hidden="true" />
            {{ t('imageGeneration.actions.settings') }}
          </router-link>
        </div>
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

      <div v-else class="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
        <!-- Prompt and model controls -->
        <section class="card p-5 sm:p-6">
          <div class="mb-5 flex flex-wrap items-center gap-2 border-b border-gray-100 pb-4 dark:border-dark-700">
            <Icon name="cog" size="md" class="text-primary-500" aria-hidden="true" />
            <h2 class="font-semibold text-gray-900 dark:text-white">{{ t('imageGeneration.controls.title') }}</h2>
            <div class="ml-auto inline-flex rounded-lg border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-700 dark:bg-dark-900" role="tablist">
              <button
                type="button"
                class="inline-flex min-h-8 items-center gap-1.5 rounded-md px-2.5 text-xs font-medium transition-colors"
                :class="mode === 'generate' ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-700 dark:text-primary-300' : 'text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-dark-100'"
                :disabled="generating"
                role="tab"
                :aria-selected="mode === 'generate'"
                @click="mode = 'generate'"
              >
                <Icon name="sparkles" size="xs" aria-hidden="true" />
                {{ t('imageGeneration.modes.generate') }}
              </button>
              <button
                type="button"
                class="inline-flex min-h-8 items-center gap-1.5 rounded-md px-2.5 text-xs font-medium transition-colors"
                :class="mode === 'edit' ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-700 dark:text-primary-300' : 'text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-dark-100'"
                :disabled="generating"
                role="tab"
                :aria-selected="mode === 'edit'"
                @click="mode = 'edit'"
              >
                <Icon name="edit" size="xs" aria-hidden="true" />
                {{ t('imageGeneration.modes.edit') }}
              </button>
            </div>
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

            <div v-if="mode === 'edit'" class="rounded-xl border border-gray-100 bg-gray-50/80 p-3 dark:border-dark-700 dark:bg-dark-900/50">
              <div class="mb-2 flex items-center justify-between gap-3">
                <label class="input-label mb-0">{{ t('imageGeneration.edit.source') }}</label>
                <span class="text-xs text-gray-400 dark:text-dark-500">{{ editSources.length }}/{{ MAX_EDIT_FILES }}</span>
              </div>
              <input
                ref="editFileInput"
                type="file"
                accept="image/png,image/jpeg,image/webp"
                multiple
                class="hidden"
                :disabled="generating || editSources.length >= MAX_EDIT_FILES"
                @change="handleEditFiles"
              />
              <button type="button" class="btn btn-secondary btn-sm" :disabled="generating || editSources.length >= MAX_EDIT_FILES" @click="openEditFilePicker">
                <Icon name="upload" size="sm" aria-hidden="true" />
                {{ t('imageGeneration.edit.chooseFiles') }}
              </button>
              <div v-if="editSources.length" class="mt-3 grid grid-cols-4 gap-2">
                <figure v-for="(source, index) in editSources" :key="`${source.file.name}-${source.file.size}-${index}`" class="group relative aspect-square overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
                  <img :src="source.src" :alt="t('imageGeneration.edit.sourceAlt', { index: index + 1 })" class="h-full w-full object-contain" />
                  <button
                    type="button"
                    class="absolute right-1 top-1 inline-flex h-6 w-6 items-center justify-center rounded-md bg-black/60 text-white opacity-0 transition-opacity hover:bg-black/80 focus:opacity-100 focus:outline-none focus:ring-2 focus:ring-white group-hover:opacity-100"
                    :title="t('imageGeneration.edit.removeSource')"
                    @click="removeEditSource(index)"
                  >
                    <Icon name="x" size="xs" aria-hidden="true" />
                    <span class="sr-only">{{ t('imageGeneration.edit.removeSource') }}</span>
                  </button>
                </figure>
              </div>
              <p v-if="editSourceError" class="mt-2 text-xs text-red-500" role="alert">{{ editSourceError }}</p>
            </div>

            <div>
              <div class="mb-1 flex items-center justify-between gap-3">
                <label class="input-label mb-0" for="image-generation-prompt">
                  {{ t('imageGeneration.controls.prompt') }}
                  <span class="text-red-500">*</span>
                </label>
                <button
                  type="button"
                  class="btn btn-ghost btn-sm px-2"
                  :disabled="optimizing || generating || !prompt.trim()"
                  :title="t('imageGeneration.actions.optimizePrompt')"
                  @click="optimizePrompt"
                >
                  <Icon v-if="!optimizing" name="sparkles" size="sm" aria-hidden="true" />
                  <Icon v-else name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
                  {{ optimizing ? t('imageGeneration.actions.optimizing') : t('imageGeneration.actions.optimizePrompt') }}
                </button>
              </div>
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
              <button
                v-if="previousPrompt"
                type="button"
                class="mt-2 text-xs text-primary-600 hover:underline dark:text-primary-400"
                :disabled="optimizing || generating"
                @click="restorePreviousPrompt"
              >
                {{ t('imageGeneration.actions.restorePrompt') }}
              </button>
              <div class="mt-1 flex items-start justify-between gap-2 text-xs">
                <span v-if="promptError || optimizationError" class="text-red-500">{{ promptError || optimizationError }}</span>
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

            <button type="button" class="btn btn-primary w-full" :disabled="generating || !currentModel || !canSubmit" @click="submit">
              <Icon v-if="!generating" :name="mode === 'edit' ? 'edit' : 'sparkles'" size="sm" aria-hidden="true" />
              <Icon v-else name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
              {{ generating ? t(mode === 'edit' ? 'imageGeneration.actions.editing' : 'imageGeneration.actions.generating') : t(mode === 'edit' ? 'imageGeneration.actions.edit' : 'imageGeneration.actions.generate') }}
            </button>
          </div>
        </section>

        <!-- Generated images -->
        <section class="card flex h-full min-h-[520px] flex-col p-5 sm:p-6">
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

          <div v-if="generating" class="flex min-h-[400px] flex-1 flex-col items-center justify-center text-center">
            <LoadingSpinner size="lg" />
            <p class="mt-4 text-sm text-gray-600 dark:text-gray-300">{{ t('imageGeneration.results.generating') }}</p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ t('imageGeneration.results.generatingHint') }}</p>
          </div>

          <div v-else-if="generationError" class="flex min-h-[400px] flex-1 flex-col items-center justify-center text-center">
            <Icon name="exclamationTriangle" size="xl" class="mb-4 text-amber-500" aria-hidden="true" />
            <p class="max-w-md text-sm text-red-600 dark:text-red-300" role="alert">{{ generationError }}</p>
            <button type="button" class="btn btn-secondary mt-5" @click="submit">
              <Icon name="refresh" size="sm" aria-hidden="true" />
              {{ t('imageGeneration.actions.retry') }}
            </button>
          </div>

          <div v-else-if="images.length === 0" class="flex min-h-[400px] flex-1 flex-col items-center justify-center text-center">
            <Icon name="sparkles" size="xl" class="mb-4 text-gray-300 dark:text-dark-600" aria-hidden="true" />
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('imageGeneration.results.empty') }}</p>
          </div>

          <div v-else :class="['grid gap-4', imageGridClass]">
            <figure v-for="(image, index) in images" :key="`${image.src}-${index}`" class="group overflow-hidden rounded-xl border border-gray-100 bg-gray-50 dark:border-dark-700 dark:bg-dark-900/50">
              <div
                class="relative flex aspect-square cursor-zoom-in items-center justify-center overflow-hidden bg-gray-100 dark:bg-dark-900"
                role="button"
                tabindex="0"
                @click="openPreview(index)"
                @keydown.enter.prevent="openPreview(index)"
                @keydown.space.prevent="openPreview(index)"
              >
                <img
                  :src="image.src"
                  :alt="t('imageGeneration.results.imageAlt', { index: index + 1 })"
                  class="h-full w-full object-contain"
                  loading="lazy"
                />
                <button
                  type="button"
                  class="absolute right-3 top-3 z-10 inline-flex h-9 w-9 items-center justify-center rounded-lg bg-black/60 text-white shadow hover:bg-black/75 focus:outline-none focus:ring-2 focus:ring-white/80"
                  :title="t('imageGeneration.actions.download')"
                  :disabled="downloadLoadingIndex === index"
                  @pointerdown.stop
                  @click.stop="downloadImage(image)"
                >
                  <Icon :name="downloadLoadingIndex === index ? 'refresh' : 'download'" size="sm" :class="downloadLoadingIndex === index ? 'animate-spin' : undefined" aria-hidden="true" />
                  <span class="sr-only">{{ t('imageGeneration.actions.download') }}</span>
                </button>
                <button
                  type="button"
                  class="absolute right-14 top-3 z-10 inline-flex h-9 w-9 items-center justify-center rounded-lg bg-black/60 text-white shadow hover:bg-black/75 focus:outline-none focus:ring-2 focus:ring-white/80"
                  :title="t('imageGeneration.actions.useForEdit')"
                  :disabled="sourceLoadingIndex === index"
                  @pointerdown.stop
                  @click.stop="useImageForEdit(image, index)"
                >
                  <Icon v-if="sourceLoadingIndex !== index" name="edit" size="sm" aria-hidden="true" />
                  <Icon v-else name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
                  <span class="sr-only">{{ t('imageGeneration.actions.useForEdit') }}</span>
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
    <BaseDialog
      :show="previewIndex !== null"
      :title="t('imageGeneration.results.previewTitle')"
      width="full"
      :close-on-click-outside="true"
      :show-close-button="true"
      @close="closePreview"
    >
      <div v-if="previewImage" class="flex flex-col items-center gap-4">
        <div class="relative flex min-h-[50vh] w-full items-center justify-center rounded-lg bg-gray-950 p-2">
          <img
            :src="previewImage.src"
            :alt="t('imageGeneration.results.imageAlt', { index: (previewIndex ?? 0) + 1 })"
            class="max-h-[70vh] max-w-full object-contain"
          />
          <button
            v-if="images.length > 1"
            type="button"
            class="absolute left-3 top-1/2 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-black/60 text-white hover:bg-black/80 focus:outline-none focus:ring-2 focus:ring-white"
            :title="t('imageGeneration.results.previous')"
            @click="showPreviousPreview"
          >
            <Icon name="chevronLeft" size="md" aria-hidden="true" />
            <span class="sr-only">{{ t('imageGeneration.results.previous') }}</span>
          </button>
          <button
            v-if="images.length > 1"
            type="button"
            class="absolute right-3 top-1/2 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-black/60 text-white hover:bg-black/80 focus:outline-none focus:ring-2 focus:ring-white"
            :title="t('imageGeneration.results.next')"
            @click="showNextPreview"
          >
            <Icon name="chevronRight" size="md" aria-hidden="true" />
            <span class="sr-only">{{ t('imageGeneration.results.next') }}</span>
          </button>
        </div>
        <div class="flex items-center gap-3">
          <span class="text-xs text-gray-500 dark:text-dark-400">{{ (previewIndex ?? 0) + 1 }} / {{ images.length }}</span>
          <button type="button" class="btn btn-primary btn-sm" :disabled="downloadLoadingIndex === previewIndex" @click="downloadImage(previewImage)">
            <Icon :name="downloadLoadingIndex === previewIndex ? 'refresh' : 'download'" size="sm" :class="downloadLoadingIndex === previewIndex ? 'animate-spin' : undefined" aria-hidden="true" />
            {{ t('imageGeneration.actions.download') }}
          </button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="sourceLoadingIndex === previewIndex" @click="previewImage && useImageForEdit(previewImage, previewIndex ?? 0)">
            <Icon name="edit" size="sm" aria-hidden="true" />
            {{ t('imageGeneration.actions.useForEdit') }}
          </button>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import { sanitizeUrl } from '@/utils/url'
import { editImages, generateImages, loadImageGenerationConfig, loadImageGenerationOptions, optimizeImagePrompt } from '../api/imageGeneration'
import {
  imageSource,
  validateCustomImageSize,
  type DisplayImage,
  type ImageGenerationBackground,
  type ImageGenerationFormat,
  type ImageGenerationGroupOption,
  type ImageGenerationRequest,
  type ImageGenerationModelOption,
  type ImageGenerationModeration,
  type ImageGenerationQuality
} from '../types/imageGeneration'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const MAX_PROMPT_LENGTH = 10_000
const CUSTOM_SIZE_VALUE = '__custom__'
const MAX_EDIT_FILES = 4
const MAX_EDIT_FILE_BYTES = 20 * 1024 * 1024
const EDIT_IMAGE_TYPES = new Set(['image/png', 'image/jpeg', 'image/webp'])

const groups = ref<ImageGenerationGroupOption[]>([])
const loadingOptions = ref(false)
const pageError = ref('')
const generationError = ref('')
const generating = ref(false)
const optimizing = ref(false)
const promptError = ref('')
const optimizationError = ref('')
const mode = ref<'generate' | 'edit'>('generate')
const editSources = ref<EditSourceImage[]>([])
const editSourceError = ref('')
const sourceLoadingIndex = ref<number | null>(null)
const downloadLoadingIndex = ref<number | null>(null)
const editFileInput = ref<HTMLInputElement | null>(null)

const selectedGroupId = ref<number | null>(null)
const selectedModel = ref('')
const prompt = ref('')
const previousPrompt = ref('')
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
const previewIndex = ref<number | null>(null)

interface EditSourceImage {
  file: File
  src: string
  objectUrl: boolean
}

let optionsController: AbortController | null = null
let generationController: AbortController | null = null
let optimizationController: AbortController | null = null

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
const maxGenerationCount = computed(() => Math.min(9, Math.max(1, currentModel.value?.max_n ?? 9)))
const canSubmit = computed(() => mode.value === 'generate' || editSources.value.length > 0)
const isCompressionVisible = computed(() => currentModel.value?.supports_compression !== false && (outputFormat.value === 'jpeg' || outputFormat.value === 'webp'))
const customSizeError = computed(() => {
  const constraints = currentModel.value?.custom_size
  if (size.value !== CUSTOM_SIZE_VALUE || !constraints) return null
  return validateCustomImageSize(customWidth.value, customHeight.value, constraints)
})
const previewImage = computed(() => previewIndex.value === null ? null : images.value[previewIndex.value] ?? null)
const imageGridClass = computed(() => {
  if (images.value.length <= 1) return 'grid-cols-1'
  if (images.value.length === 2) return 'grid-cols-2'
  if (images.value.length >= 5) return 'grid-cols-3'
  return 'grid-cols-2'
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
  if (!group?.models.some((model) => model.name === selectedModel.value)) {
    selectedModel.value = group?.models[0]?.name ?? ''
  }
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
    const [optionsResult, configResult] = await Promise.allSettled([
      loadImageGenerationOptions({ signal: controller.signal }),
      loadImageGenerationConfig({ signal: controller.signal })
    ])
    if (optionsResult.status === 'rejected') throw optionsResult.reason
    const response = optionsResult.value
    const savedConfig = configResult.status === 'fulfilled' ? configResult.value.config : null
    const useServerDefaults = groups.value.length === 0
    groups.value = response.groups
    if (useServerDefaults) {
      size.value = response.defaults.size
      quality.value = response.defaults.quality
      outputFormat.value = response.defaults.output_format
      background.value = response.defaults.background
      moderation.value = response.defaults.moderation
      count.value = response.defaults.n
      if (savedConfig) {
        size.value = savedConfig.default_size || size.value
        count.value = savedConfig.default_n || count.value
        if (savedConfig.image_group_id > 0 && groups.value.some((group) => group.id === savedConfig.image_group_id)) {
          selectedGroupId.value = savedConfig.image_group_id
          selectedModel.value = savedConfig.image_model || ''
        }
      }
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

function base64ImageBlob(value: string, mimeType: string): Blob {
  const binary = window.atob(value.trim())
  const bytes = new Uint8Array(binary.length)
  for (let index = 0; index < binary.length; index += 1) bytes[index] = binary.charCodeAt(index)
  return new Blob([bytes], { type: mimeType })
}

function disposeDisplayImages(items: DisplayImage[]): void {
  items.forEach((image) => {
    if (image.objectUrl) URL.revokeObjectURL(image.src)
  })
}

function toDisplayImages(response: { data?: Array<{ b64_json?: string | null; url?: string | null; revised_prompt?: string | null; mime_type?: string | null }> }): DisplayImage[] {
  return (response.data ?? []).flatMap((image, index) => {
    const extension = outputFormat.value === 'jpeg' ? 'jpg' : outputFormat.value
    const downloadName = `sub2api-image-${Date.now()}-${index + 1}.${extension}`
    const fallbackMime = outputFormat.value === 'jpeg' ? 'image/jpeg' : `image/${outputFormat.value}`
    const mimeType = normalizedImageType(image.mime_type || fallbackMime)
    if (typeof image.b64_json === 'string' && image.b64_json.trim()) {
      try {
        const blob = base64ImageBlob(image.b64_json, mimeType)
        const src = URL.createObjectURL(blob)
        return [{ revised_prompt: image.revised_prompt, mime_type: mimeType, src, downloadName, blob, objectUrl: true }]
      } catch {
        return []
      }
    }
    const src = sanitizeUrl(imageSource(image, outputFormat.value), { allowDataUrl: true, allowRelative: true })
    if (!src) return []
    return [{ ...image, src, downloadName }]
  })
}

function buildRequestPayload(trimmedPrompt: string): ImageGenerationRequest {
  if (!selectedGroupId.value || !currentModel.value) {
    throw new Error(t('imageGeneration.errors.noModel'))
  }
  return {
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
}

async function generate(): Promise<void> {
  normalizeCount()
  generationError.value = ''
  optimizationError.value = ''
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
    const payload = buildRequestPayload(trimmedPrompt)
    const response = await generateImages(payload, { signal: controller.signal })
    const displayImages = toDisplayImages(response)
    if (displayImages.length === 0) {
      generationError.value = t('imageGeneration.errors.emptyResponse')
    } else {
      disposeDisplayImages(images.value)
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

async function edit(): Promise<void> {
  normalizeCount()
  generationError.value = ''
  optimizationError.value = ''
  editSourceError.value = ''
  const trimmedPrompt = prompt.value.trim()
  if (!trimmedPrompt) {
    promptError.value = t('imageGeneration.validation.promptRequired')
    return
  }
  if (trimmedPrompt.length > MAX_PROMPT_LENGTH) {
    promptError.value = t('imageGeneration.validation.promptTooLong')
    return
  }
  if (editSources.value.length === 0) {
    editSourceError.value = t('imageGeneration.edit.sourceRequired')
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
    const payload = buildRequestPayload(trimmedPrompt)
    const response = await editImages(payload, editSources.value.map((source) => source.file), { signal: controller.signal })
    const displayImages = toDisplayImages(response)
    if (displayImages.length === 0) {
      generationError.value = t('imageGeneration.errors.emptyResponse')
    } else {
      disposeDisplayImages(images.value)
      images.value = displayImages
    }
  } catch (cause) {
    if (!controller.signal.aborted) {
      generationError.value = errorMessage(cause, t('imageGeneration.errors.edit'))
    }
  } finally {
    if (generationController === controller) {
      generationController = null
      generating.value = false
    }
  }
}

async function submit(): Promise<void> {
  if (mode.value === 'edit') {
    await edit()
    return
  }
  await generate()
}

function normalizedImageType(value: string): string {
  const type = value.toLowerCase().split(';', 1)[0].trim()
  return type === 'image/jpg' ? 'image/jpeg' : type
}

function addEditSources(files: File[], previews?: string[]): void {
  editSourceError.value = ''
  const remaining = MAX_EDIT_FILES - editSources.value.length
  if (remaining <= 0) {
    editSourceError.value = t('imageGeneration.edit.tooManySources', { max: MAX_EDIT_FILES })
    return
  }
  const accepted = files.slice(0, remaining).flatMap((file, index): EditSourceImage[] => {
    const type = normalizedImageType(file.type)
    if (!EDIT_IMAGE_TYPES.has(type)) {
      editSourceError.value = t('imageGeneration.edit.invalidSourceType')
      return []
    }
    if (file.size <= 0 || file.size > MAX_EDIT_FILE_BYTES) {
      editSourceError.value = t('imageGeneration.edit.sourceTooLarge')
      return []
    }
    return [{
      file,
      src: previews?.[index] || URL.createObjectURL(file),
      objectUrl: !previews?.[index]
    }]
  })
  editSources.value = [...editSources.value, ...accepted]
  if (files.length > remaining) {
    editSourceError.value = t('imageGeneration.edit.tooManySources', { max: MAX_EDIT_FILES })
  }
}

function openEditFilePicker(): void {
  editFileInput.value?.click()
}

function handleEditFiles(event: Event): void {
  const input = event.target as HTMLInputElement
  addEditSources(Array.from(input.files ?? []))
  input.value = ''
}

function removeEditSource(index: number): void {
  const source = editSources.value[index]
  if (!source) return
  if (source.objectUrl) URL.revokeObjectURL(source.src)
  editSources.value.splice(index, 1)
  editSourceError.value = ''
}

async function useImageForEdit(image: DisplayImage, index: number): Promise<void> {
  if (sourceLoadingIndex.value !== null || editSources.value.length >= MAX_EDIT_FILES) {
    if (editSources.value.length >= MAX_EDIT_FILES) {
      editSourceError.value = t('imageGeneration.edit.tooManySources', { max: MAX_EDIT_FILES })
    }
    return
  }
  sourceLoadingIndex.value = index
  editSourceError.value = ''
  try {
    const blob = image.blob ?? await fetchImageBlob(image)
    const type = normalizedImageType(blob.type || image.mime_type || `image/${outputFormat.value}`)
    if (!EDIT_IMAGE_TYPES.has(type)) throw new Error('unsupported source image type')
    if (blob.size <= 0 || blob.size > MAX_EDIT_FILE_BYTES) throw new Error('source image is too large')
    const file = new File([blob], image.downloadName, { type })
    addEditSources([file])
    mode.value = 'edit'
    closePreview()
  } catch {
    editSourceError.value = t('imageGeneration.edit.sourceLoadFailed')
  } finally {
    sourceLoadingIndex.value = null
  }
}

async function optimizePrompt(): Promise<void> {
  const trimmedPrompt = prompt.value.trim()
  if (!trimmedPrompt || optimizing.value || generating.value) return
  optimizationController?.abort()
  const controller = new AbortController()
  optimizationController = controller
  optimizing.value = true
  optimizationError.value = ''
  try {
    const result = await optimizeImagePrompt(trimmedPrompt, { signal: controller.signal })
    const optimized = result.optimized_prompt.trim()
    if (!optimized) {
      optimizationError.value = t('imageGeneration.errors.optimizeEmpty')
      return
    }
    previousPrompt.value = prompt.value
    prompt.value = optimized
    promptError.value = ''
  } catch (cause) {
    if (!controller.signal.aborted) {
      optimizationError.value = errorMessage(cause, t('imageGeneration.errors.optimize'))
    }
  } finally {
    if (optimizationController === controller) {
      optimizationController = null
      optimizing.value = false
    }
  }
}

function restorePreviousPrompt(): void {
  if (!previousPrompt.value) return
  prompt.value = previousPrompt.value
  previousPrompt.value = ''
  optimizationError.value = ''
}

function clearResults(): void {
  disposeDisplayImages(images.value)
  images.value = []
  generationError.value = ''
  closePreview()
}

function openPreview(index: number): void {
  if (index >= 0 && index < images.value.length) previewIndex.value = index
}

function closePreview(): void {
  previewIndex.value = null
}

function showPreviousPreview(): void {
  if (previewIndex.value === null || images.value.length < 2) return
  previewIndex.value = (previewIndex.value - 1 + images.value.length) % images.value.length
}

function showNextPreview(): void {
  if (previewIndex.value === null || images.value.length < 2) return
  previewIndex.value = (previewIndex.value + 1) % images.value.length
}

async function downloadImage(image: DisplayImage): Promise<void> {
  const index = images.value.indexOf(image)
  if (downloadLoadingIndex.value !== null) return
  downloadLoadingIndex.value = index >= 0 ? index : previewIndex.value
  try {
    const blob = image.blob ?? await fetchImageBlob(image)
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
  } finally {
    downloadLoadingIndex.value = null
  }
}

async function fetchImageBlob(image: DisplayImage): Promise<Blob> {
  const response = await fetch(image.src, { credentials: 'omit' })
  if (!response.ok) throw new Error('image unavailable')
  const blob = await response.blob()
  if (blob.size <= 0) throw new Error('image is empty')
  return blob
}

onMounted(() => {
  void loadOptions()
})

onUnmounted(() => {
  optionsController?.abort()
  generationController?.abort()
  optimizationController?.abort()
  editSources.value.forEach((source) => {
    if (source.objectUrl) URL.revokeObjectURL(source.src)
  })
  disposeDisplayImages(images.value)
})
</script>
