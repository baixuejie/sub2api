<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-6">
      <header class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <router-link to="/image-generation" class="mb-3 inline-flex items-center gap-1 text-sm text-gray-500 hover:text-primary-600 dark:text-dark-400 dark:hover:text-primary-400">
            <Icon name="arrowLeft" size="sm" aria-hidden="true" />
            {{ t('imageGeneration.settings.back') }}
          </router-link>
          <div class="flex items-center gap-2">
            <Icon name="cog" size="lg" class="text-primary-500" aria-hidden="true" />
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('imageGeneration.settings.title') }}</h1>
          </div>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('imageGeneration.settings.description') }}</p>
        </div>
      </header>

      <div v-if="pageError" class="flex items-start gap-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-300" role="alert">
        <Icon name="exclamationCircle" size="md" class="mt-0.5 flex-shrink-0" aria-hidden="true" />
        <span class="min-w-0 flex-1">{{ pageError }}</span>
        <button type="button" class="btn btn-ghost btn-sm -my-1 px-2" @click="loadConfig">
          {{ t('imageGeneration.actions.retry') }}
        </button>
      </div>

      <div v-if="loading" class="card flex min-h-[360px] items-center justify-center p-8">
        <LoadingSpinner size="lg" />
      </div>

      <form v-else class="space-y-6" @submit.prevent="save">
        <section class="card p-5 sm:p-6">
          <div class="mb-5 flex items-start gap-3 border-b border-gray-100 pb-4 dark:border-dark-700">
            <Icon name="sparkles" size="md" class="mt-0.5 text-primary-500" aria-hidden="true" />
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">{{ t('imageGeneration.settings.promptTitle') }}</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('imageGeneration.settings.promptHint') }}</p>
            </div>
          </div>
          <div class="grid gap-4 md:grid-cols-3">
            <div>
              <label class="input-label" for="image-settings-prompt-group">{{ t('imageGeneration.settings.group') }}</label>
              <Select id="image-settings-prompt-group" v-model="config.prompt_group_id" :options="promptGroupOptions" searchable />
            </div>
            <div>
              <label class="input-label" for="image-settings-prompt-model">{{ t('imageGeneration.settings.model') }}</label>
              <Select id="image-settings-prompt-model" v-model="config.prompt_model" :options="promptModelOptions" searchable :disabled="!config.prompt_group_id" />
            </div>
            <div>
              <label class="input-label" for="image-settings-prompt-key">{{ t('imageGeneration.settings.apiKey') }}</label>
              <Select id="image-settings-prompt-key" v-model="config.prompt_api_key_id" :options="promptKeyOptions" searchable :disabled="!config.prompt_group_id" />
            </div>
          </div>
          <p v-if="promptKeyOptions.length === 0" class="input-hint mt-3 text-amber-600 dark:text-amber-400">{{ t('imageGeneration.settings.noPromptKey') }}</p>
        </section>

        <section class="card p-5 sm:p-6">
          <div class="mb-5 flex items-start gap-3 border-b border-gray-100 pb-4 dark:border-dark-700">
            <Icon name="grid" size="md" class="mt-0.5 text-primary-500" aria-hidden="true" />
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">{{ t('imageGeneration.settings.imageTitle') }}</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('imageGeneration.settings.imageHint') }}</p>
            </div>
          </div>
          <div class="grid gap-4 md:grid-cols-3">
            <div>
              <label class="input-label" for="image-settings-image-group">{{ t('imageGeneration.settings.group') }}</label>
              <Select id="image-settings-image-group" v-model="config.image_group_id" :options="imageGroupOptions" searchable />
            </div>
            <div>
              <label class="input-label" for="image-settings-image-model">{{ t('imageGeneration.settings.model') }}</label>
              <Select id="image-settings-image-model" v-model="config.image_model" :options="imageModelOptions" searchable :disabled="!config.image_group_id" />
            </div>
            <div>
              <label class="input-label" for="image-settings-image-key">{{ t('imageGeneration.settings.apiKey') }}</label>
              <Select id="image-settings-image-key" v-model="config.image_api_key_id" :options="imageKeyOptions" searchable :disabled="!config.image_group_id" />
            </div>
          </div>
          <p v-if="imageKeyOptions.length === 0" class="input-hint mt-3 text-amber-600 dark:text-amber-400">{{ t('imageGeneration.settings.noImageKey') }}</p>

          <div class="mt-6 grid gap-4 sm:grid-cols-2">
            <div>
              <label class="input-label" for="image-settings-size">{{ t('imageGeneration.settings.defaultSize') }}</label>
              <Select id="image-settings-size" v-model="config.default_size" :options="sizeOptions" :disabled="!config.image_model" />
            </div>
            <div>
              <label class="input-label" for="image-settings-count">{{ t('imageGeneration.settings.defaultCount') }}</label>
              <input
                id="image-settings-count"
                v-model.number="config.default_n"
                class="input"
                type="number"
                min="1"
                max="9"
                step="1"
              />
              <p class="input-hint">{{ t('imageGeneration.settings.defaultCountHint') }}</p>
            </div>
          </div>
        </section>

        <div v-if="saveError" class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-300" role="alert">
          {{ saveError }}
        </div>
        <div v-if="saved" class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-950/30 dark:text-emerald-300" role="status">
          {{ t('imageGeneration.settings.saved') }}
        </div>
        <div class="flex justify-end">
          <button type="submit" class="btn btn-primary" :disabled="saving">
            <Icon v-if="!saving" name="check" size="sm" aria-hidden="true" />
            <Icon v-else name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
            {{ saving ? t('imageGeneration.settings.saving') : t('imageGeneration.settings.save') }}
          </button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import { useI18n } from 'vue-i18n'
import { loadImageGenerationConfig, saveImageGenerationConfig } from '../api/imageGeneration'
import {
  DEFAULT_IMAGE_GENERATION_CONFIG,
  type ImageGenerationConfig,
  type ImageGenerationConfigGroupOption,
  type ImageGenerationConfigOptionsResponse
} from '../types/imageGeneration'

const { t } = useI18n()
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const pageError = ref('')
const saveError = ref('')
const groups = ref<ImageGenerationConfigOptionsResponse>({
  config: { ...DEFAULT_IMAGE_GENERATION_CONFIG },
  prompt_groups: [],
  image_groups: [],
  api_keys: []
})
const config = reactive<ImageGenerationConfig>({ ...DEFAULT_IMAGE_GENERATION_CONFIG })
let controller: AbortController | null = null

const promptGroupOptions = computed<SelectOption[]>(() => groupOptions(groups.value.prompt_groups))
const imageGroupOptions = computed<SelectOption[]>(() => groupOptions(groups.value.image_groups))
const promptGroup = computed(() => findGroup(groups.value.prompt_groups, config.prompt_group_id))
const imageGroup = computed(() => findGroup(groups.value.image_groups, config.image_group_id))
const promptModelOptions = computed<SelectOption[]>(() => modelOptions(promptGroup.value))
const imageModelOptions = computed<SelectOption[]>(() => modelOptions(imageGroup.value))
const promptKeyOptions = computed<SelectOption[]>(() => keyOptions(config.prompt_group_id, false))
const imageKeyOptions = computed<SelectOption[]>(() => keyOptions(config.image_group_id, true))
const sizeOptions = computed<SelectOption[]>(() => {
  const values = config.image_model.toLowerCase() === 'gpt-image-2'
    ? ['auto', '1024x1024', '1536x1024', '1024x1536', '2048x2048', '3072x2048', '2048x3072']
    : ['auto', '1024x1024', '1536x1024', '1024x1536']
  return values.map((value) => ({ value, label: value }))
})

function groupOptions(list: ImageGenerationConfigGroupOption[]): SelectOption[] {
  return list.map((group) => ({ value: group.id, label: group.name, description: group.description || undefined }))
}

function findGroup(list: ImageGenerationConfigGroupOption[], id: number): ImageGenerationConfigGroupOption | null {
  return list.find((group) => group.id === id) ?? null
}

function modelOptions(group: ImageGenerationConfigGroupOption | null): SelectOption[] {
  return (group?.models ?? []).map((model) => ({ value: model.name, label: model.name }))
}

function keyOptions(groupId: number, imageOnly: boolean): SelectOption[] {
  return groups.value.api_keys
    .filter((key) => key.group_id === groupId && (!imageOnly || key.image_enabled))
    .map((key) => ({ value: key.id, label: `${key.name} (${key.masked_key})` }))
}

function applyConfig(value: ImageGenerationConfig): void {
  Object.assign(config, value)
}

watch(() => config.prompt_group_id, () => {
  if (!promptGroup.value?.models.some((model) => model.name === config.prompt_model)) {
    config.prompt_model = promptGroup.value?.models[0]?.name ?? ''
  }
  if (!promptKeyOptions.value.some((option) => option.value === config.prompt_api_key_id)) {
    config.prompt_api_key_id = Number(promptKeyOptions.value[0]?.value ?? 0)
  }
})

watch(() => config.image_group_id, () => {
  if (!imageGroup.value?.models.some((model) => model.name === config.image_model)) {
    config.image_model = imageGroup.value?.models.find((model) => model.name.toLowerCase() === 'gpt-image-2')?.name ?? imageGroup.value?.models[0]?.name ?? ''
  }
  if (!imageKeyOptions.value.some((option) => option.value === config.image_api_key_id)) {
    config.image_api_key_id = Number(imageKeyOptions.value[0]?.value ?? 0)
  }
})

watch(() => config.image_model, () => {
  if (!sizeOptions.value.some((option) => option.value === config.default_size)) {
    config.default_size = String(sizeOptions.value[0]?.value ?? '1024x1024')
  }
})

function errorMessage(cause: unknown, fallback: string): string {
  if (cause && typeof cause === 'object' && 'message' in cause) {
    const message = (cause as { message?: unknown }).message
    if (typeof message === 'string' && message.trim()) return message
  }
  return fallback
}

async function loadConfig(): Promise<void> {
  controller?.abort()
  const current = new AbortController()
  controller = current
  loading.value = true
  pageError.value = ''
  try {
    const result = await loadImageGenerationConfig({ signal: current.signal })
    groups.value = result
    applyConfig(result.config)
  } catch (cause) {
    if (!current.signal.aborted) pageError.value = errorMessage(cause, t('imageGeneration.settings.loadError'))
  } finally {
    if (controller === current) {
      controller = null
      loading.value = false
    }
  }
}

async function save(): Promise<void> {
  saving.value = true
  saved.value = false
  saveError.value = ''
  const payload = { ...config, default_n: Math.min(9, Math.max(1, Math.round(Number(config.default_n) || 1))) }
  try {
    const result = await saveImageGenerationConfig(payload)
    groups.value = result
    applyConfig(result.config)
    saved.value = true
  } catch (cause) {
    saveError.value = errorMessage(cause, t('imageGeneration.settings.saveError'))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadConfig()
})

onUnmounted(() => {
  controller?.abort()
})
</script>
