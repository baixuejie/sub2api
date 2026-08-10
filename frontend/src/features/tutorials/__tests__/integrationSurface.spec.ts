import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const here = dirname(fileURLToPath(import.meta.url))
const read = (path: string) => readFileSync(resolve(here, path), 'utf8')

describe('Tutorials extension integration surface', () => {
  it('keeps the page in the extension and embeds the configured surfaces', () => {
    const view = read('../views/TutorialsView.vue')
    const api = read('../api/tutorials.ts')
    const settings = read('../components/TutorialVideoSettings.vue')
    const adminSettings = read('../../../views/admin/SettingsView.vue')

    expect(view).toContain("from '../api/tutorials'")
    expect(view).toContain('TUTORIAL_DOCUMENT_URL')
    expect(view).toContain('target="_blank"')
    expect(view).toContain('tutorials.videos.empty')
    expect(view).toContain("ref<'documents' | 'videos'>('documents')")
    expect(view).toContain('role="tablist"')
    expect(view).toContain("activePanel === 'documents'")
    expect(view).toContain('role="tabpanel"')
    expect(view).toContain('if (!settings)')
    expect(api).toContain('tutorial_videos')
    expect(settings).toContain('defineModel<TutorialVideoSetting[]>')
    expect(adminSettings).toContain(
      '<TutorialVideoSettings v-model="form.tutorial_videos" />'
    )
    expect(adminSettings).toContain(
      'from "@/features/tutorials/components/TutorialVideoSettings.vue"'
    )
  })

  it('exposes a public route and bottom sidebar entry for all users', () => {
    const router = read('../../../router/index.ts')
    const sidebar = read('../../../components/layout/AppSidebar.vue')

    expect(router).toContain("path: '/tutorials'")
    expect(router).toContain("component: () => import('@/features/tutorials/views/TutorialsView.vue')")
    expect(router).toContain('requiresAuth: false')
    expect(router).toContain("'/legal', '/tutorials'")
    expect(sidebar).toContain('to="/tutorials"')
    expect(sidebar).toContain("t('nav.tutorials')")
  })
})
