import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const here = dirname(fileURLToPath(import.meta.url))
const read = (path: string) => readFileSync(resolve(here, path), 'utf8')

describe('Model Plaza extension integration surface', () => {
  it('points the public route at the isolated feature view', () => {
    const router = read('../../../router/index.ts')
    const routeStart = router.indexOf("path: '/model-plaza'")
    const route = router.slice(routeStart, router.indexOf("path: '/dashboard'", routeStart))

    expect(route).toContain("component: () => import('@/features/model-plaza/views/ModelPlazaView.vue')")
    expect(route).toContain('requiresAuth: false')
    expect(route).toContain("titleKey: 'modelPlaza.title'")
  })

  it('keeps the extension boundary around the new page and adapter', () => {
    const view = read('../views/ModelPlazaView.vue')
    const adapter = read('../api/modelPlaza.ts')

    expect(view).toContain("from '../api/modelPlaza'")
    expect(view).toContain("from '../components/ExtensionNavBar.vue'")
    expect(adapter).toContain("'/model-plaza'")
    expect(adapter).not.toContain("../components/modelPlaza")
  })

  it('keeps actual prices as the default and forwards the official-price toggle', () => {
    const explorer = read('../components/PlazaExplorer.vue')

    expect(explorer).toContain('const showOfficialPrice = ref(false)')
    expect(explorer).toContain('v-model="showOfficialPrice"')
    expect(explorer).toContain(':show-official-price="showOfficialPrice"')
  })

  it('keeps the enabled plaza in the header without duplicating it in the sidebar', () => {
    const sidebar = read('../../../components/layout/AppSidebar.vue')
    const header = read('../../../components/layout/AppHeader.vue')

    expect(sidebar).not.toContain("path: '/model-plaza'")
    expect(sidebar).not.toContain('FeatureFlags.modelPlaza')
    expect(header).toContain('const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))')
    expect(header).toContain(":to=\"{ path: '/model-plaza', query: { embedded: '1' } }\"")
    expect(header).toContain("t('nav.modelPlaza')")
  })
})
