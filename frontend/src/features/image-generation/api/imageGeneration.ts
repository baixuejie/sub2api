import { apiClient } from '@/api/client'
import {
  normalizeImageGenerationOptions,
  normalizeImageGenerationConfigOptions,
  type ImageGenerationConfig,
  type ImageGenerationConfigOptionsResponse,
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

  if (Array.isArray(data)) {
    return { data: data as GeneratedImage[] }
  }
  if (data && typeof data === 'object') {
    const response = data as Record<string, unknown>
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
  loadConfig: loadImageGenerationConfig,
  saveConfig: saveImageGenerationConfig,
  optimize: optimizeImagePrompt
}

export default imageGenerationAPI
