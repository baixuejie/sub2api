import { apiClient } from '@/api/client'
import {
  normalizeImageGenerationOptions,
  type ImageGenerationOptionsResponse,
  type ImageGenerationRequest,
  type ImageGenerationResponse,
  type GeneratedImage
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

export const imageGenerationAPI = {
  loadOptions: loadImageGenerationOptions,
  generate: generateImages
}

export default imageGenerationAPI
