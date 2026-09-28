import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import Gallery from '../views/PelicanGalleryView.vue'

const api = vi.hoisted(() => ({
  list: vi.fn(),
  status: vi.fn(),
  groups: vi.fn(),
  save: vi.fn(),
  artifact: vi.fn()
}))
vi.mock('../api', () => ({ pelicanAPI: api }))
vi.mock('vue-i18n', async () => {
  const { ref } = await import('vue')
  return { useI18n: () => ({ t: (key: string) => key, locale: ref('zh') }) }
})
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ isAdmin: true }) }))
vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: { template: '<main><slot /></main>' }
}))

const run = {
  id: 1,
  scheduled_for: '2026-09-27T07:00:00Z',
  status: 'succeeded',
  topic_id: 'pelican-ski',
  groups: [
    { id: 1, name: 'A' },
    { id: 2, name: 'B' }
  ],
  model: 'gpt-6-astra'
}
let wrapper: VueWrapper | undefined
function render() {
  wrapper = mount(Gallery, {
    global: {
      stubs: {
        RouterLink: { template: '<a><slot /></a>' },
        PelicanCard: {
          props: ['run', 'active'],
          emits: ['open', 'visibility', 'unavailable'],
          template:
            '<button class="creation" @click="$emit(\'open\', run)">{{ run.groups.map(g => g.name).join(",") }}</button>'
        },
        PelicanPreviewDialog: {
          props: ['run'],
          template: '<div class="dialog">{{ run?.id }}</div>'
        }
      }
    }
  })
  return wrapper
}
beforeEach(() => {
  vi.resetAllMocks()
  api.list.mockResolvedValue({ items: [run], next_cursor: '' })
  api.status.mockResolvedValue({
    enabled: true,
    next_run_at: null,
    last_success_at: null,
    interval_seconds: 3600
  })
  api.groups.mockResolvedValue(run.groups)
})
afterEach(() => {
  wrapper?.unmount()
  vi.useRealTimers()
})

describe('gallery reads never trigger generation', () => {
  it('shows one shared work with multiple labels and opens the same ID', async () => {
    const view = render()
    await flushPromises()
    expect(view.findAll('.creation')).toHaveLength(1)
    expect(view.find('.creation').text()).toBe('A,B')
    await view.find('.creation').trigger('click')
    expect(view.find('.dialog').text()).toBe('1')
    expect(api.save).not.toHaveBeenCalled()
  })

  it('omits failed runs and preserves successful cards when refreshing fails', async () => {
    api.list.mockResolvedValueOnce({
      items: [run, { ...run, id: 2, status: 'failed' }],
      next_cursor: ''
    })
    const view = render()
    await flushPromises()
    expect(view.findAll('.creation')).toHaveLength(1)
    api.list.mockRejectedValueOnce(new Error('offline'))
    await view
      .findAll('button')
      .find((button) => button.text() === 'pelican.refresh')!
      .trigger('click')
    await flushPromises()
    expect(view.findAll('.creation')).toHaveLength(1)
    expect(view.text()).not.toContain('offline')
    expect(api.save).not.toHaveBeenCalled()
  })

  it('ignores an old response after changing topic filters', async () => {
    let resolveOld!: (value: unknown) => void
    api.list.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveOld = resolve
      })
    )
    const view = render()
    await flushPromises()
    api.list.mockResolvedValueOnce({
      items: [{ ...run, id: 3, groups: [{ id: 3, name: 'new' }] }],
      next_cursor: ''
    })
    await view.find('select').setValue('wukong-airplane')
    await flushPromises()
    resolveOld({ items: [run], next_cursor: '' })
    await flushPromises()
    expect(view.find('.creation').text()).toBe('new')
  })
})
