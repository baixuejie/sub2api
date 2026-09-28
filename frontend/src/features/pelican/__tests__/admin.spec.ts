import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import Admin from '../views/PelicanAdminView.vue'

const api = vi.hoisted(() => ({
  config: vi.fn(),
  list: vi.fn(),
  save: vi.fn(),
  clearKey: vi.fn(),
  runNow: vi.fn(),
  detail: vi.fn(),
  rebuildPreview: vi.fn(),
  source: vi.fn()
}))
vi.mock('../api', () => ({ pelicanAPI: api }))
vi.mock('vue-i18n', async () => {
  const { ref } = await import('vue')
  return { useI18n: () => ({ t: (key: string) => key, locale: ref('zh') }) }
})
vi.mock('@/api/admin/groups', () => ({
  getAll: async () => [
    { id: 1, name: 'A' },
    { id: 2, name: 'B' }
  ]
}))
vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: { template: '<main><slot /></main>' }
}))
const config = {
  revision: 1,
  enabled: false,
  selected_group_ids: [],
  topic_mode: 'rotate',
  fixed_topic_id: 'pelican-ski',
  max_output_tokens: 16384,
  timeout_seconds: 300,
  retention_days: 30,
  interval_minutes: 60,
  next_run_at: null,
  key_configured: false,
  key_masked: '',
  encryption_ready: true,
  key_unavailable: false,
  model: 'gpt-6-astra',
  reasoning_effort: 'high'
}
let wrapper: VueWrapper | undefined
beforeEach(() => {
  vi.resetAllMocks()
  api.config.mockResolvedValue(config)
  api.list.mockResolvedValue({ items: [], next_cursor: '' })
  api.save.mockResolvedValue({ ...config, revision: 2, key_configured: true })
  api.runNow.mockResolvedValue({ id: 7, status: 'running' })
})
afterEach(() => wrapper?.unmount())

it('saves one key and several labels without any generation or editable model fields', async () => {
  wrapper = mount(Admin, { global: { stubs: { RouterLink: true, BaseDialog: true } } })
  await flushPromises()
  await wrapper.find('#pelican-key').setValue('designated-test-key')
  await wrapper.find('input[value="1"]').setValue(true)
  await wrapper.find('input[value="2"]').setValue(true)
  await wrapper.find('form').trigger('submit')
  await flushPromises()
  expect(api.save).toHaveBeenCalledTimes(1)
  expect(api.save.mock.calls[0][0]).toMatchObject({
    api_key: 'designated-test-key',
    selected_group_ids: [1, 2]
  })
  expect(api.save.mock.calls[0][0]).not.toHaveProperty('model')
  expect(api.save.mock.calls[0][0]).not.toHaveProperty('reasoning_effort')
  expect(api.runNow).not.toHaveBeenCalled()
  expect((wrapper.find('#pelican-key').element as HTMLInputElement).value).toBe('')
  expect(wrapper.text()).not.toContain('designated-test-key')
})

it('lets an administrator submit an immediate generation when a key is configured', async () => {
  api.config.mockResolvedValue({ ...config, key_configured: true })
  wrapper = mount(Admin, { global: { stubs: { RouterLink: true, BaseDialog: true } } })
  await flushPromises()
  const button = wrapper.findAll('button').find((item) => item.text() === 'pelican.runNow')
  expect(button).toBeDefined()
  await button!.trigger('click')
  await flushPromises()
  expect(api.runNow).toHaveBeenCalledTimes(1)
  expect(wrapper.find('[data-testid="pelican-run-now"]').attributes('disabled')).toBeDefined()
})

it.each([60, 30, 10])('persists a %i minute interval without triggering a run', async (minutes) => {
  wrapper = mount(Admin, { global: { stubs: { RouterLink: true, BaseDialog: true } } })
  await flushPromises()
  await wrapper.find('#pelican-interval').setValue(String(minutes))
  await wrapper.find('form').trigger('submit')
  await flushPromises()
  expect(api.save.mock.calls[0][0].interval_minutes).toBe(minutes)
  expect(api.runNow).not.toHaveBeenCalled()
})

it('requires saved settings and a usable key before running', async () => {
  api.config.mockResolvedValue({ ...config, key_configured: true })
  wrapper = mount(Admin, { global: { stubs: { RouterLink: true, BaseDialog: true } } })
  await flushPromises()
  await wrapper.find('#pelican-interval').setValue('10')
  expect(wrapper.find('[data-testid="pelican-run-now"]').attributes('disabled')).toBeDefined()
  expect(api.runNow).not.toHaveBeenCalled()
})

it.each([409, 429])('explains a rejected manual request (%i)', async (status) => {
  api.config.mockResolvedValue({ ...config, key_configured: true })
  api.runNow.mockRejectedValueOnce({ status })
  wrapper = mount(Admin, { global: { stubs: { RouterLink: true, BaseDialog: true } } })
  await flushPromises()
  await wrapper.find('[data-testid="pelican-run-now"]').trigger('click')
  await flushPromises()
  expect(wrapper.find('[role="alert"]').text()).toBe(status === 409 ? 'pelican.runActive' : 'pelican.runCooldown')
})

it('loads blocked original source and rebuilds its preview without a generation request', async () => {
  const run = { id: 42, status: 'preview_blocked', topic_id: 'pelican-ski', groups: [], scheduled_for: '2026-09-28T16:00:00Z', error_code: 'preview_unavailable' }
  const detail = { run, prompt: 'draw', request: {}, usage: null, skipped_hours: 0, source_available: true, raw_path: '42/source.txt', preview_path: '', preview_notes: ['not_html_or_svg'] }
  api.list.mockResolvedValue({ items: [run], next_cursor: '' })
  api.detail.mockResolvedValue(detail)
  api.source.mockResolvedValue({ source: '<svg><circle r="4"/></svg>' })
  api.rebuildPreview.mockResolvedValue({ ...detail, run: { ...run, status: 'succeeded' }, preview_notes: [] })
  wrapper = mount(Admin, { global: { stubs: { RouterLink: true, BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' } } } })
  await flushPromises()
  await wrapper.findAll('button').find((button) => button.text() === 'pelican.detail')!.trigger('click')
  await flushPromises()
  expect(api.source).toHaveBeenCalledWith(42)
  expect(wrapper.text()).toContain('<svg><circle')
  await wrapper.findAll('button').find((button) => button.text() === 'pelican.rebuildPreview')!.trigger('click')
  await flushPromises()
  expect(api.rebuildPreview).toHaveBeenCalledWith(42)
  expect(api.runNow).not.toHaveBeenCalled()
})
