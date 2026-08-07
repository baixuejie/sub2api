import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const here = dirname(fileURLToPath(import.meta.url))
const read = (path: string) => readFileSync(resolve(here, path), 'utf8')

describe('Image Generation extension integration surface', () => {
  it('keeps the page under the isolated feature directory', () => {
    const view = read('../views/ImageGenerationView.vue')
    const api = read('../api/imageGeneration.ts')

    expect(view).toContain("from '../api/imageGeneration'")
    expect(view).toContain("from '../types/imageGeneration'")
    expect(api).toContain("'/image-generation/options'")
    expect(api).toContain("'/image-generation/generate'")
    expect(api).toContain("'/image-generation/config'")
    expect(api).toContain("'/image-generation/optimize'")
    expect(api).toContain('timeout: IMAGE_GENERATION_TIMEOUT_MS')
    expect(view).not.toMatch(/PromptTemplate|promptTemplate|模板|AcademicPromptPicker/)
  })

  it('exposes an authenticated route and regular-user menu entry', () => {
    const router = read('../../../router/index.ts')
    const sidebar = read('../../../components/layout/AppSidebar.vue')
    const zhCommon = read('../../../i18n/locales/zh/common.ts')
    const routeStart = router.indexOf("path: '/image-generation'")
    const route = router.slice(routeStart, router.indexOf("path: '/usage'", routeStart))

    expect(route).toContain("component: () => import('@/features/image-generation/views/ImageGenerationView.vue')")
    expect(route).toContain('requiresAuth: true')
    expect(route).toContain('requiresAdmin: false')
    expect(router).toContain("path: '/image-generation/settings'")
    expect(sidebar).toContain("path: '/image-generation'")
    expect(sidebar).toContain("label: t('nav.imageGeneration')")
    expect(zhCommon).toContain("imageGeneration: '图片创作'")
  })
})
