/**
 * Contracts for the authenticated image-generation extension.
 *
 * The browser only receives the model capabilities that the current user is
 * allowed to use. Upstream credentials, account selection and billing remain
 * entirely server-side.
 */

export type ImageGenerationQuality = 'auto' | 'low' | 'medium' | 'high'
export type ImageGenerationFormat = 'png' | 'jpeg' | 'webp'
export type ImageGenerationBackground = 'auto' | 'opaque' | 'transparent'
export type ImageGenerationModeration = 'auto' | 'low'

export const DEFAULT_IMAGE_SIZES = ['1024x1024', '1536x1024', '1024x1536'] as const
export const DEFAULT_IMAGE_QUALITIES: ImageGenerationQuality[] = ['auto', 'low', 'medium', 'high']
export const DEFAULT_IMAGE_FORMATS: ImageGenerationFormat[] = ['png', 'jpeg', 'webp']
export const DEFAULT_IMAGE_BACKGROUNDS: ImageGenerationBackground[] = ['auto', 'opaque', 'transparent']
export const DEFAULT_IMAGE_MODERATIONS: ImageGenerationModeration[] = ['auto', 'low']

export interface ImageGenerationModelOption {
  /** Public model identifier. The backend validates it against the group. */
  name: string
  /** Kept for forward compatibility with options providers that expose id. */
  id?: string
  sizes: string[]
  qualities: ImageGenerationQuality[]
  output_formats: ImageGenerationFormat[]
  backgrounds: ImageGenerationBackground[]
  moderations: ImageGenerationModeration[]
  max_n: number
  supports_compression: boolean
  custom_size?: ImageGenerationCustomSizeConstraints | null
}

export interface ImageGenerationCustomSizeConstraints {
  min_pixels: number
  max_pixels: number
  max_edge: number
  edge_multiple: number
  max_aspect_ratio: number
}

export interface ImageGenerationGroupOption {
  id: number
  name: string
  description?: string | null
  platform?: string
  models: ImageGenerationModelOption[]
}

export interface ImageGenerationOptionsResponse {
  groups: ImageGenerationGroupOption[]
  defaults: ImageGenerationDefaults
}

export interface ImageGenerationDefaults {
  size: string
  quality: ImageGenerationQuality
  output_format: ImageGenerationFormat
  background: ImageGenerationBackground
  moderation: ImageGenerationModeration
  n: number
}

export interface ImageGenerationRequest {
  group_id: number
  model: string
  prompt: string
  n: number
  size: string
  quality: ImageGenerationQuality
  output_format: ImageGenerationFormat
  output_compression?: number
  background: ImageGenerationBackground
  moderation: ImageGenerationModeration
}

export interface GeneratedImage {
  b64_json?: string | null
  url?: string | null
  revised_prompt?: string | null
  mime_type?: string | null
}

export interface ImageGenerationResponse {
  created?: number
  data?: GeneratedImage[]
  /** Some gateway versions expose the array as `images`. */
  images?: GeneratedImage[]
}

/** Multipart payload fields shared by generation and image editing. */
export type ImageEditRequest = ImageGenerationRequest

export interface ImageGenerationConfig {
  version: number
  prompt_group_id: number
  prompt_model: string
  prompt_api_key_id: number
  image_group_id: number
  image_model: string
  image_api_key_id: number
  default_size: string
  default_n: number
}

export interface ImageGenerationConfigModelOption {
  name: string
}

export interface ImageGenerationConfigGroupOption {
  id: number
  name: string
  description?: string | null
  platform?: string
  models: ImageGenerationConfigModelOption[]
}

export interface ImageGenerationConfigApiKeyOption {
  id: number
  name: string
  masked_key: string
  group_id: number
  group_name: string
  image_enabled: boolean
  status: string
}

export interface ImageGenerationConfigOptionsResponse {
  config: ImageGenerationConfig
  prompt_groups: ImageGenerationConfigGroupOption[]
  image_groups: ImageGenerationConfigGroupOption[]
  api_keys: ImageGenerationConfigApiKeyOption[]
}

export interface PromptOptimizationResponse {
  original_prompt: string
  optimized_prompt: string
}

export interface DisplayImage extends GeneratedImage {
  src: string
  downloadName: string
  blob?: Blob
  objectUrl?: boolean
}

export const DEFAULT_IMAGE_GENERATION_CONFIG: ImageGenerationConfig = {
  version: 1,
  prompt_group_id: 0,
  prompt_model: '',
  prompt_api_key_id: 0,
  image_group_id: 0,
  image_model: 'gpt-image-2',
  image_api_key_id: 0,
  default_size: '1024x1024',
  default_n: 1
}

const QUALITY_VALUES = new Set<ImageGenerationQuality>(DEFAULT_IMAGE_QUALITIES)
const FORMAT_VALUES = new Set<ImageGenerationFormat>(DEFAULT_IMAGE_FORMATS)
const BACKGROUND_VALUES = new Set<ImageGenerationBackground>(DEFAULT_IMAGE_BACKGROUNDS)
const MODERATION_VALUES = new Set<ImageGenerationModeration>(DEFAULT_IMAGE_MODERATIONS)

function stringList(value: unknown, fallback: readonly string[]): string[] {
  if (!Array.isArray(value)) return [...fallback]
  const result = value.filter((item): item is string => typeof item === 'string' && item.trim().length > 0)
  return result.length ? [...new Set(result)] : [...fallback]
}

function enumList<T extends string>(value: unknown, allowed: Set<T>, fallback: readonly T[]): T[] {
  if (!Array.isArray(value)) return [...fallback]
  const result = value.filter((item): item is T => typeof item === 'string' && allowed.has(item as T))
  return result.length ? [...new Set(result)] : [...fallback]
}

/** Normalize options so a partially upgraded backend cannot break the form. */
export function normalizeImageGenerationOptions(value: unknown): ImageGenerationOptionsResponse {
  const source = value && typeof value === 'object'
    ? value as { groups?: unknown; defaults?: unknown }
    : {}
  const groups = Array.isArray(source.groups) ? source.groups : []
  const rawDefaults = source.defaults && typeof source.defaults === 'object'
    ? source.defaults as Record<string, unknown>
    : {}
  const defaultQuality = typeof rawDefaults.quality === 'string' && QUALITY_VALUES.has(rawDefaults.quality as ImageGenerationQuality)
    ? rawDefaults.quality as ImageGenerationQuality
    : 'auto'
  const defaultFormat = typeof rawDefaults.output_format === 'string' && FORMAT_VALUES.has(rawDefaults.output_format as ImageGenerationFormat)
    ? rawDefaults.output_format as ImageGenerationFormat
    : 'png'
  const defaultBackground = typeof rawDefaults.background === 'string' && BACKGROUND_VALUES.has(rawDefaults.background as ImageGenerationBackground)
    ? rawDefaults.background as ImageGenerationBackground
    : 'auto'
  const defaultModeration = typeof rawDefaults.moderation === 'string' && MODERATION_VALUES.has(rawDefaults.moderation as ImageGenerationModeration)
    ? rawDefaults.moderation as ImageGenerationModeration
    : 'auto'
  const defaultCount = typeof rawDefaults.n === 'number' && Number.isFinite(rawDefaults.n)
    ? Math.min(9, Math.max(1, Math.round(rawDefaults.n)))
    : 1

  return {
    groups: groups.flatMap((raw): ImageGenerationGroupOption[] => {
      if (!raw || typeof raw !== 'object') return []
      const item = raw as Record<string, unknown>
      const id = typeof item.id === 'number' ? item.id : Number(item.id)
      if (!Number.isFinite(id)) return []
      const name = typeof item.name === 'string' && item.name.trim() ? item.name : `Group ${id}`
      const models = Array.isArray(item.models) ? item.models : []
      return [{
        id,
        name,
        description: typeof item.description === 'string' ? item.description : null,
        platform: typeof item.platform === 'string' ? item.platform : undefined,
        models: models.flatMap((rawModel): ImageGenerationModelOption[] => {
          if (!rawModel || typeof rawModel !== 'object') return []
          const model = rawModel as Record<string, unknown>
          const modelName = typeof model.name === 'string' && model.name.trim()
            ? model.name
            : typeof model.id === 'string' && model.id.trim()
              ? model.id
              : ''
          if (!modelName) return []
          return [{
            name: modelName,
            id: typeof model.id === 'string' ? model.id : modelName,
            sizes: stringList(model.sizes, DEFAULT_IMAGE_SIZES),
            qualities: enumList(model.qualities, QUALITY_VALUES, DEFAULT_IMAGE_QUALITIES),
            output_formats: enumList(model.output_formats, FORMAT_VALUES, DEFAULT_IMAGE_FORMATS),
            backgrounds: enumList(model.backgrounds, BACKGROUND_VALUES, DEFAULT_IMAGE_BACKGROUNDS),
            moderations: enumList(model.moderations, MODERATION_VALUES, DEFAULT_IMAGE_MODERATIONS),
            max_n: typeof model.max_n === 'number' && Number.isFinite(model.max_n)
              ? Math.min(9, Math.max(1, Math.round(model.max_n)))
              : 9,
            supports_compression: model.supports_compression !== false,
            custom_size: normalizeCustomSize(model.custom_size)
          }]
        })
      }]
    }),
    defaults: {
      size: typeof rawDefaults.size === 'string' && rawDefaults.size.trim()
        ? rawDefaults.size
        : DEFAULT_IMAGE_SIZES[0],
      quality: defaultQuality,
      output_format: defaultFormat,
      background: defaultBackground,
      moderation: defaultModeration,
      n: defaultCount
    }
  }
}

function normalizeConfigGroupList(value: unknown): ImageGenerationConfigGroupOption[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((raw): ImageGenerationConfigGroupOption[] => {
    if (!raw || typeof raw !== 'object') return []
    const source = raw as Record<string, unknown>
    const id = typeof source.id === 'number' ? source.id : Number(source.id)
    if (!Number.isFinite(id) || id <= 0) return []
    const models = Array.isArray(source.models)
      ? source.models.flatMap((rawModel): ImageGenerationConfigModelOption[] => {
        if (!rawModel || typeof rawModel !== 'object') return []
        const model = rawModel as Record<string, unknown>
        return typeof model.name === 'string' && model.name.trim()
          ? [{ name: model.name.trim() }]
          : []
      })
      : []
    if (models.length === 0) return []
    return [{
      id,
      name: typeof source.name === 'string' && source.name.trim() ? source.name : `Group ${id}`,
      description: typeof source.description === 'string' ? source.description : null,
      platform: typeof source.platform === 'string' ? source.platform : undefined,
      models
    }]
  })
}

/** Normalize the user-scoped configuration payload and discard malformed secrets. */
export function normalizeImageGenerationConfigOptions(value: unknown): ImageGenerationConfigOptionsResponse {
  const source = value && typeof value === 'object' ? value as Record<string, unknown> : {}
  const rawConfig = source.config && typeof source.config === 'object'
    ? source.config as Record<string, unknown>
    : {}
  const numberValue = (key: string): number => {
    const item = rawConfig[key]
    const numeric = typeof item === 'number' ? item : Number(item)
    return Number.isFinite(numeric) ? Math.round(numeric) : 0
  }
  const config: ImageGenerationConfig = {
    version: 1,
    prompt_group_id: numberValue('prompt_group_id'),
    prompt_model: typeof rawConfig.prompt_model === 'string' ? rawConfig.prompt_model : '',
    prompt_api_key_id: numberValue('prompt_api_key_id'),
    image_group_id: numberValue('image_group_id'),
    image_model: typeof rawConfig.image_model === 'string' && rawConfig.image_model.trim() ? rawConfig.image_model : 'gpt-image-2',
    image_api_key_id: numberValue('image_api_key_id'),
    default_size: typeof rawConfig.default_size === 'string' && rawConfig.default_size.trim() ? rawConfig.default_size : '1024x1024',
    default_n: Math.min(9, Math.max(1, numberValue('default_n') || 1))
  }
  const apiKeys = Array.isArray(source.api_keys)
    ? source.api_keys.flatMap((raw): ImageGenerationConfigApiKeyOption[] => {
      if (!raw || typeof raw !== 'object') return []
      const item = raw as Record<string, unknown>
      const id = typeof item.id === 'number' ? item.id : Number(item.id)
      const groupId = typeof item.group_id === 'number' ? item.group_id : Number(item.group_id)
      if (!Number.isFinite(id) || !Number.isFinite(groupId) || id <= 0 || groupId <= 0) return []
      return [{
        id,
        name: typeof item.name === 'string' && item.name.trim() ? item.name : `API Key #${id}`,
        masked_key: typeof item.masked_key === 'string' ? item.masked_key : '',
        group_id: groupId,
        group_name: typeof item.group_name === 'string' ? item.group_name : '',
        image_enabled: item.image_enabled === true,
        status: typeof item.status === 'string' ? item.status : 'active'
      }]
    })
    : []
  return {
    config,
    prompt_groups: normalizeConfigGroupList(source.prompt_groups),
    image_groups: normalizeConfigGroupList(source.image_groups),
    api_keys: apiKeys
  }
}

function normalizeCustomSize(value: unknown): ImageGenerationCustomSizeConstraints | null {
  if (!value || typeof value !== 'object') return null
  const source = value as Record<string, unknown>
  const numeric = (key: keyof ImageGenerationCustomSizeConstraints): number | null => {
    const item = source[key]
    return typeof item === 'number' && Number.isFinite(item) && item > 0 ? item : null
  }
  const minPixels = numeric('min_pixels')
  const maxPixels = numeric('max_pixels')
  const maxEdge = numeric('max_edge')
  const edgeMultiple = numeric('edge_multiple')
  const maxAspectRatio = numeric('max_aspect_ratio')
  if (minPixels === null || maxPixels === null || maxEdge === null || edgeMultiple === null || maxAspectRatio === null) {
    return null
  }
  return { min_pixels: minPixels, max_pixels: maxPixels, max_edge: maxEdge, edge_multiple: edgeMultiple, max_aspect_ratio: maxAspectRatio }
}

export function validateCustomImageSize(
  width: number,
  height: number,
  constraints: ImageGenerationCustomSizeConstraints
): string | null {
  if (!Number.isInteger(width) || !Number.isInteger(height) || width <= 0 || height <= 0) {
    return 'positive_integer'
  }
  if (width % constraints.edge_multiple !== 0 || height % constraints.edge_multiple !== 0) {
    return 'edge_multiple'
  }
  if (width > constraints.max_edge || height > constraints.max_edge) {
    return 'max_edge'
  }
  const pixels = width * height
  if (pixels < constraints.min_pixels || pixels > constraints.max_pixels) {
    return 'pixel_range'
  }
  const longEdge = Math.max(width, height)
  const shortEdge = Math.min(width, height)
  if (longEdge / shortEdge > constraints.max_aspect_ratio) {
    return 'aspect_ratio'
  }
  return null
}

/** Return a single safe image source from an upstream image item. */
export function imageSource(image: GeneratedImage, outputFormat: ImageGenerationFormat): string {
  if (typeof image.url === 'string' && image.url.trim()) return image.url.trim()
  if (typeof image.b64_json === 'string' && image.b64_json.trim()) {
    const fallbackMime = outputFormat === 'jpeg' ? 'image/jpeg' : `image/${outputFormat}`
    const mime = image.mime_type === 'image/png' || image.mime_type === 'image/jpeg' || image.mime_type === 'image/webp'
      ? image.mime_type
      : fallbackMime
    return `data:${mime};base64,${image.b64_json.trim()}`
  }
  return ''
}
