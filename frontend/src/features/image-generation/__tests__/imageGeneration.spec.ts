import { describe, expect, it } from 'vitest'
import {
  imageSource,
  normalizeImageGenerationOptions,
  validateCustomImageSize
} from '../types/imageGeneration'

describe('image-generation extension contracts', () => {
  it('normalizes capabilities and drops malformed groups or models', () => {
    const result = normalizeImageGenerationOptions({
      groups: [
        {
          id: 7,
          name: 'OpenAI 图片组',
          models: [
            {
              name: 'gpt-image-1',
              sizes: ['1024x1024'],
              output_formats: ['png'],
              max_n: 4,
              custom_size: {
                min_pixels: 655360,
                max_pixels: 8294400,
                max_edge: 3840,
                edge_multiple: 16,
                max_aspect_ratio: 3
              }
            },
            { name: '', sizes: ['bad'] },
            null
          ]
        },
        { id: 'invalid', name: 'bad', models: [] },
        null
      ]
    })

    expect(result.groups).toHaveLength(1)
    expect(result.groups[0].models.map((model) => model.name)).toEqual(['gpt-image-1'])
    expect(result.groups[0].models[0].qualities).toEqual(['auto', 'low', 'medium', 'high'])
    expect(result.groups[0].models[0].output_formats).toEqual(['png'])
    expect(result.groups[0].models[0].backgrounds).toEqual(['auto', 'opaque', 'transparent'])
    expect(result.groups[0].models[0].max_n).toBe(4)
    expect(result.groups[0].models[0].custom_size?.max_edge).toBe(3840)
    expect(result.defaults).toEqual({
      size: '1024x1024',
      quality: 'auto',
      output_format: 'png',
      background: 'auto',
      moderation: 'auto',
      n: 1
    })
  })

  it('validates model-provided custom dimensions before submission', () => {
    const constraints = {
      min_pixels: 655360,
      max_pixels: 8294400,
      max_edge: 3840,
      edge_multiple: 16,
      max_aspect_ratio: 3
    }
    expect(validateCustomImageSize(2048, 2048, constraints)).toBeNull()
    expect(validateCustomImageSize(1000, 1000, constraints)).toBe('edge_multiple')
    expect(validateCustomImageSize(3840, 1280, constraints)).toBeNull()
    expect(validateCustomImageSize(3840, 1024, constraints)).toBe('aspect_ratio')
  })

  it('builds a data URL for base64 results and preserves upstream URLs', () => {
    expect(imageSource({ b64_json: 'abc' }, 'png')).toBe('data:image/png;base64,abc')
    expect(imageSource({ b64_json: 'abc' }, 'jpeg')).toBe('data:image/jpeg;base64,abc')
    expect(imageSource({ b64_json: 'abc', mime_type: 'image/svg+xml' }, 'webp')).toBe('data:image/webp;base64,abc')
    expect(imageSource({ url: 'https://cdn.example/image.png', b64_json: 'ignored' }, 'png')).toBe('https://cdn.example/image.png')
    expect(imageSource({}, 'webp')).toBe('')
  })
})
