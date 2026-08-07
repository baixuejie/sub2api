import { apiClient } from '@/api/client'
import {
  normalizeImageGenerationOptions,
  normalizeImageGenerationConfigOptions,
  type ImageGenerationConfig,
  type ImageGenerationConfigOptionsResponse,
  type ImageEditRequest,
  type ImageGenerationOptionsResponse,
  type ImageGenerationRequest,
  type ImageGenerationResponse,
  type GeneratedImage,
  type PromptOptimizationResponse
} from '../types/imageGeneration'

const IMAGE_GENERATION_TIMEOUT_MS = 5 * 60 * 1000

/** Load only the image groups and model capabilities allowed for the session. */
export async function loadImageGenerationOptions(options?: { signal?: AbortSignal }): Promise<ImageGenerationOptionsResponse> {
  const { data } = await apiClient.get<unknown>('/image-generation/options', {
    signal: options?.signal
  })
  return normalizeImageGenerationOptions(data)
}

/**
 * Submit a prompt to the authenticated backend extension.
 *
 * No API key, upstream URL, account id or pricing information is accepted by
 * this client. The backend resolves those details from the authenticated user
 * and selected group.
 */
export async function generateImages(
  payload: ImageGenerationRequest,
  options?: { signal?: AbortSignal }
): Promise<ImageGenerationResponse> {
  const { data } = await apiClient.post<unknown>('/image-generation/generate', payload, {
    signal: options?.signal,
    timeout: IMAGE_GENERATION_TIMEOUT_MS
  })

  return normalizeImageGenerationResponse(data)
}

/**
 * Submit one or more reference images to the authenticated Extension.
 * FormData is constructed here so the browser never needs an upstream key.
 */
export async function editImages(
  payload: ImageEditRequest,
  files: File[],
  options?: { signal?: AbortSignal }
): Promise<ImageGenerationResponse> {
  const formData = buildImageEditFormData(payload, files)

  const { data } = await apiClient.post<unknown>('/image-generation/edit', formData, {
    signal: options?.signal,
    timeout: IMAGE_GENERATION_TIMEOUT_MS,
    headers: { 'Content-Type': undefined }
  })
  return normalizeImageGenerationResponse(data)
}

export function buildImageEditFormData(payload: ImageEditRequest, files: File[]): FormData {
  const formData = new FormData()
  formData.append('group_id', String(payload.group_id))
  formData.append('model', payload.model)
  formData.append('prompt', payload.prompt)
  formData.append('n', String(payload.n))
  formData.append('size', payload.size)
  formData.append('quality', payload.quality)
  formData.append('output_format', payload.output_format)
  if (typeof payload.output_compression === 'number') {
    formData.append('output_compression', String(payload.output_compression))
  }
  formData.append('background', payload.background)
  formData.append('moderation', payload.moderation)
  files.forEach((file) => formData.append('image', file, file.name || 'image.png'))
  return formData
}

export async function loadImageGenerationConfig(options?: { signal?: AbortSignal }): Promise<ImageGenerationConfigOptionsResponse> {
  const { data } = await apiClient.get<unknown>('/image-generation/config', {
    signal: options?.signal
  })
  return normalizeImageGenerationConfigOptions(data)
}

export async function saveImageGenerationConfig(
  config: ImageGenerationConfig,
  options?: { signal?: AbortSignal }
): Promise<ImageGenerationConfigOptionsResponse> {
  const { data } = await apiClient.put<unknown>('/image-generation/config', config, {
    signal: options?.signal
  })
  return normalizeImageGenerationConfigOptions(data)
}

export async function optimizeImagePrompt(
  prompt: string,
  options?: { signal?: AbortSignal }
): Promise<PromptOptimizationResponse> {
  const { data } = await apiClient.post<unknown>('/image-generation/optimize', { prompt }, {
    signal: options?.signal,
    timeout: 90_000
  })
  const source = data && typeof data === 'object' ? data as Record<string, unknown> : {}
  return {
    original_prompt: typeof source.original_prompt === 'string' ? source.original_prompt : prompt,
    optimized_prompt: typeof source.optimized_prompt === 'string' ? source.optimized_prompt : ''
  }
}

export const imageGenerationAPI = {
  loadOptions: loadImageGenerationOptions,
  generate: generateImages,
  edit: editImages,
  loadConfig: loadImageGenerationConfig,
  saveConfig: saveImageGenerationConfig,
  optimize: optimizeImagePrompt
}

function normalizeImageGenerationResponse(value: unknown): ImageGenerationResponse {
  if (Array.isArray(value)) {
    return { data: value as GeneratedImage[] }
  }
  if (value && typeof value === 'object') {
    const response = value as Record<string, unknown>
    const images = Array.isArray(response.data)
      ? response.data as GeneratedImage[]
      : Array.isArray(response.images)
        ? response.images as GeneratedImage[]
        : []
    return {
      ...(typeof response.created === 'number' ? { created: response.created } : {}),
      data: images
    }
  }
  return { data: [] }
}

export default imageGenerationAPI
